package services

import (
	"io/ioutil"
	"net/http"
	"time"
)

func GetProxy(url string) (error, int, []byte) {
	req, errNewRequest := http.NewRequest("GET", url, nil)
	if errNewRequest != nil {
		return errNewRequest, 0, nil
	}
	// req.Header.Set("Secret_key", config.GetFenceSecretKey())
	client := &http.Client{
		Timeout: time.Second * 10}
	resp, errRequest := client.Do(req)
	if errRequest != nil {
		return errRequest, 0, nil
	}
	defer resp.Body.Close()

	byteBody, errForward := ioutil.ReadAll(resp.Body)
	if errForward != nil {
		return errForward, 0, nil
	}
	// str := fmt.Sprintf("%s", byteBody)
	return nil, resp.StatusCode, byteBody
}
