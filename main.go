package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Aloe-Corporation/logs"
	"github.com/FloRichardAloeCorp/gateway/internal/configuration"
	"github.com/FloRichardAloeCorp/gateway/internal/proxy"
	"github.com/FloRichardAloeCorp/gateway/internal/service"
	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var log = logs.Get()

const (
	PREFIX_ENV          = "GATEWAY"
	ENV_CONFIG          = PREFIX_ENV + "_CONFIG"
	DEFAULT_PATH_CONFIG = "/config/"
)

func main() {
	log.Info("loading configuration...")
	configFilePath, present := os.LookupEnv(ENV_CONFIG)
	if !present {
		configFilePath = DEFAULT_PATH_CONFIG
	}
	config, err := configuration.LoadConf(configFilePath, "GATEWAY")
	if err != nil {
		panic(err)
	}
	log.Info("configuration loaded")

	log.Info("proxy package initialization...")
	proxy.Init()
	log.Info("proxy package initialized")

	router := gin.New()
	router.Use(ginzap.RecoveryWithZap(log, true))
	router.Use(ginzap.Ginzap(log, time.RFC3339, true))
	router.Use(cors.New(cors.Config{
		AllowOrigins:     config.Server.Cors.AllowOrigins,
		AllowMethods:     config.Server.Cors.AllowMethods,
		AllowHeaders:     config.Server.Cors.AllowHeaders,
		ExposeHeaders:    config.Server.Cors.ExposeHeaders,
		AllowCredentials: config.Server.Cors.AllowCredentials,
		MaxAge:           config.Server.Cors.MaxAge,
	}))

	log.Info("Creating endpoints...")
	for _, serviceConf := range config.Services {
		service, err := service.New(serviceConf)
		if err != nil {
			panic(err)
		}
		service.AttachEndpoints(router)
	}
	log.Info("endpoints created")

	addrGin := ":" + strconv.Itoa(config.Server.Port)
	srv := &http.Server{
		ReadHeaderTimeout: time.Millisecond,
		Addr:              addrGin,
		Handler:           router,
	}

	go RunGin(addrGin, router)

	WaitSignalShutdown(srv)
}

func RunGin(addr string, engine *gin.Engine) {
	log.Info("REST API listening on : "+addr,
		zap.String("package", "main"))

	log.Error(engine.Run(addr).Error(),
		zap.String("package", "main"))
}

func WaitSignalShutdown(srv *http.Server) {
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutdown Server ...")

	// Time to wait before close forcing
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server Shutdown: ", zap.Error(err))
	}

	log.Info("Server exiting")
}
