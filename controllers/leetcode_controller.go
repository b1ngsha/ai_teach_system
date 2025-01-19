package controllers

import (
	"ai_teach_system/models"
	"ai_teach_system/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LeetCodeController struct {
	db      *gorm.DB
	service services.LeetCodeServiceInterface
}

type RunTestCaseRequest struct {
	DataInput  string `json:"data_input"`
	Lang       string `json:"lang"`
	QuestionId string `json:"question_id"`
	TypedCode  string `json:"typed_code"`
}

type SubmitRequest struct {
	Lang       string `json:"lang"`
	QuestionId string `json:"question_id"`
	TypedCode  string `json:"typed_code"`
}

func NewLeetCodeController(db *gorm.DB, service services.LeetCodeServiceInterface) *LeetCodeController {
	return &LeetCodeController{
		db:      db,
		service: service,
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

func (c *LeetCodeController) RunTestCase(ctx *gin.Context) {
	var req RunTestCaseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	questionIdInt, _ := strconv.Atoi(req.QuestionId)
	var problem models.Problem
	c.db.Model(&models.Problem{}).Where("leetcode_id = ?", questionIdInt).First(&problem)
	result, err := c.service.RunTestCase(problem.TitleSlug, req.QuestionId, req.TypedCode, req.DataInput, req.Lang)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (c *LeetCodeController) Submit(ctx *gin.Context) {
	var req SubmitRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	questionIdInt, _ := strconv.Atoi(req.QuestionId)
	var problem models.Problem
	c.db.Model(&models.Problem{}).Where("leetcode_id = ?", questionIdInt).First(&problem)

	result, err := c.service.Submit(problem.TitleSlug, req.Lang, req.QuestionId, req.TypedCode)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
