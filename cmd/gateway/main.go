package main

import (
	"log"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

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
	notifURL := os.Getenv("NOTIFICATIONS_SERVICE_URL")
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	r := gin.Default()
	// Настраиваем CORS для разработки
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{frontendURL}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	// ---------- FRONTEND --------------------------------------------------------
	r.NoRoute(func(c *gin.Context) {
		// Проксируем все остальные запросы на фронтенд
		remote, err := url.Parse(frontendURL)
		if err != nil {
			log.Fatalf("невалидный адрес фронтенда: %v", err)
		}
		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	// ---------- AUTH --------------------------------------------------------
	r.Any("/auth/*proxyPath", reverseProxy(authURL))

	// ---------- CONFIG – публичные -----------------------------------------
	r.GET("/config/components", reverseProxyPath(configURL, "/components"))
	r.GET("/config/compatible", reverseProxyPath(configURL, "/compatible"))
	r.GET("/config/usecases", reverseProxyPath(configURL, "/usecases"))
	r.POST("/config/generate", reverseProxyPath(configURL, "/generate"))

	r.GET("/config/usecase/:name", func(c *gin.Context) {
		c.Request.URL.Path = "/usecase/" + c.Param("name")
		proxyKeepPath(configURL)(c)
	})
	r.POST("/config/usecase/:name/generate", func(c *gin.Context) {
		c.Request.URL.Path = "/usecase/" + c.Param("name") + "/generate"
		proxyKeepPath(configURL)(c)
	})

	// ---------- CONFIG – защищённые (JWT) ----------------------------------
	cfgSec := r.Group("/config", middleware.AuthMiddleware(jwtSecret))
	{
		cfgSec.POST("/newconfig", proxyStripPrefix(configURL, "/config"))
		cfgSec.GET("/userconf", proxyStripPrefix(configURL, "/config"))
		cfgSec.PUT("/newconfig/:configId", proxyStripPrefix(configURL, "/config"))
		cfgSec.DELETE("/newconfig/:configId", proxyStripPrefix(configURL, "/config"))
	}

	// ---------- AGGREGATOR – защищённые ------------------------------------
	offers := r.Group("/offers", middleware.AuthMiddleware(jwtSecret))
	{
		offers.Any("", proxyKeepPath(agrURL))
		offers.Any("/*proxyPath", proxyKeepPath(agrURL))
	}

	// **НОВЫЙ БЛОК**: /subscriptions → Notifications-service
	subs := r.Group("/subscriptions", middleware.AuthMiddleware(jwtSecret))
	{
		// POST /subscriptions
		subs.POST("", proxyKeepPath(notifURL))
		// GET  /subscriptions
		subs.GET("", proxyKeepPath(notifURL))
		// DELETE /subscriptions/:componentId
		subs.DELETE("/:componentId", proxyKeepPath(notifURL))
	}

	// ---------- NOTIFICATIONS – защищённые ---------------------------------
	notifications := r.Group("/notifications", middleware.AuthMiddleware(jwtSecret))
	{
		// этот маршрут поймает запрос GET /notifications
		notifications.Any("", proxyKeepPath(notifURL))
		// а этот — все вложенные, например /notifications/count и /notifications/{id}/read
		notifications.Any("/*proxyPath", proxyKeepPath(notifURL))
	}

	log.Println("Gateway запущен на :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("не удалось запустить gateway: %v", err)
	}

}

// ============================= proxy helpers ===============================

func reverseProxy(target string) gin.HandlerFunc {
	remote, err := url.Parse(target)
	if err != nil {
		log.Fatalf("невалидный адрес сервиса: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)

	return func(c *gin.Context) {
		c.Request.URL.Path = c.Param("proxyPath")
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func reverseProxyPath(target, path string) gin.HandlerFunc {
	remote, err := url.Parse(target)
	if err != nil {
		log.Fatalf("невалидный адрес сервиса: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)

	return func(c *gin.Context) {
		c.Request.URL.Path = path
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

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

// proxyStripPrefix удаляет prefix из пути перед проксированием
func proxyStripPrefix(target, prefix string) gin.HandlerFunc {
	remote, err := url.Parse(target)
	if err != nil {
		log.Fatalf("invalid proxy url: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)

	return func(c *gin.Context) {
		c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, prefix)
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
