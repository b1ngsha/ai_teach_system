package api_test

import (
	"ai_teach_system/controllers"
	"ai_teach_system/services"
	"ai_teach_system/tests/mocks"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai_teach_system/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupAITest() (*gin.Engine, *mocks.MockAIService) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockService := mocks.NewMockAIService()
	controller := &controllers.AIController{
		Service: mockService,
	}

	r.POST("/api/ai/generate-code", controller.GenerateCode)

	return r, mockService
}

func TestGenerateCode(t *testing.T) {
	r, mockService := setupAITest()

	tests := []struct {
		name       string
		request    controllers.GenerateCodeRequest
		mockCode   string
		mockError  error
		wantStatus int
		wantError  bool
	}{
		{
			name: "successful code generation",
			request: controllers.GenerateCodeRequest{
				Title:           "Two Sum",
				Language:        "JavaScript",
				Content:         "Given an array of integers nums and an integer target...",
				SampleTestcases: "[2,7,11,15]\n9",
			},
			mockCode: `function twoSum(nums, target) {
    const map = new Map();
    for (let i = 0; i < nums.length; i++) {
        const complement = target - nums[i];
        if (map.has(complement)) {
            return [map.get(complement), i];
        }
        map.set(nums[i], i);
    }
    return [];
}`,
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name: "missing required fields",
			request: controllers.GenerateCodeRequest{
				Title: "Two Sum",
				// Language field missing
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockCode != "" {
				mockService.GenerateCodeFunc = func(req *services.CodeGenerationRequest) (string, error) {
					return tt.mockCode, tt.mockError
				}
			}

			jsonData, err := json.Marshal(tt.request)
			assert.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/ai/generate-code", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response utils.Response
			err = json.Unmarshal(w.Body.Bytes(), &response)

			if tt.wantError {
				assert.False(t, response.Result)
			} else {
				assert.NoError(t, err)
				assert.True(t, response.Result)
				assert.Empty(t, response.Message)

				data := response.Data.(map[string]interface{})
				assert.NotEmpty(t, data["code"])
				assert.Equal(t, tt.mockCode, data["code"])
			}
		})
	}
}
