package callservices

import (
	"encoding/json"
	"log"
	"net/http"
	"start/config"
	"start/controllers/request"
	"start/controllers/response"
)

func PushMessInSocket(body request.MessSocketBody) {
	data, err := json.Marshal(body)
	urlResult := config.GetGolfSocketURL() + "socket/send"

	if err != nil {
		log.Println("Mess err: ", err)
	}
	serverHeader := make(http.Header)
	serverHeader.Add("Content-Type", "application/json")
	resp, _ := writeOut(urlResult, data, serverHeader, "POST")

	mess := response.MessSocketRes{}
	err1 := json.Unmarshal(resp, &mess)
	if err1 != nil {
		log.Println("Mess err: ", err1)
	}
}
