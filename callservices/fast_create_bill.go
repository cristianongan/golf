package callservices

import (
	"encoding/json"
	"log"
	"net/http"
	"start/config"
	"start/controllers/response"
)

func CreateFastBankCredit(body response.FastBillBody) (bool, response.FastBillRes) {
	data, err := json.Marshal(body)
	urlResult := config.GetGolfPartnerURL() + "fast/bank-credit"

	if err != nil {
		log.Println("Customer err: ", err)
		return false, response.FastBillRes{}
	}
	serverHeader := make(http.Header)
	serverHeader.Add("Content-Type", "application/json")
	resp, ok := writeOut(urlResult, data, serverHeader, "POST")

	customers := response.FastBillRes{}
	err1 := json.Unmarshal(resp, &customers)
	if err1 != nil {
		return false, customers
	}

	return ok, customers
}

func CreateFastCashVoucher(body response.FastBillBody) (bool, response.FastBillRes) {
	data, err := json.Marshal(body)
	urlResult := config.GetGolfPartnerURL() + "fast/cash-voucher"

	if err != nil {
		log.Println("Customer err: ", err)
		return false, response.FastBillRes{}
	}
	serverHeader := make(http.Header)
	serverHeader.Add("Content-Type", "application/json")
	resp, ok := writeOut(urlResult, data, serverHeader, "POST")

	customers := response.FastBillRes{}
	err1 := json.Unmarshal(resp, &customers)
	if err1 != nil {
		return false, customers
	}

	return ok, customers
}
