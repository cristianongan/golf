package services

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	urlLib "net/url"
	"start/config"
	"strconv"
	"strings"
	"time"
)

type SMSData struct {
	Destination  string `xml:"Destination"`
	Sender       string `xml:"Sender"`
	KeywordName  string `xml:"KeywordName"`
	OutContent   string `xml:"OutContent"`
	ChargingFlag string `xml:"ChargingFlag"`
	MOSeqNo      string `xml:"MOSeqNo"`
	ContentType  string `xml:"ContentType"`
	LocalTime    string `xml:"LocalTime"`
	UserName     string `xml:"UserName"`
	Password     string `xml:"Password"`
}

type Envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Text    string   `xml:",chardata"`
	S       string   `xml:"S,attr"`
	Body    struct {
		Text           string `xml:",chardata"`
		SendMTResponse struct {
			Text   string `xml:",chardata"`
			Ns0    string `xml:"ns0,attr"`
			Return string `xml:"return"`
		} `xml:"sendMTResponse"`
	} `xml:"Body"`
}

type VNPaySMSBody struct {
	MessageId      string `json:"messageId"`   // Phone
	Destination    string `json:"destination"` // Phone
	Sender         string `json:"sender"`      // BrandName
	Keyword        string `json:"keyword"`
	ShortMessage   string `json:"shortMessage"`
	IsEncrypt      int    `json:"isEncrypt"` //0
	Type           int    `json:"type"`
	RequestTime    int64  `json:"requestTime"`
	PartnerCode    string `json:"partnerCode"` // UserName
	SercretKey     string `json:"sercretKey"`  // Password
	EncryptMessage string `json:"encryptMessage"`
}

type VNPaySmsResponse struct {
	MessageId     string `json:"messageId"`
	Status        string `json:"status"`
	Description   string `json:"description"`
	IsMnp         int    `json:"isMnp"`
	ProviderId    string `json:"providerId"`
	ProviderIdOrg string `json:"providerIdOrg"`
}

func (item Envelope) HandleCodeResult() error {
	switch item.Body.SendMTResponse.Return {
	case "00|Success":
		return nil
	default:
		return errors.New(item.Body.SendMTResponse.Return)
	}
}

/*
 Send vnpay sms v2
*/
func VNPaySendSmsV2(phone, message string) (string, error) {
	phoneSend := strings.ReplaceAll(phone, "+84", "0")
	url := config.GetVNPayUrl()
	messageId := phone + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	body := VNPaySMSBody{
		MessageId:      messageId,
		Destination:    phoneSend,
		Sender:         config.GetVNPaySender(),
		Keyword:        config.GetVNPayKeyword(),
		ShortMessage:   message,
		IsEncrypt:      0,
		Type:           0,
		RequestTime:    time.Now().Unix(),
		PartnerCode:    config.GetVNPayUserName(),
		SercretKey:     config.GetVNPayPassword(),
		EncryptMessage: "",
	}

	bodyBytes, errB := json.Marshal(body)
	if errB != nil {
		return "Marshal object error", errB
	}

	bodyStr := string(bodyBytes)

	log.Println("body string", bodyStr)

	httpMethod := "POST"
	req, err := http.NewRequest(httpMethod, url, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Println("VNPaySendSmsV2 Error on creating request object. ", err.Error())
		return bodyStr, err
	}
	req.Header.Add("Content-Type", "application/json")

	proxyUrl, _ := urlLib.Parse(config.GetUrlCaroProxy())
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	// client := &http.Client{
	// 	Timeout: time.Second * constants.TIMEOUT,
	// }
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on dispatching request. ", err.Error())
		return bodyStr, err
	}
	defer resp.Body.Close()

	byteResp, errForward := ioutil.ReadAll(resp.Body)
	if errForward != nil {
		return "error parse vnpay sms v2", errForward
	}
	resModel := VNPaySmsResponse{}
	_ = json.Unmarshal(byteResp, &resModel)

	log.Println("send sms vnpay v2 response", string(byteResp))

	if resModel.Status == "00" {
		//Success
		return bodyStr, nil
	}

	return bodyStr, errors.New(string(byteResp))
}

func mapStatusToDes(status string) string {
	statusDesc := ""
	switch status {
	case "00":
		statusDesc = "Thành công"
		break
	case "01":
		statusDesc = "Sai số điện thoại"
		break
	case "02":
		statusDesc = "Độ dài không hợp lệ"
		break
	case "04":
		statusDesc = "Sai thông tin xác thực"
		break
	case "05":
		statusDesc = "Mất kết nối đến nhà cung cấp"
		break
	case "06":
		statusDesc = "IP không được phép truy cập"
		break
	case "08":
		statusDesc = "Timeout"
		break
	case "11":
		statusDesc = "Sai loại tin nhắn"
		break
	case "12":
		statusDesc = "Không hỗ trợ tin Unicode"
		break
	case "80":
		statusDesc = "Không tìm thấy đối tác"
		break
	default:
		statusDesc = "Lỗi ngoại lệ"
		break
	}

	return statusDesc
}
