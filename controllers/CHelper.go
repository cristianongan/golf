package controllers

import (
	"database/sql/driver"
	"encoding/json"
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	model_service "start/models/service"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CHelper struct{}

func (_ *CHelper) AppLog(c *gin.Context, prof models.CmsUser) {
	var body map[string]interface{}

	if bindErr := c.BindJSON(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	okResponse(c, "ok")
}

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

func (_ *CHelper) CreateAddCustomer(c *gin.Context, prof models.CmsUser) {
	body := ListCustomerT{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	db := datasources.GetDatabase()

	// note create_at > 1672735789

	for _, v := range body {
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

	okResponse(c, "ok")
}

// / Member card
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

func (_ *CHelper) CreateMemberCard(c *gin.Context, prof models.CmsUser) {
	body := ListMemberCardT{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	db := datasources.GetDatabase()

	// note create_at > 1672735789

	for _, v := range body {
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

	okResponse(c, "ok")
}

// Proshop
type ProshopT struct {
	ProshopId   string  `json:"proshop_id"`
	AccountCode string  `json:"account_code"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Unit        string  `json:"unit"`
	Type        string  `json:"type"`
	GroupCode   string  `json:"group_code"`
	VieName     string  `json:"vie_name"`
}

type ListProshopT []ProshopT

func (item *ListProshopT) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListProshopT) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func (_ *CHelper) CreateProshop(c *gin.Context, prof models.CmsUser) {
	body := ListProshopT{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	db := datasources.GetDatabase()

	// note create_at > 1672735789

	for _, v := range body {
		proshop := model_service.Proshop{
			PartnerUid:  "CHI-LINH",
			CourseUid:   "CHI-LINH-01",
			ProShopId:   v.ProshopId,
			AccountCode: v.AccountCode,
			Name:        v.Name,
			VieName:     v.VieName,
			Price:       v.Price,
			Unit:        v.Unit,
			Type:        v.Type,
			GroupCode:   v.GroupCode,
		}

		proshop.Create(db)
	}

	okResponse(c, "ok")
}

// FB
type FBT struct {
	FBCode      string  `json:"fb_code"`
	AccountCode string  `json:"account_code"`
	Name        string  `json:"name"`
	EnglishName string  `json:"english_name"`
	VieName     string  `json:"vie_name"`
	Price       float64 `json:"price"`
	Unit        string  `json:"unit"`
	Type        string  `json:"type"`
	GroupCode   string  `json:"group_code"`
}

type ListFBT []FBT

func (item *ListFBT) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListFBT) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func (_ *CHelper) CreateFB(c *gin.Context, prof models.CmsUser) {
	body := ListFBT{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	db := datasources.GetDatabase()

	// note create_at > 1672735789

	for _, v := range body {
		fb := model_service.FoodBeverage{
			PartnerUid:  "CHI-LINH",
			CourseUid:   "CHI-LINH-01",
			FBCode:      v.FBCode,
			AccountCode: v.AccountCode,
			Name:        v.Name,
			EnglishName: v.EnglishName,
			VieName:     v.VieName,
			Price:       v.Price,
			Unit:        v.Unit,
			Type:        v.Type,
			GroupCode:   v.GroupCode,
		}

		fb.Create(db)
	}

	okResponse(c, "ok")
}
