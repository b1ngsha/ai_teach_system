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

type GetProblemListRequest struct {
	Difficulty       models.ProblemDifficulty `json:"difficulty"`
	KnowledgePointID uint                     `json:"knowledge_point_id"`
}

type SetKnowledgePointProblemsRequest struct {
	ProblemsIDs      []uint `json:"problem_ids"`
}

func (c *ProblemController) GetProblemList(ctx *gin.Context) {
	courseID, err := strconv.ParseUint(ctx.Param("course_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error("无效的课程ID"))
		return
	}

	var req GetProblemListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && err != io.EOF {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	userID := ctx.GetUint("userID")
	response, err := c.service.GetProblemList(uint(courseID), userID, req.Difficulty, req.KnowledgePointID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprint("获取题目列表失败: %v", err)))
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

func (c *ProblemController) SetKnowledgePointProblems(ctx *gin.Context) {
	knowledgePointID, err := strconv.ParseUint(ctx.Param("knowledge_point_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error("无效的知识点ID"))
		return
	}

	var req SetKnowledgePointProblemsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	result, err := c.service.SetKnowledgePointProblems(uint(knowledgePointID), req.ProblemsIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("设置课程题目失败: %v", err)))
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
	problems, err := c.service.GetKnowledgePointProblems(uint(knowledgePointID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("获取知识点题目失败: %v", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success(problems))
}
