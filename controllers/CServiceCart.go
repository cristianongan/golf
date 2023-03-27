package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	model_service "start/models/service"
	"start/utils"
	"start/utils/response_message"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type CServiceCart struct{}

// Thêm sản phẩm vào giỏ hàng
func (_ CServiceCart) AddItemServiceToCart(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.AddItemServiceCartBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate golf bag
	dateDisplay, _ := utils.GetBookingDateFromTimestamp(time.Now().Unix())

	booking := model_booking.Booking{}
	booking.PartnerUid = body.PartnerUid
	booking.CourseUid = body.CourseUid
	booking.Bag = body.GolfBag
	booking.BookingDate = dateDisplay
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Booking "+err.Error())
		return
	}

	if booking.BagStatus != constants.BAG_STATUS_WAITING && booking.BagStatus != constants.BAG_STATUS_IN_COURSE && booking.BagStatus != constants.BAG_STATUS_TIMEOUT {
		response_message.BadRequest(c, "Bag status invalid")
		return
	}

	if *booking.LockBill {
		response_message.BadRequestDynamicKey(c, "BAG_BE_LOCK", "Bag lock")
		return
	}

	// validate kiosk
	kiosk := model_service.Kiosk{}
	kiosk.Id = body.ServiceId
	if err := kiosk.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Kiosk "+err.Error())
		return
	}

	// validate quantity
	// inventory := kiosk_inventory.InventoryItem{}
	// inventory.PartnerUid = body.PartnerUid
	// inventory.CourseUid = body.CourseUid
	// inventory.ServiceId = body.ServiceId
	// inventory.Code = body.ItemCode

	// if err := inventory.FindFirst(db); err != nil {
	// 	response_message.BadRequest(c, "Inventory "+err.Error())
	// 	return
	// }

	// // Kiểm tra số lượng hàng tồn trong kho
	// if body.Quantity > inventory.Quantity {
	// 	response_message.BadRequest(c, "The quantity of goods in stock is not enough")
	// 	return
	// }

	// // Update số lượng hàng tồn trong kho
	// inventory.Quantity -= body.Quantity
	// if err := inventory.Update(db); err != nil {
	// 	response_message.BadRequest(c, err.Error())
	// 	return
	// }

	// create cart item
	serviceCartItem := model_booking.BookingServiceItem{}

	// validate item code by group
	if kiosk.ServiceType == constants.GROUP_FB {
		fb := model_service.FoodBeverage{}
		fb.PartnerUid = body.PartnerUid
		fb.CourseUid = body.CourseUid
		fb.FBCode = body.ItemCode

		if err := fb.FindFirst(db); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
		// add infor cart item
		serviceCartItem.ItemId = fb.Id
		serviceCartItem.Type = kiosk.KioskType
		serviceCartItem.Location = kiosk.KioskName
		serviceCartItem.GroupCode = fb.GroupCode
		serviceCartItem.Name = fb.VieName
		serviceCartItem.EngName = fb.EnglishName
		serviceCartItem.UnitPrice = int64(fb.Price)
		serviceCartItem.Unit = fb.Unit
	}

	if kiosk.ServiceType == constants.GROUP_PROSHOP {
		proshop := model_service.Proshop{}
		proshop.PartnerUid = body.PartnerUid
		proshop.CourseUid = body.CourseUid
		proshop.ProShopId = body.ItemCode

		if err := proshop.FindFirst(db); err != nil {
			response_message.BadRequest(c, "Proshop "+err.Error())
			return
		}
		// add infor cart item
		serviceCartItem.ItemId = proshop.Id
		serviceCartItem.Type = kiosk.KioskType
		serviceCartItem.Location = kiosk.KioskName
		serviceCartItem.GroupCode = proshop.GroupCode
		serviceCartItem.Name = proshop.VieName
		serviceCartItem.EngName = proshop.EnglishName
		serviceCartItem.UnitPrice = int64(proshop.Price)
		serviceCartItem.Unit = proshop.Unit
	}

	// check service cart
	serviceCart := models.ServiceCart{}
	serviceCart.PartnerUid = body.PartnerUid
	serviceCart.CourseUid = body.CourseUid

	applyDate := utils.GetDateLocal()

	if body.BillId != 0 {
		serviceCart.Id = body.BillId
	} else {
		serviceCart.GolfBag = body.GolfBag
		serviceCart.BookingUid = booking.Uid
		serviceCart.BookingDate = datatypes.Date(applyDate)
		serviceCart.ServiceId = body.ServiceId
		serviceCart.BillCode = constants.BILL_NONE
		serviceCart.StaffOrder = prof.FullName
		serviceCart.BillStatus = constants.POS_BILL_STATUS_PENDING
		serviceCart.ServiceType = kiosk.KioskType
		serviceCart.PlayerName = booking.CustomerName
	}

	err := serviceCart.FindFirst(db)
	// no cart
	if err != nil {
		// create cart
		serviceCart.Amount = body.Quantity * serviceCartItem.UnitPrice
		if err := serviceCart.Create(db); err != nil {
			response_message.InternalServerError(c, "Create cart "+err.Error())
			return
		}
	} else {
		//Kiểm tra trạng thái bill
		// if serviceCart.BillStatus == constants.POS_BILL_STATUS_OUT {
		// 	response_message.BadRequest(c, "Bill status invalid")
		// 	return
		// }
		// update tổng giá bill
		serviceCart.Amount += body.Quantity * serviceCartItem.UnitPrice
		if err := serviceCart.Update(db); err != nil {
			response_message.InternalServerError(c, "Update cart "+err.Error())
			return
		}
	}

	// add infor cart item
	serviceCartItem.PartnerUid = body.PartnerUid
	serviceCartItem.CourseUid = body.CourseUid
	serviceCartItem.ServiceType = kiosk.ServiceType
	serviceCartItem.Bag = booking.Bag
	serviceCartItem.BillCode = booking.BillCode
	serviceCartItem.BookingUid = booking.Uid
	serviceCartItem.PlayerName = booking.CustomerName
	serviceCartItem.ServiceId = strconv.Itoa(int(serviceCart.ServiceId))
	serviceCartItem.ServiceBill = serviceCart.Id
	serviceCartItem.ItemCode = body.ItemCode
	serviceCartItem.Quality = int(body.Quantity)
	serviceCartItem.Amount = body.Quantity * serviceCartItem.UnitPrice
	serviceCartItem.UserAction = prof.UserName

	if err := serviceCartItem.Create(db); err != nil {
		response_message.InternalServerError(c, "Create item "+err.Error())
		return
	}

	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_POS,
		Action:      constants.OP_LOG_ACTION_ADD_ITEM,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{},
		ValueNew:    models.JsonDataLog{Data: serviceCartItem},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}

	if serviceCartItem.ServiceType == constants.KIOSK_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_KIOSK
	} else if serviceCartItem.ServiceType == constants.MINI_B_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_MINI_BAR
	} else if serviceCartItem.ServiceType == constants.PROSHOP_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_PROSHOP
	}

	go createOperationLog(opLog)

	c.JSON(200, serviceCart)
}

func (_ CServiceCart) AddItemRentalToCart(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.AddItemRentalCartBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate golf bag
	dateDisplay, _ := utils.GetBookingDateFromTimestamp(time.Now().Unix())

	booking := model_booking.Booking{}
	booking.PartnerUid = body.PartnerUid
	booking.CourseUid = body.CourseUid
	booking.Bag = body.GolfBag
	booking.BookingDate = dateDisplay
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Booking "+err.Error())
		return
	}

	if booking.BagStatus != constants.BAG_STATUS_WAITING && booking.BagStatus != constants.BAG_STATUS_IN_COURSE && booking.BagStatus != constants.BAG_STATUS_TIMEOUT {
		response_message.BadRequest(c, "Bag status invalid")
		return
	}

	if *booking.LockBill {
		response_message.BadRequestDynamicKey(c, "BAG_BE_LOCK", "Bag lock")
		return
	}

	// validate kiosk
	kiosk := model_service.Kiosk{}
	kiosk.Id = body.ServiceId
	if err := kiosk.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Kiosk "+err.Error())
		return
	}

	// create cart item
	serviceCartItem := model_booking.BookingServiceItem{}

	// add infor cart item
	serviceCartItem.Type = kiosk.KioskType
	if body.LocationType != "" {
		serviceCartItem.Location = body.LocationType
	} else {
		serviceCartItem.Location = kiosk.KioskName
	}
	serviceCartItem.Name = body.Name
	serviceCartItem.UnitPrice = body.Price

	if body.ItemCode != "" {
		rental := model_service.Rental{}
		rental.PartnerUid = body.PartnerUid
		rental.CourseUid = body.CourseUid
		rental.RentalId = body.ItemCode

		if err := rental.FindFirst(db); err != nil {
			response_message.BadRequest(c, "Proshop "+err.Error())
			return
		}

		serviceCartItem.ItemId = rental.Id
		serviceCartItem.GroupCode = rental.GroupCode
		serviceCartItem.EngName = rental.EnglishName
		serviceCartItem.Unit = rental.Unit
	}

	if body.Hole > 0 {
		serviceCartItem.Hole = body.Hole
	}

	// check service cart
	serviceCart := models.ServiceCart{}
	serviceCart.PartnerUid = body.PartnerUid
	serviceCart.CourseUid = body.CourseUid

	if body.BillId != 0 {
		serviceCart.Id = body.BillId
	} else {
		applyDate := utils.GetDateLocal()

		serviceCart.GolfBag = body.GolfBag
		serviceCart.BookingUid = booking.Uid
		serviceCart.BookingDate = datatypes.Date(applyDate)
		serviceCart.ServiceId = body.ServiceId
		serviceCart.BillCode = constants.BILL_NONE
		serviceCart.StaffOrder = prof.FullName
		serviceCart.BillStatus = constants.POS_BILL_STATUS_PENDING
	}

	err := serviceCart.FindFirst(db)
	// no cart
	if err != nil {
		// create cart
		serviceCart.RentalStatus = constants.POS_RETAL_STATUS_RENT
		serviceCart.Amount = body.Quantity * serviceCartItem.UnitPrice
		serviceCart.CaddieCode = body.CaddieCode
		serviceCart.ServiceType = kiosk.KioskType
		serviceCart.PlayerName = booking.CustomerName

		if err := serviceCart.Create(db); err != nil {
			response_message.InternalServerError(c, "Create cart "+err.Error())
			return
		}
	} else {
		//Kiểm tra trạng thái bill
		// if serviceCart.BillStatus == constants.POS_BILL_STATUS_OUT {
		// 	response_message.BadRequest(c, "Bill status invalid")
		// 	return
		// }
		// update tổng giá bill
		serviceCart.Amount += body.Quantity * serviceCartItem.UnitPrice
		if err := serviceCart.Update(db); err != nil {
			response_message.InternalServerError(c, "Update cart "+err.Error())
			return
		}
	}

	// add infor cart item
	serviceCartItem.PartnerUid = body.PartnerUid
	serviceCartItem.CourseUid = body.CourseUid
	serviceCartItem.Bag = booking.Bag
	serviceCartItem.BillCode = booking.BillCode
	serviceCartItem.BookingUid = booking.Uid
	serviceCartItem.PlayerName = booking.CustomerName
	serviceCartItem.ServiceId = strconv.Itoa(int(serviceCart.ServiceId))
	serviceCartItem.ServiceBill = serviceCart.Id
	serviceCartItem.ItemCode = body.ItemCode
	serviceCartItem.Quality = int(body.Quantity)
	serviceCartItem.Amount = body.Quantity * serviceCartItem.UnitPrice
	serviceCartItem.UserAction = prof.UserName

	if body.ServiceType != "" {
		serviceCartItem.ServiceType = body.ServiceType
	} else {
		serviceCartItem.ServiceType = kiosk.ServiceType
	}

	if err := serviceCartItem.Create(db); err != nil {
		response_message.InternalServerError(c, "Create item "+err.Error())
		return
	}

	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_GO,
		Action:      constants.OP_LOG_ACTION_ADD_ITEM,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{},
		ValueNew:    models.JsonDataLog{Data: serviceCartItem},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}

	if body.LocationType == "GO" {
		opLog.Module = constants.OP_LOG_MODULE_GO
		opLog.Function = constants.OP_LOG_FUNCTION_COURSE_INFO_IN_COURSE
	} else {
		if serviceCartItem.Type == constants.RENTAL_SETTING {
			opLog.Function = constants.OP_LOG_FUNCTION_GOLF_CLUB_RENTAL
		} else if serviceCartItem.Type == constants.DRIVING_SETTING {
			opLog.Function = constants.OP_LOG_FUNCTION_DRIVING
		}

		opLog.Module = constants.OP_LOG_MODULE_POS
	}

	go createOperationLog(opLog)

	c.JSON(200, serviceCart)
}

func (_ CServiceCart) AddDiscountToItem(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.AddDiscountServiceItemBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate cart item
	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.Id = body.ItemId

	if err := serviceCartItem.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate cart
	serviceCart := models.ServiceCart{}
	serviceCart.Id = serviceCartItem.ServiceBill

	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validaet booking
	bookingR := model_booking.Booking{}
	bookingR.Uid = serviceCart.BookingUid
	booking, errF := bookingR.FindFirstByUId(db)
	if errF != nil {
		response_message.InternalServerError(c, "Booking "+errF.Error())
		return
	}

	if booking.BagStatus != constants.BAG_STATUS_WAITING && booking.BagStatus != constants.BAG_STATUS_IN_COURSE && booking.BagStatus != constants.BAG_STATUS_TIMEOUT {
		response_message.BadRequest(c, "Bag status invalid")
		return
	}

	if *booking.LockBill {
		response_message.BadRequestDynamicKey(c, "BAG_BE_LOCK", "Bag lock")
		return
	}

	// Update amount
	if body.DiscountType == constants.ITEM_BILL_DISCOUNT_BY_PERCENT {
		amountDiscont := (int64(serviceCartItem.Quality) * serviceCartItem.UnitPrice) * (100 - body.DiscountPrice) / 100

		serviceCart.Amount = serviceCart.Amount - serviceCartItem.Amount + amountDiscont
		serviceCartItem.Amount = amountDiscont

	} else if body.DiscountType == constants.ITEM_BILL_DISCOUNT_BY_PRICE {
		var amountDiscont int64

		totalPrice := serviceCartItem.Quality * int(serviceCartItem.UnitPrice)
		amountRaw := int64(totalPrice) - body.DiscountPrice

		if amountRaw > 0 {
			amountDiscont = amountRaw
			serviceCart.Amount = serviceCart.Amount - serviceCartItem.Amount + amountRaw
		} else {
			serviceCart.Amount = serviceCart.Amount - serviceCartItem.Amount
			amountDiscont = 0
		}
		serviceCartItem.Amount = amountDiscont
	}

	serviceCartItem.DiscountType = body.DiscountType
	serviceCartItem.DiscountValue = body.DiscountPrice
	serviceCartItem.DiscountReason = body.DiscountReason

	if err := serviceCartItem.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	if err := serviceCart.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	//Update giá nếu bill active
	if serviceCart.BillStatus != constants.POS_BILL_STATUS_PENDING &&
		serviceCart.BillStatus != constants.POS_BILL_STATUS_OUT &&
		serviceCart.BillStatus != constants.RES_BILL_STATUS_ORDER &&
		serviceCart.BillStatus != constants.RES_BILL_STATUS_BOOKING &&
		serviceCart.BillStatus != constants.RES_BILL_STATUS_CANCEL {
		//Update lại giá trong booking
		updatePriceWithServiceItem(&booking, prof)
	}

	go addLog(c, prof, serviceCartItem, constants.OP_LOG_ACTION_ADD_DISCOUNT)

	okRes(c)
}

func (_ CServiceCart) AddDiscountToBill(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.AddDiscountBillBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate cart
	serviceCart := models.ServiceCart{}
	serviceCart.Id = body.BillId

	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validaet booking
	bookingR := model_booking.Booking{}
	bookingR.Uid = serviceCart.BookingUid
	booking, errF := bookingR.FindFirstByUId(db)
	if errF != nil {
		response_message.InternalServerError(c, "Booking "+errF.Error())
		return
	}

	if booking.BagStatus != constants.BAG_STATUS_WAITING && booking.BagStatus != constants.BAG_STATUS_IN_COURSE && booking.BagStatus != constants.BAG_STATUS_TIMEOUT {
		response_message.BadRequest(c, "Bag status invalid")
		return
	}

	if *booking.LockBill {
		response_message.BadRequestDynamicKey(c, "BAG_BE_LOCK", "Bag lock")
		return
	}

	// Update service cart
	serviceCart.DiscountType = body.DiscountType
	serviceCart.DiscountValue = body.DiscountPrice
	serviceCart.DiscountReason = body.DiscountReason

	if err := serviceCart.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	//Update giá nếu bill active
	if serviceCart.BillStatus == constants.POS_BILL_STATUS_ACTIVE &&
		serviceCart.BillStatus != constants.RES_BILL_STATUS_ORDER &&
		serviceCart.BillStatus != constants.RES_BILL_STATUS_BOOKING &&
		serviceCart.BillStatus != constants.RES_BILL_STATUS_CANCEL {
		//Update lại giá trong booking
		updatePriceWithServiceItem(&booking, prof)
	}

	okRes(c)
}

func (_ CServiceCart) GetItemInCart(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	query := request.GetItemServiceCartBody{}
	if bindErr := c.ShouldBind(&query); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	page := models.Page{
		Limit:   query.PageRequest.Limit,
		Page:    query.PageRequest.Page,
		SortBy:  query.PageRequest.SortBy,
		SortDir: query.PageRequest.SortDir,
	}

	bookingDate, _ := time.Parse(constants.DATE_FORMAT, query.BookingDate)

	serviceCart := models.ServiceCart{}
	serviceCart.PartnerUid = query.PartnerUid
	serviceCart.CourseUid = query.CourseUid
	serviceCart.ServiceId = query.ServiceId
	serviceCart.GolfBag = query.GolfBag
	serviceCart.BookingDate = datatypes.Date(bookingDate)
	serviceCart.Id = query.BillId
	serviceCart.BillStatus = query.BillStatus

	if err := serviceCart.FindFirst(db); err != nil {
		res := response.PageResponse{
			Total: 0,
			Data:  nil,
		}

		c.JSON(200, res)
		return
	}

	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.ServiceBill = serviceCart.Id

	list, total, err := serviceCartItem.FindList(db, page)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	dataRes := map[string]interface{}{
		"cart_infor": serviceCart,
		"list_item":  list,
	}

	res := response.PageResponse{
		Total: total,
		Data:  dataRes,
	}

	c.JSON(200, res)
}

func (_ CServiceCart) GetBestItemInKiosk(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	query := request.GetBestItemBody{}
	if bindErr := c.ShouldBind(&query); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	page := models.Page{
		Limit:   query.PageRequest.Limit,
		Page:    query.PageRequest.Page,
		SortBy:  query.PageRequest.SortBy,
		SortDir: query.PageRequest.SortDir,
	}

	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.PartnerUid = query.PartnerUid
	serviceCartItem.CourseUid = query.CourseUid
	serviceCartItem.ServiceId = strconv.Itoa(int(query.ServiceId))
	serviceCartItem.GroupCode = query.GroupCode

	list, total, err := serviceCartItem.FindBestCartItem(db, page)

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

func (_ CServiceCart) GetBestGroupInKiosk(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	query := request.GetBestGroupBody{}
	if bindErr := c.ShouldBind(&query); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	page := models.Page{
		Limit:   query.PageRequest.Limit,
		Page:    query.PageRequest.Page,
		SortBy:  query.PageRequest.SortBy,
		SortDir: query.PageRequest.SortDir,
	}

	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.PartnerUid = query.PartnerUid
	serviceCartItem.CourseUid = query.CourseUid
	serviceCartItem.ServiceId = query.ServiceId

	list, total, err := serviceCartItem.FindBestGroup(db, page)

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

func (_ CServiceCart) GetListCart(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	query := request.GetServiceCartBody{}
	if bindErr := c.ShouldBind(&query); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	page := models.Page{
		Limit:   query.PageRequest.Limit,
		Page:    query.PageRequest.Page,
		SortBy:  query.PageRequest.SortBy,
		SortDir: query.PageRequest.SortDir,
	}

	bookingDate, _ := time.Parse(constants.DATE_FORMAT, query.BookingDate)

	serviceCart := models.ServiceCart{}
	serviceCart.PartnerUid = query.PartnerUid
	serviceCart.CourseUid = query.CourseUid
	serviceCart.ServiceId = query.ServiceId
	serviceCart.BookingDate = datatypes.Date(bookingDate)
	serviceCart.GolfBag = query.GolfBag

	list, total, err := serviceCart.FindList(db, page)

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

func (_ CServiceCart) GetListRentalCart(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	query := request.GetServiceCartRentalBody{}
	if bindErr := c.ShouldBind(&query); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	page := models.Page{
		Limit:   query.PageRequest.Limit,
		Page:    query.PageRequest.Page,
		SortBy:  query.PageRequest.SortBy,
		SortDir: query.PageRequest.SortDir,
	}

	bookingDate, _ := time.Parse(constants.DATE_FORMAT, query.BookingDate)

	serviceCart := models.ServiceCart{}
	serviceCart.PartnerUid = query.PartnerUid
	serviceCart.CourseUid = query.CourseUid
	serviceCart.ServiceId = query.ServiceId
	serviceCart.BookingDate = datatypes.Date(bookingDate)
	serviceCart.RentalStatus = query.RentalStatus
	serviceCart.GolfBag = query.GolfBag

	list, total, err := serviceCart.FindList(db, page)

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

func (_ CServiceCart) UpdateItemCart(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.UpdateServiceCartBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	if body.PartnerUid != "" && body.PartnerUid != prof.PartnerUid {
		response_message.BadRequest(c, "invalid params")
		return
	}
	if body.CourseUid != "" && body.CourseUid != prof.CourseUid {
		response_message.BadRequest(c, "invalid params")
		return
	}

	// validate cart item
	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.Id = body.CartItemId
	serviceCartItem.PartnerUid = prof.PartnerUid
	serviceCartItem.CourseUid = prof.CourseUid

	if err := serviceCartItem.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate golf bag
	dateDisplay, _ := utils.GetBookingDateFromTimestamp(time.Now().Unix())

	booking := model_booking.Booking{}
	booking.PartnerUid = body.PartnerUid
	booking.CourseUid = body.CourseUid
	booking.Bag = serviceCartItem.Bag
	booking.BookingDate = dateDisplay
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Booking "+err.Error())
		return
	}

	// Check bag status
	if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		response_message.BadRequest(c, "Bag check out")
		return
	}

	if *booking.LockBill {
		response_message.BadRequestDynamicKey(c, "BAG_BE_LOCK", "Bag lock")
		return
	}

	// validate cart
	serviceCart := models.ServiceCart{}
	serviceCart.Id = serviceCartItem.ServiceBill

	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Kiểm tra trạng thái bill
	// if serviceCart.BillStatus != constants.POS_BILL_STATUS_PENDING {
	// 	response_message.BadRequest(c, "Bill status invalid")
	// 	return
	// }

	if body.Quantity > 0 {
		// if serviceCartItem.Type != constants.RENTAL_SETTING &&
		// 	serviceCartItem.Type != constants.DRIVING_SETTING {
		// 	// validate quantity
		// 	inventory := kiosk_inventory.InventoryItem{}
		// 	inventory.PartnerUid = body.PartnerUid
		// 	inventory.CourseUid = body.CourseUid
		// 	inventory.ServiceId = serviceCart.ServiceId
		// 	inventory.Code = serviceCartItem.ItemCode

		// 	if err := inventory.FindFirst(db); err != nil {
		// 		response_message.BadRequest(c, err.Error())
		// 		return
		// 	}

		// 	// Kiểm tra số lượng hàng tồn trong kho
		// 	if body.Quantity > inventory.Quantity+int64(serviceCartItem.Quality) {
		// 		response_message.BadRequest(c, "The quantity of goods in stock is not enough")
		// 		return
		// 	}

		// 	// Update số lượng hàng tồn trong kho
		// 	inventory.Quantity = inventory.Quantity + int64(serviceCartItem.Quality) - body.Quantity
		// 	if err := inventory.Update(db); err != nil {
		// 		response_message.BadRequest(c, err.Error())
		// 		return
		// 	}
		// }

		// Update amount
		if serviceCartItem.DiscountType == constants.ITEM_BILL_DISCOUNT_BY_PERCENT {
			amountDiscont := (((body.Quantity - int64(serviceCartItem.Quality)) * serviceCartItem.UnitPrice) * (100 - serviceCartItem.DiscountValue)) / 100

			serviceCart.Amount = serviceCart.Amount + amountDiscont
		} else if serviceCartItem.DiscountType == constants.ITEM_BILL_DISCOUNT_BY_PRICE {
			var amountDiscont int64

			amountRaw := (body.Quantity * serviceCartItem.UnitPrice) - serviceCartItem.DiscountValue

			if amountRaw > 0 {
				serviceCart.Amount = serviceCart.Amount + serviceCartItem.DiscountValue - (int64(serviceCartItem.Quality) * serviceCartItem.UnitPrice)
				amountDiscont = amountRaw
			} else {
				amountDiscont = 0
			}
			serviceCart.Amount += amountDiscont
		} else {
			serviceCart.Amount += (body.Quantity * serviceCartItem.UnitPrice) - (int64(serviceCartItem.Quality) * serviceCartItem.UnitPrice)
		}

		if err := serviceCart.Update(db); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		// update service item
		serviceCartItem.Quality = int(body.Quantity)
		serviceCartItem.Amount = body.Quantity * serviceCartItem.UnitPrice
		// Update amount
		if serviceCartItem.DiscountType == constants.ITEM_BILL_DISCOUNT_BY_PERCENT {
			amountDiscont := (serviceCartItem.Amount * serviceCartItem.DiscountValue) / 100

			serviceCartItem.Amount = serviceCartItem.Amount - amountDiscont
		} else if serviceCartItem.DiscountType == constants.ITEM_BILL_DISCOUNT_BY_PRICE {
			serviceCartItem.Amount = serviceCartItem.Amount - serviceCartItem.DiscountValue
		}
	}

	if body.Note != "" {
		serviceCartItem.Input = body.Note
	}

	if err := serviceCartItem.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	//Update giá nếu bill active
	if serviceCart.BillStatus == constants.POS_BILL_STATUS_ACTIVE {
		//Update lại giá trong booking
		updatePriceWithServiceItem(&booking, prof)
	}

	okRes(c)
}

func (_ CServiceCart) DeleteItemInCart(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	idRequest := c.Param("id")
	id, errId := strconv.ParseInt(idRequest, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	// validate cart item
	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.Id = id
	serviceCartItem.PartnerUid = prof.PartnerUid
	serviceCartItem.CourseUid = prof.CourseUid

	if err := serviceCartItem.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate golf bag
	dateDisplay, _ := utils.GetBookingDateFromTimestamp(time.Now().Unix())

	booking := model_booking.Booking{}
	booking.PartnerUid = serviceCartItem.PartnerUid
	booking.CourseUid = serviceCartItem.CourseUid
	booking.Bag = serviceCartItem.Bag
	booking.BookingDate = dateDisplay
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Booking "+err.Error())
		return
	}

	if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		response_message.BadRequest(c, "Bag status invalid")
		return
	}

	if *booking.LockBill {
		response_message.BadRequestDynamicKey(c, "BAG_BE_LOCK", "Bag lock")
		return
	}

	// validate cart
	serviceCart := models.ServiceCart{}
	serviceCart.Id = serviceCartItem.ServiceBill

	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// if serviceCartItem.Type != constants.RENTAL_SETTING &&
	// 	serviceCartItem.Type != constants.DRIVING_SETTING {
	// 	// validate quantity
	// 	inventory := kiosk_inventory.InventoryItem{}
	// 	inventory.PartnerUid = serviceCartItem.PartnerUid
	// 	inventory.CourseUid = serviceCartItem.CourseUid
	// 	inventory.ServiceId = serviceCart.ServiceId
	// 	inventory.Code = serviceCartItem.ItemCode

	// 	if err := inventory.FindFirst(db); err != nil {
	// 		response_message.BadRequest(c, err.Error())
	// 		return
	// 	}

	// 	// Update số lượng hàng tồn trong kho
	// 	inventory.Quantity += int64(serviceCartItem.Quality)
	// 	if err := inventory.Update(db); err != nil {
	// 		response_message.BadRequest(c, err.Error())
	// 		return
	// 	}
	// }

	// update service cart
	serviceCart.Amount -= serviceCartItem.Amount
	if err := serviceCart.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// Delete Item
	if err := serviceCartItem.Delete(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	if serviceCart.BillStatus == constants.RES_BILL_STATUS_ACTIVE {
		//Update lại giá trong booking
		updatePriceWithServiceItem(&booking, prof)
	}

	go addLog(c, prof, serviceCartItem, constants.OP_LOG_ACTION_DELETE_SERVICE_ITEM)

	okRes(c)
}

func (_ CServiceCart) CreateBill(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CreateBillCodeBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate golf bag
	dateDisplay, _ := utils.GetBookingDateFromTimestamp(time.Now().Unix())

	booking := model_booking.Booking{}
	booking.PartnerUid = body.PartnerUid
	booking.CourseUid = body.CourseUid
	booking.Bag = body.GolfBag
	booking.BookingDate = dateDisplay
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	applyDate := utils.GetDateLocal()
	serviceCart := models.ServiceCart{}
	serviceCart.PartnerUid = body.PartnerUid
	serviceCart.CourseUid = body.CourseUid
	serviceCart.ServiceId = body.ServiceId
	serviceCart.GolfBag = body.GolfBag
	serviceCart.BookingDate = datatypes.Date(applyDate)
	serviceCart.BillCode = constants.BILL_NONE

	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		response_message.BadRequest(c, "Bag status invalid")
		return
	}

	if *booking.LockBill {
		response_message.BadRequestDynamicKey(c, "BAG_BE_LOCK", "Bag lock")
		return
	}

	serviceCart.BillCode = utils.GetTimeNow().Format("20060102150405")

	if err := serviceCart.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	createExportBillInventory(c, prof, serviceCart, serviceCart.BillCode)

	okRes(c)
}

func (_ CServiceCart) MoveItemToOtherCart(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.MoveItemToOtherServiceCartBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate golf bag
	dateDisplay, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())

	booking := model_booking.Booking{}
	booking.Bag = body.GolfBag
	booking.BookingDate = dateDisplay
	booking.AddedRound = setBoolForCursor(false)

	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find booking target "+err.Error())
		return
	}

	if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		response_message.BadRequest(c, "Bag status invalid")
		return
	}

	if *booking.LockBill {
		response_message.BadRequestDynamicKey(c, "BAG_BE_LOCK", "Bag lock")
		return
	}

	// validate cart code
	sourceServiceCart := models.ServiceCart{}
	sourceServiceCart.Id = body.ServiceCartId

	if err := sourceServiceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find bill source "+err.Error())
		return
	}

	if sourceServiceCart.GolfBag == body.GolfBag {
		response_message.BadRequest(c, "Bag transfer is not the same as current bag")
		return
	}

	// validate golf bag source
	bookingS := model_booking.Booking{}
	bookingS.Uid = sourceServiceCart.BookingUid
	bookingSource, errB := bookingS.FindFirstByUId(db)
	if errB != nil {
		response_message.BadRequest(c, "Booking "+errB.Error())
		return
	}

	if bookingS.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		response_message.BadRequest(c, "Bag status invalid")
		return
	}

	if *bookingS.LockBill {
		response_message.BadRequestDynamicKey(c, "BAG_BE_LOCK", "Bag lock")
		return
	}

	// validate cart by golf bag
	applyDate := utils.GetDateLocal()

	targetServiceCart := models.ServiceCart{}
	targetServiceCart.PartnerUid = body.PartnerUid
	targetServiceCart.CourseUid = body.CourseUid
	targetServiceCart.GolfBag = body.GolfBag
	targetServiceCart.BookingDate = datatypes.Date(applyDate)
	targetServiceCart.ServiceId = sourceServiceCart.ServiceId
	targetServiceCart.ServiceType = sourceServiceCart.ServiceType
	targetServiceCart.BillStatus = constants.POS_BILL_STATUS_PENDING

	err := targetServiceCart.FindFirst(db)

	// no cart
	if err != nil {
		// create cart
		targetServiceCart.BookingUid = booking.Uid
		targetServiceCart.StaffOrder = prof.FullName
		targetServiceCart.BillCode = constants.BILL_NONE
		targetServiceCart.BillCode = utils.GetTimeNow().Format("20060102150405")

		if err := targetServiceCart.Create(db); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	hasError := false
	var totalAmount int64 = 0
	var errFor error

	for _, cartItemId := range body.CartItemIdList {
		serviceCartItem := model_booking.BookingServiceItem{}
		serviceCartItem.Id = cartItemId
		serviceCartItem.ServiceBill = sourceServiceCart.Id

		if err := serviceCartItem.FindFirst(db); err != nil {
			continue
		}

		serviceCartItem.ServiceBill = targetServiceCart.Id
		serviceCartItem.Bag = booking.Bag
		serviceCartItem.BillCode = booking.BillCode
		serviceCartItem.BookingUid = booking.Uid
		serviceCartItem.PlayerName = booking.CustomerName
		totalAmount += serviceCartItem.Amount

		if errFor = serviceCartItem.Update(db); errFor != nil {
			hasError = true
			break
		}
	}

	if hasError {
		response_message.InternalServerError(c, errFor.Error())
		return
	}

	// Update amount target bill
	targetServiceCart.Amount += totalAmount
	targetServiceCart.BillStatus = constants.POS_BILL_STATUS_ACTIVE
	if err := targetServiceCart.Update(db); err != nil {
		response_message.InternalServerError(c, "Update target cart "+err.Error())
		return
	}

	if targetServiceCart.BillStatus == constants.POS_BILL_STATUS_ACTIVE {
		updatePriceWithServiceItem(&booking, prof)
	}

	// Update amount target bill
	sourceServiceCart.Amount = sourceServiceCart.Amount - totalAmount

	if sourceServiceCart.BillStatus == constants.POS_BILL_STATUS_ACTIVE {

		//Update lại giá trong booking
		updatePriceWithServiceItem(&bookingSource, prof)
	}

	//
	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.ServiceBill = sourceServiceCart.Id

	list, err := serviceCartItem.FindAll(db)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	if len(list) == 0 {
		sourceServiceCart.BillStatus = constants.POS_BILL_STATUS_TRANSFER
	}

	if err := sourceServiceCart.Update(db); err != nil {
		response_message.InternalServerError(c, "Update target cart "+err.Error())
		return
	}

	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_POS,
		Action:      constants.OP_LOG_ACTION_TRANSFER,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{Data: sourceServiceCart},
		ValueNew:    models.JsonDataLog{Data: targetServiceCart},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}

	if targetServiceCart.ServiceType == constants.KIOSK_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_KIOSK
	} else if targetServiceCart.ServiceType == constants.MINI_B_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_MINI_BAR
	} else if targetServiceCart.ServiceType == constants.PROSHOP_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_PROSHOP
	} else if targetServiceCart.ServiceType == constants.RENTAL_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_GOLF_CLUB_RENTAL
	} else if targetServiceCart.ServiceType == constants.DRIVING_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_DRIVING
	}

	go createOperationLog(opLog)

	okRes(c)
}

func (_ CServiceCart) DeleteCart(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	idRequest := c.Param("id")
	id, errId := strconv.ParseInt(idRequest, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	// validate cart
	serviceCart := models.ServiceCart{}
	serviceCart.Id = id
	serviceCart.PartnerUid = prof.PartnerUid
	serviceCart.CourseUid = prof.CourseUid

	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	//old data
	dataOld := serviceCart

	// validate golf bag
	bookingR := model_booking.Booking{}
	bookingR.Uid = serviceCart.BookingUid
	booking, errF := bookingR.FindFirstByUId(db)
	if errF != nil {
		response_message.InternalServerError(c, "Booking "+errF.Error())
		return
	}

	// Check bag status
	if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		response_message.BadRequest(c, "Bag check out")
		return
	}

	if *booking.LockBill {
		response_message.BadRequestDynamicKey(c, "BAG_BE_LOCK", "Bag lock")
		return
	}

	serviceCart.BillStatus = constants.POS_BILL_STATUS_OUT

	if serviceCart.BillCode == constants.BILL_NONE {
		serviceCart.BillCode = utils.GetTimeNow().Format("20060102150405")
	}

	if serviceCart.RentalStatus == constants.POS_RETAL_STATUS_RENT {
		serviceCart.RentalStatus = constants.POS_RETAL_STATUS_CANCEL
	}

	if err := serviceCart.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	//Update lại giá trong booking
	updatePriceWithServiceItem(&booking, prof)

	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_POS,
		Action:      constants.OP_LOG_ACTION_DELETE_BILL,
		Body:        models.JsonDataLog{Data: idRequest},
		ValueOld:    models.JsonDataLog{Data: dataOld},
		ValueNew:    models.JsonDataLog{Data: serviceCart},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}

	if serviceCart.ServiceType == constants.KIOSK_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_KIOSK
	} else if serviceCart.ServiceType == constants.MINI_B_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_MINI_BAR
	} else if serviceCart.ServiceType == constants.PROSHOP_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_PROSHOP
	} else if serviceCart.ServiceType == constants.RENTAL_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_GOLF_CLUB_RENTAL
	} else if serviceCart.ServiceType == constants.DRIVING_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_DRIVING
	}

	go createOperationLog(opLog)

	okRes(c)
}

func createExportBillInventory(c *gin.Context, prof models.CmsUser, serviceCart models.ServiceCart, code string) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.PartnerUid = serviceCart.PartnerUid
	serviceCartItem.CourseUid = serviceCart.CourseUid
	serviceCartItem.ServiceBill = serviceCart.Id

	listItemInBill, _ := serviceCartItem.FindAll(db)

	if len(listItemInBill) > 0 {
		bodyOutputBill := request.CreateOutputBillBody{}
		bodyOutputBill.PartnerUid = serviceCart.PartnerUid
		bodyOutputBill.CourseUid = serviceCart.CourseUid
		bodyOutputBill.ServiceId = serviceCart.ServiceId

		service := model_service.Kiosk{}
		service.Id = serviceCart.Id
		if err := service.FindFirst(db); err != nil {
			bodyOutputBill.ServiceName = service.KioskName
		}

		bodyOutputBill.UserExport = prof.UserName
		bodyOutputBill.Bag = serviceCart.GolfBag
		bodyOutputBill.CustomerName = serviceCartItem.PlayerName
		lisItem := []request.KioskInventoryItemBody{}

		for _, data := range listItemInBill {
			inputItem := request.KioskInventoryItemBody{}
			inputItem.Quantity = int64(data.Quality)
			inputItem.ItemCode = data.ItemCode
			inputItem.ItemName = data.Name
			inputItem.UserUpdate = prof.UserName
			inputItem.Unit = data.Unit
			inputItem.GroupCode = data.GroupCode
			inputItem.Price = float64(data.UnitPrice)
			lisItem = append(lisItem, inputItem)
		}

		bodyOutputBill.ListItem = lisItem

		cKioskOutputInventory := CKioskOutputInventory{}
		cKioskOutputInventory.MethodOutputBill(c, prof, bodyOutputBill, constants.KIOSK_BILL_INVENTORY_SELL, code, constants.KIOSK_BILL_INVENTORY_APPROVED)
	}
}

func (_ CServiceCart) CreateNewGuest(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CreateNewGuestBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Get key redis
	key := datasources.GetRedisKeyLockerCreateGuestName()

	// check cache
	bagClone, errCache := datasources.GetCache(key)
	if errCache != nil {
		datasources.SetCache(key, "100000", -1)
		bagClone = "100000"
	}

	for {
		// validate golf bag
		dateDisplay, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())

		booking := model_booking.Booking{}
		booking.Bag = bagClone
		booking.BookingDate = dateDisplay
		if err := booking.FindFirst(db); err == nil {
			bag, err := strconv.ParseInt(bagClone, 10, 64)
			if err != nil {
				response_message.BadRequest(c, err.Error())
				return
			}

			bag += 1
			bagClone = strconv.FormatInt(bag, 10)
			datasources.SetCache(key, bagClone, -1)
		} else {
			break
		}
	}

	// Booking Uid
	bookingUid := uuid.New()
	bUid := body.CourseUid + "-" + utils.HashCodeUuid(bookingUid.String())
	billCode := utils.HashCodeUuid(bookingUid.String())

	// Check group Fee
	golfFeeR := models.GolfFee{
		PartnerUid:   body.PartnerUid,
		CourseUid:    body.CourseUid,
		CustomerType: constants.CUSTOMER_TYPE_NONE_GOLF,
	}

	err := golfFeeR.FindFirstWithCusType(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// Create booking
	booking := model_booking.Booking{
		PartnerUid:     body.PartnerUid,
		CourseUid:      body.CourseUid,
		Bag:            bagClone,
		BillCode:       billCode,
		BookingDate:    utils.GetTimeNow().Format("02/01/2006"),
		BagStatus:      constants.BAG_STATUS_WAITING,
		InitType:       constants.BOOKING_INIT_TYPE_CHECKIN,
		CheckInTime:    utils.GetTimeNow().Unix(),
		CustomerName:   body.GuestName,
		CustomerType:   constants.CUSTOMER_TYPE_NONE_GOLF,
		GuestStyle:     golfFeeR.GuestStyle,
		GuestStyleName: golfFeeR.GuestStyleName,
	}

	errC := booking.Create(db, bUid)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_POS,
		Function:    constants.OP_LOG_FUNCTION_BOOKING,
		Action:      constants.OP_LOG_ACTION_CREATE_BAG,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{},
		ValueNew:    models.JsonDataLog{Data: booking},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}

	go createOperationLog(opLog)

	c.JSON(200, booking)
}

// Chốt order
func (_ CServiceCart) FinishOrder(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.FinishOrderBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// validate body
	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate bill
	serviceCart := models.ServiceCart{}
	serviceCart.Id = body.BillId
	serviceCart.PartnerUid = prof.PartnerUid
	serviceCart.CourseUid = prof.CourseUid
	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find service Cart "+err.Error())
		return
	}

	//Kiểm tra trạng thái bill
	if serviceCart.BillStatus == constants.POS_BILL_STATUS_OUT {
		response_message.BadRequest(c, "Bill status invalid")
		return
	}

	// validate golf bag
	bookingR := model_booking.Booking{}
	bookingR.PartnerUid = serviceCart.PartnerUid
	bookingR.CourseUid = serviceCart.CourseUid
	bookingR.Uid = serviceCart.BookingUid
	booking, err := bookingR.FindFirstByUId(db)

	if err != nil {
		response_message.BadRequest(c, "Booking "+err.Error())
		return
	}

	// Check bag status
	if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		response_message.BadRequest(c, "Bag check out")
		return
	}

	if *booking.LockBill {
		response_message.BadRequestDynamicKey(c, "BAG_BE_LOCK", "Bag lock")
		return
	}

	// Update trạng thái
	serviceCart.BillStatus = constants.POS_BILL_STATUS_ACTIVE
	if err := serviceCart.Update(db); err != nil {
		response_message.BadRequest(c, "Update service Cart "+err.Error())
		return
	}

	//Update lại giá trong booking
	updatePriceWithServiceItem(&booking, prof)

	okRes(c)
}

// Chuyển trạng thái
func (_ CServiceCart) UndoStatus(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.FinishOrderBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// validate body
	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate bill
	serviceCart := models.ServiceCart{}
	serviceCart.Id = body.BillId
	serviceCart.PartnerUid = prof.PartnerUid
	serviceCart.CourseUid = prof.CourseUid
	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find service Cart "+err.Error())
		return
	}

	// validate golf bag
	bookingR := model_booking.Booking{}
	bookingR.Uid = serviceCart.BookingUid
	booking, errF := bookingR.FindFirstByUId(db)
	if errF != nil {
		response_message.InternalServerError(c, "Booking "+errF.Error())
		return
	}

	//old data
	dataOld := serviceCart

	// Update trạng thái
	serviceCart.BillStatus = constants.POS_BILL_STATUS_PENDING
	if err := serviceCart.Update(db); err != nil {
		response_message.BadRequest(c, "Update service Cart "+err.Error())
		return
	}

	//Update lại giá trong booking
	updatePriceWithServiceItem(&booking, prof)

	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_POS,
		Action:      constants.OP_LOG_ACTION_UNDO_BILL,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{Data: dataOld},
		ValueNew:    models.JsonDataLog{Data: serviceCart},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}

	if serviceCart.ServiceType == constants.KIOSK_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_KIOSK
	} else if serviceCart.ServiceType == constants.MINI_B_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_MINI_BAR
	} else if serviceCart.ServiceType == constants.PROSHOP_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_PROSHOP
	} else if serviceCart.ServiceType == constants.RENTAL_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_GOLF_CLUB_RENTAL
	} else if serviceCart.ServiceType == constants.DRIVING_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_DRIVING
	}

	go createOperationLog(opLog)

	okRes(c)
}

// Chuyển trạng thái thuê đô
func (_ CServiceCart) ChangeRentalStatus(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.FinishOrderBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// validate body
	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate bill
	serviceCart := models.ServiceCart{}
	serviceCart.Id = body.BillId
	serviceCart.PartnerUid = prof.PartnerUid
	serviceCart.CourseUid = prof.CourseUid
	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find service Cart "+err.Error())
		return
	}

	//old data
	dataOld := serviceCart

	// validate golf bag
	bookingR := model_booking.Booking{}
	bookingR.Uid = serviceCart.BookingUid
	booking, errF := bookingR.FindFirstByUId(db)
	if errF != nil {
		response_message.InternalServerError(c, "Booking "+errF.Error())
		return
	}

	// Update trạng thái
	serviceCart.RentalStatus = constants.POS_RETAL_STATUS_RETURN
	if err := serviceCart.Update(db); err != nil {
		response_message.BadRequest(c, "Update service Cart "+err.Error())
		return
	}

	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_POS,
		Action:      constants.OP_LOG_ACTION_RETURN,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{Data: dataOld},
		ValueNew:    models.JsonDataLog{Data: serviceCart},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}

	if serviceCart.ServiceType == constants.RENTAL_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_GOLF_CLUB_RENTAL
	} else if serviceCart.ServiceType == constants.DRIVING_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_DRIVING
	}

	okRes(c)
}

func addLog(c *gin.Context, prof models.CmsUser, serviceCartItem model_booking.BookingServiceItem, action string) {
	opLog := models.OperationLog{
		PartnerUid:  serviceCartItem.PartnerUid,
		CourseUid:   serviceCartItem.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_POS,
		Action:      action,
		Body:        models.JsonDataLog{Data: serviceCartItem},
		ValueOld:    models.JsonDataLog{},
		ValueNew:    models.JsonDataLog{Data: serviceCartItem},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         serviceCartItem.Bag,
		BookingDate: utils.GetCurrentDay1(),
		BillCode:    serviceCartItem.BillCode,
		BookingUid:  serviceCartItem.BookingUid,
	}

	if serviceCartItem.Type == constants.RENTAL_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_GOLF_CLUB_RENTAL
	}

	if serviceCartItem.Type == constants.DRIVING_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_DRIVING
	}

	if serviceCartItem.Type == constants.PROSHOP_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_PROSHOP
	}

	if serviceCartItem.Type == constants.KIOSK_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_DRIVING
	}

	if serviceCartItem.Type == constants.MINI_B_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_MINIBAR
	}

	createOperationLog(opLog)
}

func (_ CServiceCart) UpdateBillAmout(itemId int64, parnerId, courseId string) {
	db := datasources.GetDatabaseWithPartner(parnerId)
	// validate cart item
	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.Id = itemId
	serviceCartItem.PartnerUid = parnerId
	serviceCartItem.CourseUid = courseId

	if err := serviceCartItem.FindFirst(db); err != nil {
		return
	}

	// validate cart
	serviceCart := models.ServiceCart{}
	serviceCart.Id = serviceCartItem.ServiceBill

	if err := serviceCart.FindFirst(db); err != nil {
		return
	}

	// update service cart
	serviceCart.Amount -= serviceCartItem.Amount
	if err := serviceCart.Update(db); err != nil {
		return
	}
}
