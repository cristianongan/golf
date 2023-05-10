package services

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"start/config"
	"start/constants"
	"time"
)

type EkycUpdateBody struct {
	D           string `json:"d"`
	S           string `json:"s"`
	SelfieImage string `json:"selfieImage"`
}

type EkycDataModel struct {
	Sid            string `json:"sid"`
	IdNumber       string `json:"idNumber"`
	SelfieCheckSum string `json:"selfieCheckSum"`
	Timestamp      string `json:"timestamp"`
	RequestId      string `json:"requestId"`
}

func CallEkyc(urlFull string, bBody []byte, dataModel EkycUpdateBody, imgFile multipart.File) (error, int, []byte) {
	req, errNewRequest := http.NewRequest("POST", urlFull, bytes.NewBuffer(bBody))
	if errNewRequest != nil {
		return errNewRequest, 0, nil
	}

	form := url.Values{}
	form.Add("d", dataModel.D)
	form.Add("s", dataModel.S)

	req.PostForm = form

	req.Header.Add("Authorization", config.GetEkycAuthKey())
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Second * constants.TIMEOUT,
	}
	log.Println("CallEkyc test 01")
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

func EkycUpdateImage(bBody []byte, dataModel EkycUpdateBody, imgFile multipart.File) (error, int) {

	url := config.GetEkycUrl() + config.GetEkycUpdate()

	// shortResp := ShortResp{}
	log.Println("EkycUpdateImage url", url)

	err, statusCode, dataByte := CallEkyc(url, bBody, dataModel, imgFile)
	if err != nil {
		return err, statusCode
	}

	if statusCode != 200 && statusCode != 201 {
		return errors.New("EkycUpdateImage error status code"), statusCode
	}

	log.Println("EkycUpdateImage dataByte ", string(dataByte))

	// errUn := json.Unmarshal(dataByte, &shortResp)
	// if errUn != nil {
	// 	return errUn, statusCode, shortResp
	// }

	return nil, statusCode
}
