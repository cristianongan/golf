package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const BaseURL = "https://api-ssl.bitly.com/v4"

type Shortener interface {
	//Short A url
	Shorten(url string) (*ShortResp, error)
}
type bitly struct {
	Token   string
	Client  *http.Client
	baseURL string
}

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

func (b bitly) Shorten(url string) (*ShortResp, error) {
	buff := new(bytes.Buffer)
	err := json.NewEncoder(buff).Encode(ShortReq{URL: url})
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/shorten", b.baseURL), buff)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", b.Token))
	resp, err := b.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, resp.Status)
	}

	ret := &ShortResp{}
	err = json.NewDecoder(resp.Body).Decode(ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// New creates a new bitly shortner
func New(token string) Shortener {
	c := http.DefaultClient
	return &bitly{
		baseURL: BaseURL,
		Token:   token,
		Client:  c,
	}
}
