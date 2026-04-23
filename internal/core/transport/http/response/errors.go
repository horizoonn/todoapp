package core_http_response

type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}
