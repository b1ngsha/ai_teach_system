package mocks

import (
	"ai_teach_system/services"
)

type MockAIService struct {
	GenerateCodeFunc func(req *services.CodeGenerationRequest) (string, error)
}

func NewMockAIService() *MockAIService {
	return &MockAIService{
		GenerateCodeFunc: func(req *services.CodeGenerationRequest) (string, error) {
			return `function twoSum(nums, target) {
    const map = new Map();
    for (let i = 0; i < nums.length; i++) {
        const complement = target - nums[i];
        if (map.has(complement)) {
            return [map.get(complement), i];
        }
        map.set(nums[i], i);
    }
    return [];
}`, nil
		},
	}
}

func (m *MockAIService) GenerateCode(req *services.CodeGenerationRequest) (string, error) {
	return m.GenerateCodeFunc(req)
}
