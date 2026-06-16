package main

import (
	"bazi/internal/config"
	"bazi/internal/handler"
	"bazi/internal/middleware"
	"bazi/internal/model"
	"bazi/internal/service/bazi"
	"bazi/internal/service/fortune"
	"bazi/internal/service/interpretation"
	"bazi/internal/service/localrag"
	"bazi/internal/service/rag"
	"bazi/internal/service/ragflow"
	"bazi/internal/service/ziwei"
	"bazi/internal/store"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func seedAdminIfEmpty(us *store.DBUserStore) {
	existing, _ := us.FindByUsername("admin")
	if existing != nil {
		return
	}
	admin := &model.User{Username: "admin", Email: "admin@bazi.com"}
	if err := admin.SetPassword("admin"); err != nil {
		log.Printf("Failed to hash admin password: %v", err)
		return
	}
	if err := us.Create(admin); err != nil {
		log.Printf("Failed to seed admin user: %v", err)
		return
	}
	log.Println("Seeded default admin user (username=admin, password=admin)")
}

func main() {
	cfg := config.Load()
	middleware.InitJWT(cfg.JWTSecret)

	db := initDatabase(cfg)
	cs := store.NewDBChartStore(db)
	fs := store.NewDBFortuneStore(db)
	feedbackStore := store.NewDBFeedbackStore(db)
	us := store.NewDBUserStore(db)

	seedAdminIfEmpty(us)

	baziSvc := &bazi.BaziService{}
	parser := &bazi.InputParser{}
	engine := fortune.NewFortuneEngine()
	ziweiSvc := ziwei.NewZiWeiService()
	retriever := buildRAGRetriever(cfg)
	interpretSvc := &interpretation.Service{
		Charts:    cs,
		Bazi:      baziSvc,
		Retriever: retriever,
		MinScore:  cfg.RAGMinScore,
		TopK:      cfg.RAGTopK,
	}

	r := gin.Default()
	allowOrigin := os.Getenv("CORS_ORIGIN")
	if allowOrigin == "" {
		allowOrigin = "*"
	}
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	auth := &handler.AuthHandler{Store: us}
	r.POST("/api/auth/register", auth.Register)
	r.POST("/api/auth/login", auth.Login)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		ch := &handler.ChartHandler{Parser: parser, Bazi: baziSvc, Store: cs}
		api.POST("/chart", ch.Chart)
		fh := &handler.FortuneHandler{Engine: engine, ChartStore: cs}
		api.POST("/fortune", fh.CalculateDaily)
		wh := &handler.WeeklyFortuneHandler{Engine: engine, Charts: cs}
		api.POST("/fortune/weekly", middleware.ETag(), wh.Weekly)
		mh := &handler.MonthlyFortuneHandler{Engine: engine, ChartStore: cs}
		api.POST("/fortune/monthly", middleware.ETag(), mh.HandleMonthly)
		ah := &handler.AIStubHandler{}
		api.POST("/fortune/ai", ah.AnalyzeFortune)
		api.GET("/auth/me", auth.Me)
		hh := &handler.HistoryHandler{Charts: cs, FortuneHistory: fs, Bazi: baziSvc}
		api.GET("/charts", hh.ListCharts)
		api.GET("/charts/:id", hh.GetChart)
		api.GET("/fortune/history", hh.FortuneHistoryList)
		handler.RegisterZiWeiRoutesWithStore(api, ziweiSvc, cs)
		handler.RegisterZiWeiPeriodRoutes(api, ziweiSvc, cs)
		handler.RegisterInterpretationRoutes(api, interpretSvc)
		handler.RegisterFeedbackRoutes(api, cs, feedbackStore)
	}

	log.Printf("Server starting on :%s", cfg.ServerPort)
	r.Run(":" + cfg.ServerPort)
}

func buildRAGRetriever(cfg *config.Config) rag.Retriever {
	if cfg == nil || !cfg.RAGEnabled {
		return nil
	}
	switch cfg.RAGProvider {
	case "", "sqlite_fts5", "local", "sqlite":
		return localrag.NewRetriever(localrag.Config{
			Enabled:   cfg.RAGEnabled,
			IndexPath: cfg.LocalRAGIndexPath,
			MinScore:  cfg.RAGMinScore,
			TopK:      cfg.RAGTopK,
		})
	case "ragflow":
		return ragflow.NewClient(ragflow.Config{
			Enabled:        cfg.RAGEnabled || cfg.RAGFlowEnabled,
			BaseURL:        cfg.RAGFlowBaseURL,
			APIKey:         cfg.RAGFlowAPIKey,
			DatasetID:      cfg.RAGFlowBaziDatasetID,
			TimeoutSeconds: cfg.RAGTimeoutSeconds,
			MinScore:       cfg.RAGMinScore,
			TopK:           cfg.RAGTopK,
		})
	default:
		log.Printf("[WARN] Unknown RAG_PROVIDER=%q; classical interpretation will use fallback", cfg.RAGProvider)
		return nil
	}
}
