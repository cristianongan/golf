package response

import (
	"start/models"
	model_service "start/models/service"
)

type Result struct {
	Status int64  `json:"status"`
	Infor  string `json:"infor"`
}

type GetServiceRes struct {
	Result     Result                 `json:"result"`
	RentalList []model_service.Rental `json:"RentalList"`
	CaddieList []models.Caddie        `json:"CaddieList"`
	Token      string                 `json:"Token"`
	CourseCode string                 `json:"CourseCode"`
}

type CheckServiceRes struct {
	Result     Result `json:"result"`
	Token      string `json:"Token"`
	RenTalCode string `json:"RenTalCode"`
	CaddieNo   int64  `json:"CaddieNo"`
	DateStr    string `json:"DateStr"`
	TeeOffStr  string `json:"TeeOffStr"`
	CourseCode string `json:"CourseCode"`
	Qty        int64  `json:"Qty"`
}

type CaddieRes struct {
	Number   string `json:"Number"`
	FullName string `json:"FullName"`
	Phone    string `json:"Phone"`
}

type RentalRes struct {
	Code      string `json:"Code"`
	Name      string `json:"Name"`
	Unit      string `json:"Unit"`
	Price     string `json:"Price"`
	Inventory string `json:"Inventory"`
}
