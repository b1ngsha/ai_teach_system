package controllers

import (
	"ai_teach_system/models"
	"ai_teach_system/services"
	"ai_teach_system/utils"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProblemController struct {
	service *services.ProblemService
}

func NewProblemController(service *services.ProblemService) *ProblemController {
	return &ProblemController{service: service}
}

type GetCourseProblemListRequest struct {
	Difficulty       models.ProblemDifficulty `json:"difficulty"`
	KnowledgePointID uint                     `json:"knowledge_point_id"`
	TagID            uint                     `json:"tag_id"`
}

type SetKnowledgePointTagsRequest struct {
	TagIDs []uint `json:"tag_ids"`
}

type GetProblemListRequest struct {
	Difficulty string `json:"difficulty"`
	TagID      uint   `json:"tag_id"`
}

type CreateCustomProblemRequest struct {
	Title           string                   `json:"title" binding:"required"`
	Content         string                   `json:"content" binding:"required"`
	Difficulty      models.ProblemDifficulty `json:"difficulty" binding:"required"`
	SampleTestcases string                   `json:"sample_testcases" binding:"required"`
	TestCases       string                   `json:"test_cases" binding:"required"`
	TagIDs          []uint                   `json:"tag_ids" binding:"required,min=1"`
	TimeLimit       int                      `json:"time_limit" binding:"required"`
	MemoryLimit     int                      `json:"memory_limit" binding:"required"`
}

func (c *ProblemController) GetCourseProblemList(ctx *gin.Context) {
	courseID, err := strconv.ParseUint(ctx.Param("course_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error("无效的课程ID"))
		return
	}

	var req GetCourseProblemListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && err != io.EOF {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	userID := ctx.GetUint("userID")
	response, err := c.service.GetCourseProblemList(uint(courseID), userID, req.Difficulty, req.KnowledgePointID, req.TagID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("获取题目列表失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(response))
}

func (c *ProblemController) GetProblemDetail(ctx *gin.Context) {
	problemID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error("无效的题目id"))
		return
	}

	problem, err := c.service.GetProblemDetail(uint(problemID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error("获取题目详情失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(problem))
}

func (c *ProblemController) SetKnowledgePointTags(ctx *gin.Context) {
	knowledgePointID, err := strconv.ParseUint(ctx.Param("knowledge_point_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error("无效的知识点ID"))
		return
	}

	var req SetKnowledgePointTagsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	result, err := c.service.SetKnowledgePointTags(uint(knowledgePointID), req.TagIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("设置知识点标签失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(result))
}

func (c *ProblemController) GetKnowledgePointProblems(ctx *gin.Context) {
	knowledgePointID, err := strconv.ParseUint(ctx.Param("knowledge_point_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error("无效的知识点ID"))
		return
	}

	userID := ctx.GetUint("userID")
	problems, err := c.service.GetKnowledgePointProblems(userID, uint(knowledgePointID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("获取知识点题目失败: %v", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success(problems))
}

func (c *ProblemController) GetProblemList(ctx *gin.Context) {
	var req GetProblemListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	problems, err := c.service.GetProblemList(req.Difficulty, req.TagID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("获取题目列表失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(problems))
}

func (c *ProblemController) GetAllTags(ctx *gin.Context) {
	tags, err := c.service.GetAllTags()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("获取标签列表失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(tags))
}

func (c *ProblemController) CreateCustomProblem(ctx *gin.Context) {
	var req CreateCustomProblemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	// 验证难度值是否合法
	if req.Difficulty != models.ProblemDifficultyEasy &&
		req.Difficulty != models.ProblemDifficultyMedium &&
		req.Difficulty != models.ProblemDifficultyHard {
		ctx.JSON(http.StatusBadRequest, utils.Error("无效的难度值"))
		return
	}

	problem := &models.Problem{
		TitleCn:         req.Title,
		ContentCn:       req.Content,
		Difficulty:      req.Difficulty,
		SampleTestcases: req.SampleTestcases,
		TestCases:       req.TestCases,
		IsCustom:        true,
		TimeLimit:       req.TimeLimit,
		MemoryLimit:     req.MemoryLimit,
	}

	createdProblem, err := c.service.CreateCustomProblem(problem, req.TagIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("创建题目失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(map[string]interface{}{
		"id":         createdProblem.ID,
		"title":      createdProblem.TitleCn,
		"difficulty": createdProblem.Difficulty,
	}))
}
