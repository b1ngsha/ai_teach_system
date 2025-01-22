package controllers

import (
	"ai_teach_system/models"
	"ai_teach_system/services"
	"ai_teach_system/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProblemController struct {
	service *services.ProblemService
}

func NewProblemController(service *services.ProblemService) *ProblemController {
	return &ProblemController{service: service}
}

type GetProblemListRequest struct {
	Difficulty       models.ProblemDifficulty `json:"difficulty" binding:""`
	KnowledgePointID uint                     `json:"knowledge_point_id"`
}

func (c *ProblemController) GetProblemList(ctx *gin.Context) {
	var req GetProblemListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	response, err := c.service.GetProblemList(req.Difficulty, req.KnowledgePointID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprint("获取题目列表失败")))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(response))
}
