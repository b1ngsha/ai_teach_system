package controllers

import (
	"ai_teach_system/services"
	"ai_teach_system/utils"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AddCourseRequest struct {
	CourseName          string   `json:"course_name"`
	KnowledgePointNames []string `json:"knowledge_point_names"`
}

type SetKnowledgePointProblemsRequest struct {
	CourseID         uint   `json:"course_id"`
	KnowledgePointID uint   `json:"knowledge_point_id"`
	ProblemsIDs      []uint `json:"problem_ids"`
}

type SetClassCoursesRequest struct {
	CourseID uint   `json:"course_id"`
	ClassIDs []uint `json:"class_ids"`
}

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

func (c *CourseController) AddCourse(ctx *gin.Context) {
	var req AddCourseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}
	course, err := c.courseService.AddCourse(req.CourseName, req.KnowledgePointNames)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("创建课程失败: %v", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success(course))
}

func (c *CourseController) SetKnowledgePointProblems(ctx *gin.Context) {
	var req SetKnowledgePointProblemsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}
	result, err := c.courseService.SetKnowledgePointProblems(req.KnowledgePointID, req.ProblemsIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("设置课程题目失败: %v", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success(result))
}

func (c *CourseController) GetKnowledgePointProblems(ctx *gin.Context) {
	knowledgePointID, err := strconv.ParseUint(ctx.Param("knowledge_point_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error("无效的知识点ID"))
		return
	}
	problems, err := c.courseService.GetKnowledgePointProblems(uint(knowledgePointID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("获取知识点题目失败: %v", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success(problems))
}

func (c *CourseController) SetCourseClasses(ctx *gin.Context) {
	var req SetClassCoursesRequest
	if err := ctx.ShouldBindJSON(&req); err!= nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}
	result, err := c.courseService.SetCourseClasses(req.CourseID, req.ClassIDs)
	if err!= nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("设置课程班级失败: %v", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success(result))
}
