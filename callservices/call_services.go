package callservices

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Call Post API
func writeOut(url string, data []byte, header http.Header, method string) ([]byte, bool) {
	client := http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	req.Header = header
	resp, err := client.Do(req)

	if err == nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			resp.Body.Close()
		}
		return body, resp.StatusCode == 200
	}

	log.Println("writeOut error:", err)

	return []byte{}, false
}
