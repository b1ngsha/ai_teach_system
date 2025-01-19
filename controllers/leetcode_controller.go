package controllers

import (
	"ai_teach_system/models"
	"ai_teach_system/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LeetCodeController struct {
	db      *gorm.DB
	service *services.LeetCodeService
}

func NewLeetCodeController(db *gorm.DB) *LeetCodeController {
	return &LeetCodeController{
		db:      db,
		service: services.NewLeetCodeService(),
	}
}

func (c *LeetCodeController) GetProblem(ctx *gin.Context) {
	var problem models.Problem
	leetcodeID := ctx.Param("id")

	if err := c.db.Preload("Tags").Preload("KnowledgePoints").Where("leetcode_id = ?", leetcodeID).First(&problem).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Problem not found"})
		return
	}

	ctx.JSON(http.StatusOK, problem)
}
