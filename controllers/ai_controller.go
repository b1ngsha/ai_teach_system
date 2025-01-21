package controllers

import (
	"ai_teach_system/services"
	"ai_teach_system/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
	Content         string `json:"content"`
	SampleTestcases string `json:"sample_testcases"`
}

func (c *AIController) GenerateCode(ctx *gin.Context) {
	var req GenerateCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	code, err := c.Service.GenerateCode(&services.CodeGenerationRequest{
		Title:           req.Title,
		Language:        req.Language,
		Content:         req.Content,
		SampleTestcases: req.SampleTestcases,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("生成代码失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(gin.H{
		"code": code,
	}))
}
