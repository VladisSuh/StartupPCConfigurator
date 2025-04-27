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

	r := gin.Default()
	r.Use(cors.Default())

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
