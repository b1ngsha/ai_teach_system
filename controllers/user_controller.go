package controllers

import (
	"ai_teach_system/models"
	"ai_teach_system/services"
	"ai_teach_system/utils"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username  string
	Password  string
	Name      string
	StudentID string
	Class     string
}

type SelectCourseRequest struct {
	CourseID uint `json:"course_id" binding:"required"`
}

type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

type UpdateUserRequest struct {
	Username  string      `json:"username"`
	Password  string      `json:"password"`
	Name      string      `json:"name"`
	StudentID string      `json:"student_id"`
	Role      models.Role `json:"role"`
	ClassID   uint        `json:"class_id"`
}

type UserController struct {
	userService *services.UserService
	ossService  services.OSSServiceInterface
}

func NewUserController(service *services.UserService, ossService services.OSSServiceInterface) *UserController {
	return &UserController{
		userService: service,
		ossService:  ossService,
	}
}

func (c *UserController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	token, err := c.userService.Login(req.StudentID, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.Error(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(gin.H{
		"token": token,
	}))
}

func (c *UserController) Register(ctx *gin.Context) {
	var req RegisterRequest

	if err := ctx.Request.ParseMultipartForm(32 << 20); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	req.Username = ctx.Request.FormValue("username")
	req.Password = ctx.Request.FormValue("password")
	req.Name = ctx.Request.FormValue("name")
	req.StudentID = ctx.Request.FormValue("student_id")
	req.Class = ctx.Request.FormValue("class")

	if req.Username == "" || req.Password == "" || req.Name == "" || req.StudentID == "" || req.Class == "" {
		ctx.JSON(http.StatusBadRequest, utils.Error("缺少必填字段"))
		return
	}

	user, err := c.userService.Register(req.Username, req.Password, req.Name, req.StudentID, req.Class)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, utils.Success(user))
}

func (c *UserController) GetUserInfo(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	userInfo, err := c.userService.GetUserInfo(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("获取用户信息失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(userInfo))
}

func (c *UserController) GetCourseTryRecords(ctx *gin.Context) {
	var userID uint
	// 首先尝试从query中获取user_id
	userIDstr := ctx.Query("user_id")
	if userIDstr != "" {
		var err error
		var userID64 uint64
		userID64, err = strconv.ParseUint(userIDstr, 10, 32)
		userID = uint(userID64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, utils.Error("无效的 user_id 参数"))
			return
		}
	} else {
		// 如果query参数中取不到，则从context当中取
		userID = ctx.GetUint("userID")
	}

	courseID, err := strconv.ParseUint(ctx.Param("course_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error("无效的 course_id 参数"))
		return
	}

	records, err := c.userService.GetCourseTryRecords(uint(courseID), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("获取用户答题记录失败: %v", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success(records))
}

func (c *UserController) GetTryRecordDetail(ctx *gin.Context) {
	recordID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error("无效的答题记录ID"))
		return
	}

	record, err := c.userService.GetTryRecordDetail(uint(recordID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("获取用户答题详情失败: %v", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success(record))
}

func (c *UserController) GetUserListByCourseAndClass(ctx *gin.Context) {
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
	result, err := c.userService.GetUserListByCourseAndClass(uint(classID), uint(courseID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("查询学生列表失败: %v", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success(result))
}

func (c *UserController) GetUserListByClass(ctx *gin.Context) {
	classID, err := strconv.ParseUint(ctx.Param("class_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error("无效的班级ID"))
		return
	}

	result, err := c.userService.GetUserListByClass(uint(classID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("查询学生列表失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(result))
}

func (c *UserController) SelectCourse(ctx *gin.Context) {
	var req SelectCourseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	// 直接将course_id设置到上下文中
	ctx.Set("courseID", req.CourseID)

	ctx.JSON(http.StatusOK, utils.Success(nil))
}

func (c *UserController) ResetPassword(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	var req ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	err := c.userService.ResetPassword(userID, req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("重置密码失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(nil))
}

func (c *UserController) GetTryRecords(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	result, err := c.userService.GetTryRecords(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("获取作答历史记录失败: %v", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(result))
}

func (c *UserController) UpdateUser(ctx *gin.Context) {
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error("无效的用户ID"))
		return
	}

	var req UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error(err.Error()))
		return
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Password != "" {
		updates["password"] = req.Password
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.StudentID != "" {
		updates["student_id"] = req.StudentID
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.ClassID != 0 {
		updates["class_id"] = req.ClassID
	}

	if err := c.userService.UpdateUser(uint(userID), updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(nil))
}

func (c *UserController) DeleteUser(ctx *gin.Context) {
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.Error("无效的用户ID"))
		return
	}

	if err := c.userService.DeleteUser(uint(userID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success(nil))
}
