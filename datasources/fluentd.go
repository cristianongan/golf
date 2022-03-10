package datasources

import (
	"bytes"
	"encoding/json"
	"net/http"
	"start/config"
)

func FluentdSendRequestLog(data json.RawMessage) error {
	req, errNewRequest := http.NewRequest("POST", config.GetFluentdUrl(), bytes.NewBuffer(data))
	if errNewRequest != nil {
		return errNewRequest
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(config.GetFluentdUser(), config.GetFluentdPass())
	client := &http.Client{}
	resp, errRequest := client.Do(req)
	if errRequest != nil {
		return errRequest
	}
	defer resp.Body.Close()

	return nil
}
