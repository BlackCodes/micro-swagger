package router

import (
	"net/http"

	"github.com/BlackCodes/micro-swagger/web/handler"
	"github.com/gin-gonic/gin"
)

func Route(r *gin.Engine) {
	r.Use(CorsMiddleware())
	r.POST("/push/doc", handler.NewSwagger().Push)

	r.GET("/push/get/:project", handler.NewSwagger().Get)
}

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		// 核心处理方式
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS, POST, PUT, DELETE")
		c.Set("content-type", "application/json")
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, gin.H{})
		}

		c.Next()
	}
}
