package middlewares

import (
	"log"

	"github.com/gin-gonic/gin"
)

func T1AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("auth t1 middleware")
		c.Next()
	}
}

func T2AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("auth t2 middleware")
		c.Next()
	}
}
