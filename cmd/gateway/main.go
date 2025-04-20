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
	// Считываем JWT секрет из переменной окружения или хардкодим
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "secret_key"
	}

	router := gin.Default()

	// CORS (по необходимости можно расширить)
	router.Use(cors.Default())

	// Публичный /auth/** — без авторизации
	router.Any("/auth/*proxyPath", reverseProxy("http://localhost:8001"))

	// Группа с авторизацией
	authGroup := router.Group("/")
	authGroup.Use(middleware.AuthMiddleware(jwtSecret))
	{
		authGroup.Any("/config/*proxyPath", reverseProxy("http://localhost:8002"))
		authGroup.Any("/offers/*proxyPath", reverseProxy("http://localhost:8003"))
	}

	// Запуск
	log.Println("Gateway запущен на :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Не удалось запустить gateway: %v", err)
	}
}

// reverseProxy возвращает gin.HandlerFunc, которая проксирует запрос
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
