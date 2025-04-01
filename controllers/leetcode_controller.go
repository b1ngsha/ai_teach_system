package controllers

import (
	"ai_teach_system/services"
	"ai_teach_system/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LeetCodeController struct {
	service services.LeetCodeServiceInterface
}

type RunTestCaseRequest struct {
	Lang               string `json:"lang"`
	LeetcodeQuestionId int    `json:"leetcode_question_id"`
	TypedCode          string `json:"typed_code"`
}

type SubmitRequest struct {
	Lang               string `json:"lang"`
	KnowledgePointID   uint   `json:"knowledge_point_id"`
	LeetcodeQuestionId int    `json:"leetcode_question_id"`
	TypedCode          string `json:"typed_code"`
}

type CheckRequest struct {
	RunCodeID string `json:"run_code_id"`
	Test      bool   `json:"test"` // 是否进行检查测试用例
}

func NewLeetCodeController(service services.LeetCodeServiceInterface) *LeetCodeController {
	return &LeetCodeController{
		service: service,
	}
}

func (c *LeetCodeController) RunTestCase(ctx *gin.Context) {
	var req RunTestCaseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}
	userID := ctx.GetUint("userID")

	result, err := c.service.RunTestCase(userID, req.LeetcodeQuestionId, req.TypedCode, req.Lang)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("运行测试用例失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(result))
}

func (c *LeetCodeController) Submit(ctx *gin.Context) {
	var req SubmitRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}
	userID := ctx.GetUint("userID")

	result, err := c.service.Submit(userID, req.Lang, req.KnowledgePointID, req.LeetcodeQuestionId, req.TypedCode)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("提交代码失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(result))
}

func (c *LeetCodeController) Check(ctx *gin.Context) {
	var req CheckRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}
	userID := ctx.GetUint("userID")

	result, err := c.service.Check(userID, req.RunCodeID, req.Test)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("检查代码运行结果失败: %v", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success(result))
}
