package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"scratch/tocken"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Router struct {
	*gin.Engine
}

func NewRouter(scratchHandler ScratchHandler, tockenMaker tocken.Maker) (*Router, error) {
	env := os.Getenv("APP_ENV")
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
		logFile, _ := os.Create("gin.log")
		gin.DefaultWriter = io.Writer(logFile)
	}

	// Setup CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	config.AllowHeaders = []string{"*"}
	config.AllowBrowserExtensions = true
	config.AllowMethods = []string{"*"}

	router := gin.New()
	router.RedirectTrailingSlash = false

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": []string{"Invalid Path"},
			"errorno": []string{"INV1"},
		})
	})

	router.Use(
		gin.LoggerWithFormatter(customLogger),
		gin.Recovery(),
		cors.New(config),
	)

	// Group routes
	api := router.Group("/api")
	{
		api.POST("/create", scratchHandler.CreateScratchHandler)

		authRoutes := router.Group("/auth").Use(AuthMiddleware(tockenMaker))
		{
			authRoutes.GET("/:name", scratchHandler.FetchScratchHandler)
		}

	}

	return &Router{router}, nil
}

func (r *Router) Serve(listenAddr string) error {
	return r.Run(listenAddr)
}

func customLogger(param gin.LogFormatterParams) string {
	return fmt.Sprintf("[%s] - %s \"%s %s %s %d %s [%s]\"\n",
		param.TimeStamp.Format(time.RFC1123),
		param.ClientIP,
		param.Method,
		param.Path,
		param.Request.Proto,
		param.StatusCode,
		param.Latency.Round(time.Millisecond),
		param.Request.UserAgent(),
	)
}

// func ValidateContentType(allowedTypes []string) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		contentType := c.GetHeader("Content-Type")

// 		valid := false
// 		for _, allowed := range allowedTypes {
// 			if contentType == allowed {
// 				valid = true
// 				break
// 			}
// 		}

// 		if !valid {
// 			c.JSON(http.StatusUnsupportedMediaType, gin.H{
// 				"success": false,
// 				"message": []string{"Invalid Content-Type. Supported types are: " + fmt.Sprintf("%v", allowedTypes)},
// 				"errorno": []string{"USP1"},
// 			})
// 			c.Abort()
// 			return
// 		}
// 		c.Next()
// 	}
// }
