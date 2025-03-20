package controllers

import (
	"ai_teach_system/services"
	"ai_teach_system/utils"
	"fmt"
	"mime/multipart"
	"net/http"

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
	Avatar    *multipart.FileHeader
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

	var avatarURL string
	file, err := ctx.FormFile("avatar")
	if err == nil {
		if !utils.IsValidImageFile(file.Filename) {
			ctx.JSON(http.StatusBadRequest, utils.Error("只支持上传图片文件"))
			return
		}

		avatarURL, err = c.ossService.UploadAvatar(file)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, utils.Error(fmt.Sprintf("上传头像失败: %v", err)))
			return
		}
	}

	user, err := c.userService.Register(req.Username, req.Password, req.Name, req.StudentID, req.Class, avatarURL)
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
