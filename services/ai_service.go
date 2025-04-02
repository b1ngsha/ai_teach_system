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
	GenerateHint(title, content, sampleTestCases string) (string, error)
	CorrectCode(recordID, problemID uint, lang, typedCode string) (string, error)
	AnalyzeCode(recordID, problemID uint, lang, typedCode string) (string, error)
	Chat(problemID uint, typedCode, question string) (string, error)
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

func (s *AIService) GenerateHint(title, content, sampleTestCases string) (string, error) {
	prompt := fmt.Sprintf(`你是一个大学算法课的老师，现在有算法题具体信息如下：

题目：%s
题目内容：%s

示例测试用例：
%v

请生成符合以下要求的作答提示文字：
1. 这段提示需要具有引导作用，不要给出过于详细的作答思路描述，只需要给出大体的思考方向，例如使用某一种算法，引导出问题的解决思路即可
2. 时空复杂度最优
3. 可读性良好
4. 务必使用中文描述`, title, content, sampleTestCases)

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

func (s *AIService) CorrectCode(recordID, problemID uint, language, typedCode string) (string, error) {
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

	err = s.db.Model(&models.UserProblem{}).Where("id = ?", recordID).Update("corrected_code", completion.Choices[0].Message.Content).Error
	if err != nil {
		return "", fmt.Errorf("set corrected_code error: %v", err)
	}

	return completion.Choices[0].Message.Content, nil
}

func (s *AIService) AnalyzeCode(recordID, problemID uint, language, typedCode string) (string, error) {
	var problem models.Problem
	err := s.db.Model(&models.Problem{}).First(&problem, problemID).Error
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`你是一个大学算法课的老师，请分析以下错误代码：

题目：%s
编程语言：%s
题目内容：%s
当前已有代码：%s

示例测试用例：
%v

请生成代码和题目分析，并确保分为两个点进行输出：
第一点为指出代码的错误原因（指定标题为“错误分析”）、
第二点为分析本题目所涉及的计算机领域的知识点（指定标题为“AI讲师分析”），
注意，不要返回正确的代码示例，仅仅进行分析即可。

同时，请一定确保你生成的响应格式如下（在花括号内填入具体的内容）：
**错误分析**：
{错误分析}

**AI讲师分析**：
{AI讲师分析}`, problem.Title, language, problem.Content, typedCode, problem.SampleTestcases)
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

	err = s.db.Model(&models.UserProblem{}).Where("id = ?", recordID).Update("wrong_reason_and_analyze", completion.Choices[0].Message.Content).Error
	if err != nil {
		return "", fmt.Errorf("set wrong_reason_and_analyze error: %v", err)
	}

	return completion.Choices[0].Message.Content, nil
}

func (s *AIService) Chat(problemID uint, typedCode, question string) (string, error) {
	var problem models.Problem
	err := s.db.Model(&models.Problem{}).First(&problem, problemID).Error
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`你是一个大学算法课的AI助教，请基于以下题目信息，回答学生的问题：

题目：%s
题目内容：%s
示例测试用例：
%v

学生问题：%s
学生当前代码：%s

请提供专业、准确、有教育意义的回答，帮助学生理解题目和相关知识点。`, problem.Title, problem.Content, problem.SampleTestcases, typedCode, question)

	completion, err := s.client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是一个大学算法课的AI助教，你的任务是帮助学生理解算法题目，解答他们的疑问，并提供有教育意义的指导。"),
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
