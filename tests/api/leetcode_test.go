package api_test

import (
	"ai_teach_system/controllers"
	"ai_teach_system/models"
	"ai_teach_system/tests"
	"ai_teach_system/tests/mocks"
	"ai_teach_system/utils"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupLeetCodeTest() (*gin.Engine, *gorm.DB, func()) {
	gin.SetMode(gin.TestMode)
	db, cleanup := tests.SetupTestDB()

	r := gin.New()
	mockService := mocks.NewMockLeetCodeService()
	controller := controllers.NewLeetCodeController(mockService)

	leetcode := r.Group("/api/leetcode")
	{
		leetcode.POST("/interpret_solution", controller.RunTestCase)
		leetcode.POST("/submit", controller.Submit)
		leetcode.GET("/check/:id", controller.Check)
	}

	return r, db, cleanup
}

func TestRunTestCase(t *testing.T) {
	r, db, cleanup := setupLeetCodeTest()
	defer cleanup()

	// Create a test problem
	problem := models.Problem{
		LeetcodeID: 1,
		Title:      "Two Sum",
		TitleSlug:  "two-sum",
		Difficulty: "Easy",
	}
	db.Create(&problem)

	tests := []struct {
		name       string
		request    controllers.RunTestCaseRequest
		wantStatus int
	}{
		{
			name: "valid test case",
			request: controllers.RunTestCaseRequest{
				LeetcodeQuestionId: 1,
				Lang:               "javascript",
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

			req := httptest.NewRequest("POST", "/api/leetcode/interpret_solution", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response utils.Response
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.True(t, response.Result)

			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, "runcode_11223344", response.Data.(map[string]interface{})["interpret_id"])
			}
		})
	}
}

func TestSubmit(t *testing.T) {
	r, db, cleanup := setupLeetCodeTest()
	defer cleanup()

	// Create a test problem
	problem := models.Problem{
		LeetcodeID: 1,
		Title:      "Two Sum",
		TitleSlug:  "two-sum",
		Difficulty: "Easy",
	}
	db.Create(&problem)

	tests := []struct {
		name       string
		request    controllers.SubmitRequest
		wantStatus int
	}{
		{
			name: "valid submission",
			request: controllers.SubmitRequest{
				LeetcodeQuestionId: 1,
				Lang:               "javascript",
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

			req := httptest.NewRequest("POST", "/api/leetcode/submit", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response utils.Response
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.True(t, response.Result)

			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, float64(594247274), response.Data.(map[string]interface{})["submission_id"])
			}
		})
	}
}

func TestCheck(t *testing.T) {
	r, _, cleanup := setupLeetCodeTest()
	defer cleanup()

	tests := []struct {
		name       string
		runCodeID  string
		wantStatus int
		wantResult bool
	}{
		{
			name:       "valid check",
			runCodeID:  "runcode_1737896487.348871_X8Uj2Q8p6H",
			wantStatus: http.StatusOK,
			wantResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/leetcode/check/"+tt.runCodeID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response utils.Response
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantResult, response.Result)

			if tt.wantStatus == http.StatusOK {
				data := response.Data.(map[string]interface{})
				assert.Equal(t, "SUCCESS", data["state"])
			}
		})
	}
}
