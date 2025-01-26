package api_test

import (
	"ai_teach_system/controllers"
	"ai_teach_system/models"
	"ai_teach_system/services"
	"ai_teach_system/tests"
	"ai_teach_system/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupProblemTest() (*gin.Engine, *gorm.DB, func()) {
	gin.SetMode(gin.TestMode)
	db, cleanup := tests.SetupTestDB()

	r := gin.New()
	problemService := services.NewProblemService(db)
	problemController := controllers.NewProblemController(problemService)

	auth := r.Group("/api/problems")
	auth.Use(func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Next()
	})
	auth.POST("", problemController.GetProblemList)
	auth.GET("/:id", problemController.GetProblemDetail)
	return r, db, cleanup
}

func TestGetProblemList(t *testing.T) {
	r, db, cleanup := setupProblemTest()
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
		TitleSlug:  "Two Sum",
		LeetcodeID: 1,
		Difficulty: models.ProblemDifficultyEasy,
	}
	db.Create(problem)
	db.Model(problem).Association("Tags").Append(tag)

	userProblem := &models.UserProblem{
		UserID:    user.ID,
		ProblemID: problem.ID,
		Status:    models.ProblemStatusTried,
	}
	db.Create(&userProblem)

	requestBody := map[string]interface{}{
		"knowledge_point_id": point.ID,
		"difficulty":         problem.Difficulty,
	}
	jsonData, err := json.Marshal(requestBody)
	body := bytes.NewBuffer(jsonData)
	req := httptest.NewRequest("POST", "/api/problems", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Result)

	data := response.Data.([]interface{})
	assert.Len(t, data, 1)

	difficulty := models.ProblemDifficulty(data[0].(map[string]interface{})["difficulty"].(string))
	assert.Equal(t, problem.Difficulty, difficulty)

	leetcodeID := data[0].(map[string]interface{})["leetcode_id"].(float64)
	assert.Equal(t, problem.LeetcodeID, int(leetcodeID))

	status := models.ProblemStatus(data[0].(map[string]interface{})["status"].(string))
	assert.Equal(t, userProblem.Status, status)
}

func TestGetProblemDetail(t *testing.T) {
	r, db, cleanup := setupProblemTest()
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
		TitleSlug:  "Two Sum",
		LeetcodeID: 1,
		Difficulty: models.ProblemDifficultyEasy,
	}
	db.Create(problem)
	db.Model(problem).Association("Tags").Append(tag)

	req := httptest.NewRequest("GET", "/api/problems"+fmt.Sprintf("/%d", problem.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Result)

	data := response.Data.(map[string]interface{})
	assert.Equal(t, problem.TitleSlug, data["title_slug"])
	assert.Equal(t, problem.Difficulty, models.ProblemDifficulty(data["difficulty"].(string)))

	tags := data["tags"].([]interface{})
	assert.Equal(t, tag.Name, tags[0].(map[string]interface{})["name"])
}
