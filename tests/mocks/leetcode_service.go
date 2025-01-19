package mocks

import (
	"ai_teach_system/models"
)

type MockLeetCodeService struct {
	Problems []*models.Problem
}

func NewMockLeetCodeService() *MockLeetCodeService {
	return &MockLeetCodeService{
		Problems: []*models.Problem{
			{
				LeetcodeID: 1,
				Title:      "Two Sum",
				TitleSlug:  "two-sum",
				Content:    "Given an array of integers...",
				Difficulty: "Easy",
				Tags: []models.Tag{
					{Name: "Array"},
					{Name: "Hash Table"},
				},
				SampleTestcases: "[2,7,11,15]\n9",
			},
			{
				LeetcodeID: 2,
				Title:      "Add Two Numbers",
				TitleSlug:  "add-two-numbers",
				Content:    "You are given two non-empty linked lists...",
				Difficulty: "Medium",
				Tags: []models.Tag{
					{Name: "Linked List"},
					{Name: "Math"},
				},
				SampleTestcases: "[2,4,3]\n[5,6,4]",
			},
		},
	}
}

func (m *MockLeetCodeService) FetchAllProblems() ([]*models.Problem, error) {
	return m.Problems, nil
}

func (m *MockLeetCodeService) RunTestCase(titleSlug string, questionId string, code string, testCase string, lang string) (map[string]interface{}, error) {
	resp := map[string]interface{}{
		"interpret_id":          "runcode_11223344",
		"test_case":             "[2,7,11,15]\n9\n[3,2,4]\n6\n[3,3]\n6",
		"interpret_expected_id": "runcode_11223344",
	}
	return resp, nil
}

func (m *MockLeetCodeService) Submit(titleSlug string, lang string, question_id string, code string) (map[string]interface{}, error) {
	resp := map[string]interface{}{
		"submission_id": 594247274,
	}
	return resp, nil
}
