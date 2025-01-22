package controllers

import (
	"ai_teach_system/services"
	"ai_teach_system/utils"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CourseController struct {
	courseService *services.CourseService
}

func NewCourseController(service *services.CourseService) *CourseController {
	return &CourseController{
		courseService: service,
	}
}

func (c *CourseController) GetCourseDetail(ctx *gin.Context) {
	courseID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error("无效的课程ID"))
		return
	}
	userID := ctx.GetUint("userID")
	course, points, err := c.courseService.GetCourseDetail(uint(courseID), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("获取课程详情失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(gin.H{
		"course": course,
		"points": points,
	}))
}
