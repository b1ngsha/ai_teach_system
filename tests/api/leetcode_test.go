package api_test

import (
	"ai_teach_system/controllers"
	"ai_teach_system/models"
	"ai_teach_system/tests"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestLeetCodeController_GetProblem(t *testing.T) {
	db, cleanup := tests.SetupTestDB()
	defer cleanup()

	db.AutoMigrate(&models.Problem{})

	problem := models.Problem{
		LeetcodeID: 1,
		Title:      "Two Sum",
		Difficulty: "Easy",
	}
	db.Create(&problem)

	controller := controllers.NewLeetCodeController(db)

	router := gin.Default()
	router.GET("/problems/:id", controller.GetProblem)

	server := httptest.NewServer(router)
	defer server.Close()

	response := map[string]interface{}{}
	httpClient := resty.New()
	httpClient.R().SetResult(&response).Get(fmt.Sprintf("%s/problems/1", server.URL))

	leetcodeID := int(response["leetcode_id"].(float64))

	assert.Equal(t, 1, leetcodeID)
	assert.Equal(t, "Two Sum", response["title"])
	assert.Equal(t, "Easy", response["difficulty"])
}
