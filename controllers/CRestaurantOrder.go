package controllers

import (
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
		item.ItemStatus = constants.RES_STATUS_CANCEL

		if err := item.Update(db); err != nil {
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

	list, total, err := serviceCart.FindList(db, page)

	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	listData := make([]map[string]interface{}, len(list))

	for i, data := range list {
		//find all item in bill
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
	booking.Bag = serviceCart.GolfBag
	booking.BookingDate = time.Now().Format("02/01/2006")
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find booking "+err.Error())
		return
	}

	// validate kiosk
	kiosk := model_service.Kiosk{}
	kiosk.Id = serviceCart.ServiceId
	if err := kiosk.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find kiosk "+err.Error())
		return
	}

	// create cart item
	serviceCartItem := model_booking.BookingServiceItem{}

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
		serviceCartItem.Type = kiosk.KioskType
		serviceCartItem.GroupCode = fbSet.GroupCode
		serviceCartItem.Name = fbSet.SetName
		serviceCartItem.UnitPrice = int64(fbSet.Price)

		// add item res
		for _, v := range fbSet.FBList {
			fb := model_service.FoodBeverage{}
			fb.PartnerUid = body.PartnerUid
			fb.CourseUid = body.CourseUid
			fb.FBCode = v

			if err := fb.FindFirst(db); err != nil {
				response_message.BadRequest(c, "Find fb in combo "+err.Error())
				return
			}

			item := models.RestaurantItem{
				Type:          fb.Type,
				ItemName:      fb.VieName,
				ItemComboName: fbSet.SetName,
				ItemCode:      fb.FBCode,
				ItemUnit:      fb.Unit,
			}

			restaurantItems = append(restaurantItems, item)
		}
	} else {
		fb := model_service.FoodBeverage{}
		fb.PartnerUid = body.PartnerUid
		fb.CourseUid = body.CourseUid
		fb.FBCode = body.ItemCode

		if err := fb.FindFirst(db); err != nil {
			response_message.BadRequest(c, "Find fb "+err.Error())
			return
		}

		// add infor cart item
		serviceCartItem.Type = kiosk.KioskType
		serviceCartItem.GroupCode = fb.GroupCode
		serviceCartItem.Name = fb.VieName
		serviceCartItem.EngName = fb.EnglishName
		serviceCartItem.UnitPrice = int64(fb.Price)
		serviceCartItem.Unit = fb.Unit

		// add infor res item
		item := models.RestaurantItem{
			Type:     fb.Type,
			ItemName: fb.VieName,
			ItemCode: fb.FBCode,
			ItemUnit: fb.Unit,
		}

		restaurantItems = append(restaurantItems, item)
	}

	// update service cart
	serviceCart.Amount += (int64(body.Quantity) * serviceCartItem.UnitPrice)
	if err := serviceCart.Update(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// add infor cart item
	serviceCartItem.PartnerUid = body.PartnerUid
	serviceCartItem.CourseUid = body.CourseUid
	serviceCartItem.Bag = booking.Bag
	serviceCartItem.BillCode = booking.BillCode
	serviceCartItem.BookingUid = booking.Uid
	serviceCartItem.PlayerName = booking.CustomerName
	serviceCartItem.ServiceId = strconv.Itoa(int(serviceCart.ServiceId))
	serviceCartItem.ServiceBill = body.BillId
	serviceCartItem.ItemCode = body.ItemCode
	serviceCartItem.Quality = body.Quantity
	serviceCartItem.Amount = int64(body.Quantity) * serviceCartItem.UnitPrice
	serviceCartItem.UserAction = prof.UserName

	if err := serviceCartItem.Create(db); err != nil {
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
		v.Quantity = body.Quantity
		v.QuantityProgress = body.Quantity

		// Đổi trạng thái món khi đã có bill code
		if serviceCart.BillCode != "NONE" {
			v.ItemStatus = constants.RES_STATUS_PROCESS
		}

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

	// validate res item
	restaurantItem := models.RestaurantItem{}
	restaurantItem.PartnerUid = body.PartnerUid
	restaurantItem.CourseUid = body.CourseUid
	restaurantItem.ServiceId = serviceCart.ServiceId
	restaurantItem.BillId = serviceCart.Id
	restaurantItem.ItemId = serviceCartItem.Id

	if err := restaurantItem.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find res item"+err.Error())
		return
	}

	// update service cart
	serviceCart.Amount += (int64(body.Quantity) * serviceCartItem.UnitPrice) - (int64(serviceCartItem.Quality) * serviceCartItem.UnitPrice)
	if err := serviceCart.Update(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// update service item
	serviceCartItem.Quality = int(body.Quantity)
	serviceCartItem.Amount = int64(body.Quantity) * serviceCartItem.UnitPrice
	serviceCartItem.Input = body.Note

	if err := serviceCartItem.Update(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// update res item
	restaurantItem.Quantity = body.Quantity
	restaurantItem.QuantityProgress = body.Quantity
	restaurantItem.ItemNote = body.Note

	if err := restaurantItem.Update(db); err != nil {
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

	// Update trạng thái khi trả hết món
	if resItem.QuantityProgress-1 == 0 {
		resItem.ItemStatus = constants.RES_STATUS_DONE
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
	serviceCart.Note = body.Note

	if err := serviceCart.Create(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// add item
	for _, item := range body.ListOrderItem {
		// create cart item
		serviceCartItem := model_booking.BookingServiceItem{}

		// add res item with combo
		restaurantItems := []models.RestaurantItem{}

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

			// add infor cart item
			serviceCartItem.Type = kiosk.KioskType
			serviceCartItem.GroupCode = fbSet.GroupCode
			serviceCartItem.Name = fbSet.SetName
			serviceCartItem.UnitPrice = int64(fbSet.Price)

			// add item res
			for _, v := range fbSet.FBList {
				fb := model_service.FoodBeverage{}
				fb.PartnerUid = body.PartnerUid
				fb.CourseUid = body.CourseUid
				fb.FBCode = v

				if err := fb.FindFirst(db); err != nil {
					response_message.BadRequest(c, "Find fb in combo "+err.Error())
					return
				}

				item := models.RestaurantItem{
					Type:          fb.Type,
					ItemName:      fb.VieName,
					ItemComboName: fbSet.SetName,
					ItemCode:      fb.FBCode,
					ItemUnit:      fb.Unit,
				}

				restaurantItems = append(restaurantItems, item)
			}
		} else {
			fb := model_service.FoodBeverage{}
			fb.PartnerUid = body.PartnerUid
			fb.CourseUid = body.CourseUid
			fb.FBCode = item.ItemCode

			if err := fb.FindFirst(db); err != nil {
				response_message.BadRequest(c, "Find fb "+err.Error())
				return
			}

			// add infor cart item
			serviceCartItem.Type = kiosk.KioskType
			serviceCartItem.GroupCode = fb.GroupCode
			serviceCartItem.Name = fb.VieName
			serviceCartItem.UnitPrice = int64(fb.Price)
			serviceCartItem.Unit = fb.Unit

			// add infor res item
			item := models.RestaurantItem{
				Type:     fb.Type,
				ItemName: fb.VieName,
				ItemCode: fb.FBCode,
				ItemUnit: fb.Unit,
			}

			restaurantItems = append(restaurantItems, item)
		}

		// update amount service cart
		// update service cart
		serviceCart.Amount += (int64(item.Quantity) * serviceCartItem.UnitPrice)

		// add infor cart item
		serviceCartItem.PartnerUid = body.PartnerUid
		serviceCartItem.CourseUid = body.CourseUid
		serviceCartItem.ServiceId = strconv.Itoa(int(serviceCart.ServiceId))
		serviceCartItem.ServiceBill = serviceCart.Id
		serviceCartItem.ItemCode = item.ItemCode
		serviceCartItem.Quality = item.Quantity
		serviceCartItem.Amount = int64(item.Quantity) * serviceCartItem.UnitPrice
		serviceCartItem.UserAction = prof.UserName

		if err := serviceCartItem.Create(db); err != nil {
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
			v.Quantity = item.Quantity
			v.QuantityProgress = item.Quantity

			// Đổi trạng thái món khi đã có bill code
			if serviceCart.BillCode != "NONE" {
				v.ItemStatus = constants.RES_STATUS_PROCESS
			}

			if err := v.Create(db); err != nil {
				response_message.BadRequest(c, err.Error())
				return
			}
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
