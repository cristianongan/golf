package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"start/config"
	"start/constants"
	"time"
)

type ShortLinkGenBody struct {
	FullLink  string `json:"full_link"` // full_link
	Signature string `json:"signature"` // signature
}

type ShortLinkGenRes struct {
	Short string `json:"short"` // short
}

func GenShortLink(fullLink string) (ShortLinkGenRes, error) {

	secret := config.GetShortLinkSecretKey()
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(fullLink))
	checksum := hex.EncodeToString(h.Sum(nil))

	body := ShortLinkGenBody{
		FullLink:  fullLink,
		Signature: checksum,
	}

	bodyBytes, errB := json.Marshal(body)
	respModel := ShortLinkGenRes{}
	if errB != nil {
		return respModel, errB
	}

	url := config.GetShortLinkUrlBe()

	bodyStr := string(bodyBytes)

	log.Println("body string", bodyStr)

	httpMethod := "POST"
	req, err := http.NewRequest(httpMethod, url, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Println("GenShortLink Error on creating request object. ", err.Error())
		return respModel, err
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Second * constants.TIMEOUT,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on dispatching request. ", err.Error())
		return respModel, err
	}
	defer resp.Body.Close()

	byteResp, errForward := ioutil.ReadAll(resp.Body)
	if errForward != nil {
		return respModel, errForward
	}
	log.Println("GenShortLink send sms v2 response", string(byteResp))

	errUnM := json.Unmarshal(byteResp, &respModel)
	if errUnM != nil {
		return respModel, errUnM
	}

	return respModel, nil

}
