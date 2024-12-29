package controllers

import (
	"ai_teach_system/models"
	"ai_teach_system/services"
	"log"
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

func (c *LeetCodeController) FetchProblem(ctx *gin.Context) {
	titleSlug := ctx.Param("titleSlug")

	problem, err := c.service.FetchProblemDetail(titleSlug)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 保存到数据库
	if err := c.db.Create(problem).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save problem"})
		return
	}

	ctx.JSON(http.StatusOK, problem)
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

func (c *LeetCodeController) FetchAllProblems(ctx *gin.Context) {
	problems, err := c.service.FetchAllProblems()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 批量保存到数据库
	for _, problem := range problems {
		if err := c.db.Create(problem).Error; err != nil {
			// 记录错误但继续处理
			log.Printf("保存题目失败 %d: %v", problem.LeetcodeID, err)
			continue
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Successfully fetched all problems",
		"count":   len(problems),
	})
}
