package cron

import (
	"io/ioutil"
	"net/http"
)

// =========================== cron util ===========================
func requestToCron(url string) (error, int, []byte) {
	req, errNewRequest := http.NewRequest("POST", url, nil)
	if errNewRequest != nil {
		return errNewRequest, 0, nil
	}

	client := &http.Client{}
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
