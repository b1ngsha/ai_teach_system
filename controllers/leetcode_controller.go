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
	LeetcodeQuestionId string `json:"leetcode_question_id"`
	TypedCode          string `json:"typed_code"`
}

type SubmitRequest struct {
	Lang               string `json:"lang"`
	LeetcodeQuestionId string `json:"leetcode_question_id"`
	TypedCode          string `json:"typed_code"`
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

	result, err := c.service.RunTestCase(req.LeetcodeQuestionId, req.TypedCode, req.Lang)
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

	result, err := c.service.Submit(req.Lang, req.LeetcodeQuestionId, req.TypedCode)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("提交代码失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(result))
}
