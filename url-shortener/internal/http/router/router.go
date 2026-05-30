package router

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"url-shortener/internal/http/handlers"
	"url-shortener/internal/http/middleware"
	"url-shortener/internal/service"
)

func New(linkService *service.LinkService, authService *service.AuthService, logger *zap.Logger) *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},
	}))

	r.GET("/health", handlers.Check)

	linkHandler := handlers.NewLinkHandler(linkService, logger)
	authHandler := handlers.NewAuthHandler(authService, logger)

	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	api := r.Group("/api")
	api.Use(middleware.AuthRequired(authService, logger))
	api.GET("/links", linkHandler.ListUserLinks)
	api.GET("/links/:code/stats", linkHandler.GetLinkStats)
	api.GET("/links/:code/qr", linkHandler.GetLinkQR)
	api.PATCH("/links/:code", linkHandler.UpdateLink)
	api.DELETE("/links/:code", linkHandler.DeleteLink)

	r.POST("/shorten", middleware.OptionalAuth(authService, logger), linkHandler.Shorten)

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if c.Request.Method == http.MethodGet && !strings.HasPrefix(path, "/api/") && !strings.HasPrefix(path, "/auth/") {
			linkHandler.Redirect(c)
			return
		}

		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "not_found",
				"message": "Route not found",
			},
		})
	})

	return r
}
