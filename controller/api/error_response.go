package api

// ErrorResponse presents an error response of REST API
type ErrorResponse struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details"`
}
