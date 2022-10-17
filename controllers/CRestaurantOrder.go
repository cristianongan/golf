package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	model_service "start/models/service"
	"start/utils/response_message"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/datatypes"
)

type CRestaurantOrder struct{}

// Tạo hóa đơn cho nhà hàng
func (_ CRestaurantOrder) CreateRestaurantOrder(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CreateRestaurantOrderBody{}
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

	// validate golf bag
	booking := model_booking.Booking{}
	booking.PartnerUid = body.PartnerUid
	booking.CourseUid = body.CourseUid
	booking.Bag = body.GolfBag
	booking.BookingDate = time.Now().Format("02/01/2006")
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find booking "+err.Error())
		return
	}

	if booking.BagStatus != constants.BAG_STATUS_WAITING && booking.BagStatus != constants.BAG_STATUS_IN_COURSE && booking.BagStatus != constants.BAG_STATUS_TIMEOUT {
		response_message.BadRequest(c, "Bag status invalid")
		return
	}

	// validate kiosk
	kiosk := model_service.Kiosk{}
	kiosk.Id = body.ServiceId
	if err := kiosk.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find kiosk "+err.Error())
		return
	}

	// Tạo đơn order
	serviceCart := models.ServiceCart{}

	serviceCart.Type = body.Type
	serviceCart.TypeCode = body.TypeCode

	if body.Type == constants.RES_TYPE_TABLE {
		serviceCart.NumberGuest = body.NumberGuest
		serviceCart.ResFloor = body.Floor
	}

	serviceCart.PartnerUid = body.PartnerUid
	serviceCart.CourseUid = body.CourseUid
	serviceCart.GolfBag = body.GolfBag
	serviceCart.BookingUid = booking.Uid
	serviceCart.BookingDate = datatypes.Date(time.Now().UTC())
	serviceCart.ServiceId = body.ServiceId
	serviceCart.ServiceType = kiosk.KioskType
	serviceCart.BillCode = constants.BILL_NONE
	serviceCart.BillStatus = constants.RES_STATUS_ORDER
	serviceCart.StaffOrder = prof.FullName
	serviceCart.PlayerName = booking.CustomerName

	if err := serviceCart.Create(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	c.JSON(200, serviceCart)
}

// Tạo mã đơn
func (_ CRestaurantOrder) CreateBill(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CreateBillOrderBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	serviceCart := models.ServiceCart{}
	serviceCart.Id = body.BillId

	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	if serviceCart.BillCode == constants.BILL_NONE {
		serviceCart.BillCode = "OD-" + strconv.Itoa(int(body.BillId))
		serviceCart.TimeProcess = time.Now().Unix()
		serviceCart.BillStatus = constants.RES_STATUS_PROCESS

		if err := serviceCart.Update(db); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
	}

	//find all item in bill
	restaurantItem := models.RestaurantItem{}
	restaurantItem.BillId = body.BillId
	restaurantItem.ItemStatus = constants.RES_STATUS_ORDER

	list, err := restaurantItem.FindAll(db)

	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	for _, item := range list {
		item.ItemStatus = constants.RES_STATUS_PROCESS

		if err := item.Update(db); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
	}

	// createExportBillInventory(c, prof, serviceCart, serviceCart.BillCode)

	c.JSON(200, serviceCart)
}

// Hủy đơn
func (_ CRestaurantOrder) DeleteRestaurantOrder(c *gin.Context, prof models.CmsUser) {
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

	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate golf bag
	booking := model_booking.Booking{}
	booking.Uid = serviceCart.BookingUid

	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Booking "+err.Error())
		return
	}

	serviceCart.BillStatus = constants.RES_BILL_STATUS_CANCEL

	if err := serviceCart.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	//find all item in bill
	restaurantItem := models.RestaurantItem{}
	restaurantItem.BillId = serviceCart.Id

	list, err := restaurantItem.FindAll(db)

	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	for _, item := range list {
		if err := item.Delete(db); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
	}

	okRes(c)
}

func (_ CRestaurantOrder) GetListBill(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	query := request.GetListBillBody{}
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
	serviceCart.BillStatus = query.BillStatus
	serviceCart.BillCode = query.BillCode
	serviceCart.TypeCode = query.Table
	serviceCart.Type = query.Type
	serviceCart.ResFloor = query.Floor
	serviceCart.PlayerName = query.CustomerName
	serviceCart.GolfBag = query.GolfBag

	list, total, err := serviceCart.FindList(db, page)

	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	listData := make([]map[string]interface{}, len(list))

	for i, data := range list {
		//find all item in bill
		serviceCartItem := model_booking.BookingServiceItem{}
		serviceCartItem.ServiceBill = data.Id

		listItem, err := serviceCartItem.FindAll(db)

		if err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}

		//find all res item in bill
		restaurantItem := models.RestaurantItem{}
		restaurantItem.BillId = data.Id

		listResItem, err := restaurantItem.FindAll(db)

		if err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}

		// Add infor to response
		listData[i] = map[string]interface{}{
			"bill_infor": data,
			"list_item":  listItem,
			"menu":       listResItem,
		}
	}

	res := response.PageResponse{
		Total: total,
		Data:  listData,
	}

	c.JSON(200, res)
}

// Thêm sản phẩm vào hóa đơn
func (_ CRestaurantOrder) AddItemOrder(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.AddItemOrderBody{}
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

	// validate restaurant order
	serviceCart := models.ServiceCart{}
	serviceCart.Id = body.BillId
	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find service Cart "+err.Error())
		return
	}

	// validate golf bag
	booking := model_booking.Booking{}
	booking.PartnerUid = body.PartnerUid
	booking.CourseUid = body.CourseUid
	booking.Uid = serviceCart.BookingUid
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Booking "+err.Error())
		return
	}

	// validate kiosk
	kiosk := model_service.Kiosk{}
	kiosk.Id = serviceCart.ServiceId
	if err := kiosk.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Kiosk "+err.Error())
		return
	}

	// create cart item
	serviceCartItem := model_booking.BookingServiceItem{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		Bag:         serviceCart.GolfBag,
		BookingUid:  serviceCart.BookingUid,
		BillCode:    booking.BillCode,
		PlayerName:  serviceCart.PlayerName,
		ServiceId:   strconv.Itoa(int(serviceCart.ServiceId)),
		ServiceBill: body.BillId,
		ItemCode:    body.ItemCode,
		Quality:     body.Quantity,
		UserAction:  prof.UserName,
	}

	// add res item with combo
	restaurantItems := []models.RestaurantItem{}

	// validate item code by group
	if body.Type == "COMBO" {
		fbSet := model_service.FbPromotionSet{}
		fbSet.PartnerUid = body.PartnerUid
		fbSet.CourseUid = body.CourseUid
		fbSet.Code = body.ItemCode

		if err := fbSet.FindFirst(db); err != nil {
			response_message.BadRequest(c, "Find fb set "+err.Error())
			return
		}

		// add infor cart item
		serviceCartItem.Type = constants.RESTAURANT_SETTING
		serviceCartItem.Location = kiosk.KioskName
		serviceCartItem.Name = fbSet.VieName
		serviceCartItem.UnitPrice = int64(fbSet.Price)
		serviceCartItem.Amount = int64(body.Quantity) * int64(fbSet.Price)

		// add item res
		for _, v := range fbSet.FBList {
			item := models.RestaurantItem{
				Type:             v.Type,
				ItemName:         v.VieName,
				ItemComboName:    fbSet.VieName,
				ItemComboCode:    body.ItemCode,
				ItemCode:         v.FBCode,
				ItemUnit:         v.Unit,
				Quantity:         v.Quantity * body.Quantity,
				QuantityProgress: v.Quantity * body.Quantity,
			}

			restaurantItems = append(restaurantItems, item)
		}
	}

	if body.Type == "NORMAL" {
		fb := model_service.FoodBeverage{}
		fb.PartnerUid = body.PartnerUid
		fb.CourseUid = body.CourseUid
		fb.FBCode = body.ItemCode

		if err := fb.FindFirst(db); err != nil {
			response_message.BadRequest(c, "Find fb "+err.Error())
			return
		}

		// add infor cart item
		serviceCartItem.Type = constants.RESTAURANT_SETTING
		serviceCartItem.Location = kiosk.KioskName
		serviceCartItem.GroupCode = fb.GroupCode
		serviceCartItem.Name = fb.VieName
		serviceCartItem.EngName = fb.EnglishName
		serviceCartItem.UnitPrice = int64(fb.Price)
		serviceCartItem.Unit = fb.Unit
		serviceCartItem.Amount = int64(body.Quantity) * int64(fb.Price)

		// add infor res item
		item := models.RestaurantItem{
			Type:             fb.Type,
			ItemName:         fb.VieName,
			ItemCode:         fb.FBCode,
			ItemUnit:         fb.Unit,
			Quantity:         body.Quantity,
			QuantityProgress: body.Quantity,
		}

		restaurantItems = append(restaurantItems, item)
	}

	// create cart item
	if err := serviceCartItem.Create(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// update service cart
	serviceCart.Amount += (int64(body.Quantity) * serviceCartItem.UnitPrice)
	if err := serviceCart.Update(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	for _, v := range restaurantItems {
		// add infor restaurant item
		v.PartnerUid = body.PartnerUid
		v.CourseUid = body.CourseUid
		v.ServiceId = serviceCart.ServiceId
		v.OrderDate = time.Now().Format("02/01/2006")
		v.BillId = serviceCart.Id
		v.ItemId = serviceCartItem.Id
		v.ItemStatus = constants.RES_STATUS_ORDER

		if err := v.Create(db); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
	}

	okRes(c)
}

// Update sản phẩm
func (_ CRestaurantOrder) UpdateItemOrder(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.UpdateItemOrderBody{}
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

	// validate cart item
	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.Id = body.ItemId

	if err := serviceCartItem.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find item"+err.Error())
		return
	}

	// validate restaurant order
	serviceCart := models.ServiceCart{}
	serviceCart.Id = serviceCartItem.ServiceBill

	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find order"+err.Error())
		return
	}

	if serviceCart.BillStatus == constants.RES_BILL_STATUS_OUT ||
		serviceCart.BillStatus == constants.RES_BILL_STATUS_CANCEL {

		response_message.BadRequest(c, "Bill status invalid")
		return
	}

	// validate golf bag
	booking := model_booking.Booking{}
	booking.Uid = serviceCart.BookingUid
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Booking "+err.Error())
		return
	}

	if body.Quantity > 0 {
		// validate res item
		restaurantItem := models.RestaurantItem{}

		restaurantItem.PartnerUid = body.PartnerUid
		restaurantItem.CourseUid = body.CourseUid
		restaurantItem.ServiceId = serviceCart.ServiceId
		restaurantItem.BillId = serviceCart.Id
		restaurantItem.ItemId = serviceCartItem.Id

		list, err := restaurantItem.FindAll(db)

		if err != nil {
			response_message.BadRequest(c, "Find res item"+err.Error())
			return
		}

		// update service cart
		serviceCart.Amount += (int64(body.Quantity) * serviceCartItem.UnitPrice) - (int64(serviceCartItem.Quality) * serviceCartItem.UnitPrice)
		if err := serviceCart.Update(db); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}

		// update res item
		for _, v := range list {
			if v.ItemComboCode != "" {
				v.Quantity = (v.Quantity / serviceCartItem.Quality) * body.Quantity
				v.QuantityProgress = (v.Quantity / serviceCartItem.Quality) * body.Quantity
			} else {
				v.Quantity = body.Quantity
				v.QuantityProgress = body.Quantity
			}

			if body.Note != "" {
				v.ItemNote = body.Note
			}

			if err := v.Update(db); err != nil {
				response_message.BadRequest(c, err.Error())
				return
			}
		}

		// update service item
		serviceCartItem.Quality = int(body.Quantity)
		serviceCartItem.Amount = int64(body.Quantity) * serviceCartItem.UnitPrice
	}

	if body.Note != "" {
		serviceCartItem.Input = body.Note
	}

	if err := serviceCartItem.Update(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}

// Delete sản phẩm
func (_ CRestaurantOrder) DeleteItemOrder(c *gin.Context, prof models.CmsUser) {
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

	if err := serviceCartItem.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find item"+err.Error())
		return
	}

	// validate res order
	serviceCart := models.ServiceCart{}
	serviceCart.Id = serviceCartItem.ServiceBill

	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find res order"+err.Error())
		return
	}

	// validate golf bag
	booking := model_booking.Booking{}
	booking.Uid = serviceCart.BookingUid
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Booking "+err.Error())
		return
	}

	// validate res item
	restaurantItem := models.RestaurantItem{}
	restaurantItem.BillId = serviceCart.Id
	restaurantItem.ItemId = serviceCartItem.Id

	resList, err := restaurantItem.FindAll(db)
	if err != nil {
		response_message.BadRequest(c, "Find res item"+err.Error())
		return
	}

	// update service cart
	serviceCart.Amount -= serviceCartItem.Amount
	if err := serviceCart.Update(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Delete Item
	if err := serviceCartItem.Delete(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	if serviceCart.BillStatus == constants.RES_BILL_STATUS_PROCESS {
		// Update res item
		for _, item := range resList {
			item.ItemStatus = constants.RES_STATUS_CANCEL

			if err := item.Update(db); err != nil {
				response_message.BadRequest(c, err.Error())
				return
			}
		}
	} else {
		// Delete res item
		for _, item := range resList {

			if err := item.Delete(db); err != nil {
				response_message.BadRequest(c, err.Error())
				return
			}
		}

	}

	okRes(c)
}

// get list sản phẩm
func (_ CRestaurantOrder) GetListItemOrder(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	query := request.GetItemResOrderBody{}
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
	serviceCartItem.ServiceBill = query.BillId

	list, total, err := serviceCartItem.FindList(db, page)

	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	res := response.PageResponse{
		Total: total,
		Data:  list,
	}

	c.JSON(200, res)
}

// Update res item
func (_ CRestaurantOrder) UpdateResItem(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.UpdateResItemBody{}
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

	// validate restaurant item
	resItem := models.RestaurantItem{}
	resItem.Id = body.ItemId

	if err := resItem.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	//validate bill
	bill := models.ServiceCart{}

	bill.Id = resItem.BillId

	if err := bill.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Update trạng thái khi trả hết món
	if resItem.QuantityProgress-1 == 0 {
		resItem.ItemStatus = constants.RES_STATUS_DONE

		// Kiểm tra trạng thái các món
		restaurantItem := models.RestaurantItem{
			PartnerUid: resItem.PartnerUid,
			CourseUid:  resItem.CourseUid,
			ServiceId:  resItem.ServiceId,
			BillId:     resItem.BillId,
			ItemStatus: constants.RES_STATUS_PROCESS,
		}

		list, errRI := restaurantItem.FindAll(db)

		if errRI != nil {
			response_message.BadRequest(c, errRI.Error())
			return
		}

		if len(list) == 0 {
			bill.BillStatus = constants.RES_BILL_STATUS_FINISH

			if errBU := bill.Update(db); errBU != nil {
				response_message.BadRequest(c, errBU.Error())
				return
			}
		}

	}

	// Update quantity progress when finish
	resItem.QuantityProgress -= 1

	// update res item
	if err := resItem.Update(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}

// get list sản phẩm cho nhà bếp
func (_ CRestaurantOrder) GetFoodProcess(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.GetFoodProcessBody{}
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

	// validate restaurant item
	resItem := models.RestaurantItem{}
	resItem.ServiceId = body.ServiceId
	resItem.Type = body.Type
	resItem.ItemName = body.Name

	list, err := resItem.FindAllGroupBy(db)

	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	c.JSON(200, list)
}

// get list theo sản phẩm
func (_ CRestaurantOrder) GetDetailFoodProcess(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.GetDetailFoodProcessBody{}
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

	// validate restaurant item
	resItem := models.RestaurantItem{}
	resItem.ServiceId = body.ServiceId
	resItem.ItemCode = body.ItemCode
	resItem.ItemStatus = constants.RES_STATUS_PROCESS

	list, err := resItem.FindAll(db)

	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	c.JSON(200, list)
}

// Action hoàn thành all
func (_ CRestaurantOrder) FinishAllResItem(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.FinishAllResItemBody{}
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

	// validate restaurant item
	resItem := models.RestaurantItem{}
	resItem.ServiceId = body.ServiceId
	resItem.BillId = body.BillId
	resItem.ItemCode = body.ItemCode

	list, err := resItem.FindAll(db)

	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	for _, v := range list {
		v.ItemStatus = constants.RES_STATUS_DONE
		v.QuantityProgress = 0

		if err := v.Update(db); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
	}

	if body.BillId != 0 {
		// validate res order
		serviceCart := models.ServiceCart{}
		serviceCart.Id = body.BillId

		if err := serviceCart.FindFirst(db); err != nil {
			response_message.BadRequest(c, "Find res order"+err.Error())
			return
		}

		serviceCart.BillStatus = constants.RES_BILL_STATUS_FINISH
		if err := serviceCart.Update(db); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
	}

	okRes(c)
}

// Tạo booking cho nhà hàng
func (_ CRestaurantOrder) CreateRestaurantBooking(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CreateBookingRestaurantBody{}
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

	// Tạo đơn order
	serviceCart := models.ServiceCart{}
	booking := model_booking.Booking{}

	if body.GolfBag != "" {
		// validate golf bag
		booking.PartnerUid = body.PartnerUid
		booking.CourseUid = body.CourseUid
		booking.Bag = body.GolfBag
		booking.BookingDate = time.Now().Format("02/01/2006")
		if err := booking.FindFirst(db); err != nil {
			response_message.BadRequest(c, "Find booking "+err.Error())
			return
		}

		if booking.BagStatus != constants.BAG_STATUS_WAITING && booking.BagStatus != constants.BAG_STATUS_IN_COURSE && booking.BagStatus != constants.BAG_STATUS_TIMEOUT {
			response_message.BadRequest(c, "Bag status invalid")
			return
		}

		// add infor service cart
		serviceCart.GolfBag = body.GolfBag
		serviceCart.BookingUid = booking.Uid
	}

	// validate kiosk
	kiosk := model_service.Kiosk{}
	kiosk.Id = body.ServiceId
	if err := kiosk.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find kiosk "+err.Error())
		return
	}

	// validate from kiosk
	fromKiosk := model_service.Kiosk{}
	fromKiosk.Id = body.FromServiceId
	if err := fromKiosk.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find from kiosk "+err.Error())
		return
	}

	// create cart item
	itemCombos := []model_service.FbPromotionSet{}
	itemQuatityCombos := []int{}
	itemFBs := []model_service.FoodBeverage{}
	itemQuatityFBs := []int{}
	restaurantItems := []models.RestaurantItem{}

	// add item
	for _, item := range body.ListOrderItem {
		// validate item code by group
		if item.Type == "COMBO" {
			fbSet := model_service.FbPromotionSet{}
			fbSet.PartnerUid = body.PartnerUid
			fbSet.CourseUid = body.CourseUid
			fbSet.Code = item.ItemCode

			if err := fbSet.FindFirst(db); err != nil {
				response_message.BadRequest(c, "Find fb set "+err.Error())
				return
			}

			itemCombos = append(itemCombos, fbSet)
			itemQuatityCombos = append(itemQuatityCombos, item.Quantity)

		} else {
			fb := model_service.FoodBeverage{}
			fb.PartnerUid = body.PartnerUid
			fb.CourseUid = body.CourseUid
			fb.FBCode = item.ItemCode

			if err := fb.FindFirst(db); err != nil {
				response_message.BadRequest(c, "Find fb "+err.Error())
				return
			}

			itemFBs = append(itemFBs, fb)
			itemQuatityFBs = append(itemQuatityFBs, item.Quantity)
		}
	}

	// create service cart
	serviceCart.PartnerUid = body.PartnerUid
	serviceCart.CourseUid = body.CourseUid
	serviceCart.BookingDate = datatypes.Date(time.Now().UTC())
	serviceCart.ServiceId = body.ServiceId
	serviceCart.ServiceType = kiosk.KioskType
	serviceCart.BillCode = constants.BILL_NONE
	serviceCart.BillStatus = constants.RES_BILL_STATUS_BOOKING
	serviceCart.Type = constants.RES_TYPE_TABLE
	serviceCart.TypeCode = body.Table
	serviceCart.NumberGuest = body.NumberGuest
	serviceCart.ResFloor = body.Floor
	serviceCart.StaffOrder = prof.FullName
	serviceCart.PlayerName = body.PlayerName
	serviceCart.Phone = body.Phone
	serviceCart.OrderTime = body.OrderTime
	serviceCart.Note = body.Note

	if err := serviceCart.Create(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	if len(itemCombos) > 0 {
		for i, item := range itemCombos {
			quantity := itemQuatityCombos[i]
			// add infor cart item
			serviceCartItem := model_booking.BookingServiceItem{
				PartnerUid:  body.PartnerUid,
				CourseUid:   body.CourseUid,
				ServiceBill: serviceCart.Id,
				Type:        kiosk.KioskType,
				Location:    kiosk.KioskName,
				Name:        item.VieName,
				UserAction:  prof.UserName,
				PlayerName:  body.PlayerName,
			}

			serviceCartItem.UnitPrice = int64(item.Price)
			serviceCartItem.ServiceId = strconv.Itoa(int(serviceCart.ServiceId))
			serviceCartItem.ItemCode = item.Code
			serviceCartItem.Quality = quantity
			serviceCartItem.Amount = int64(quantity) * serviceCartItem.UnitPrice

			if body.GolfBag != "" {
				serviceCartItem.Bag = booking.Bag
				serviceCartItem.BookingUid = booking.Uid
				serviceCartItem.BillCode = booking.BillCode
				serviceCartItem.PlayerName = booking.CustomerName
			}

			if err := serviceCartItem.Create(db); err != nil {
				response_message.BadRequest(c, err.Error())
				return
			}

			// update amount service cart
			serviceCart.Amount += (int64(quantity) * serviceCartItem.UnitPrice)

			// add item res
			for _, v := range item.FBList {
				item := models.RestaurantItem{
					Type:             kiosk.KioskType,
					BillId:           serviceCart.Id,
					ItemId:           serviceCartItem.Id,
					ItemName:         v.VieName,
					ItemComboName:    item.VieName,
					ItemComboCode:    item.Code,
					ItemCode:         v.FBCode,
					ItemUnit:         v.Unit,
					Quantity:         v.Quantity * quantity,
					QuantityProgress: v.Quantity * quantity,
				}

				restaurantItems = append(restaurantItems, item)
			}
		}
	}

	if len(itemFBs) > 0 {
		for i, item := range itemFBs {
			quantity := itemQuatityFBs[i]
			// add infor cart item
			serviceCartItem := model_booking.BookingServiceItem{
				PartnerUid:  body.PartnerUid,
				CourseUid:   body.CourseUid,
				ServiceBill: serviceCart.Id,
				Type:        kiosk.KioskType,
				Location:    kiosk.KioskName,
				Name:        item.VieName,
				UserAction:  prof.UserName,
				PlayerName:  body.PlayerName,
			}

			serviceCartItem.UnitPrice = int64(item.Price)
			serviceCartItem.ServiceId = strconv.Itoa(int(serviceCart.ServiceId))
			serviceCartItem.ItemCode = item.FBCode
			serviceCartItem.Quality = quantity
			serviceCartItem.Amount = int64(quantity) * serviceCartItem.UnitPrice

			if body.GolfBag != "" {
				serviceCartItem.Bag = booking.Bag
				serviceCartItem.BookingUid = booking.Uid
				serviceCartItem.BillCode = booking.BillCode
				serviceCartItem.PlayerName = booking.CustomerName
			}

			if err := serviceCartItem.Create(db); err != nil {
				response_message.BadRequest(c, err.Error())
				return
			}

			// update amount service cart
			serviceCart.Amount += (int64(quantity) * serviceCartItem.UnitPrice)

			// add item res
			item := models.RestaurantItem{
				Type:             kiosk.KioskType,
				BillId:           serviceCart.Id,
				ItemId:           serviceCartItem.Id,
				ItemName:         item.VieName,
				ItemCode:         item.FBCode,
				ItemUnit:         item.Unit,
				Quantity:         quantity,
				QuantityProgress: quantity,
			}

			restaurantItems = append(restaurantItems, item)
		}
	}

	for _, v := range restaurantItems {
		// add infor restaurant item
		v.PartnerUid = body.PartnerUid
		v.CourseUid = body.CourseUid
		v.ServiceId = serviceCart.ServiceId
		v.OrderDate = time.Now().Format("02/01/2006")
		v.ItemStatus = constants.RES_STATUS_ORDER
		// v.Quantity = item.Quantity
		// v.QuantityProgress = item.Quantity

		// Đổi trạng thái món khi đã có bill code
		if serviceCart.BillCode != "NONE" {
			v.ItemStatus = constants.RES_STATUS_PROCESS
		}

		if err := v.Create(db); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
	}

	// Update
	if err := serviceCart.Update(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}

// Chốt order
func (_ CRestaurantOrder) FinishRestaurantOrder(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.FinishRestaurantOrderBody{}
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

	// validate restaurant order
	serviceCart := models.ServiceCart{}
	serviceCart.Id = body.BillId
	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find service Cart "+err.Error())
		return
	}

	// Update trạng thái
	serviceCart.BillStatus = constants.RES_BILL_STATUS_OUT
	if err := serviceCart.Update(db); err != nil {
		response_message.BadRequest(c, "Update service Cart "+err.Error())
		return
	}

	// validate golf bag
	booking := model_booking.Booking{}
	booking.PartnerUid = serviceCart.PartnerUid
	booking.CourseUid = serviceCart.CourseUid
	booking.Uid = serviceCart.BookingUid
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Booking "+err.Error())
		return
	}

	//Update lại giá trong booking
	updatePriceWithServiceItem(booking, prof)

	okRes(c)
}

// Update thông tin restaurant booking

func (_ CRestaurantOrder) UpdateRestaurantBooking(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	billIdStr := c.Param("id")
	billId, err := strconv.ParseInt(billIdStr, 10, 64)
	if err != nil || billId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	body := request.UpdateBookingRestaurantBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	serviceCart := models.ServiceCart{}
	serviceCart.Id = billId
	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find bill "+err.Error())
		return
	}

	//
	if body.NumberGuest != 0 {
		serviceCart.NumberGuest = body.NumberGuest
	}

	if body.Table != "" {
		serviceCart.TypeCode = body.Table
	}

	if body.PlayerName != "" {
		serviceCart.PlayerName = body.PlayerName
	}

	if body.Phone != "" {
		serviceCart.Phone = body.Phone
	}

	if body.GolfBag != "" {
		// validate golf bag
		booking := model_booking.Booking{}
		booking.PartnerUid = body.PartnerUid
		booking.CourseUid = body.CourseUid
		booking.Bag = body.GolfBag
		booking.BookingDate = time.Now().Format("02/01/2006")
		if err := booking.FindFirst(db); err != nil {
			response_message.BadRequest(c, "Find booking "+err.Error())
			return
		}

		if booking.BagStatus != constants.BAG_STATUS_WAITING && booking.BagStatus != constants.BAG_STATUS_IN_COURSE && booking.BagStatus != constants.BAG_STATUS_TIMEOUT {
			response_message.BadRequest(c, "Bag status invalid")
			return
		}

		// add infor service cart
		serviceCart.GolfBag = body.GolfBag
		serviceCart.BookingUid = booking.Uid
	}

	if err := serviceCart.Update(db); err != nil {
		response_message.BadRequest(c, "Update bill "+err.Error())
		return
	}

	okRes(c)
}
