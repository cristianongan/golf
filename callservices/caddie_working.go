package callservices

import (
	"encoding/json"
	"log"
	"net/http"
	"start/config"
	"start/controllers/request"
	"start/controllers/response"
)

func GetDetailCaddieWorking(body request.GetDetalCaddieWorkingSyncBody) (bool, response.CaddieWorkingSyncRes) {
	data, err := json.Marshal(body)
	urlResult := config.GetGolfPartnerURL() + "caddie-working/detail"
	if err != nil {
		log.Println("CaddieWorkingSync err: ", err)
		return false, response.CaddieWorkingSyncRes{}
	}
	serverHeader := make(http.Header)
	serverHeader.Add("Content-Type", "application/json")
	resp, ok := writeOut(urlResult, data, serverHeader, "POST")

	caddieWorking := response.CaddieWorkingSyncRes{}
	err1 := json.Unmarshal(resp, &caddieWorking)
	if err1 != nil || len(caddieWorking.Data) <= 0 {
		return false, caddieWorking
	}

	return ok, caddieWorking
}

func ImportCaddieWorking(body request.CreateCaddieWorkingReq) (map[string]interface{}, string) {
	data, err := json.Marshal(body)
	urlResult := config.GetGolfPartnerURL() + "caddie-working/import"
	if err != nil {
		log.Println("CaddieWorkingSync err: ", err)
		return map[string]interface{}{}, ""
	}
	serverHeader := make(http.Header)
	serverHeader.Add("Content-Type", "application/json")
	resp, _ := writeOut(urlResult, data, serverHeader, "POST")

	caddieWorking := map[string]interface{}{}
	err1 := json.Unmarshal(resp, &caddieWorking)
	if err1 != nil {
		return caddieWorking, string(resp)
	}

	return caddieWorking, string(resp)
}
