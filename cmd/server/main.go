package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/fahimulhaque/tts-one-click/internal/api"
	"github.com/fahimulhaque/tts-one-click/internal/config"
	"github.com/fahimulhaque/tts-one-click/internal/tts"
	"github.com/fahimulhaque/tts-one-click/pkg/logger"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config file")
	dev := flag.Bool("dev", false, "development mode (skip python subprocess)")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	cfg.DevMode = *dev

	log := logger.New(cfg.DevMode)
	defer log.Sync()

	mgr := tts.NewManager(cfg, log)
	if !cfg.DevMode {
		tts.FreePort(cfg.ServerPort, log)
		if err := mgr.Start(); err != nil {
			log.Sugar().Fatalf("start python server: %v", err)
		}
		defer mgr.Stop()
	}

	if !cfg.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	h := api.NewHandlers(mgr.BaseURL())
	api.RegisterRoutes(r, h)

	// Serve React static files
	r.Static("/assets", "./web/dist/assets")
	r.NoRoute(func(c *gin.Context) {
		c.File("./web/dist/index.html")
	})

	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Sugar().Infof("TTS-One-Click running at http://localhost%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Sugar().Fatalf("server error: %v", err)
	}
}
