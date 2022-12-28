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
	"start/utils"
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
	dateDisplay, _ := utils.GetBookingDateFromTimestamp(time.Now().Unix())

	booking := model_booking.Booking{}
	booking.PartnerUid = body.PartnerUid
	booking.CourseUid = body.CourseUid
	booking.Bag = body.GolfBag
	booking.BookingDate = dateDisplay
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find booking "+err.Error())
		return
	}

	if booking.BagStatus != constants.BAG_STATUS_WAITING && booking.BagStatus != constants.BAG_STATUS_IN_COURSE && booking.BagStatus != constants.BAG_STATUS_TIMEOUT {
		response_message.BadRequest(c, "Bag status invalid")
		return
	}

	if booking.LockBill == setBoolForCursor(true) {
		response_message.BadRequest(c, "Bag lock")
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
	serviceCart.BookingDate = datatypes.Date(time.Now().Local())
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

	// validate golf bag
	booking := model_booking.Booking{}
	booking.Uid = serviceCart.BookingUid

	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Booking "+err.Error())
		return
	}

	if serviceCart.BillCode == constants.BILL_NONE {
		serviceCart.BillCode = "OD-" + strconv.Itoa(int(body.BillId))
		serviceCart.TimeProcess = time.Now().Unix()
		serviceCart.BillStatus = constants.RES_STATUS_PROCESS
		// serviceCart.TotalMoveKitchen += 1

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

	if len(list) > 0 {
		if serviceCart.BillStatus == constants.RES_BILL_STATUS_FINISH {
			serviceCart.TimeProcess = time.Now().Unix()
			serviceCart.BillStatus = constants.RES_STATUS_PROCESS
		}
		// Update số lần move kitchen
		serviceCart.TotalMoveKitchen += 1

		if err := serviceCart.Update(db); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
	}

	for _, item := range list {
		item.ItemStatus = constants.RES_STATUS_PROCESS
		item.MoveKitchenTimes = serviceCart.TotalMoveKitchen

		if err := item.Update(db); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
	}

	//Update lại giá trong booking
	updatePriceWithServiceItem(booking, prof)
	createExportBillInventory(c, prof, serviceCart, serviceCart.BillCode)

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
	serviceCart.PartnerUid = prof.PartnerUid
	serviceCart.CourseUid = prof.CourseUid

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

	//Update lại giá trong booking
	updatePriceWithServiceItem(booking, prof)

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
	serviceCart.FromService = query.FromService

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
	if body.Type == constants.SERVICE_ITEM_RES_COMBO {
		fbSet := model_service.FbPromotionSet{}
		fbSet.PartnerUid = body.PartnerUid
		fbSet.CourseUid = body.CourseUid
		fbSet.Code = body.ItemCode

		if err := fbSet.FindFirst(db); err != nil {
			response_message.BadRequest(c, "Find fb set "+err.Error())
			return
		}

		// add infor cart item
		serviceCartItem.ItemId = fbSet.Id
		serviceCartItem.ServiceType = kiosk.ServiceType
		serviceCartItem.Type = kiosk.KioskType
		serviceCartItem.Location = kiosk.KioskName
		serviceCartItem.ItemType = constants.SERVICE_ITEM_RES_COMBO
		serviceCartItem.Name = fbSet.VieName
		serviceCartItem.EngName = fbSet.EnglishName
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

	if body.Type == constants.SERVICE_ITEM_RES_NORMAL {
		fb := model_service.FoodBeverage{}
		fb.PartnerUid = body.PartnerUid
		fb.CourseUid = body.CourseUid
		fb.FBCode = body.ItemCode

		if err := fb.FindFirst(db); err != nil {
			response_message.BadRequest(c, "Find fb "+err.Error())
			return
		}

		// add infor cart item
		serviceCartItem.ItemId = fb.Id
		serviceCartItem.ServiceType = kiosk.ServiceType
		serviceCartItem.Type = kiosk.KioskType
		serviceCartItem.Location = kiosk.KioskName
		serviceCartItem.GroupCode = fb.GroupCode
		serviceCartItem.ItemType = constants.SERVICE_ITEM_RES_NORMAL
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
		v.OrderDate = time.Now().Format(constants.DATE_FORMAT_1)
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
	serviceCartItem.PartnerUid = prof.PartnerUid
	serviceCartItem.CourseUid = prof.CourseUid

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
			if body.Quantity > 0 {
				if v.ItemComboCode != "" {
					v.Quantity = (v.Quantity / serviceCartItem.Quality) * body.Quantity
					v.QuantityProgress = (v.QuantityProgress / serviceCartItem.Quality) * body.Quantity
				} else {
					v.Quantity = body.Quantity
					v.QuantityProgress = body.Quantity
				}
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
	serviceCartItem.PartnerUid = prof.PartnerUid
	serviceCartItem.CourseUid = prof.CourseUid

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

	list, total, err := serviceCartItem.FindListWithStatus(db, page)

	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	for _, item := range list {
		// Kiểm tra trạng thái các món
		if item["order_counts"].(int64) > 0 {
			item["item_status"] = constants.RES_STATUS_ORDER
		} else if item["process_counts"].(int64) > 0 {
			item["item_status"] = constants.RES_STATUS_PROCESS
		} else {
			item["item_status"] = constants.RES_STATUS_DONE
		}
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
	resItem.ItemCode = body.ItemCode
	resItem.BillId = body.BillId
	resItem.ItemStatus = constants.RES_STATUS_PROCESS
	resItem.PartnerUid = prof.PartnerUid
	resItem.CourseUid = prof.CourseUid

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
	}

	// Update quantity progress when finish
	resItem.QuantityProgress -= 1

	// update res item
	if err := resItem.Update(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

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
	resItem.ItemStatus = constants.RES_STATUS_PROCESS

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
		dateDisplay, _ := utils.GetBookingDateFromTimestamp(time.Now().Unix())

		booking.PartnerUid = body.PartnerUid
		booking.CourseUid = body.CourseUid
		booking.Bag = body.GolfBag
		booking.BookingDate = dateDisplay
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
	if body.FromServiceId != 0 {
		fromKiosk.Id = body.FromServiceId
		if err := fromKiosk.FindFirst(db); err != nil {
			response_message.BadRequest(c, "Find from kiosk "+err.Error())
			return
		}

		serviceCart.FromService = body.FromServiceId
		serviceCart.FromServiceName = fromKiosk.KioskName
	}

	// create cart item
	itemCombos := []model_service.FbPromotionSet{}
	itemQuatityCombos := []int{}
	itemFBs := []model_service.FoodBeverage{}
	itemQuatityFBs := []int{}

	// validate item
	if len(body.ListOrderItem) > 0 {
		for _, item := range body.ListOrderItem {
			// validate item code by group
			if item.Type == constants.SERVICE_ITEM_RES_COMBO {
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
	}

	// create service cart
	serviceCart.PartnerUid = body.PartnerUid
	serviceCart.CourseUid = body.CourseUid
	serviceCart.BookingDate = datatypes.Date(time.Now().Local())
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
				EngName:     item.EnglishName,
				UserAction:  prof.UserName,
				PlayerName:  body.PlayerName,
				ServiceType: kiosk.ServiceType,
				ItemId:      item.Id,
			}

			serviceCartItem.UnitPrice = int64(item.Price)
			serviceCartItem.ServiceId = strconv.Itoa(int(serviceCart.ServiceId))
			serviceCartItem.ItemCode = item.Code
			serviceCartItem.ItemType = constants.SERVICE_ITEM_RES_COMBO
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
		}
	}

	if len(itemFBs) > 0 {
		for i, item := range itemFBs {
			quantity := itemQuatityFBs[i]
			// add infor cart item
			serviceCartItem := model_booking.BookingServiceItem{
				PartnerUid:  body.PartnerUid,
				CourseUid:   body.CourseUid,
				ItemId:      item.Id,
				ServiceType: kiosk.ServiceType,
				ServiceBill: serviceCart.Id,
				GroupCode:   item.GroupCode,
				Type:        kiosk.KioskType,
				Location:    kiosk.KioskName,
				Name:        item.VieName,
				EngName:     item.EnglishName,
				Unit:        item.Unit,
				UserAction:  prof.UserName,
				PlayerName:  body.PlayerName,
			}

			serviceCartItem.UnitPrice = int64(item.Price)
			serviceCartItem.ServiceId = strconv.Itoa(int(serviceCart.ServiceId))
			serviceCartItem.ItemCode = item.FBCode
			serviceCartItem.ItemType = constants.SERVICE_ITEM_RES_NORMAL
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
	serviceCart.PartnerUid = prof.PartnerUid
	serviceCart.CourseUid = prof.CourseUid

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
	serviceCart.PartnerUid = prof.PartnerUid
	serviceCart.CourseUid = prof.CourseUid
	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find bill "+err.Error())
		return
	}

	// validate kiosk
	kiosk := model_service.Kiosk{}
	kiosk.Id = serviceCart.ServiceId
	if err := kiosk.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find kiosk "+err.Error())
		return
	}

	if body.PlayerName != "" {
		serviceCart.PlayerName = body.PlayerName
	}

	if body.Phone != "" {
		serviceCart.Phone = body.Phone
	}

	booking := model_booking.Booking{}
	if body.GolfBag != "" {
		// validate golf bag
		dateDisplay, _ := utils.GetBookingDateFromTimestamp(time.Now().Unix())

		booking := model_booking.Booking{}
		booking.PartnerUid = body.PartnerUid
		booking.CourseUid = body.CourseUid
		booking.Bag = body.GolfBag
		booking.BookingDate = dateDisplay
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

	if len(body.ListOrderItem) > 0 {
		serviceItemF := model_booking.BookingServiceItem{
			ServiceBill: billId,
		}

		serviceItems, err := serviceItemF.FindAll(db)

		if err != nil {
			response_message.BadRequest(c, "Find All"+err.Error())
			return
		}

		// Xóa các item cũ
		for _, serviceItem := range serviceItems {
			if err := serviceItem.Delete(db); err != nil {
				response_message.BadRequest(c, "Delete service item "+err.Error())
				return
			}
		}

		// create cart item
		itemCombos := []model_service.FbPromotionSet{}
		itemQuatityCombos := []int{}
		itemFBs := []model_service.FoodBeverage{}
		itemQuatityFBs := []int{}

		// validate item
		for _, item := range body.ListOrderItem {
			// validate item code by group
			if item.Type == constants.SERVICE_ITEM_RES_COMBO {
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

		if len(itemCombos) > 0 {
			for i, item := range itemCombos {
				quantity := itemQuatityCombos[i]
				// add infor cart item
				serviceCartItem := model_booking.BookingServiceItem{
					PartnerUid:  body.PartnerUid,
					CourseUid:   body.CourseUid,
					ServiceBill: serviceCart.Id,
					ItemId:      item.Id,
					ServiceType: kiosk.ServiceType,
					Type:        kiosk.KioskType,
					Location:    kiosk.KioskName,
					Name:        item.VieName,
					EngName:     item.EnglishName,
					UserAction:  prof.UserName,
					PlayerName:  body.PlayerName,
				}

				serviceCartItem.UnitPrice = int64(item.Price)
				serviceCartItem.ServiceId = strconv.Itoa(int(serviceCart.ServiceId))
				serviceCartItem.ItemCode = item.Code
				serviceCartItem.ItemType = constants.SERVICE_ITEM_RES_COMBO
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
					ItemId:      item.Id,
					ServiceType: kiosk.ServiceType,
					Type:        kiosk.KioskType,
					EngName:     item.EnglishName,
					Unit:        item.Unit,
					Location:    kiosk.KioskName,
					Name:        item.VieName,
					UserAction:  prof.UserName,
					PlayerName:  body.PlayerName,
				}

				serviceCartItem.UnitPrice = int64(item.Price)
				serviceCartItem.ServiceId = strconv.Itoa(int(serviceCart.ServiceId))
				serviceCartItem.ItemCode = item.FBCode
				serviceCartItem.ItemType = constants.SERVICE_ITEM_RES_NORMAL
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
			}
		}

	}

	serviceCart.TypeCode = body.Table
	serviceCart.NumberGuest = body.NumberGuest
	serviceCart.Note = body.Note
	serviceCart.ResFloor = body.Floor

	if err := serviceCart.Update(db); err != nil {
		response_message.BadRequest(c, "Update bill "+err.Error())
		return
	}

	okRes(c)
}

// Confrim thông tin restaurant booking

func (_ CRestaurantOrder) ConfrimRestaurantBooking(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	billIdStr := c.Param("id")
	billId, err := strconv.ParseInt(billIdStr, 10, 64)
	if err != nil || billId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	body := request.ConfrimBookingRestaurantBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	serviceCart := models.ServiceCart{}
	serviceCart.Id = billId
	serviceCart.PartnerUid = prof.PartnerUid
	serviceCart.CourseUid = prof.CourseUid
	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find bill "+err.Error())
		return
	}

	booking := model_booking.Booking{}

	if body.GolfBag != "" {
		// validate golf bag
		dateDisplay, _ := utils.GetBookingDateFromTimestamp(time.Now().Unix())

		booking.PartnerUid = serviceCart.PartnerUid
		booking.CourseUid = serviceCart.CourseUid
		booking.Bag = body.GolfBag
		booking.BookingDate = dateDisplay
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

	// Find all service item with bill id
	serviceItemF := model_booking.BookingServiceItem{
		ServiceBill: billId,
	}

	serviceItems, err := serviceItemF.FindAll(db)

	if err != nil {
		response_message.BadRequest(c, "Find All"+err.Error())
		return
	}

	// create item res theo service item
	restaurantItems := []models.RestaurantItem{}

	for _, serviceItem := range serviceItems {
		if serviceItem.ItemType == constants.SERVICE_ITEM_RES_COMBO {
			fbSet := model_service.FbPromotionSet{}
			fbSet.PartnerUid = serviceCart.PartnerUid
			fbSet.CourseUid = serviceCart.CourseUid
			fbSet.Code = serviceItem.ItemCode

			if err := fbSet.FindFirst(db); err != nil {
				response_message.BadRequest(c, "Find fb set "+err.Error())
				return
			}

			// add item res
			for _, v := range fbSet.FBList {
				item := models.RestaurantItem{
					Type:             serviceItem.Type,
					BillId:           serviceCart.Id,
					ItemId:           serviceItem.Id,
					ItemName:         v.VieName,
					ItemComboName:    fbSet.VieName,
					ItemComboCode:    fbSet.Code,
					ItemCode:         v.FBCode,
					ItemUnit:         v.Unit,
					Quantity:         v.Quantity * serviceItem.Quality,
					QuantityProgress: v.Quantity * serviceItem.Quality,
				}

				restaurantItems = append(restaurantItems, item)
			}
		} else {
			// add item res
			item := models.RestaurantItem{
				Type:             serviceItem.Type,
				BillId:           serviceCart.Id,
				ItemId:           serviceItem.Id,
				ItemName:         serviceItem.Name,
				ItemCode:         serviceItem.ItemCode,
				ItemUnit:         serviceItem.Unit,
				Quantity:         serviceItem.Quality,
				QuantityProgress: serviceItem.Quality,
			}

			restaurantItems = append(restaurantItems, item)
		}

		if body.GolfBag != "" {
			serviceItem.Bag = booking.Bag
			serviceItem.BookingUid = booking.Uid
			serviceItem.BillCode = booking.BillCode
			serviceItem.PlayerName = booking.CustomerName

			if err := serviceItem.Update(db); err != nil {
				response_message.BadRequest(c, "Update service item "+err.Error())
				return
			}
		}
	}

	if len(restaurantItems) > 0 {
		for _, v := range restaurantItems {
			// add infor restaurant item
			v.PartnerUid = serviceCart.PartnerUid
			v.CourseUid = serviceCart.CourseUid
			v.ServiceId = serviceCart.ServiceId
			v.OrderDate = time.Now().Format(constants.DATE_FORMAT_1)
			v.ItemStatus = constants.RES_STATUS_ORDER

			if err := v.Create(db); err != nil {
				response_message.BadRequest(c, err.Error())
				return
			}
		}
	}

	serviceCart.BillStatus = constants.RES_BILL_STATUS_ORDER

	if err := serviceCart.Update(db); err != nil {
		response_message.BadRequest(c, "Update bill "+err.Error())
		return
	}

	okRes(c)
}

func (_ CRestaurantOrder) TransferItem(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.TransferItemBody{}
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
	booking.Bag = body.GolfBag
	booking.BookingDate = dateDisplay
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find booking target "+err.Error())
		return
	}

	if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		response_message.BadRequest(c, "Bag status invalid")
		return
	}

	// validate cart code
	sourceServiceCart := models.ServiceCart{}
	sourceServiceCart.Id = body.ServiceCartId

	if err := sourceServiceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find bill source "+err.Error())
		return
	}

	// validate golf bag source
	bookingSource := model_booking.Booking{}
	bookingSource.Bag = sourceServiceCart.GolfBag
	bookingSource.BookingDate = dateDisplay
	if err := bookingSource.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find booking source "+err.Error())
		return
	}

	// validate cart by golf bag
	targetServiceCart := models.ServiceCart{}
	targetServiceCart.PartnerUid = body.PartnerUid
	targetServiceCart.CourseUid = body.CourseUid
	targetServiceCart.GolfBag = body.GolfBag
	targetServiceCart.BookingDate = datatypes.Date(time.Now().UTC())
	targetServiceCart.ServiceId = sourceServiceCart.ServiceId
	targetServiceCart.ServiceType = sourceServiceCart.ServiceType
	targetServiceCart.BillStatus = sourceServiceCart.BillStatus
	targetServiceCart.BookingUid = booking.Uid
	targetServiceCart.StaffOrder = prof.FullName
	targetServiceCart.BillCode = constants.BILL_NONE

	// create cart
	if err := targetServiceCart.Create(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
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
		totalAmount += (serviceCartItem.Amount - serviceCartItem.DiscountValue)

		if errFor = serviceCartItem.Update(db); errFor != nil {
			hasError = true
			break
		}

		restaurantItem := models.RestaurantItem{}

		restaurantItem.ItemId = serviceCartItem.Id
		restaurantItem.BillId = targetServiceCart.Id

		if errFor = restaurantItem.UpdateBatchBillId(db); errFor != nil {
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

	if targetServiceCart.BillStatus != constants.RES_BILL_STATUS_BOOKING &&
		targetServiceCart.BillStatus != constants.RES_BILL_STATUS_ORDER {
		targetServiceCart.BillCode = "OD-" + strconv.Itoa(int(targetServiceCart.Id))
	}

	if err := targetServiceCart.Update(db); err != nil {
		response_message.InternalServerError(c, "Update target cart "+err.Error())
		return
	}

	if targetServiceCart.BillStatus == constants.RES_BILL_STATUS_PROCESS ||
		targetServiceCart.BillStatus == constants.RES_BILL_STATUS_ACTIVE ||
		targetServiceCart.BillStatus == constants.RES_BILL_STATUS_FINISH ||
		targetServiceCart.BillStatus == constants.RES_BILL_STATUS_OUT {

		//Update lại giá trong booking
		updatePriceWithServiceItem(booking, prof)
	}

	// Update amount target bill
	sourceServiceCart.Amount = sourceServiceCart.Amount - totalAmount

	if sourceServiceCart.BillStatus == constants.RES_BILL_STATUS_PROCESS ||
		sourceServiceCart.BillStatus == constants.RES_BILL_STATUS_ACTIVE ||
		sourceServiceCart.BillStatus == constants.RES_BILL_STATUS_FINISH ||
		sourceServiceCart.BillStatus == constants.RES_BILL_STATUS_OUT {

		//Update lại giá trong booking
		updatePriceWithServiceItem(bookingSource, prof)
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
		sourceServiceCart.BillStatus = constants.RES_BILL_STATUS_TRANSFER
	}

	if err := sourceServiceCart.Update(db); err != nil {
		response_message.InternalServerError(c, "Update target cart "+err.Error())
		return
	}

	okRes(c)
}
