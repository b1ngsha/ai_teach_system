package services

import (
	"ai_teach_system/constants"
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type AIService struct {
	client *openai.Client
}

func NewAIService() *AIService {
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("QWEN_API_KEY")),
		option.WithBaseURL(constants.QwenHost),
	)

	return &AIService{
		client: client,
	}
}

type CodeGenerationRequest struct {
	Title           string // 题目
	Language        string // 编程语言
	Content         string // 题目内容
	SampleTestcases string // 示例测试用例
}

func (s *AIService) GenerateCode(req *CodeGenerationRequest) (string, error) {
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

请直接返回代码，不需要其他解释。`, req.Title, req.Language, req.Content, req.SampleTestcases)

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
