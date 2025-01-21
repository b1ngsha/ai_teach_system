package utils

// Response 统一响应结构
type Response struct {
	Result  bool        `json:"result"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func Success(data interface{}) *Response {
	return &Response{
		Result: true,
		Data:   data,
	}
}

func Error(message string) *Response {
	return &Response{
		Result:  false,
		Message: message,
	}
}
