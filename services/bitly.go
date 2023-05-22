package services

type ShortReq struct {
	URL       string `json:"long_url"`
	Domain    string `json:"domain,omitempty"`
	GroupGUID string `json:"group_guid,omitempty"`
}
type ShortResp struct {
	References map[string]string `json:"references"`
	Archived   bool              `json:"archived"`
	Tags       []string          `json:"tags"`
	CreatedAt  string            `json:"created_at"`
	Title      string            `json:"title"`
	Deeplinks  []struct {
		Bitlink     string `json:"bitlink"`
		InstallURL  string `json:"install_url"`
		Created     string `json:"created"`
		AppURIPath  string `json:"app_uri_path"`
		Modified    string `json:"modified"`
		InstallType string `json:"install_type"`
		AppGUID     string `json:"app_guid"`
		GUID        string `json:"guid"`
		Os          string `json:"os"`
		BrandGUID   string `json:"brand_guid"`
	} `json:"deeplinks"`
	CreatedBy      string   `json:"created_by"`
	LongURL        string   `json:"long_url"`
	ClientID       string   `json:"client_id"`
	CustomBitlinks []string `json:"custom_bitlinks"`
	URL            string   `json:"link"`
	ID             string   `json:"id"`
}

// func CallBitly(url string, bBody []byte) (error, int, []byte) {
// 	req, errNewRequest := http.NewRequest("POST", url, bytes.NewBuffer(bBody))
// 	if errNewRequest != nil {
// 		return errNewRequest, 0, nil
// 	}
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.GetBitlyToken()))
// 	client := &http.Client{
// 		Timeout: time.Second * constants.TIMEOUT,
// 	}
// 	resp, errRequest := client.Do(req)
// 	if errRequest != nil {
// 		return errRequest, 0, nil
// 	}
// 	defer resp.Body.Close()

// 	byteBody, errForward := ioutil.ReadAll(resp.Body)
// 	if errForward != nil {
// 		return errForward, 0, nil
// 	}
// 	log.Println("CallBitly response ", string(byteBody))
// 	return nil, resp.StatusCode, byteBody
// }

// func BitlyShorten(bBody []byte) (error, int, ShortResp) {

// 	url := config.GetBitlyUrl() + "shorten"

// 	shortResp := ShortResp{}

// 	err, statusCode, dataByte := CallBitly(url, bBody)
// 	if err != nil {
// 		return err, statusCode, shortResp
// 	}

// 	if statusCode != 200 && statusCode != 201 {
// 		return errors.New("BitlyShorten error status code"), statusCode, shortResp
// 	}

// 	errUn := json.Unmarshal(dataByte, &shortResp)
// 	if errUn != nil {
// 		return errUn, statusCode, shortResp
// 	}

// 	return nil, statusCode, shortResp
// }
