package api_test

import (
	"ai_teach_system/controllers"
	"ai_teach_system/models"
	"ai_teach_system/services"
	"ai_teach_system/tests"
	"ai_teach_system/tests/mocks"
	"ai_teach_system/utils"
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupUserTest() (*gin.Engine, *gorm.DB, func()) {
	gin.SetMode(gin.TestMode)
	db, cleanup := tests.SetupTestDB()

	r := gin.New()
	userService := services.NewUserService(db)
	mockOSSService := &mocks.MockOSSService{
		UploadAvatarFunc: func(file *multipart.FileHeader) (string, error) {
			return "https://example.com/avatars/test.jpg", nil
		},
	}
	userController := controllers.NewUserController(userService, mockOSSService)

	// 公开路由
	r.POST("/api/users/register", userController.Register)
	r.POST("/api/users/login", userController.Login)

	// 需要认证的路由
	auth := r.Group("/api/users")
	auth.Use(func(c *gin.Context) {
		c.Set("userID", uint(1)) // 在测试中模拟认证用户
		c.Next()
	})
	auth.GET("", userController.GetUserInfo)

	return r, db, cleanup
}

func TestUserRegister(t *testing.T) {
	r, db, cleanup := setupUserTest()
	defer cleanup()

	tests := []struct {
		name       string
		payload    controllers.RegisterRequest
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid registration",
			payload: controllers.RegisterRequest{
				Username:  "testuser",
				Password:  "password123",
				Name:      "Test User",
				StudentID: "2024001",
				Class:     "CS-01",
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name: "duplicate username",
			payload: controllers.RegisterRequest{
				Username:  "testuser",
				Password:  "password123",
				Name:      "Test User 2",
				StudentID: "2024002",
				Class:     "CS-01",
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "empty username",
			payload: controllers.RegisterRequest{
				Username:  "",
				Password:  "password123",
				Name:      "Test User",
				StudentID: "2024003",
				Class:     "CS-01",
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			_ = writer.WriteField("username", tt.payload.Username)
			_ = writer.WriteField("password", tt.payload.Password)
			_ = writer.WriteField("name", tt.payload.Name)
			_ = writer.WriteField("student_id", tt.payload.StudentID)
			_ = writer.WriteField("class", tt.payload.Class)

			part, err := writer.CreateFormFile("avatar", "test_avatar.jpg")
			assert.NoError(t, err)
			_, err = part.Write([]byte("fake image content"))
			assert.NoError(t, err)

			err = writer.Close()
			assert.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/users/register", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantErr {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response["result"].(bool))
			} else {
				var count int64
				db.Model(&models.User{}).Where("username = ?", tt.payload.Username).Count(&count)
				assert.Equal(t, int64(1), count)

				var user models.User
				err := db.Where("username = ?", tt.payload.Username).First(&user).Error
				assert.NoError(t, err)
				assert.NotEmpty(t, user.Avatar)
				assert.Contains(t, user.Avatar, "https://example.com/avatars/test.jpg")
			}
		})
	}
}

func TestUserLogin(t *testing.T) {
	r, _, cleanup := setupUserTest()
	defer cleanup()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("username", "testuser")
	_ = writer.WriteField("password", "password123")
	_ = writer.WriteField("name", "Test User")
	_ = writer.WriteField("student_id", "2024001")
	_ = writer.WriteField("class", "CS-01")

	part, err := writer.CreateFormFile("avatar", "test_avatar.jpg")
	assert.NoError(t, err)
	_, err = part.Write([]byte("fake image content"))
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	// 注册测试用户
	req := httptest.NewRequest("POST", "/api/users/register", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	tests := []struct {
		name       string
		studentID  string
		password   string
		wantStatus int
		wantToken  bool
	}{
		{
			name:       "valid login",
			studentID:  "2024001",
			password:   "password123",
			wantStatus: http.StatusOK,
			wantToken:  true,
		},
		{
			name:       "invalid password",
			studentID:  "2024001",
			password:   "wrongpassword",
			wantStatus: http.StatusUnauthorized,
			wantToken:  false,
		},
		{
			name:       "non-existent user",
			studentID:  "2024002",
			password:   "password123",
			wantStatus: http.StatusUnauthorized,
			wantToken:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loginData := map[string]string{
				"student_id": tt.studentID,
				"password":   tt.password,
			}

			jsonData, err := json.Marshal(loginData)
			assert.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/users/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response utils.Response
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.wantToken {
				assert.True(t, response.Result)
				assert.Empty(t, response.Message)
				data := response.Data.(map[string]interface{})
				assert.NotEmpty(t, data["token"])
			} else {
				assert.False(t, response.Result)
				assert.NotEmpty(t, response.Message)
				assert.Nil(t, response.Data)
			}
		})
	}
}

func TestGetUserProgress(t *testing.T) {
	r, db, cleanup := setupUserTest()
	defer cleanup()

	class := &models.Class{
		Name: "test_class",
	}

	user := &models.User{
		Username:  "testuser",
		Avatar:    "https://example.com/avatar.jpg",
		Name:      "Test User",
		StudentID: "2024001",
		Class:     *class,
		ClassID:   class.ID,
	}
	db.Create(user)

	problem := &models.Problem{
		Title:      "Two Sum",
		LeetcodeID: 1,
		TitleSlug:  "two-sum",
		Difficulty: "Easy",
	}
	db.Create(problem)

	userProblem := &models.UserProblem{
		UserID:    user.ID,
		ProblemID: problem.ID,
		Status:    models.ProblemStatusSolved,
	}
	db.Create(userProblem)

	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Result)
	assert.Empty(t, response.Message)

	data := response.Data.(map[string]interface{})
	assert.Equal(t, "testuser", data["username"])
	assert.Equal(t, "https://example.com/avatar.jpg", data["avatar"])
	assert.Equal(t, float64(1), data["solved_problems"])
	assert.Equal(t, float64(100), data["learn_progress"])
}
