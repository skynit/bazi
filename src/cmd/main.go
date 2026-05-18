package main

import (
	"bazi/internal/handler"
	"bazi/internal/middleware"
	"bazi/internal/model"
	"bazi/internal/service"
	"log"
	"os"
	"sync"
	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
)

// === Mock stores for local dev ===
type memUserStore struct {
	mu    sync.Mutex
	users map[uint]*model.User
	next  uint
}
func newMemUserStore() *memUserStore { return &memUserStore{users: map[uint]*model.User{}, next: 1} }
func (s *memUserStore) Create(u *model.User) error { s.mu.Lock(); defer s.mu.Unlock(); u.ID = s.next; s.next++; s.users[u.ID] = u; return nil }
func (s *memUserStore) FindByUsername(name string) (*model.User, error) { s.mu.Lock(); defer s.mu.Unlock(); for _, u := range s.users { if u.Username == name { return u, nil } }; return nil, nil }
func (s *memUserStore) FindByID(id uint) (*model.User, error) { s.mu.Lock(); defer s.mu.Unlock(); if u, ok := s.users[id]; ok { return u, nil }; return nil, nil }

type memChartStore struct {
	mu     sync.Mutex
	charts map[uint]*model.BirthChart
	next   uint
}
func newMemChartStore() *memChartStore { return &memChartStore{charts: map[uint]*model.BirthChart{}, next: 1} }
func (s *memChartStore) Create(c *model.BirthChart) error { s.mu.Lock(); defer s.mu.Unlock(); c.ID = s.next; s.next++; s.charts[c.ID] = c; return nil }
func (s *memChartStore) FindByID(id uint) (*model.BirthChart, error) { s.mu.Lock(); defer s.mu.Unlock(); if c, ok := s.charts[id]; ok { return c, nil }; return nil, nil }
func (s *memChartStore) ListByUser(uid uint, page, size int) ([]model.BirthChart, int64, error) { s.mu.Lock(); defer s.mu.Unlock(); var r []model.BirthChart; for _, c := range s.charts { if c.UserID == uid { r = append(r, *c) } }; return r, int64(len(r)), nil }

type memFortuneStore struct{}
func (s *memFortuneStore) ListByChartID(cid uint, page, size int) ([]model.HistoryResponse, int64, error) { return nil, 0, nil }

func main() {
	jwt := os.Getenv("JWT_SECRET")
	if jwt == "" { jwt = "dev-secret" }
	middleware.InitJWT(jwt)
	
	log.Println("Starting server with mock stores (no MySQL needed)")

	us := newMemUserStore()
	cs := newMemChartStore()
	// Seed admin account
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	us.Create(&model.User{Username: "admin", Email: "admin@bazi.com", PasswordHash: string(hash)})
	fs := &memFortuneStore{}

	baziSvc := &service.BaziService{}
	parser := &service.InputParser{}
	engine := service.NewFortuneEngine()
	ziweiSvc := service.NewZiWeiService()

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" { c.AbortWithStatus(204); return }
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status":"ok"}) })

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
		api.POST("/fortune/weekly", wh.Weekly)
		mh := &handler.MonthlyFortuneHandler{Engine: engine, ChartStore: cs}
		api.POST("/fortune/monthly", mh.HandleMonthly)
		ah := &handler.AIStubHandler{}
		api.POST("/fortune/ai", ah.AnalyzeFortune)
		api.GET("/auth/me", auth.Me)
		hh := &handler.HistoryHandler{Charts: cs, FortuneHistory: fs}
		api.GET("/charts", hh.ListCharts)
		api.GET("/charts/:id", hh.GetChart)
		api.GET("/fortune/history", hh.FortuneHistoryList)
		handler.RegisterZiWeiRoutes(r, ziweiSvc)
		handler.RegisterZiWeiPeriodRoutes(r, ziweiSvc, cs)
	}

	log.Println("Server starting on :8080")
	r.Run(":8080")
}
