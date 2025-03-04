package mocks

type MockAIService struct {
}

func NewMockAIService() *MockAIService {
	return &MockAIService{}
}

func (m *MockAIService) GenerateCode(title string, language string, content string, sampleTestCases string) (string, error) {
	return `function twoSum(nums, target) {
    const map = new Map();
    for (let i = 0; i < nums.length; i++) {
        const complement = target - nums[i];
        if (map.has(complement)) {
            return [map.get(complement), i];
        }
        map.set(nums[i], i);
    }
    return [];
}`, nil
}

func (m *MockAIService) CorrectCode(problemID uint, lang string, typedCode string) (string, error) {
	return `class Solution:
		def twoSum(self, nums: List[int], target: int) -> List[int]:
			hashtable = dict()
			for i, num in enumerate(nums):
				if target - num in hashtable:
					return [hashtable[target - num], i]
				# AI Comment：将hashtable[nums[i]] = i改为hashtable[num] = i以避免重复访问nums[i]
				# hashtable[nums[i]] = i
				hashtable[num] = i  # 修改原因：直接使用num变量，减少对列表的索引操作，提高代码效率和可读性`, nil
}

func (m *MockAIService) AnalyzeCode(problemID uint, lang string, typedCode string) (string, error) {
	return `**错误分析**
		这段代码没有实现找到加起来等于目标和的两个数这一逻辑。
		**AI讲师分析**
		这道题包含了理解哈希表和它们在减少时间复杂度上的用法。`, nil
}
