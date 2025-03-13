package api_test

import (
	"ai_teach_system/controllers"
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
	r.POST("/api/ai/correct-code", controller.CorrectCode)
	r.POST("/api/ai/analyze-code", controller.AnalyzeCode)
	r.POST("/api/ai/chat", controller.Chat)

	return r, mockService
}

func TestGenerateCode(t *testing.T) {
	r, _ := setupAITest()

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

func TestCorrectCode(t *testing.T) {
	r, _ := setupAITest()

	tests := []struct {
		name       string
		request    controllers.CorrectCodeRequest
		mockCode   string
		mockError  error
		wantStatus int
		wantError  bool
	}{
		{
			name: "successful code correction",
			request: controllers.CorrectCodeRequest{
				ProblemID: 1,
				Language:  "Python",
				TypedCode: "def twoSum(nums, target): pass",
			},
			mockCode: `class Solution:
		def twoSum(self, nums: List[int], target: int) -> List[int]:
			hashtable = dict()
			for i, num in enumerate(nums):
				if target - num in hashtable:
					return [hashtable[target - num], i]
				# AI Comment：将hashtable[nums[i]] = i改为hashtable[num] = i以避免重复访问nums[i]
				# hashtable[nums[i]] = i
				hashtable[num] = i  # 修改原因：直接使用num变量，减少对列表的索引操作，提高代码效率和可读性`,
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name: "missing required fields",
			request: controllers.CorrectCodeRequest{
				ProblemID: 1,
				// Language field missing
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.request)
			assert.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/ai/correct-code", bytes.NewBuffer(jsonData))
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

func TestAnalyzeCode(t *testing.T) {
	r, _ := setupAITest()

	tests := []struct {
		name        string
		request     controllers.AnalyzeCodeRequest
		mockMessage string
		mockError   error
		wantStatus  int
		wantError   bool
	}{
		{
			name: "successful code analysis",
			request: controllers.AnalyzeCodeRequest{
				ProblemID: 1,
				Language:  "Python",
				TypedCode: "def twoSum(nums, target): pass",
			},
			mockMessage: `**错误分析**
		这段代码没有实现找到加起来等于目标和的两个数这一逻辑。
		**AI讲师分析**
		这道题包含了理解哈希表和它们在减少时间复杂度上的用法。`,
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name: "missing required fields",
			request: controllers.AnalyzeCodeRequest{
				ProblemID: 1,
				// Language field missing
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.request)
			assert.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/ai/analyze-code", bytes.NewBuffer(jsonData))
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
				assert.NotEmpty(t, data["message"])
				assert.Equal(t, tt.mockMessage, data["message"])
			}
		})
	}
}

func TestChat(t *testing.T) {
	r, _ := setupAITest()

	tests := []struct {
		name        string
		request     controllers.ChatRequest
		mockMessage string
		mockError   error
		wantStatus  int
		wantError   bool
	}{
		{
			name: "successful question answering",
			request: controllers.ChatRequest{
				ProblemID: 1,
				TypedCode: "def twoSum(nums, target): pass",
				Question:  "这道题目的核心思想是什么？",
			},
			mockMessage: `这道题目是经典的"两数之和"问题，需要在数组中找到两个数，使它们的和等于目标值。

关于你的问题，这道题的核心思想是使用哈希表来降低时间复杂度。传统的暴力解法需要O(n²)的时间复杂度，而使用哈希表可以将时间复杂度降低到O(n)。

哈希表的作用是记录已经遍历过的元素及其索引，这样当我们遍历到一个新元素时，可以在O(1)的时间内查找是否存在一个已经遍历过的元素，使得两者之和等于目标值。`,
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name: "missing required fields",
			request: controllers.ChatRequest{
				ProblemID: 1,
				TypedCode: "def twoSum(nums, target): pass",
				// Question field missing
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.request)
			assert.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/ai/chat", bytes.NewBuffer(jsonData))
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
				assert.NotEmpty(t, data["message"])
				assert.Equal(t, tt.mockMessage, data["message"])
			}
		})
	}
}
