package controllers

import (
	"ai_teach_system/models"
	"ai_teach_system/services"
	"ai_teach_system/utils"
	"fmt"
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

func (c *LeetCodeController) RunTestCase(ctx *gin.Context) {
	var req RunTestCaseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	questionIdInt, _ := strconv.Atoi(req.QuestionId)
	var problem models.Problem
	c.db.Model(&models.Problem{}).Where("leetcode_id = ?", questionIdInt).First(&problem)
	result, err := c.service.RunTestCase(problem.TitleSlug, req.QuestionId, req.TypedCode, req.DataInput, req.Lang)
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

	questionIdInt, _ := strconv.Atoi(req.QuestionId)
	var problem models.Problem
	c.db.Model(&models.Problem{}).Where("leetcode_id = ?", questionIdInt).First(&problem)

	result, err := c.service.Submit(problem.TitleSlug, req.Lang, req.QuestionId, req.TypedCode)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("提交代码失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(result))
}
