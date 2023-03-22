package controllers

import (
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CBookingWaiting struct{}

func (_ *CBookingWaiting) CreateBookingWaiting(c *gin.Context, prof models.CmsUser) {
	body := request.CreateBookingWaitingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	bookingWaiting := model_booking.BookingWaiting{}
	createBookingWaitingCommon(body, c, prof, db)

	errC := bookingWaiting.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	c.JSON(200, bookingWaiting)
}

func createBookingWaitingCommon(body request.CreateBookingWaitingBody, c *gin.Context, prof models.CmsUser, db *gorm.DB) *model_booking.BookingWaiting {
	bookingWaiting := model_booking.BookingWaiting{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		TeeType:     body.TeeType,
		TeePath:     body.TeePath,
		TeeTime:     body.TeeTime,
		TurnTime:    body.TurnTime,
		CmsUser:     prof.UserName,
		Hole:        body.Hole,
		BookingCode: body.BookingCode,
		CourseType:  body.CourseType,
	}

	// Check Guest of member, check member có còn slot đi cùng không
	guestStyle := ""

	if body.BookingDate != "" {
		bookingWaiting.BookingDate = body.BookingDate
	} else {
		dateDisplay, errDate := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
		if errDate == nil {
			bookingWaiting.BookingDate = dateDisplay
		} else {
			log.Println("booking date display err ", errDate.Error())
		}
	}

	//Check duplicated
	// isDuplicated, _ := bookingWaiting.IsDuplicated(db, true, true)
	// if isDuplicated {
	// 	response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
	// 	return nil
	// }

	// Member Card
	// Check xem booking guest hay booking member
	if body.MemberCardUid != nil && *body.MemberCardUid != "" {
		// Get Member Card
		var memberCard = models.MemberCard{}
		memberCard.Uid = *body.MemberCardUid
		errFind := memberCard.FindFirst(db)
		if errFind != nil {
			response_message.BadRequest(c, errFind.Error())
			return nil
		}

		owner, errOwner := memberCard.GetOwner(db)
		if errOwner != nil {
			response_message.BadRequest(c, errOwner.Error())
			return nil
		}

		bookingWaiting.MemberCardUid = *body.MemberCardUid
		bookingWaiting.CardId = memberCard.CardId
		bookingWaiting.CustomerName = owner.Name
		bookingWaiting.CustomerUid = owner.Uid
		bookingWaiting.CustomerType = owner.Type

		cus := convertToCustomerSqlIntoBooking(owner)
		bookingWaiting.CustomerInfo = &cus

		guestStyle = memberCard.GetGuestStyle(db)
	} else {
		bookingWaiting.CustomerName = body.CustomerName
	}

	if body.MemberUidOfGuest != nil && *body.MemberUidOfGuest != "" {
		var memberCardOfGuest = models.MemberCard{}
		memberCardOfGuest.Uid = body.MemberNameOfGuest
		errFind := memberCardOfGuest.FindFirst(db)
		if errFind != nil {
			response_message.BadRequest(c, errFind.Error())
			return nil
		}

		owner, errOwner := memberCardOfGuest.GetOwner(db)
		if errOwner != nil {
			response_message.BadRequest(c, errOwner.Error())
			return nil
		}

		bookingWaiting.MemberUidOfGuest = *body.MemberUidOfGuest
		bookingWaiting.MemberNameOfGuest = owner.Name
	}

	//Agency id
	if body.AgencyId > 0 {
		agency := models.Agency{}
		agency.Id = body.AgencyId
		errFindAgency := agency.FindFirst(db)
		if errFindAgency != nil || agency.Id == 0 {
			response_message.BadRequest(c, errFindAgency.Error())
			return nil
		}

		agencyBooking := cloneToAgencyBooking(agency)
		bookingWaiting.AgencyInfo = &agencyBooking
		bookingWaiting.AgencyId = body.AgencyId
		bookingWaiting.CustomerType = agency.Type
		guestStyle = agency.GuestStyle
	}

	if body.CustomerUid != "" {
		//check customer
		customer := models.CustomerUser{}
		customer.Uid = body.CustomerUid
		errFindCus := customer.FindFirst(db)
		if errFindCus != nil || customer.Uid == "" {
			response_message.BadRequest(c, "customer"+errFindCus.Error())
			return nil
		}

		bookingWaiting.CustomerName = customer.Name

		cus := convertToCustomerSqlIntoBooking(customer)
		bookingWaiting.CustomerInfo = &cus
		bookingWaiting.CustomerUid = body.CustomerUid
	}

	bookingWaiting.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())

	// Nếu guestyle truyền lên khác với gs của agency or member thì lấy gs truyền lên
	if body.GuestStyle != "" && guestStyle != body.GuestStyle {
		guestStyle = body.GuestStyle
	}

	if guestStyle != "" {
		golfFeeModel := models.GolfFee{
			PartnerUid: body.PartnerUid,
			CourseUid:  body.CourseUid,
			GuestStyle: guestStyle,
		}

		if errGS := golfFeeModel.FindFirst(db); errGS != nil {
			response_message.InternalServerError(c, "guest style not found ")
		}

		// Lấy phí bởi Guest style với ngày tạo
		golfFee := models.GolfFee{}
		var errFindGF error

		if bookingWaiting.BookingDate != "" {
			timeDate, _ := time.Parse(constants.DATE_FORMAT_1, bookingWaiting.BookingDate)
			golfFee, errFindGF = golfFeeModel.GetGuestStyleOnTime(db, timeDate)
			if errFindGF != nil {
				response_message.InternalServerError(c, "golf fee err "+errFindGF.Error())
			}
		} else {
			golfFee, errFindGF = golfFeeModel.GetGuestStyleOnDay(db)
			if errFindGF != nil {
				response_message.InternalServerError(c, "golf fee err "+errFindGF.Error())
			}
		}

		bookingWaiting.GuestStyle = guestStyle
		bookingWaiting.GuestStyleName = golfFee.GuestStyleName
	}

	// Update caddie
	if body.CaddieCode != nil && *body.CaddieCode != "" {
		caddieList := models.CaddieList{}
		caddieList.CourseUid = body.CourseUid
		caddieList.CaddieCode = *body.CaddieCode
		caddieNew, err := caddieList.FindFirst(db)
		if err != nil {
			response_message.BadRequestFreeMessage(c, "Caddie "+err.Error())
			return nil
		}

		// check caddie booking
		bookingWaiting.CaddieBooking = caddieNew.Code
	}

	if body.CustomerName != "" {
		bookingWaiting.CustomerName = body.CustomerName
	}

	if body.CustomerBookingName != "" {
		bookingWaiting.CustomerBookingName = body.CustomerBookingName
	}

	if body.CustomerBookingPhone != "" {
		bookingWaiting.CustomerBookingPhone = body.CustomerBookingPhone
	}

	if body.BookingCode == "" {
		bookingCode := utils.HashCodeUuid(uuid.New().String())
		bookingWaiting.BookingCode = bookingCode
	} else {
		bookingWaiting.BookingCode = body.BookingCode
	}

	return &bookingWaiting
}

func (cBooking *CBookingWaiting) CreateBookingWaitingList(c *gin.Context, prof models.CmsUser) {
	bodyRequest := request.CreateBatchBookingWaitingBody{}
	if bindErr := c.ShouldBind(&bodyRequest); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	bookingCode := utils.HashCodeUuid(uuid.New().String())
	for index, body := range bodyRequest.BookingList {
		if body.BookingCode == "" {
			bodyRequest.BookingList[index].BookingCode = bookingCode
		} else {
			bodyRequest.BookingList[index].BookingCode = body.BookingCode
		}
	}

	listBookingWaiting := []model_booking.BookingWaiting{}
	for _, body := range bodyRequest.BookingList {
		booking := createBookingWaitingCommon(body, c, prof, db)
		if booking != nil {
			listBookingWaiting = append(listBookingWaiting, *booking)
		}
	}

	errCreate := db.Create(&listBookingWaiting).Error

	if errCreate != nil {
		response_message.BadRequestFreeMessage(c, errCreate.Error())
		return
	}

	okRes(c)
}

func (_ *CBookingWaiting) GetBookingWaitingList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingWaitingForm{}
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

	bookingWaitingRequest := model_booking.BookingWaiting{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	if form.PlayerName != "" {
		bookingWaitingRequest.CustomerName = form.PlayerName
	}

	if form.BookingDate != "" {
		bookingWaitingRequest.BookingDate = form.BookingDate
	}

	if form.BookingCode != "" {
		bookingWaitingRequest.BookingCode = form.BookingCode
	}

	list, total, err := bookingWaitingRequest.FindList(db, page)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := response.PageResponse{
		Total: total,
		Data:  list,
	}

	c.JSON(200, res)
}

func (_ *CBookingWaiting) DeleteBookingWaiting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	bookingIdStr := c.Param("id")
	bookingId, errId := strconv.ParseInt(bookingIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	bookingWaitingRequest := model_booking.BookingWaiting{}
	bookingWaitingRequest.Id = bookingId
	bookingWaitingRequest.PartnerUid = prof.PartnerUid
	bookingWaitingRequest.CourseUid = prof.CourseUid
	errF := bookingWaitingRequest.FindFirst(db)

	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	err := bookingWaitingRequest.Delete(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	opLog := models.OperationLog{
		PartnerUid:  prof.PartnerUid,
		CourseUid:   prof.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Function:    constants.OP_LOG_FUNCTION_WAITTING_LIST,
		Action:      constants.OP_LOG_ACTION_DELETE,
		Body:        models.JsonDataLog{Data: bookingIdStr},
		ValueOld:    models.JsonDataLog{Data: bookingWaitingRequest},
		ValueNew:    models.JsonDataLog{},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		BookingDate: bookingWaitingRequest.BookingDate,
	}
	go createOperationLog(opLog)

	okRes(c)
}

func (_ *CBookingWaiting) DeleteBookingWaitingByBookingCode(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	bodyRequest := request.DeleteBookingWaiting{}
	if bindErr := c.ShouldBind(&bodyRequest); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	bookingWaitingRequest := model_booking.BookingWaiting{}
	bookingWaitingRequest.PartnerUid = bodyRequest.PartnerUid
	bookingWaitingRequest.CourseUid = bodyRequest.CourseUid
	bookingWaitingRequest.TeeTime = bodyRequest.TeeTime
	bookingWaitingRequest.TeeType = bodyRequest.TeeType
	bookingWaitingRequest.CourseType = bodyRequest.CourseType
	bookingWaitingRequest.BookingCode = bodyRequest.BookingCode
	bookingWaitingRequest.BookingDate = bodyRequest.BookingDate

	errF := bookingWaitingRequest.DeleteByBookingCode(db)

	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	opLog := models.OperationLog{
		PartnerUid:  prof.PartnerUid,
		CourseUid:   prof.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Function:    constants.OP_LOG_FUNCTION_WAITTING_LIST,
		Action:      constants.OP_LOG_ACTION_DELETE,
		Body:        models.JsonDataLog{Data: bodyRequest},
		ValueOld:    models.JsonDataLog{Data: bookingWaitingRequest},
		ValueNew:    models.JsonDataLog{},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		BookingDate: bookingWaitingRequest.BookingDate,
	}
	go createOperationLog(opLog)

	okRes(c)
}

func (_ *CBookingWaiting) UpdateBookingWaiting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	var bodyRequest request.CreateBatchBookingWaitingBody
	if bindErr := c.ShouldBind(&bodyRequest); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	for _, body := range bodyRequest.BookingList {
		bookingWaiting := model_booking.BookingWaiting{}
		bookingWaiting.Id = body.Id
		bookingWaiting.PartnerUid = prof.PartnerUid
		bookingWaiting.CourseUid = prof.CourseUid

		errF := bookingWaiting.FindFirst(db)
		if errF != nil {
			response_message.BadRequest(c, errF.Error())
			return
		}

		// Data old
		// oldData := bookingWaiting

		// Check Guest of member, check member có còn slot đi cùng không
		guestStyle := ""

		if body.BookingDate != "" {
			bookingWaiting.BookingDate = body.BookingDate
		} else {
			dateDisplay, errDate := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())
			if errDate == nil {
				bookingWaiting.BookingDate = dateDisplay
			} else {
				log.Println("booking date display err ", errDate.Error())
			}
		}

		// Member Card
		// Check xem booking guest hay booking member
		if body.MemberCardUid != nil {
			if *body.MemberCardUid != "" {
				if *body.MemberCardUid != bookingWaiting.MemberCardUid {
					// Get Member Card
					var memberCard = models.MemberCard{}
					memberCard.Uid = *body.MemberCardUid
					errFind := memberCard.FindFirst(db)
					if errFind != nil {
						response_message.BadRequest(c, errFind.Error())
						return
					}

					owner, errOwner := memberCard.GetOwner(db)
					if errOwner != nil {
						response_message.BadRequest(c, errOwner.Error())
						return
					}

					bookingWaiting.MemberCardUid = *body.MemberCardUid
					bookingWaiting.CardId = memberCard.CardId
					bookingWaiting.CustomerName = owner.Name
					bookingWaiting.CustomerUid = owner.Uid
					bookingWaiting.CustomerType = owner.Type

					cus := convertToCustomerSqlIntoBooking(owner)
					bookingWaiting.CustomerInfo = &cus

					guestStyle = memberCard.GetGuestStyle(db)
				}
			} else {
				bookingWaiting.MemberCardUid = ""
				bookingWaiting.CardId = ""
				bookingWaiting.CustomerUid = ""
				bookingWaiting.CustomerType = ""
				bookingWaiting.CustomerInfo = &model_booking.CustomerInfo{}
			}
		} else {
			bookingWaiting.CustomerName = body.CustomerName
		}

		if body.MemberCardUid != nil {
			if *body.MemberCardUid != "" {
				if *body.MemberCardUid != bookingWaiting.MemberCardUid {
					var memberCardOfGuest = models.MemberCard{}
					memberCardOfGuest.Uid = body.MemberNameOfGuest
					errFind := memberCardOfGuest.FindFirst(db)
					if errFind != nil {
						response_message.BadRequest(c, errFind.Error())
						return
					}

					owner, errOwner := memberCardOfGuest.GetOwner(db)
					if errOwner != nil {
						response_message.BadRequest(c, errOwner.Error())
						return
					}

					bookingWaiting.MemberUidOfGuest = *body.MemberUidOfGuest
					bookingWaiting.MemberNameOfGuest = owner.Name
				}
			} else {
				bookingWaiting.MemberUidOfGuest = ""
				bookingWaiting.MemberNameOfGuest = ""
			}
		}

		//Agency id
		if body.AgencyId > 0 && body.AgencyId != bookingWaiting.AgencyId {
			agency := models.Agency{}
			agency.Id = body.AgencyId
			errFindAgency := agency.FindFirst(db)
			if errFindAgency != nil || agency.Id == 0 {
				response_message.BadRequest(c, errFindAgency.Error())
				return
			}

			agencyBooking := cloneToAgencyBooking(agency)
			bookingWaiting.AgencyInfo = &agencyBooking
			bookingWaiting.AgencyId = body.AgencyId
			bookingWaiting.CustomerType = agency.Type
			guestStyle = agency.GuestStyle
		}

		if body.CustomerUid != "" && body.CustomerUid != bookingWaiting.CustomerUid {
			//check customer
			customer := models.CustomerUser{}
			customer.Uid = body.CustomerUid
			errFindCus := customer.FindFirst(db)
			if errFindCus != nil || customer.Uid == "" {
				response_message.BadRequest(c, "customer"+errFindCus.Error())
				return
			}

			bookingWaiting.CustomerName = customer.Name

			cus := convertToCustomerSqlIntoBooking(customer)
			bookingWaiting.CustomerInfo = &cus
			bookingWaiting.CustomerUid = body.CustomerUid
		}

		bookingWaiting.CmsUserLog = getBookingCmsUserLog(prof.UserName, utils.GetTimeNow().Unix())

		// Nếu guestyle truyền lên khác với gs của agency or member thì lấy gs truyền lên
		if body.GuestStyle != "" && guestStyle != body.GuestStyle {
			guestStyle = body.GuestStyle
		}

		if guestStyle != "" {
			golfFeeModel := models.GolfFee{
				PartnerUid: body.PartnerUid,
				CourseUid:  body.CourseUid,
				GuestStyle: guestStyle,
			}

			if errGS := golfFeeModel.FindFirst(db); errGS != nil {
				response_message.InternalServerError(c, "guest style not found ")
			}

			// Lấy phí bởi Guest style với ngày tạo
			golfFee := models.GolfFee{}
			var errFindGF error

			if bookingWaiting.BookingDate != "" {
				timeDate, _ := time.Parse(constants.DATE_FORMAT_1, bookingWaiting.BookingDate)
				golfFee, errFindGF = golfFeeModel.GetGuestStyleOnTime(db, timeDate)
				if errFindGF != nil {
					response_message.InternalServerError(c, "golf fee err "+errFindGF.Error())
				}
			} else {
				golfFee, errFindGF = golfFeeModel.GetGuestStyleOnDay(db)
				if errFindGF != nil {
					response_message.InternalServerError(c, "golf fee err "+errFindGF.Error())
				}
			}

			bookingWaiting.GuestStyle = guestStyle
			bookingWaiting.GuestStyleName = golfFee.GuestStyleName
		}

		// Update caddie
		if body.CaddieCode != nil && *body.CaddieCode != "" {
			caddieList := models.CaddieList{}
			caddieList.CourseUid = body.CourseUid
			caddieList.CaddieCode = *body.CaddieCode
			caddieNew, err := caddieList.FindFirst(db)
			if err != nil {
				response_message.BadRequestFreeMessage(c, "Caddie "+err.Error())
				return
			}

			// check caddie booking
			bookingWaiting.CaddieBooking = caddieNew.Code
		}

		if body.CustomerName != "" {
			bookingWaiting.CustomerName = body.CustomerName
		}

		if body.CustomerBookingName != "" {
			bookingWaiting.CustomerBookingName = body.CustomerBookingName
		}

		if body.CustomerBookingPhone != "" {
			bookingWaiting.CustomerBookingPhone = body.CustomerBookingPhone
		}

		if body.BookingCode == "" {
			bookingCode := utils.HashCodeUuid(uuid.New().String())
			bookingWaiting.BookingCode = bookingCode
		} else {
			bookingWaiting.BookingCode = body.BookingCode
		}

		err := bookingWaiting.Update(db)
		if err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	// opLog := models.OperationLog{
	// 	PartnerUid:  body.PartnerUid,
	// 	CourseUid:   body.CourseUid,
	// 	UserName:    prof.UserName,
	// 	UserUid:     prof.Uid,
	// 	Module:      constants.OP_LOG_MODULE_RECEPTION,
	// 	Function:    constants.OP_LOG_FUNCTION_WAITTING_LIST,
	// 	Action:      constants.OP_LOG_ACTION_UPDATE,
	// 	Body:        models.JsonDataLog{Data: body},
	// 	ValueOld:    models.JsonDataLog{Data: oldData},
	// 	ValueNew:    models.JsonDataLog{Data: bookingWaiting},
	// 	Path:        c.Request.URL.Path,
	// 	Method:      c.Request.Method,
	// 	BookingDate: bookingWaiting.BookingDate,
	// }
	// go createOperationLog(opLog)

	okRes(c)
}
