package services

import (
	"ai_teach_system/constants"
	"ai_teach_system/models"
	"fmt"
	"log"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type LeetCodeServiceInterface interface {
	FetchAllProblems() ([]*models.Problem, error)
	RunTestCase(titleSlug string, questionId string, code string, testCase string, lang string) (map[string]interface{}, error)
	Submit(titleSlug string, lang string, question_id string, code string) (map[string]interface{}, error)
}

type LeetCodeService struct {
	Client *resty.Client
}

type GraphQLQuery struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

func NewLeetCodeService() *LeetCodeService {
	client := resty.New().
		SetBaseURL(constants.LeetCodeHost).
		SetHeader("Content-Type", "application/json")

	return &LeetCodeService{
		Client: client,
	}
}

func (s *LeetCodeService) FetchAllProblems() ([]*models.Problem, error) {
	problems := make([]*models.Problem, 0)
	pageSize := 100
	skip := 0
	hasMore := true

	for hasMore {
		query := `
		query problemsetQuestionList($categorySlug: String, $limit: Int, $skip: Int, $filters: QuestionListFilterInput) {
			problemsetQuestionList(
				categorySlug: $categorySlug
				limit: $limit
				skip: $skip
				filters: $filters
			) {
				hasMore
				total
				questions {
					acRate
					difficulty
					freqBar
					frontendQuestionId
					isFavor
					paidOnly
					solutionNum
					status
					title
					titleCn
					titleSlug
					topicTags {
						name
						nameTranslated
						id
						slug
					}
				}
			}
		}`

		graphqlQuery := GraphQLQuery{
			Query: query,
			Variables: map[string]interface{}{
				"limit":        pageSize,
				"skip":         skip,
				"filters":      map[string]interface{}{},
				"categorySlug": "all-code-essentials",
			},
			OperationName: "problemsetQuestionList",
		}

		var result map[string]interface{}
		_, err := s.Client.R().
			SetBody(graphqlQuery).
			SetResult(&result).
			Post("/graphql")

		if err != nil {
			return nil, err
		}

		data := result["data"].(map[string]interface{})
		problemList := data["problemsetQuestionList"].(map[string]interface{})
		questions := problemList["questions"].([]interface{})
		hasMore = problemList["hasMore"].(bool)

		for _, q := range questions {
			question := q.(map[string]interface{})
			titleSlug := question["titleSlug"].(string)

			problem, err := s.FetchProblemDetail(titleSlug)
			if err != nil {
				log.Printf("获取题目详情失败 %s: %v", titleSlug, err)
				continue
			}

			// 处理标签
			if tags, ok := question["topicTags"].([]interface{}); ok {
				for _, t := range tags {
					tag := t.(map[string]interface{})
					problem.Tags = append(problem.Tags, models.Tag{
						Name: tag["name"].(string),
					})
				}
			}

			problems = append(problems, problem)
		}

		log.Printf("已获取 %d 题，当前页 %d 条记录，是否还有更多：%v", len(problems), len(questions), hasMore)

		skip += pageSize
	}

	return problems, nil
}

func (s *LeetCodeService) FetchProblemDetail(titleSlug string) (*models.Problem, error) {
	query := `
	query questionData($titleSlug: String!) {
		question(titleSlug: $titleSlug) {
			questionId
			title
			titleSlug
			content
			difficulty
			sampleTestCase
		}
	}`

	graphqlQuery := GraphQLQuery{
		Query: query,
		Variables: map[string]interface{}{
			"titleSlug": titleSlug,
		},
		OperationName: "questionData",
	}

	var result map[string]interface{}
	_, err := s.Client.R().
		SetBody(graphqlQuery).
		SetResult(&result).
		Post("/graphql")

	if err != nil {
		return nil, err
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data is not a map[string]interface{}")
	}

	question := data["question"].(map[string]interface{})

	leetcodeID, err := strconv.Atoi(question["questionId"].(string))
	if err != nil {
		return nil, err
	}

	// 适配vip题目无法抓取题目内容的情况
	content, ok := question["content"].(string)
	if !ok {
		content = ""
	}

	problem := &models.Problem{
		LeetcodeID:      leetcodeID,
		Title:           question["title"].(string),
		TitleSlug:       titleSlug,
		Difficulty:      models.ProblemDifficulty(question["difficulty"].(string)),
		Content:         content,
		SampleTestcases: question["sampleTestCase"].(string),
	}

	return problem, nil
}

func (s *LeetCodeService) RunTestCase(titleSlug string, questionId string, code string, testCase string, lang string) (map[string]interface{}, error) {
	body := &map[string]string{
		"data_input":  testCase,
		"lang":        lang,
		"question_id": questionId,
		"typed_code":  code,
	}

	var result map[string]interface{}
	path := fmt.Sprintf("/problems/%s/interpret_solution", titleSlug)
	_, err := s.Client.R().
		SetBody(body).
		SetResult(&result).
		Post(path)

	if err != nil {
		return nil, err
	}

	data := result["data"].(map[string]interface{})
	return data, nil
}

func (s *LeetCodeService) Submit(titleSlug string, lang string, question_id string, code string) (map[string]interface{}, error) {
	body := &map[string]string{
		"lang":        lang,
		"question_id": question_id,
		"typed_code":  code,
	}
	var result map[string]interface{}
	path := fmt.Sprintf("/problems/%s/submit/", titleSlug)
	_, err := s.Client.R().
		SetBody(body).
		SetResult(&result).
		Post(path)

	if err != nil {
		return nil, err
	}

	data := result["data"].(map[string]interface{})
	return data, nil
}
