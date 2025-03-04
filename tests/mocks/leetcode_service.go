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

func (m *MockLeetCodeService) RunTestCase(userID uint, questionId string, code string, lang string) (map[string]interface{}, error) {
	resp := map[string]interface{}{
		"interpret_id":          "runcode_11223344",
		"test_case":             "[2,7,11,15]\n9\n[3,2,4]\n6\n[3,3]\n6",
		"interpret_expected_id": "runcode_11223344",
	}
	return resp, nil
}

func (m *MockLeetCodeService) Submit(userID uint, lang string, question_id string, code string) (map[string]interface{}, error) {
	resp := map[string]interface{}{
		"submission_id": 594247274,
	}
	return resp, nil
}

func (c *MockLeetCodeService) Check(userID uint, runCodeID string) (map[string]interface{}, error) {
	resp := map[string]interface{}{
		"code_output":               []string{},
		"compare_result":            "1",
		"correct_answer":            true,
		"display_runtime":           "0",
		"elapsed_time":              124,
		"expected_code_answer":      []string{"[0,1]", ""},
		"expected_code_output":      []string{},
		"expected_display_runtime":  "0",
		"expected_elapsed_time":     13,
		"expected_lang":             "cpp",
		"expected_memory":           8432000,
		"expected_run_success":      true,
		"expected_status_code":      10,
		"expected_status_runtime":   "0",
		"expected_std_output_list":  []string{"", ""},
		"expected_task_finish_time": 1737895297914,
		"expected_task_name":        "judger.interprettask.Interpret",
		"fast_submit":               false,
		"lang":                      "python3",
		"memory":                    17916000,
		"memory_percentile":         nil,
		"pretty_lang":               "Python3",
		"run_success":               true,
		"runtime_percentile":        nil,
		"state":                     "SUCCESS",
		"status_code":               10,
		"status_memory":             "17.5 MB",
		"status_msg":                "Accepted",
		"status_runtime":            "0 ms",
		"std_output_list":           []string{"", ""},
		"submission_id":             "runcode_1737896487.348871_X8Uj2Q8p6H",
		"task_finish_time":          1737896487587,
		"task_name":                 "judger.runcodetask.RunCode",
		"total_correct":             1,
		"total_testcases":           1,
	}
	return resp, nil
}
