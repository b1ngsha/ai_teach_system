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
	course, points, skillAnalysis, overview, err := c.courseService.GetCourseDetail(uint(courseID), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("获取课程详情失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(gin.H{
		"course":         course,
		"points":         points,
		"skill_analysis": skillAnalysis,
		"overview":       overview,
	}))
}

func (c *CourseController) GetKnowledgePoints(ctx *gin.Context) {
	courseID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error("无效的课程ID"))
		return
	}

	points, err := c.courseService.GetKnowledgePoints(uint(courseID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("获取知识点列表失败: %v", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success(points))
}

func (c *CourseController) GetCourseList(ctx *gin.Context) {
	courseNames, err := c.courseService.GetCourseList()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("获取课程列表失败: %v", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success(gin.H{
		"course_names": courseNames,
	}))
}

func (c *CourseController) GetUserListByCourseAndClass(ctx *gin.Context) {
	courseID, err := strconv.ParseUint(ctx.Param("course_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error("无效的课程ID"))
		return
	}
	classID, err := strconv.ParseUint(ctx.Param("class_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error("无效的班级ID"))
		return
	}
	result, err := c.courseService.GetUserListByCourseAndClass(uint(classID), uint(courseID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("查询学生列表失败: %v", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success(result))
}
