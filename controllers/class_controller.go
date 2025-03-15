package controllers

import (
	"ai_teach_system/services"
	"ai_teach_system/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ClassController struct {
	classService *services.ClassService
}

func NewClassController(service *services.ClassService) *ClassController {
	return &ClassController{
		classService: service,
	}
}

func (c *ClassController) GetClassList(ctx *gin.Context) {
	classNames, err := c.classService.GetClassList()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("获取班级列表失败: %v", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success(gin.H{
		"class_names": classNames,
	}))
}
