package model

// ErrorResponse adalah SATU-SATUNYA bentuk kegagalan yang dikirim ke client.
// Ditulis oleh ErrorHandler terpusat (config/app.go), tidak oleh handler mana pun.
type ErrorResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}
