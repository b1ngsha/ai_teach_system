package controllers

import (
	"ai_teach_system/services"
	"ai_teach_system/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CodeGenerationRequest struct {
	Title           string // 题目
	Language        string // 编程语言
	Content         string // 题目内容
	SampleTestcases string // 示例测试用例
}

type AIController struct {
	Service services.AIServiceInterface
}

func NewAIController(service services.AIServiceInterface) *AIController {
	return &AIController{
		Service: service,
	}
}

type GenerateCodeRequest struct {
	Title           string `json:"title" binding:"required"`
	Language        string `json:"language" binding:"required"`
	Content         string `json:"content" binding:"required"`
	SampleTestcases string `json:"sample_testcases" binding:"required"`
}

type CorrectCodeRequest struct {
	ProblemID uint   `json:"problem_id" binding:"required"`
	Language  string `json:"language" binding:"required"`
	TypedCode string `json:"typed_code" binding:"required"`
}

type AnalyzeCodeRequest struct {
	ProblemID uint   `json:"problem_id" binding:"required"`
	Language  string `json:"language" binding:"required"`
	TypedCode string `json:"typed_code" binding:"required"`
}

type ChatRequest struct {
	ProblemID uint   `json:"problem_id" binding:"required"`
	Question  string `json:"question" binding:"required"`
	TypedCode string `json:"typed_code" binding:"required"`
}

func (c *AIController) GenerateCode(ctx *gin.Context) {
	var req GenerateCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	code, err := c.Service.GenerateCode(req.Title, req.Language, req.Content, req.SampleTestcases)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("生成代码失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(gin.H{
		"code": code,
	}))
}

func (c *AIController) CorrectCode(ctx *gin.Context) {
	var req CorrectCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	code, err := c.Service.CorrectCode(req.ProblemID, req.Language, req.TypedCode)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("生成代码失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(gin.H{
		"code": code,
	}))
}

func (c *AIController) AnalyzeCode(ctx *gin.Context) {
	var req AnalyzeCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	message, err := c.Service.AnalyzeCode(req.ProblemID, req.Language, req.TypedCode)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("生成代码失败: %v", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success(gin.H{
		"message": message,
	}))
}

func (c *AIController) Chat(ctx *gin.Context) {
	var req ChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	message, err := c.Service.Chat(req.ProblemID, req.TypedCode, req.Question)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("问答异常: %v", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success(gin.H{
		"message": message,
	}))
}
