package types

import "encoding/json"

type Response struct {
	Success int    `json:"success"`
	Message string `json:"message"`
}

func (r *Response) UnsafeMarshal() string {
	body, _ := json.Marshal(r)
	return string(body)
}

func ErrorResponse(message string) string {
	res := Response{
		Success: 0,
		Message: message,
	}
	return res.UnsafeMarshal()
}

func StringResponse(message string) string {
	res := Response{
		Success: 1,
		Message: message,
	}
	return res.UnsafeMarshal()
}
