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
