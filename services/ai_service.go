package services

import (
	"ai_teach_system/constants"
	"ai_teach_system/models"
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"gorm.io/gorm"
)

type AIServiceInterface interface {
	GenerateCode(title string, language string, content string, sampleTestCases string) (string, error)
	CorrectCode(problemID uint, lang string, typedCode string) (string, error)
	AnalyzeCode(problemID uint, lang string, typedCode string) (string, error)
}

type AIService struct {
	client *openai.Client
	db     *gorm.DB
}

func NewAIService(db *gorm.DB) *AIService {
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("QWEN_API_KEY")),
		option.WithBaseURL(constants.QwenHost),
	)

	return &AIService{
		client: client,
		db:     db,
	}
}

func (s *AIService) GenerateCode(title string, language string, content string, sampleTestCases string) (string, error) {
	prompt := fmt.Sprintf(`作为一个专业的算法工程师，请根据以下要求生成代码解答问题：

题目：%s
编程语言：%s
题目内容：%s

示例测试用例：
%v

请生成符合要求的代码，并确保：
1. 代码正确性
2. 代码时空复杂度最优
3. 代码可读性

请直接返回代码，不需要其他解释。`, title, language, content, sampleTestCases)

	completion, err := s.client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是一个专业的算法工程师，精通各种编程语言和算法。请生成最优时空复杂度的代码来解决问题。"),
			openai.UserMessage(prompt),
		}),
		Model: openai.F("qwen-plus"),
	})

	if err != nil {
		return "", fmt.Errorf("AI service error: %v", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no response from AI service")
	}

	return completion.Choices[0].Message.Content, nil
}

func (s *AIService) CorrectCode(problemID uint, language string, typedCode string) (string, error) {
	var problem models.Problem
	err := s.db.Model(&models.Problem{}).First(&problem, problemID).Error
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`作为一个专业的算法工程师，请修改以下已有代码解答问题：

题目：%s
编程语言：%s
题目内容：%s
当前已有代码：%s

示例测试用例：
%v

请生成符合要求的代码，并确保：
1. 代码正确性
2. 代码时空复杂度最优
3. 代码可读性
4. 尽量在已有代码的基础上进行修改，非必要情况下请勿大规模修改代码逻辑
5. 请不要删除被修改的代码片段，而是将修改后的代码片段添加到被修改的代码片段下方，并添加注释说明修改原因，该注释需要以“AI Comment：”作为前缀，注释中不要出现prompt相关的内容

只需要返回修改后代码，不需要其他解释，不需要测试用例的示例，也不需要用markdown格式来返回代码，直接返回即可。`, problem.Title, language, problem.Content, typedCode, problem.SampleTestcases)
	completion, err := s.client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是一个专业的算法工程师，精通各种编程语言和算法。请生成最优时空复杂度的代码来解决问题。"),
			openai.UserMessage(prompt),
		}),
		Model: openai.F("qwen-plus"),
	})

	if err != nil {
		return "", fmt.Errorf("AI service error: %v", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no response from AI service")
	}

	return completion.Choices[0].Message.Content, nil
}

func (s *AIService) AnalyzeCode(problemID uint, language string, typedCode string) (string, error) {
	var problem models.Problem
	err := s.db.Model(&models.Problem{}).First(&problem, problemID).Error
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`作为一个专业的算法工程师，请分析以下错误代码：

题目：%s
编程语言：%s
题目内容：%s
当前已有代码：%s

示例测试用例：
%v

请生成代码和题目分析，并确保分为两个点进行输出：第一点为指出代码的错误原因（指定标题为“错误分析”）、第二点为分析本题目所涉及的计算机领域的知识点（指定标题为“AI讲师分析”），注意，不要返回正确的代码示例，仅仅进行分析即可`, problem.Title, language, problem.Content, typedCode, problem.SampleTestcases)
	completion, err := s.client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是一个大学的算法课老师，请对同学们的错误代码片段和对应的题目进行分析。"),
			openai.UserMessage(prompt),
		}),
		Model: openai.F("qwen-plus"),
	})

	if err != nil {
		return "", fmt.Errorf("AI service error: %v", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no response from AI service")
	}

	return completion.Choices[0].Message.Content, nil
}
