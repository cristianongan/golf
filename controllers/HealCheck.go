package controllers

import (
	"github.com/gin-gonic/gin"
)

type HealCheck struct{}

func (_ *HealCheck) HealCheck(c *gin.Context) {
	okResponse(c, gin.H{"message": "0.02"})
}

func (_ *HealCheck) CheckIp(c *gin.Context) {
	c.String(200, c.ClientIP())
}
