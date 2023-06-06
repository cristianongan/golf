package response

import "start/models"

type Result struct {
	Status int64  `json:"status"`
	Infor  string `json:"infor"`
}

type GetServiceRes struct {
	Result     Result             `json:"result"`
	RentalList []RentalRes        `json:"RentalList"`
	CaddieList []models.CaddieRes `json:"CaddieList"`
	Token      string             `json:"Token"`
	CourseCode string             `json:"CourseCode"`
}

type CheckServiceRes struct {
	Result     Result `json:"result"`
	Token      string `json:"Token"`
	RenTalCode string `json:"RenTalCode"`
	CaddieNo   string `json:"CaddieNo"`
	DateStr    string `json:"DateStr"`
	TeeOffStr  string `json:"TeeOffStr"`
	CourseCode string `json:"CourseCode"`
	Qty        int64  `json:"Qty"`
}
type RentalRes struct {
	Code      string  `json:"Code"`
	Name      string  `json:"Name"`
	Unit      string  `json:"Unit"`
	Price     float64 `json:"Price"`
	Inventory string  `json:"Inventory"`
}

type ServiceFeeRes struct {
	Result        Result       `json:"result"`
	DateStr       string       `json:"DateStr"`
	CourseCode    string       `json:"CourseCode"`
	RentalFee     ServiceInfor `json:"rental_fee"`
	PrivateCarFee ServiceInfor `json:"private_car_fee"`
	OddCarFee     ServiceInfor `json:"odd_car_fee"`
	CaddieFee     ServiceInfor `json:"caddie_fee"`
}

type ServiceInfor struct {
	Fee  int64  `json:"fee"`
	Name string `json:"name"`
}
