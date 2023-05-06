package services

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"start/config"
	"start/constants"
	"time"
)

type EkycUpdateBody struct {
	D           string `json:"d"`
	S           string `json:"s"`
	SelfieImage string `json:"selfieImage"`
}

func CallEkyc(url string, bBody []byte) (error, int, []byte) {
	req, errNewRequest := http.NewRequest("POST", url, bytes.NewBuffer(bBody))
	if errNewRequest != nil {
		return errNewRequest, 0, nil
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Second * constants.TIMEOUT,
	}
	resp, errRequest := client.Do(req)
	if errRequest != nil {
		return errRequest, 0, nil
	}
	defer resp.Body.Close()

	byteBody, errForward := ioutil.ReadAll(resp.Body)

	log.Println("CallEkyc byteBody ", string(byteBody))

	if errForward != nil {
		return errForward, 0, nil
	}
	log.Println("CallEkyc response ", string(byteBody))
	return nil, resp.StatusCode, byteBody
}

func EkycUpdateImage(bBody []byte) (error, int) {

	url := config.GetEkycUrl() + config.GetEkycUpdate()

	// shortResp := ShortResp{}

	err, statusCode, dataByte := CallEkyc(url, bBody)
	if err != nil {
		return err, statusCode
	}

	if statusCode != 200 && statusCode != 201 {
		return errors.New("BitlyShorten error status code"), statusCode
	}

	log.Println("EkycUpdateImage dataByte ", string(dataByte))

	// errUn := json.Unmarshal(dataByte, &shortResp)
	// if errUn != nil {
	// 	return errUn, statusCode, shortResp
	// }

	return nil, statusCode
}
