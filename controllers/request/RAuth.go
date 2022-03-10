package request

type LoginBody struct {
	OsCode       string `json:"os_code" binding:"required"`
	Phone        string `json:"phone" binding:"required"`
	Name         string `json:"name"`
	TimestampStr string `json:"timestamp_str" binding:"required"`
	Signature    string `json:"signature" binding:"required"`
	Type         string `json:"type"` //SDK
}
