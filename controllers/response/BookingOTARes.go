package response

import "encoding/json"

type CancelBookOTARes struct {
	Result      ResultOTA `json:"result"`
	BookingCode string    `json:"BookingCode"`
}

func (r *BookingOTARes) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type BookingOTARes struct {
	Result       ResultOTA `json:"result"`
	Token        string    `json:"Token"`
	EmailConfirm string    `json:"EmailConfirm"`
	CourseCode   string    `json:"CourseCode"`
	TeeOffStr    string    `json:"TeeOffStr"`
	DateStr      string    `json:"DateStr"`
	Part         int64     `json:"Part"`
	Tee          int64     `json:"Tee"`
	IsMainCourse bool      `json:"isMainCourse"`
	NumBook      int64     `json:"NumBook"`
	Holes        int64     `json:"Holes"`
	PlayerName   string    `json:"PlayerName"`
	Contact      string    `json:"Contact"`
	Note         string    `json:"Note"`
	BookingCode  string    `json:"BookingCode"`
	BookOtaID    string    `json:"BookOtaID"`
	GreenFee     int64     `json:"GreenFee"`
	CaddieFee    int64     `json:"CaddieFee"`
	BuggyFee     int64     `json:"BuggyFee"`
	CardID       string    `json:"CardID"`
	AgentCode    string    `json:"AgentCode"`
	BookAgent    string    `json:"BookAgent"`
	GuestStyle   string    `json:"GuestStyle"`
	Rental       string    `json:"Rental"`
	Caddies      string    `json:"Caddies"`
}

type ResultOTA struct {
	Status int64  `json:"status"`
	Infor  string `json:"infor"`
}

type ResultLockTeeTimeOTA struct {
	Status  int64  `json:"status"`
	Infor   string `json:"infor"`
	NumBook int    `json:"NumBook,omitempty"`
}

func UnmarshalWelcome(data []byte) (TeeTimeOTA, error) {
	var r TeeTimeOTA
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TeeTimeOTA) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type TeeTypeOTARes struct {
	TeeType   string `json:"TeeType"`
	Name      string `json:"Name"`
	ImageLink string `json:"ImageLink"`
	Note      string `json:"Note"`
}
type GetTeeTimeOTAResponse struct {
	Result        ResultOTA       `json:"result"`
	Data          []TeeTimeOTA    `json:"data"`
	IsMainCourse  bool            `json:"isMainCourse"`
	Token         interface{}     `json:"Token"`
	TeeTypeInfo   []TeeTypeOTARes `json:"TeeTypeInfo"`
	CourseCode    string          `json:"CourseCode"`
	OTACode       string          `json:"OTA_Code"`
	GuestCode     string          `json:"Guest_Code"`
	Date          string          `json:"Date"`
	GolfPriceRate string          `json:"GolfPriceRate"`
	NumTeeTime    int64           `json:"NumTeeTime"`
}

type TeeTimeOTA struct {
	TeeOffStr    string `json:"TeeOffStr"`
	DateStr      string `json:"DateStr"`
	TeeOff       string `json:"TeeOff"`
	Part         int64  `json:"Part"`
	TimeIndex    int64  `json:"TimeIndex"`
	Tee          int64  `json:"Tee"`
	NumBook      int64  `json:"NumBook"`
	IsMainCourse bool   `json:"isMainCourse"`
	TeeType      string `json:"TeeType"`
	// Play1        interface{} `json:"Play1"`
	// Play2        interface{} `json:"Play2"`
	// Play3        interface{} `json:"Play3"`
	// Play4        interface{} `json:"Play4"`
	// Contact1     interface{} `json:"Contact1"`
	// Contact2     interface{} `json:"Contact2"`
	// Contact3     interface{} `json:"Contact3"`
	// Contact4     interface{} `json:"Contact4"`
	// Note1        interface{} `json:"Note1"`
	// Note2        interface{} `json:"Note2"`
	// Note3        interface{} `json:"Note3"`
	// Note4        interface{} `json:"Note4"`
	// Id1          int64       `json:"ID1"`
	// Id2          int64       `json:"ID2"`
	// Id3          int64       `json:"ID3"`
	// Id4          int64       `json:"ID4"`
	// IsWaiting    bool        `json:"IsWaiting"`
	// Islock       bool        `json:"Islock"`
	// LockReson    interface{} `json:"LockReson"`
	GreenFee  int64 `json:"GreenFee"`
	CaddieFee int64 `json:"CaddieFee"`
	BuggyFee  int64 `json:"BuggyFee"`
	Holes     int64 `json:"Holes"`
}
type TeeTimePartOTA struct {
	IsHideTeePart bool
	StartPart     string
	EndPart       string
}
type TeeTimeStatus struct {
	Result       ResultOTA   `json:"result"`
	Token        interface{} `json:"Token"`
	IsMainCourse bool        `json:"isMainCourse"`
	Edit         bool        `json:"Edit"`
	CreateUser   interface{} `json:"CreateUser"`
	CourseCode   string      `json:"CourseCode"`
	Locktime     int64       `json:"Locktime"`
	DateStr      string      `json:"DateStr"`
	TeeTimeOTA
}

type LockTeeTimeRes struct {
	Result       ResultLockTeeTimeOTA `json:"result"`
	Token        interface{}          `json:"Token"`
	IsMainCourse bool                 `json:"isMainCourse"`
	Edit         bool                 `json:"Edit"`
	CreateUser   interface{}          `json:"CreateUser"`
	CourseCode   string               `json:"CourseCode"`
	Locktime     int64                `json:"Locktime"`
	DateStr      string               `json:"DateStr"`
	TeeTimeOTA
}
