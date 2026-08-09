package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"p-mon-backend/internal/api"
	"p-mon-backend/internal/auth"
	"p-mon-backend/internal/config"
	"p-mon-backend/internal/db"
	"p-mon-backend/internal/scheduler"
	"p-mon-backend/internal/ws"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load("config.json"); err != nil {
		log.Println("Starting with default config, could not load config.json:", err)
	}

	if err := db.InitDB(config.Cfg.DBPath); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	if err := db.Migrate("internal/db/schema.sql"); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	hub := ws.NewHub()
	ws.DefaultHub = hub
	go hub.Run()

	scheduler.Start()

	r := gin.Default()

	// CORS config
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/ws", ws.HandleWS(hub))
	r.GET("/install/agent.sh", api.ServeInstallScript)

	v1 := r.Group("/api/v1")
	{
		// Public Auth
		v1.POST("/auth/register", api.Register)
		v1.POST("/auth/login", api.Login)

		// Agent receiver
		v1.POST("/agent/:key", api.AgentReceive)

		// Protected Routes
		protected := v1.Group("")
		protected.Use(auth.JWTMiddleware())
		{
			protected.GET("/auth/me", api.Me)
			protected.POST("/auth/logout", api.Logout)

			// Servers
			protected.GET("/servers", api.ListServers)
			protected.POST("/servers", api.CreateServer)
			protected.GET("/servers/:id", api.GetServer)
			protected.PUT("/servers/:id", api.UpdateServer)
			protected.DELETE("/servers/:id", api.DeleteServer)
			protected.GET("/servers/:id/history", api.GetServerHistory)
			protected.GET("/servers/:id/incidents", api.GetServerIncidents)

			// Websites
			protected.GET("/websites", api.ListWebsites)
			protected.POST("/websites", api.CreateWebsite)
			protected.GET("/websites/:id", api.GetWebsite)
			protected.PUT("/websites/:id", api.UpdateWebsite)
			protected.DELETE("/websites/:id", api.DeleteWebsite)
			protected.GET("/websites/:id/history", api.GetWebsiteHistory)
			protected.GET("/websites/:id/incidents", api.GetWebsiteIncidents)

			// PAPI WhatsApp panels
			protected.GET("/papi/panels", api.ListPapiPanels)
			protected.POST("/papi/panels", api.CreatePapiPanel)
			protected.GET("/papi/panels/:id", api.GetPapiPanel)
			protected.PUT("/papi/panels/:id", api.UpdatePapiPanel)
			protected.DELETE("/papi/panels/:id", api.DeletePapiPanel)
			protected.POST("/papi/panels/:id/check", api.CheckPapiPanelNow)
			protected.GET("/papi/panels/:id/incidents", api.GetPapiPanelIncidents)

			// Checks
			protected.GET("/checks", api.ListChecks)
			protected.POST("/checks", api.CreateCheck)
			protected.GET("/checks/:id", api.GetCheck)
			protected.PUT("/checks/:id", api.UpdateCheck)
			protected.DELETE("/checks/:id", api.DeleteCheck)

			// Incidents
			protected.GET("/incidents", api.ListIncidents)
			protected.PUT("/incidents/:id/acknowledge", api.AcknowledgeIncident)
			protected.PUT("/incidents/:id/resolve", api.ResolveIncident)
			protected.PUT("/incidents/:id/ignore", api.IgnoreIncident)

			// Settings / Channels
			protected.GET("/settings", api.GetSettings)
			protected.PUT("/settings", api.UpdateSettings)
			protected.GET("/settings/channels", api.ListChannels)
			protected.POST("/settings/channels", api.CreateChannel)
			protected.PUT("/settings/channels/:id", api.UpdateChannel)
			protected.DELETE("/settings/channels/:id", api.DeleteChannel)
			protected.POST("/settings/channels/:id/test", api.TestChannel)

			// Alert Rules
			protected.GET("/alert-rules", api.ListRules)
			protected.POST("/alert-rules", api.CreateRule)
			protected.PUT("/alert-rules/:id", api.UpdateRule)
			protected.DELETE("/alert-rules/:id", api.DeleteRule)

			// Alert Templates
			protected.GET("/alert-templates", api.ListTemplates)
			protected.POST("/alert-templates", api.CreateTemplate)
			protected.PUT("/alert-templates/:id", api.UpdateTemplate)
			protected.DELETE("/alert-templates/:id", api.DeleteTemplate)
			protected.POST("/alert-templates/:alert_type/reset", api.ResetTemplate)

			// Company (whitelabel) + internal users
			protected.GET("/company", api.GetCompany)
			protected.PUT("/company", api.UpdateCompany)
			protected.GET("/company/users", api.ListCompanyUsers)
			protected.POST("/company/users", api.CreateCompanyUser)
			protected.PUT("/company/users/:id", api.UpdateCompanyUser)
			protected.DELETE("/company/users/:id", api.DeleteCompanyUser)
		}
	}

	addr := fmt.Sprintf("%s:%d", config.Cfg.Host, config.Cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		log.Printf("Listening on %s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
