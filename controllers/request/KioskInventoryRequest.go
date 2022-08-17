package request

type KioskInventoryInputItemBody struct {
	Code     string `json:"code"`
	Quantity int64  `json:"quantity"`
}

type KioskInventoryOutputItemBody struct {
	Code     string `json:"code"`
	Quantity int64  `json:"quantity"`
}

type KioskInventoryCreateItemBody struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
