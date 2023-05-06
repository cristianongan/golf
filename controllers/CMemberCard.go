package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CMemberCard struct{}

// ======================== eKyc ===========================

/*
List member cho app thu thap
*/
func (_ *CMemberCard) EKycUpdateImageMemberCard(c *gin.Context, prof models.CmsUser) {
	// db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

}

/*
List member cho app thu thap
*/
func (_ *CMemberCard) EKycGetListMemberCard(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListMemberCardEKycAppThuThapForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	memberCardR := models.MemberCard{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	memberCardR.Status = constants.STATUS_ENABLE
	list, err := memberCardR.FindListForEkycAppThuThap(db, page, form.Search)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okResponse(c, list)
}

func (_ *CMemberCard) CreateMemberCard(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := models.MemberCard{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if !body.IsValidated() {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	// Check Member Card Type Exit
	mcType := models.MemberCardType{}
	mcType.Id = body.McTypeId
	errFind := mcType.FindFirst(db)
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	// Check Owner Invalid
	owner := models.CustomerUser{}

	if body.OwnerUid == "" {
		owner.Uid = body.OwnerUid
		errFind = owner.FindFirst(db)
		if errFind != nil {
			response_message.BadRequest(c, errFind.Error())
			return
		}
	}

	// Check duplicated
	if body.IsDuplicated(db) {
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	memberCard := models.MemberCard{
		CardId:   body.CardId,
		McTypeId: body.McTypeId,
	}

	memberCard.PartnerUid = body.PartnerUid
	memberCard.CourseUid = body.CourseUid

	memberCard.OwnerUid = body.OwnerUid
	memberCard.ValidDate = body.ValidDate
	memberCard.ExpDate = body.ExpDate
	memberCard.Note = body.Note
	memberCard.ReasonUnactive = body.ReasonUnactive
	memberCard.ChipCode = body.ChipCode
	memberCard.StartPrecial = body.StartPrecial
	memberCard.EndPrecial = body.EndPrecial

	memberCard.PriceCode = body.PriceCode
	memberCard.GreenFee = body.GreenFee
	memberCard.CaddieFee = body.CaddieFee
	memberCard.BuggyFee = body.BuggyFee
	memberCard.AdjustPlayCount = body.AdjustPlayCount
	memberCard.AnnualType = body.AnnualType
	memberCard.Float = mcType.Float

	if mcType.Subject == constants.MEMBER_CARD_BASE_SUBJECT_COMPANY {
		// Check Company Exit
		company := models.Company{}
		company.Id = body.CompanyId
		errFind := company.FindFirst(db)
		if errFind != nil {
			response_message.BadRequest(c, errFind.Error())
			return
		}

		memberCard.CompanyId = body.CompanyId
		memberCard.CompanyName = company.Name

		if memberCard.Float == 1 {
			memberCard.OwnerUid = ""
		}
	}

	if mcType.Subject == constants.MEMBER_CARD_BASE_SUBJECT_FAMILY {
		// Check customer ralation ship
		owner.Uid = body.MemberConnect
		errFind = owner.FindFirst(db)
		if errFind != nil {
			response_message.BadRequest(c, errFind.Error())
			return
		}

		memberCard.MemberConnect = body.MemberConnect
		memberCard.Relationship = body.Relationship
	}

	errC := memberCard.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  memberCard.PartnerUid,
		CourseUid:   memberCard.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_CUSTOMER,
		Function:    constants.OP_LOG_FUNCTION_MEMBER_CARD,
		Action:      constants.OP_LOG_ACTION_CREATE,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{},
		ValueNew:    models.JsonDataLog{Data: memberCard},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         "",
		BookingDate: utils.GetCurrentDay1(),
		BillCode:    "",
		BookingUid:  "",
	}
	go createOperationLog(opLog)

	okResponse(c, memberCard)
}

func (_ *CMemberCard) GetListMemberCard(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListMemberCardForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	memberCardR := models.MemberCard{
		PartnerUid:    form.PartnerUid,
		CourseUid:     form.CourseUid,
		McTypeId:      form.McTypeId,
		OwnerUid:      form.OwnerUid,
		CardId:        form.CardId,
		MemberConnect: form.MemberConnect,
	}
	memberCardR.Status = form.Status
	list, total, err := memberCardR.FindList(db, page, form.PlayerName)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// Find list Golf Fee
	golfFee := models.GolfFee{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	listFee, errGF := golfFee.FindAll(db)
	if errGF == nil && len(listFee) > 0 {
		for i, v := range list {
			for j := 0; j < len(listFee); j++ {
				if listFee[j].GuestStyle == v["guest_style"] {
					list[i]["guest_style_name"] = listFee[j].GuestStyleName
					break
				}
			}
		}
	}

	res := map[string]interface{}{
		"total": total,
		"data":  list,
	}

	okResponse(c, res)
}

func (_ *CMemberCard) UpdateMemberCard(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	memberCardUidStr := c.Param("uid")
	if memberCardUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	memberCard := models.MemberCard{}
	memberCard.Uid = memberCardUidStr
	memberCard.PartnerUid = prof.PartnerUid
	memberCard.CourseUid = prof.CourseUid
	errF := memberCard.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	oldValue := memberCard.CloneMemberCard()

	body := models.MemberCard{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Check Member Card Type Exit
	mcType := models.MemberCardType{}
	mcType.Id = body.McTypeId
	errFind := mcType.FindFirst(db)
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	// if body.OwnerUid != "" {
	memberCard.OwnerUid = body.OwnerUid
	// }
	if body.Status != "" {
		memberCard.Status = body.Status
	}
	if body.ReasonUnactive != "" {
		memberCard.ReasonUnactive = body.ReasonUnactive
	}
	memberCard.McTypeId = body.McTypeId
	memberCard.ExpDate = body.ExpDate
	memberCard.PriceCode = body.PriceCode
	memberCard.GreenFee = body.GreenFee
	memberCard.CaddieFee = body.CaddieFee
	memberCard.BuggyFee = body.BuggyFee
	memberCard.Note = body.Note
	memberCard.ValidDate = body.ValidDate
	memberCard.StartPrecial = body.StartPrecial
	memberCard.EndPrecial = body.EndPrecial
	memberCard.AdjustPlayCount = body.AdjustPlayCount
	memberCard.Float = body.Float
	memberCard.PromotionCode = body.PromotionCode
	memberCard.UserEdit = body.UserEdit
	memberCard.AnnualType = body.AnnualType

	if mcType.Subject == constants.MEMBER_CARD_BASE_SUBJECT_COMPANY {
		// Check Company Exit
		company := models.Company{}
		company.Id = body.CompanyId
		errFind := company.FindFirst(db)
		if errFind != nil {
			response_message.BadRequest(c, errFind.Error())
			return
		}

		memberCard.CompanyId = body.CompanyId
		memberCard.CompanyName = company.Name

		if memberCard.Float == 1 {
			memberCard.OwnerUid = ""
		}
	}

	if mcType.Subject == constants.MEMBER_CARD_BASE_SUBJECT_FAMILY {
		memberCard.MemberConnect = body.MemberConnect
		memberCard.Relationship = body.Relationship
	}

	errUdp := memberCard.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  memberCard.PartnerUid,
		CourseUid:   memberCard.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_CUSTOMER,
		Function:    constants.OP_LOG_FUNCTION_MEMBER_CARD,
		Action:      constants.OP_LOG_ACTION_UPDATE,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{Data: oldValue},
		ValueNew:    models.JsonDataLog{Data: memberCard},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         "",
		BookingDate: utils.GetCurrentDay1(),
		BillCode:    "",
		BookingUid:  "",
	}
	go createOperationLog(opLog)

	okResponse(c, memberCard)
}

func (_ *CMemberCard) DeleteMemberCard(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	memberCardUidStr := c.Param("uid")
	if memberCardUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	member := models.MemberCard{}
	member.Uid = memberCardUidStr
	member.PartnerUid = prof.PartnerUid
	member.CourseUid = prof.CourseUid
	errF := member.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := member.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  member.PartnerUid,
		CourseUid:   member.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_CUSTOMER,
		Function:    constants.OP_LOG_FUNCTION_MEMBER_CARD,
		Action:      constants.OP_LOG_ACTION_DELETE,
		Body:        models.JsonDataLog{},
		ValueOld:    models.JsonDataLog{Data: member},
		ValueNew:    models.JsonDataLog{},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         "",
		BookingDate: utils.GetCurrentDay1(),
		BillCode:    "",
		BookingUid:  "",
	}
	go createOperationLog(opLog)

	okRes(c)
}

func (_ *CMemberCard) GetDetail(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	memberCardUidStr := c.Param("uid")
	if memberCardUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	memberCard := models.MemberCard{}
	memberCard.Uid = memberCardUidStr
	errF := memberCard.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	memberDetailRes, errFind := memberCard.FindDetail(db)
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	if memberDetailRes.PartnerUid != prof.PartnerUid || memberDetailRes.CourseUid != prof.CourseUid {
		response_message.Forbidden(c, "forbidden")
		return
	}

	okResponse(c, memberDetailRes)
}

func (_ *CMemberCard) UnactiveMemberCard(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.LockMemberCardBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	memberCard := models.MemberCard{}
	memberCard.Uid = body.MemberCardUid
	errF := memberCard.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	memberCard.Status = body.Status
	memberCard.ReasonUnactive = body.ReasonUnactive

	errUdp := memberCard.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, memberCard)
}
func (_ *CMemberCard) MarkContactCustomer(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.MarkContactCustomerBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	memberCard := models.MemberCard{}
	memberCard.Uid = body.MemberCardUid
	errF := memberCard.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	memberCard.IsContacted = *body.IsContacted

	errUdp := memberCard.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, memberCard)
}

func (_ *CMemberCard) UnMarkContactCustomer(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.UnMarkContactCustomerBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	rMemberCard := models.MemberCard{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		IsContacted: 1,
	}

	list, _, _ := rMemberCard.FindAllMemberCardContacted(db)

	for index, _ := range list {
		list[index].IsContacted = 0
	}

	rMemberCard.BatchUpdate(db, list)
	okRes(c)
}

func (_ *CMemberCard) MarkAllContactCustomer(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.UnMarkContactCustomerBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	rMemberCard := models.MemberCard{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		IsContacted: 0,
	}

	list, _, _ := rMemberCard.FindAllMemberCardContacted(db)

	for index, _ := range list {
		list[index].IsContacted = 1
	}

	rMemberCard.BatchUpdate(db, list)
	okRes(c)
}
