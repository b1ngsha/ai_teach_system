package api_test

import (
	"ai_teach_system/controllers"
	"ai_teach_system/models"
	"ai_teach_system/services"
	"ai_teach_system/tests"
	"ai_teach_system/tests/mocks"
	"encoding/json"
	"fmt"
	"net/http"
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

	controller := controllers.NewLeetCodeController(db, services.NewLeetCodeService())

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

func TestLeetCodeController_RunTestCase(t *testing.T) {
	db, cleanup := tests.SetupTestDB()
	defer cleanup()

	db.AutoMigrate(&models.Problem{})

	// 创建测试题目
	problem := models.Problem{
		LeetcodeID: 1,
		Title:      "Two Sum",
		TitleSlug:  "two-sum",
		Difficulty: "Easy",
	}
	db.Create(&problem)

	controller := controllers.NewLeetCodeController(db, mocks.NewMockLeetCodeService())

	router := gin.Default()
	router.POST("/leetcode/interpret_solution", controller.RunTestCase)

	server := httptest.NewServer(router)
	defer server.Close()

	tests := []struct {
		name       string
		request    controllers.RunTestCaseRequest
		wantStatus int
	}{
		{
			name: "valid test case",
			request: controllers.RunTestCaseRequest{
				QuestionId: "1",
				Lang:       "javascript",
				TypedCode: `var twoSum = function(nums, target) {
					const map = new Map();
					for (let i = 0; i < nums.length; i++) {
						const complement = target - nums[i];
						if (map.has(complement)) {
							return [map.get(complement), i];
						}
						map.set(nums[i], i);
					}
					return [];
				};`,
				DataInput: "[2,7,11,15]\n9",
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.request)
			assert.NoError(t, err)

			resp, err := resty.New().R().
				SetHeader("Content-Type", "application/json").
				SetBody(jsonData).
				Post(fmt.Sprintf("%s/leetcode/interpret_solution", server.URL))

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode())
		})
	}
}

func TestLeetCodeController_Submit(t *testing.T) {
	db, cleanup := tests.SetupTestDB()
	defer cleanup()

	db.AutoMigrate(&models.Problem{})

	// 创建测试题目
	problem := models.Problem{
		LeetcodeID: 1,
		Title:      "Two Sum",
		TitleSlug:  "two-sum",
		Difficulty: "Easy",
	}
	db.Create(&problem)

	controller := controllers.NewLeetCodeController(db, mocks.NewMockLeetCodeService())

	router := gin.Default()
	router.POST("/leetcode/submit", controller.Submit)

	server := httptest.NewServer(router)
	defer server.Close()

	tests := []struct {
		name       string
		request    controllers.SubmitRequest
		wantStatus int
	}{
		{
			name: "valid submission",
			request: controllers.SubmitRequest{
				QuestionId: "1",
				Lang:       "javascript",
				TypedCode: `var twoSum = function(nums, target) {
					const map = new Map();
					for (let i = 0; i < nums.length; i++) {
						const complement = target - nums[i];
						if (map.has(complement)) {
							return [map.get(complement), i];
						}
						map.set(nums[i], i);
					}
					return [];
				};`,
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.request)
			assert.NoError(t, err)

			resp, err := resty.New().R().
				SetHeader("Content-Type", "application/json").
				SetBody(jsonData).
				Post(fmt.Sprintf("%s/leetcode/submit", server.URL))

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode())
		})
	}
}
