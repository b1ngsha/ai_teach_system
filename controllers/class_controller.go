package controllers

import (
	"ai_teach_system/services"
	"ai_teach_system/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AddClassRequest struct {
	ClassName string `json:"class_name"`
}

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

func (c *ClassController) AddClass(ctx *gin.Context) {
	var req AddClassRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}
	class, err := c.classService.AddClass(req.ClassName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("创建班级失败: %v", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success(gin.H{
		"class_id":   class.ID,
		"class_name": class.Name,
	}))
}
