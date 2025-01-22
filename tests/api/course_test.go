package api_test

import (
	"ai_teach_system/controllers"
	"ai_teach_system/models"
	"ai_teach_system/services"
	"ai_teach_system/tests"
	"ai_teach_system/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupCourseTest() (*gin.Engine, *gorm.DB, func()) {
	gin.SetMode(gin.TestMode)
	db, cleanup := tests.SetupTestDB()

	r := gin.New()
	courseService := services.NewCourseService(db)
	courseController := controllers.NewCourseController(courseService)

	auth := r.Group("/api/courses")
	auth.Use(func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Next()
	})
	auth.GET("/:id", courseController.GetCourseDetail)

	return r, db, cleanup
}

func TestGetCourseDetail(t *testing.T) {
	r, db, cleanup := setupCourseTest()
	defer cleanup()

	user := &models.User{
		Username:  "testuser",
		Name:      "Test User",
		StudentID: "2024001",
		Class:     "CS-01",
	}
	db.Create(user)

	course := &models.Course{
		Name: "数据结构与算法",
	}
	db.Create(course)

	point := &models.KnowledgePoint{
		Name:     "数组",
		CourseID: course.ID,
	}
	db.Create(point)

	tag := &models.Tag{
		Name:             "数组操作",
		KnowledgePointID: point.ID,
	}
	db.Create(tag)

	problem := &models.Problem{
		Title:      "Two Sum",
		LeetcodeID: 1,
		Difficulty: "Easy",
	}
	db.Create(problem)
	db.Model(problem).Association("Tags").Append(tag)

	userProblem := &models.UserProblem{
		UserID:    user.ID,
		ProblemID: problem.ID,
		Status:    models.ProblemStatusTried,
	}
	db.Create(&userProblem)

	req := httptest.NewRequest("GET", "/api/courses/"+fmt.Sprint(course.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Result)

	data := response.Data.(map[string]interface{})
	courseData := data["course"].(map[string]interface{})
	points := data["points"].([]interface{})

	assert.Equal(t, "数据结构与算法", courseData["name"])
	assert.Len(t, points, 1)

	point0 := points[0].(map[string]interface{})
	assert.Equal(t, "数组", point0["name"])
	assert.Equal(t, float64(1), point0["problem_count"])

	tags := point0["tags"].([]interface{})
	assert.Len(t, tags, 1)
	assert.Equal(t, "数组操作", tags[0].(map[string]interface{})["name"])

	skillAnalysis := data["skill_analysis"].([]interface{})
	assert.Len(t, skillAnalysis, 1)
	analysis0 := skillAnalysis[0].(map[string]interface{})
	assert.Equal(t, "数组", analysis0["knowledge_point"])
	assert.Equal(t, float64(0), analysis0["correct_rate"])
	assert.Equal(t, float64(1), analysis0["total_attempts"])
	assert.Equal(t, float64(0), analysis0["correct_count"])

	overview := data["overview"].(map[string]interface{})
	assert.Equal(t, float64(1), overview["total_problems"])
	assert.Equal(t, float64(1), overview["attempted_problems"])
	assert.Equal(t, float64(0), overview["correct_rate"])
	assert.Equal(t, float64(1), overview["wrong_problems"])
}
