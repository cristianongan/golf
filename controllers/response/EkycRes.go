package response

type EkycBaseResponse struct {
	Code string      `json:"code"`
	Desc string      `json:"desc"`
	Data interface{} `json:"data"`
}
