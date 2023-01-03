package server

import (
	"database/sql/driver"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"start/constants"
	"start/datasources"
	"start/models"
	"strconv"
)

// ====== Customer User =========
type CustomerT struct {
	Name string `json:"name"`
	Dob  string `json:"dob"`
	Note string `json:"note"`
}

type ListCustomerT []CustomerT

func (item *ListCustomerT) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListCustomerT) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func importDataCustomerUser() {
	// Get data from local
	f, errF := os.Open("customer.json")
	if errF != nil {
		log.Println("TestReadData errF", errF.Error())
		return
	}
	defer f.Close()

	byteCustomerValue, errRA := ioutil.ReadAll(f)

	if errRA != nil {
		log.Println("TestReadData errRA", errRA.Error())
		return
	}

	listData := ListCustomerT{}

	errUnM := json.Unmarshal(byteCustomerValue, &listData)
	if errUnM != nil {
		log.Println("TestReadData errUnM", errUnM.Error())
		return
	}
	log.Println("ok")
	// log.Print(listData)
	db := datasources.GetDatabase()

	for _, v := range listData {
		dobInt, _ := strconv.ParseInt(v.Dob, 10, 64)
		log.Println(v.Name)
		customerUser := models.CustomerUser{
			PartnerUid: "CHI-LINH",
			CourseUid:  "CHI-LINH-01",
			Name:       v.Name,
			Note:       v.Note,
			Type:       constants.BOOKING_CUSTOMER_TYPE_MEMBER,
			Dob:        dobInt,
		}

		customerUser.Create(db)
	}

}

// ====== Member Card Type =========

type MemberCardTypeT struct {
	Name               string `json:"name"`
	Subject            string `json:"subject"`
	Type               string `json:"type"`
	GuestStyleOfGuest  string `json:"guest_style_of_guest"`
	NormalDayTakeGuest string `json:"normal_day_take_guest"`
	WeekendTakeGuest   string `json:"weekend_take_guest"`
}

type ListMemberCardTypeT []MemberCardTypeT

func (item *ListMemberCardTypeT) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListMemberCardTypeT) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func importDataMemberCardType() {
	// Get data from local
	f, errF := os.Open("member_card_type.json")
	if errF != nil {
		log.Println("TestReadData errF", errF.Error())
		return
	}
	defer f.Close()

	byteMemberCardTypeValue, errRA := ioutil.ReadAll(f)

	if errRA != nil {
		log.Println("TestReadData errRA", errRA.Error())
		return
	}

	listData := ListMemberCardTypeT{}

	errUnM := json.Unmarshal(byteMemberCardTypeValue, &listData)
	if errUnM != nil {
		log.Println("TestReadData errUnM", errUnM.Error())
		return
	}
	log.Println("ok")
	// log.Print(listData)
	db := datasources.GetDatabase()

	for _, v := range listData {
		log.Println(v.Name)
		memberCardType := models.MemberCardType{
			PartnerUid:         "CHI-LINH",
			CourseUid:          "CHI-LINH-01",
			Name:               v.Name,
			Type:               v.Type,
			Subject:            v.Subject,
			GuestStyleOfGuest:  v.GuestStyleOfGuest,
			NormalDayTakeGuest: v.NormalDayTakeGuest,
			WeekendTakeGuest:   v.WeekendTakeGuest,
		}

		memberCardType.Create(db)
	}

}

// ====== Member Card =========

type MemberCardT struct {
	CardId    string `json:"card_id"`
	ValidDate string `json:"valid_date"`
	ExpDate   string `json:"exp_date"`
	McTypeId  int64  `json:"mc_type_id"`
	OwnerUid  string `json:"owner_uid"`
}

type ListMemberCardT []MemberCardT

func (item *ListMemberCardT) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListMemberCardT) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func importDataMemberCard() {
	// Get data from local
	f, errF := os.Open("member_card.json")
	if errF != nil {
		log.Println("TestReadData errF", errF.Error())
		return
	}
	defer f.Close()

	byteMemberCardValue, errRA := ioutil.ReadAll(f)

	if errRA != nil {
		log.Println("TestReadData errRA", errRA.Error())
		return
	}

	listData := ListMemberCardT{}

	errUnM := json.Unmarshal(byteMemberCardValue, &listData)
	if errUnM != nil {
		log.Println("TestReadData errUnM", errUnM.Error())
		return
	}
	log.Println("ok")
	// log.Print(listData)
	db := datasources.GetDatabase()

	for _, v := range listData {
		log.Println(v.CardId)
		validDateInt, _ := strconv.ParseInt(v.ValidDate, 10, 64)
		expDateInt, _ := strconv.ParseInt(v.ExpDate, 10, 64)
		memberCard := models.MemberCard{
			PartnerUid: "CHI-LINH",
			CourseUid:  "CHI-LINH-01",
			CardId:     v.CardId,
			ValidDate:  validDateInt,
			McTypeId:   v.McTypeId,
			OwnerUid:   v.OwnerUid,
			ExpDate:    expDateInt,
		}

		memberCard.Create(db)
	}

}

// ====== Agencies =========

type AgenciesT struct {
	Type                string `json:"type"`
	ShortName           string `json:"short_name"`
	Name                string `json:"name"`
	Province            string `json:"province"`
	PrimaryContactFirst struct {
		Name        string `json:"name"`
		JobTitle    string `json:"job_title"`
		DirectPhone string `json:"direct_phone"`
		Mail        string `json:"mail"`
	} `json:"primary_contact_first"`
	PrimaryContactSecond struct {
		Name        string `json:"name"`
		JobTitle    string `json:"job_title"`
		DirectPhone string `json:"direct_phone"`
		Mail        string `json:"mail"`
	} `json:"primary_contact_second"`
	ContractDetail struct {
		ContractNo      string `json:"contract_no"`
		Phone           string `json:"phone"`
		Email           string `json:"email"`
		TaxCode         string `json:"tax_code"`
		ContractAddress string `json:"contract_address"`
		ExpDate         string `json:"exp_date"`
		ContractDate    string `json:"contract_date"`
		OfficialAddress string `json:"official_address"`
		Rounds          string `json:"rounds"`
		Note            string `json:"note"`
	} `json:"contract_detail"`
}

type ListAgenciesT []AgenciesT

func (item *ListAgenciesT) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListAgenciesT) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func importDataAgencies() {
	// Get data from local
	f, errF := os.Open("agencies.json")
	if errF != nil {
		log.Println("TestReadData errF", errF.Error())
		return
	}
	defer f.Close()

	byteAgenciesValue, errRA := ioutil.ReadAll(f)

	if errRA != nil {
		log.Println("TestReadData errRA", errRA.Error())
		return
	}

	listData := ListAgenciesT{}

	errUnM := json.Unmarshal(byteAgenciesValue, &listData)
	if errUnM != nil {
		log.Println("TestReadData errUnM", errUnM.Error())
		return
	}
	log.Println("ok")
	// log.Print(listData)
	db := datasources.GetDatabase()

	for _, v := range listData {
		log.Println(v.Name)
		expDateInt, _ := strconv.ParseInt(v.ContractDetail.ExpDate, 10, 64)
		contractDateInt, _ := strconv.ParseInt(v.ContractDetail.ContractDate, 10, 64)
		agency := models.Agency{
			PartnerUid: "CHI-LINH",
			CourseUid:  "CHI-LINH-01",
			Type:       v.Type,
			ShortName:  v.ShortName,
			Name:       v.Name,
			Province:   v.Province,
			PrimaryContactFirst: models.AgencyContact{
				Name:        v.PrimaryContactFirst.Name,
				JobTile:     v.PrimaryContactFirst.JobTitle,
				DirectPhone: v.PrimaryContactFirst.DirectPhone,
				Mail:        v.PrimaryContactFirst.Mail,
			},
			PrimaryContactSecond: models.AgencyContact{
				Name:        v.PrimaryContactSecond.Name,
				JobTile:     v.PrimaryContactSecond.JobTitle,
				DirectPhone: v.PrimaryContactSecond.DirectPhone,
				Mail:        v.PrimaryContactSecond.Mail,
			},
			ContractDetail: models.AgencyContract{
				ContractNo:      v.ContractDetail.ContractNo,
				Phone:           v.ContractDetail.Phone,
				Email:           v.ContractDetail.Email,
				TaxCode:         v.ContractDetail.TaxCode,
				ContractAddress: v.ContractDetail.ContractAddress,
				ExpDate:         expDateInt,
				ContractDate:    contractDateInt,
				OfficialAddress: v.ContractDetail.OfficialAddress,
				Rounds:          0,
				Note:            v.ContractDetail.ContractNo,
			},
		}

		agency.Create(db)
	}

}
