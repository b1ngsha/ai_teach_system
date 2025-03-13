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

func (m *MockAIService) Chat(problemID uint, typedCode string, question string) (string, error) {
	return `这道题目是经典的"两数之和"问题，需要在数组中找到两个数，使它们的和等于目标值。

关于你的问题，这道题的核心思想是使用哈希表来降低时间复杂度。传统的暴力解法需要O(n²)的时间复杂度，而使用哈希表可以将时间复杂度降低到O(n)。

哈希表的作用是记录已经遍历过的元素及其索引，这样当我们遍历到一个新元素时，可以在O(1)的时间内查找是否存在一个已经遍历过的元素，使得两者之和等于目标值。`, nil
}
