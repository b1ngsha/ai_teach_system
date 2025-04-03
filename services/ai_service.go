package services

import (
	"ai_teach_system/constants"
	"ai_teach_system/models"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"gorm.io/gorm"
)

type AIServiceInterface interface {
	GenerateHint(title, content, sampleTestCases, modelType string) (string, error)
	CorrectCode(recordID, problemID uint, lang, typedCode string) (map[string]interface{}, error)
	AnalyzeCode(recordID, problemID uint, lang, typedCode string) (map[string]interface{}, error)
	Chat(problemID uint, typedCode, question, modelType string) (string, error)
	SuggestKnowledgePointTags(knowledgePointID uint) ([]models.Tag, error)
	JudgeCode(problemID uint, lang, code string) (map[string]interface{}, error)
}

// JudgeResult 定义判题结果的结构
type JudgeResult struct {
	Status      string  `json:"status"`      // Accepted, Wrong Answer, Time Limit Exceeded, etc.
	TimeUsed    float64 `json:"time_used"`   // 运行时间(ms)
	MemoryUsed  float64 `json:"memory_used"` // 内存使用(MB)
	TestResults []struct {
		Input          string `json:"input"`
		ExpectedOutput string `json:"expected_output"`
		ActualOutput   string `json:"actual_output"`
		Status         string `json:"status"`
		Message        string `json:"message"`
	} `json:"test_results"`
}

type AIService struct {
	clientDeepseek *openai.Client
	clientQwen     *openai.Client
	db             *gorm.DB
}

func NewAIService(db *gorm.DB) *AIService {
	clientDeepseek := openai.NewClient(
		option.WithAPIKey(os.Getenv("DEEPSEEK_API_KEY")),
		option.WithBaseURL(constants.DeepseekHost),
	)

	clientQwen := openai.NewClient(
		option.WithAPIKey(os.Getenv("QWEN_API_KEY")),
		option.WithBaseURL(constants.QwenHost),
	)

	return &AIService{
		clientDeepseek: clientDeepseek,
		clientQwen:     clientQwen,
		db:             db,
	}
}

func (s *AIService) GenerateHint(title, content, sampleTestCases, modelType string) (string, error) {
	var client *openai.Client
	if modelType == "qwen" {
		client = s.clientQwen
	} else {
		client = s.clientDeepseek
	}

	var model string
	if modelType == "qwen" {
		model = "qwen2.5-14b-instruct-1m"
	} else {
		model = "deepseek-chat"
	}

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

	completion, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是一个大学算法课的老师，你需要通过语言引导学生做出正确的答案。"),
			openai.UserMessage(prompt),
		}),
		Model: openai.F(model),
	})

	if err != nil {
		return "", fmt.Errorf("AI service error: %v", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no response from AI service")
	}

	return completion.Choices[0].Message.Content, nil
}

func (s *AIService) CorrectCode(recordID, problemID uint, language, typedCode string) (map[string]interface{}, error) {
	var problem models.Problem
	err := s.db.Model(&models.Problem{}).First(&problem, problemID).Error
	if err != nil {
		return nil, err
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
5. 请不要删除被修改的代码片段，而是将修改后的代码片段添加到被修改的代码片段下方，并添加注释说明修改原因，该注释需要以"AI Comment："作为前缀，注释中不要出现prompt相关的内容

只需要返回修改后代码，不需要其他解释，不需要测试用例的示例，也不需要用markdown格式来返回代码，直接返回即可。`, problem.Title, language, problem.Content, typedCode, problem.SampleTestcases)
	completionQwen, err := s.clientQwen.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是一个专业的算法工程师，精通各种编程语言和算法。请生成最优时空复杂度的代码来解决问题。"),
			openai.UserMessage(prompt),
		}),
		Model: openai.F("qwen2.5-14b-instruct-1m"),
	})

	if err != nil {
		return nil, fmt.Errorf("qwen AI service error: %v", err)
	}

	if len(completionQwen.Choices) == 0 {
		return nil, fmt.Errorf("no response from Qwen AI service")
	}

	completionDeepseek, err := s.clientDeepseek.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是一个专业的算法工程师，精通各种编程语言和算法。请生成最优时空复杂度的代码来解决问题。"),
			openai.UserMessage(prompt),
		}),
		Model: openai.F("deepseek-chat"),
	})

	if err != nil {
		return nil, fmt.Errorf("deepseek AI service error: %v", err)
	}

	if len(completionQwen.Choices) == 0 {
		return nil, fmt.Errorf("no response from Deepseek AI service")
	}

	err = s.db.Model(&models.UserProblem{}).Where("id = ?", recordID).Updates(map[string]interface{}{
		"deepseek_corrected_code": completionDeepseek.Choices[0].Message.Content,
		"qwen_corrected_code":     completionQwen.Choices[0].Message.Content,
	}).Error
	if err != nil {
		return nil, fmt.Errorf("set corrected_code error: %v", err)
	}

	return map[string]interface{}{
		"deepseek_corrected_code": completionDeepseek.Choices[0].Message.Content,
		"qwen_corrected_code":     completionQwen.Choices[0].Message.Content,
	}, nil
}

func (s *AIService) AnalyzeCode(recordID, problemID uint, language, typedCode string) (map[string]interface{}, error) {
	var problem models.Problem
	err := s.db.Model(&models.Problem{}).First(&problem, problemID).Error
	if err != nil {
		return nil, err
	}

	prompt := fmt.Sprintf(`你是一个大学算法课的老师，请分析以下错误代码：

题目：%s
编程语言：%s
题目内容：%s
当前已有代码：%s

示例测试用例：
%v

请生成代码和题目分析，并确保分为两个点进行输出：
第一点为指出代码的错误原因（指定标题为"错误分析"）、
第二点为分析本题目所涉及的计算机领域的知识点（指定标题为"AI讲师分析"），
注意，不要返回正确的代码示例，仅仅进行分析即可。

同时，请一定确保你生成的响应格式如下（在花括号内填入具体的内容）：
**错误分析**：
{错误分析}

**AI讲师分析**：
{AI讲师分析}`, problem.Title, language, problem.Content, typedCode, problem.SampleTestcases)
	completionQwen, err := s.clientQwen.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是一个大学的算法课老师，请对同学们的错误代码片段和对应的题目进行分析。"),
			openai.UserMessage(prompt),
		}),
		Model: openai.F("qwen2.5-14b-instruct-1m"),
	})

	if err != nil {
		return nil, fmt.Errorf("qwen AI service error: %v", err)
	}

	if len(completionQwen.Choices) == 0 {
		return nil, fmt.Errorf("no response from Qwen AI service")
	}

	completionDeepseek, err := s.clientDeepseek.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是一个大学的算法课老师，请对同学们的错误代码片段和对应的题目进行分析。"),
			openai.UserMessage(prompt),
		}),
		Model: openai.F("deepseek-chat"),
	})

	if err != nil {
		return nil, fmt.Errorf("deepseek AI service error: %v", err)
	}

	if len(completionQwen.Choices) == 0 {
		return nil, fmt.Errorf("no response from Deepseek AI service")
	}

	err = s.db.Model(&models.UserProblem{}).Where("id = ?", recordID).Updates(map[string]interface{}{
		"qwen_wrong_reason_and_analyze":     completionQwen.Choices[0].Message.Content,
		"deepseek_wrong_reason_and_analyze": completionDeepseek.Choices[0].Message.Content,
	}).Error
	if err != nil {
		return nil, fmt.Errorf("set wrong_reason_and_analyze error: %v", err)
	}

	return map[string]interface{}{
		"qwen_wrong_reason_and_analyze":     completionQwen.Choices[0].Message.Content,
		"deepseek_wrong_reason_and_analyze": completionDeepseek.Choices[0].Message.Content,
	}, nil
}

func (s *AIService) Chat(problemID uint, typedCode, question, modelType string) (string, error) {
	var problem models.Problem
	err := s.db.Model(&models.Problem{}).First(&problem, problemID).Error
	if err != nil {
		return "", err
	}

	var client *openai.Client
	if modelType == "qwen" {
		client = s.clientQwen
	} else {
		client = s.clientDeepseek
	}

	var model string
	if modelType == "qwen" {
		model = "qwen2.5-14b-instruct-1m"
	} else {
		model = "deepseek-chat"
	}

	prompt := fmt.Sprintf(`你是一个大学算法课的AI助教，请基于以下题目信息，回答学生的问题：

题目：%s
题目内容：%s
示例测试用例：
%v

学生问题：%s
学生当前代码：%s

请提供专业、准确、有教育意义的回答，帮助学生理解题目和相关知识点。`, problem.Title, problem.Content, problem.SampleTestcases, typedCode, question)

	completion, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是一个大学算法课的AI助教，你的任务是帮助学生理解算法题目，解答他们的疑问，并提供有教育意义的指导。"),
			openai.UserMessage(prompt),
		}),
		Model: openai.F(model),
	})

	if err != nil {
		return "", fmt.Errorf("AI service error: %v", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no response from AI service")
	}

	return completion.Choices[0].Message.Content, nil
}

func (s *AIService) SuggestKnowledgePointTags(knowledgePointID uint) ([]models.Tag, error) {
	var knowledgePoint models.KnowledgePoint
	if err := s.db.First(&knowledgePoint, knowledgePointID).Error; err != nil {
		return nil, fmt.Errorf("未找到知识点: %v", err)
	}

	// 获取所有已有的标签
	var existingTags []models.Tag
	if err := s.db.Model(&models.Tag{}).Find(&existingTags).Error; err != nil {
		return nil, fmt.Errorf("获取已有标签失败: %v", err)
	}

	if len(existingTags) == 0 {
		return nil, fmt.Errorf("当前暂无标签")
	}

	// 构建标签上下文
	var tagsContext string
	for i, tag := range existingTags {
		if i > 0 {
			tagsContext += "\n"
		}
		tagsContext += fmt.Sprintf("%d. %s（%s）", i+1, tag.Name, tag.NameCn)
	}

	prompt := fmt.Sprintf(`作为一个专业的计算机教育领域AI助手，请从以下已有标签中为知识点内容选择最相关的标签：

知识点：%s

已有标签列表：
%s

请从上述已有标签中选择3-5个最相关的标签。请仅返回标签的序号（每行一个数字），例如：
1
3
5

注意：
1. 只能从已有标签中选择
2. 选择最能反映知识点核心内容的标签
3. 确保选择的标签数量在3-5个之间`, knowledgePoint.Name, tagsContext)

	completion, err := s.clientQwen.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是一个专业的计算机教育领域AI助手，精通计算机课程知识点的分类和标签关联。"),
			openai.UserMessage(prompt),
		}),
		Model: openai.F("qwen2.5-14b-instruct-1m"),
	})

	if err != nil {
		return nil, fmt.Errorf("AI 服务错误: %v", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("AI 服务未返回结果")
	}

	// 解析返回的标签序号
	var selectedTags []models.Tag
	lines := strings.Split(strings.TrimSpace(completion.Choices[0].Message.Content), "\n")

	for _, line := range lines {
		index := 0
		_, err := fmt.Sscanf(strings.TrimSpace(line), "%d", &index)
		if err != nil {
			continue
		}

		// 调整为0基索引
		index--

		if index >= 0 && index < len(existingTags) {
			selectedTags = append(selectedTags, existingTags[index])
		}
	}

	if len(selectedTags) == 0 {
		return nil, fmt.Errorf("AI 未能选择合适的标签")
	}

	return selectedTags, nil
}

func (s *AIService) JudgeCode(problemID uint, lang, code string) (map[string]interface{}, error) {
	var problem models.Problem
	if err := s.db.First(&problem, problemID).Error; err != nil {
		return nil, fmt.Errorf("题目不存在: %v", err)
	}

	// 构建判题提示
	prompt := fmt.Sprintf(`作为一个专业的编程题目评测系统，请对以下代码进行评测：

题目信息：
%s

提交的代码：
%s

测试用例：
%s

判题要求：
1. 时间限制：%dms
2. 内存限制：%dMB
3. 编程语言：%s

请确保严格按照以下JSON格式返回判题结果，不要出现任何其他信息，不需要解释说明，只返回以下格式的JSON即可：
{
    "status": "判题状态(SUCCESS/FAILED/Time Limit Exceeded/Memory Limit Exceeded/Runtime Error)",
    "time_used": 实际运行时间(ms),
    "memory_used": 实际使用内存(MB),
    "test_results": [
        {
            "input": "测试用例输入",
            "expected_output": "期望输出",
            "actual_output": "实际输出",
            "status": "用例状态",
            "message": "错误信息（如果有）"
        }
    ]
}

注意：
1. 必须验证代码的正确性
2. 必须检查时间和内存限制
3. 必须测试所有测试用例
4. 必须按照规定格式返回结果
5. 对于每个测试用例，都要实际运行代码并比较结果`,
		problem.Content,
		code,
		problem.TestCases,
		problem.TimeLimit,
		problem.MemoryLimit,
		lang,
	)

	completion, err := s.clientQwen.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是一个专业的编程题目评测系统，你需要严格按照题目要求对代码进行评测，并返回规范的评测结果。"),
			openai.UserMessage(prompt),
		}),
		Model: openai.F("qwen2.5-14b-instruct-1m"),
	})

	if err != nil {
		return nil, fmt.Errorf("AI 服务错误: %v", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("AI 服务未返回结果")
	}

	// 替换掉markdown格式
	content := completion.Choices[0].Message.Content
	content = strings.ReplaceAll(content, "```json", "")
	content = strings.ReplaceAll(content, "```", "")
	content = strings.TrimSpace(content)

	// 解析 AI 返回的判题结果
	var result JudgeResult
	err = json.Unmarshal([]byte(content), &result)
	if err != nil {
		return nil, fmt.Errorf("解析判题结果失败: %v", err)
	}

	return map[string]interface{}{
		"status":       result.Status,
		"time_used":    result.TimeUsed,
		"memory_used":  result.MemoryUsed,
		"test_results": result.TestResults,
	}, nil
}
