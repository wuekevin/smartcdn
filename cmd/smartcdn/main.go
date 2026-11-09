package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/wuekevin/smartcdn/internal/application/setup"
	"github.com/wuekevin/smartcdn/internal/handler/api"
	"github.com/wuekevin/smartcdn/internal/infra/crypto"
	"github.com/wuekevin/smartcdn/internal/infra/logger"
)

func main() {
	if err := crypto.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "FATAL:", err)
		os.Exit(1)
	}

	if err := logger.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "FATAL: logger init:", err)
		os.Exit(1)
	}
	defer logger.Sync()

	log := logger.Named("main")
	log.Info("SmartCDN starting")

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(api.RequestIDMiddleware())
	r.Use(api.SecurityHeadersMiddleware())

	v1 := r.Group("/api/v1")

	// Setup routes — not protected by SetupLockMiddleware so the check/complete
	// endpoints are reachable; the handler itself enforces the lock.
	setupSvc := setup.NewService()
	setupHandler := api.NewSetupHandler(setupSvc)
	setupHandler.RegisterSetup(v1)

	// Serve embedded React SPA
	api.RegisterSPA(r, frontendFS())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Info("listening on :" + port)
	if err := r.Run(":" + port); err != nil {
		log.Sugar().Fatalf("server error: %v", err)
	}
}
