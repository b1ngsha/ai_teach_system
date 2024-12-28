package controllers

import (
	"net/http"
	"ai_teach_system/models"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	var users []models.User
	// TODO: 实现获取用户列表的逻辑

	c.JSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 实现创建用户的逻辑

	c.JSON(http.StatusCreated, user)
}
