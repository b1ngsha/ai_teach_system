package api_test

import (
	"ai_teach_system/controllers"
	"ai_teach_system/models"
	"ai_teach_system/services"
	"ai_teach_system/tests"
	"bytes"
	"encoding/json"
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
	userController := controllers.NewUserController(userService)

	r.POST("/api/user/register", userController.Register)
	r.POST("/api/user/login", userController.Login)

	return r, db, cleanup
}

func TestUserRegister(t *testing.T) {
	r, db, cleanup := setupUserTest()
	defer cleanup()

	tests := []struct {
		name       string
		payload    services.RegisterRequest
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid registration",
			payload: services.RegisterRequest{
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
			payload: services.RegisterRequest{
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
			payload: services.RegisterRequest{
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
			jsonData, err := json.Marshal(tt.payload)
			assert.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantErr {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				var count int64
				db.Model(&models.User{}).Where("username = ?", tt.payload.Username).Count(&count)
				assert.Equal(t, int64(1), count)
			}
		})
	}
}

func TestUserLogin(t *testing.T) {
	r, db, cleanup := setupUserTest()
	defer cleanup()

	testUser := services.RegisterRequest{
		Username:  "testuser",
		Password:  "password123",
		Name:      "Test User",
		StudentID: "2024001",
		Class:     "CS-01",
	}

	userService := services.NewUserService(db)
	err := userService.Register(&testUser)
	assert.NoError(t, err)

	var count int64
	db.Model(&models.User{}).Where("username = ?", "testuser").Count(&count)
	assert.Equal(t, int64(1), count)

	tests := []struct {
		name       string
		payload    services.LoginRequest
		wantStatus int
		wantToken  bool
	}{
		{
			name: "valid login",
			payload: services.LoginRequest{
				StudentID: "2024001",
				Password:  "password123",
			},
			wantStatus: http.StatusOK,
			wantToken:  true,
		},
		{
			name: "invalid password",
			payload: services.LoginRequest{
				StudentID: "2024001",
				Password:  "wrongpassword",
			},
			wantStatus: http.StatusUnauthorized,
			wantToken:  false,
		},
		{
			name: "non-existent user",
			payload: services.LoginRequest{
				StudentID: "2024002",
				Password:  "password123",
			},
			wantStatus: http.StatusUnauthorized,
			wantToken:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.payload)
			assert.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/user/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.wantToken {
				assert.Contains(t, response, "token")
				assert.NotEmpty(t, response["token"])
			} else {
				assert.Contains(t, response, "error")
			}
		})
	}
}
