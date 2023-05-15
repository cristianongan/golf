package services

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"start/config"
	"start/constants"
	"strconv"
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
	PartnerUid     string `json:"partnerUid"`
	CourseUid      string `json:"courseUid"`
	SelfieCheckSum string `json:"selfieCheckSum"`
	Timestamp      string `json:"timestamp"`
	RequestId      string `json:"requestId"`
}

func CallEkyc(urlFull string, bBody []byte, dataModel EkycUpdateBody, imgFile *multipart.File) (error, int, []byte) {
	// req, errNewRequest := http.NewRequest("POST", urlFull, bytes.NewBuffer(bBody))
	// if errNewRequest != nil {
	// 	return errNewRequest, 0, nil
	// }
	if imgFile == nil {
		return errors.New("File is nill"), 0, nil
	}

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("d", dataModel.D)
	_ = writer.WriteField("s", dataModel.S)

	timeUnix := time.Now().UnixNano()
	timeUnixStr := strconv.FormatInt(timeUnix, 10)

	fileName := "ekyc-" + timeUnixStr + ".png"
	log.Println("CallEkyc fileName", fileName)

	/*
		var buff bytes.Buffer
		io.Copy(&buff, *file)
		bFile := buff.Bytes()

		path, bytes.NewReader(bFile),
	*/

	part3, errFile3 := writer.CreateFormFile("selfieImage", filepath.Base(fileName))
	_, errFile3 = io.Copy(part3, *imgFile)
	if errFile3 != nil {
		log.Println("CallEkyc errFile3", errFile3.Error())
		return errFile3, 0, nil
	}

	err := writer.Close()
	if err != nil {
		log.Println("CallEkyc err", err.Error())
		return err, 0, nil
	}

	log.Println("CallEkyc len data", payload.Bytes())

	log.Println("CallEkyc payload", payload)
	log.Println("CallEkyc writer", writer)

	client := &http.Client{
		Timeout: time.Second * constants.TIMEOUT,
	}
	req, err := http.NewRequest("POST", urlFull, payload)

	if err != nil {
		fmt.Println(err)
		return err, 0, nil
	}
	req.Header.Add("Authorization", config.GetEkycAuthKey())

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// client := &http.Client{
	// 	Timeout: time.Second * constants.TIMEOUT,
	// }
	// log.Println("CallEkyc test 01")
	resp, errRequest := client.Do(req)
	if errRequest != nil {
		return errRequest, 0, nil
	}
	defer resp.Body.Close()

	byteBody, errForward := ioutil.ReadAll(resp.Body)

	if byteBody != nil {
		log.Println("CallEkyc response ", string(byteBody))
	}

	if errForward != nil {
		return errForward, 0, nil
	}
	return nil, resp.StatusCode, byteBody
}

func EkycUpdateImage(bBody []byte, dataModel EkycUpdateBody, imgFile *multipart.File) (error, int) {

	url := config.GetEkycUrl() + config.GetEkycUpdate()

	// shortResp := ShortResp{}
	log.Println("EkycUpdateImage url", url)

	err, statusCode, _ := CallEkyc(url, bBody, dataModel, imgFile)
	if err != nil {
		return err, statusCode
	}

	if statusCode != 200 && statusCode != 201 {
		return errors.New("EkycUpdateImage error status code"), statusCode
	}

	// log.Println("EkycUpdateImage res dataByte ", string(dataByte))

	// errUn := json.Unmarshal(dataByte, &shortResp)
	// if errUn != nil {
	// 	return errUn, statusCode, shortResp
	// }

	return nil, statusCode
}
