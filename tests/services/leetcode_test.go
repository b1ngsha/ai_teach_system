package services_test

import (
	"ai_teach_system/services"
	"ai_teach_system/tests"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeetCodeService_FetchAllProblems(t *testing.T) {
	server := tests.SetupTestLeetCodeServer()
	defer server.Close()

	service := services.NewLeetCodeService()
	service.Client.SetBaseURL(server.URL)

	problems, err := service.FetchAllProblems()
	assert.NoError(t, err)
	assert.Len(t, problems, 1)
	assert.Equal(t, "Two Sum", problems[0].Title)
	assert.Len(t, problems[0].Tags, 2)
}

func TestLeetCodeService_FetchProblemDetail(t *testing.T) {
	server := tests.SetupTestLeetCodeServer()
	defer server.Close()

	service := services.NewLeetCodeService()
	service.Client.SetBaseURL(server.URL)

	problem, err := service.FetchProblemDetail("two-sum")
	assert.NoError(t, err)
	assert.Equal(t, 1, problem.LeetcodeID)
	assert.Equal(t, "Two Sum", problem.Title)
}
