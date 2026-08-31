package main

import (
	"bazi/internal/config"
	"bazi/internal/handler"
	"bazi/internal/middleware"
	"bazi/internal/model"
	"bazi/internal/service/bazi"
	"bazi/internal/service/buyi"
	"bazi/internal/service/elementasset"
	"bazi/internal/service/fortune"
	"bazi/internal/service/interpretation"
	"bazi/internal/service/localrag"
	"bazi/internal/service/rag"
	"bazi/internal/service/ragflow"
	"bazi/internal/service/ziwei"
	"bazi/internal/store"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func seedAdminIfConfigured(us *store.DBUserStore, cfg *config.Config) {
	if us == nil || cfg == nil {
		return
	}
	if cfg.AdminUsername == "" && cfg.AdminEmail == "" && cfg.AdminPassword == "" {
		return
	}
	if cfg.AdminUsername == "" || cfg.AdminEmail == "" || len([]rune(cfg.AdminPassword)) < 12 {
		log.Println("[WARN] Admin initialization skipped: ADMIN_USERNAME, ADMIN_EMAIL and a 12+ character ADMIN_PASSWORD are required")
		return
	}
	existing, _ := us.FindByUsername(cfg.AdminUsername)
	if existing != nil {
		return
	}
	admin := &model.User{Username: cfg.AdminUsername, Email: cfg.AdminEmail}
	if err := admin.SetPassword(cfg.AdminPassword); err != nil {
		log.Printf("Failed to hash configured admin password: %v", err)
		return
	}
	if err := us.Create(admin); err != nil {
		log.Printf("Failed to seed configured admin user: %v", err)
		return
	}
	log.Printf("Seeded configured admin user %q", cfg.AdminUsername)
}

func corsMiddleware(allowOrigin string) gin.HandlerFunc {
	allowOrigin = strings.TrimSpace(allowOrigin)
	return func(c *gin.Context) {
		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
			c.Header("Vary", "Origin")
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}
		}
		c.Next()
	}
}

func main() {
	cfg := config.Load()
	middleware.InitJWT(cfg.JWTSecret)

	db := initDatabase(cfg)
	cs := store.NewDBChartStore(db)
	fs := store.NewDBFortuneStore(db)
	feedbackStore := store.NewDBFeedbackStore(db)
	buyiStore := store.NewDBBuyiStore(db)
	us := store.NewDBUserStore(db)
	elementAssetStore := store.NewDBElementAssetStore(db)
	elementAssetSvc := elementasset.New(elementAssetStore)
	if err := elementAssetStore.UpsertDefaults(elementasset.DefaultAssets()); err != nil {
		log.Printf("[WARN] Failed to seed default element assets: %v", err)
	}

	seedAdminIfConfigured(us, cfg)

	baziSvc := &bazi.BaziService{}
	buyiSvc := buyi.NewService()
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
	r.Use(corsMiddleware(cfg.CORSOrigin))

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	if err := os.MkdirAll(cfg.ElementAssetDir, 0o755); err != nil {
		log.Fatalf("Failed to prepare element asset directory: %v", err)
	}
	r.Static("/uploads/element-assets", cfg.ElementAssetDir)

	auth := &handler.AuthHandler{Store: us}
	r.POST("/api/auth/register", auth.Register)
	r.POST("/api/auth/login", auth.Login)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		ch := &handler.ChartHandler{Parser: parser, Bazi: baziSvc, Store: cs}
		api.POST("/chart/preview", ch.Preview)
		api.POST("/chart", ch.Chart)
		fh := &handler.FortuneHandler{Engine: engine, ChartStore: cs}
		api.POST("/fortune", fh.CalculateDaily)
		handler.RegisterElementAssetRoutes(api, elementAssetSvc, elementAssetStore, us, cfg.AdminUsername, cfg.ElementAssetDir)
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
		handler.RegisterBuyiRoutes(api, buyiSvc, buyiStore)
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
