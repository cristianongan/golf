package request

type MessSocketBody struct {
	Data map[string]interface{} `json:"data" binding:"required"`
	Room string                 `json:"room" binding:"required"`
}
