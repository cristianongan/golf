package controllers

import (
	"start/models"

	"github.com/gin-gonic/gin"
)

type CUser struct{}

func (_ *CUser) Test1(c *gin.Context) {
	okResponse(c, gin.H{"message": "success"})
}

func (_ *CUser) Test(c *gin.Context, prof models.User) {
	okResponse(c, gin.H{"message": "success"})
}
