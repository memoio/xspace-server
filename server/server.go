package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/memoio/xspace-server/docs"
	"github.com/memoio/xspace-server/server/router"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewServer(ctx context.Context, chain, sk, port string) (*http.Server, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20 // 8 MiB

	r.Use(router.Cors())
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Xspace Server",
		})
	})

	err := router.NewRouter(ctx, chain, sk, r.Group("/v1"))
	if err != nil {
		return nil, err
	}

	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}, nil
}
