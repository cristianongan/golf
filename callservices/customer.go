package callservices

import (
	"encoding/json"
	"log"
	"net/http"
	"start/config"
	"start/controllers/request"
	"start/controllers/response"
)

func CreateCustomer(body request.CustomerBody) (bool, response.CustomerRes) {
	data, err := json.Marshal(body)
	urlResult := config.GetGolfPartnerURL() + "fast/customer"

	if err != nil {
		log.Println("Customer err: ", err)
		return false, response.CustomerRes{}
	}
	serverHeader := make(http.Header)
	serverHeader.Add("Content-Type", "application/json")
	resp, ok := writeOut(urlResult, data, serverHeader, "POST")

	customers := response.CustomerRes{}
	err1 := json.Unmarshal(resp, &customers)
	if err1 != nil {
		return false, customers
	}

	return ok, customers
}
