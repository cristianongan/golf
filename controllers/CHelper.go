package controllers

import (
	"database/sql/driver"
	"encoding/json"
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CHelper struct{}

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
}

/// Member card
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
}
