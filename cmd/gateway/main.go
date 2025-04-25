package main

import (
	"log"
	"net/http/httputil"
	"net/url"
	"os"

	"StartupPCConfigurator/pkg/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "secret_key"
	}

	authURL := os.Getenv("AUTH_SERVICE_URL")
	configURL := os.Getenv("CONFIG_SERVICE_URL")
	agrURL := os.Getenv("AGGREGATOR_SERVICE_URL")

	router := gin.Default()
	router.Use(cors.Default())

	// 🔓 Публичный /auth/*
	router.Any("/auth/*proxyPath", reverseProxy(authURL))

	// 🔓 Публичные ручки config-сервиса
	router.GET("/config/components", reverseProxyPath(configURL, "/components"))
	router.GET("/config/compatible", reverseProxyPath(configURL, "/compatible"))
	router.GET("/config/usecases", reverseProxyPath(configURL, "/usecases"))
	router.POST("/config/generate", reverseProxyPath(configURL, "/generate"))

	router.GET("/config/usecase/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.Request.URL.Path = "/usecase/" + name
		proxyKeepPath(configURL)(c) // ← вместо reverseProxy
	})

	router.POST("/config/usecase/:name/generate", func(c *gin.Context) {
		name := c.Param("name")
		c.Request.URL.Path = "/usecase/" + name + "/generate"
		proxyKeepPath(configURL)(c) // ← вместо reverseProxy
	})

	// Защищённые ручки config-сервиса через /config-secure/*
	configProtected := router.Group("/config-secure")
	configProtected.Use(middleware.AuthMiddleware(jwtSecret))
	{
		configProtected.Any("/*proxyPath", reverseProxy(configURL))
	}

	// Защищённые ручки aggregator-сервиса через /offers/*
	offersGroup := router.Group("/offers")
	offersGroup.Use(middleware.AuthMiddleware(jwtSecret))
	{
		offersGroup.Any("/*proxyPath", reverseProxy(agrURL))
	}

	log.Println("Gateway запущен на :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Не удалось запустить gateway: %v", err)
	}
}

// reverseProxy — для маршрутов с *proxyPath
func reverseProxy(target string) gin.HandlerFunc {
	remote, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Невалидный адрес сервиса: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)

	return func(c *gin.Context) {
		c.Request.URL.Path = c.Param("proxyPath")
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// reverseProxyPath — для фиксированных путей без wildcard
func reverseProxyPath(target, path string) gin.HandlerFunc {
	remote, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Невалидный адрес сервиса: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)

	return func(c *gin.Context) {
		c.Request.URL.Path = path
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// proxyKeepPath — не переписывает путь, просто проксирует как есть
func proxyKeepPath(target string) gin.HandlerFunc {
	remote, err := url.Parse(target)
	if err != nil {
		log.Fatalf("invalid proxy url: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
