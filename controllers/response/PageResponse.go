package response

type PageResponse struct {
	Total int64       `json:"total"`
	Data  interface{} `json:"data"`
}

type PageResponseByOffset struct {
	Page  int64       `json:"page"`
	Limit int64       `json:"limit"`
	Total int64       `json:"total"`
	Data  interface{} `json:"data"`
}

type BaseResponse struct {
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	Status    string `json:"status"` //ENABLE, DISABLE, TESTING
}
