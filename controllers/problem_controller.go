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

func (c *ProblemController) GetProblemList(ctx *gin.Context) {
	var req GetProblemListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && err != io.EOF {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	userID := ctx.GetUint("userID")
	response, err := c.service.GetProblemList(userID, req.Difficulty, req.KnowledgePointID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprint("获取题目列表失败")))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(response))
}

func (c *ProblemController) GetProblemDetail(ctx *gin.Context) {
	problemID, err := strconv.ParseUint(ctx.Param("id"), 10,  32)
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
