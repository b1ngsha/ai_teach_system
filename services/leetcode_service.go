package services

import (
	"ai_teach_system/config"
	"ai_teach_system/constants"
	"ai_teach_system/models"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"
)

type LeetCodeServiceInterface interface {
	FetchAllProblems() ([]*models.Problem, error)
	RunTestCase(userID uint, questionId int, code string, lang string) (map[string]interface{}, error)
	Submit(userID uint, lang string, knowledge_point_id uint, question_id int, code string) (map[string]interface{}, error)
	Check(userID uint, runCodeID string, test bool) (map[string]interface{}, error)
}

type LeetCodeService struct {
	Client *resty.Client
	db     *gorm.DB
}

type GraphQLQuery struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

func NewLeetCodeService(db *gorm.DB) *LeetCodeService {
	client := resty.New().
		SetBaseURL(constants.LeetCodeHost).
		SetHeader("Content-Type", "application/json")

	return &LeetCodeService{
		Client: client,
		db:     db,
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

		data, ok := result["data"].(map[string]interface{})
		if !ok {
			log.Printf("%v", result["data"])
		}
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
						Name:   tag["name"].(string),
						NameCn: tag["nameTranslated"].(string),
					})
				}
			}

			problems = append(problems, problem)
		}

		log.Printf("已获取 %d 题，当前页 %d 条记录，是否还有更多：%v", len(problems), len(questions), hasMore)

		time.Sleep(3 * time.Second)

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
			translatedTitle
			titleSlug
			content
			translatedContent
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
	contentCn, ok := question["translatedContent"].(string)
	if !ok {
		contentCn = ""
	}

	translatedTitle, ok := question["translatedTitle"].(string)
	if !ok {
		translatedTitle = ""
	}

	problem := &models.Problem{
		LeetcodeID:      leetcodeID,
		Title:           question["title"].(string),
		TitleCn:         translatedTitle,
		TitleSlug:       titleSlug,
		Difficulty:      models.ProblemDifficulty(question["difficulty"].(string)),
		Content:         content,
		ContentCn:       contentCn,
		SampleTestcases: question["sampleTestCase"].(string),
	}

	return problem, nil
}

func (s *LeetCodeService) RunTestCase(userID uint, leetcodeQuestionId int, code string, lang string) (map[string]interface{}, error) {
	var problem models.Problem
	s.db.Model(&models.Problem{}).Where("leetcode_id = ?", leetcodeQuestionId).First(&problem)
	body := &map[string]interface{}{
		"data_input":  problem.SampleTestcases,
		"lang":        lang,
		"question_id": leetcodeQuestionId,
		"typed_code":  code,
	}

	var result map[string]interface{}
	path := fmt.Sprintf("/problems/%s/interpret_solution", problem.TitleSlug)
	_, err := s.Client.R().
		SetHeader("Cookie", fmt.Sprintf("LEETCODE_SESSION=%s", config.Leetcode.LeetcodeSession)).
		SetBody(body).
		SetResult(&result).
		Post(path)

	if err != nil {
		return nil, err
	}

	var tryRecord models.UserProblem
	s.db.Where(models.UserProblem{UserID: userID, ProblemID: problem.ID}).Attrs(models.UserProblem{Status: models.ProblemStatusTried}).FirstOrCreate(&tryRecord)

	return result, nil
}

func (s *LeetCodeService) Submit(userID uint, lang string, knowledge_point_id uint, leetcodeQuestionId int, code string) (map[string]interface{}, error) {
	var problem models.Problem
	s.db.Model(&models.Problem{}).Where("leetcode_id = ?", leetcodeQuestionId).First(&problem)
	body := &map[string]interface{}{
		"lang":        lang,
		"question_id": strconv.Itoa(leetcodeQuestionId),
		"typed_code":  code,
	}
	var result map[string]interface{}
	path := fmt.Sprintf("/problems/%s/submit/", problem.TitleSlug)
	_, err := s.Client.R().
		SetHeader("Cookie", fmt.Sprintf("LEETCODE_SESSION=%s", config.Leetcode.LeetcodeSession)).
		SetBody(body).
		SetResult(&result).
		Post(path)

	if err != nil {
		return nil, err
	}

	// 新增作答记录
	submissionID := result["submission_id"].(float64)
	tryRecord := models.UserProblem{
		UserID:           userID,
		KnowledgePointID: knowledge_point_id,
		ProblemID:        problem.ID,
		Status:           models.ProblemStatusTried,
		TypedCode:        code,
		SubmissionID:     submissionID,
	}
	s.db.Create(&tryRecord)

	return result, nil
}

func (s *LeetCodeService) Check(userID uint, runCodeID string, test bool) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/submissions/detail/%s/check", runCodeID)
	_, err := s.Client.R().
		SetResult(&result).
		Get(path)

	if err != nil {
		return nil, err
	}

	// 当检查提交结果时才更新提交记录状态
	if !test {
		// 修改提交记录状态
		state := result["state"].(string)
		var status models.ProblemStatus
		switch state {
		case "SUCCESS":
			status = models.ProblemStatusSolved
		case "FAILED":
			status = models.ProblemStatusFailed
		default:
			return nil, fmt.Errorf("未知状态: %s", state)
		}

		err = s.db.Model(&models.UserProblem{}).
			Where("user_id = ? AND submission_id = ?", userID, runCodeID).
			Update("status", status).
			Error
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
