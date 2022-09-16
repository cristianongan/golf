package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
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
	if err := booking.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	if booking.BagStatus != constants.BAG_STATUS_WAITING && booking.BagStatus != constants.BAG_STATUS_IN_COURSE && booking.BagStatus != constants.BAG_STATUS_TIMEOUT {
		response_message.BadRequest(c, "Bag status invalid")
		return
	}

	// validate kiosk
	kiosk := model_service.Kiosk{}
	kiosk.Id = body.ServiceId
	if err := kiosk.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Tạo đơn order
	serviceCart := models.ServiceCart{}

	serviceCart.Type = body.Type
	serviceCart.TypeCode = body.TypeCode

	if body.Type == constants.RES_TYPE_TABLE {
		serviceCart.NumberGuest = body.NumberGuest
	}

	serviceCart.PartnerUid = body.PartnerUid
	serviceCart.CourseUid = body.CourseUid
	serviceCart.GolfBag = body.GolfBag
	serviceCart.BookingUid = booking.Uid
	serviceCart.BookingDate = datatypes.Date(time.Now().UTC())
	serviceCart.ServiceId = body.ServiceId
	serviceCart.BillCode = "NONE"
	serviceCart.BillStatus = constants.RES_STATUS_ORDER
	serviceCart.StaffOrder = prof.FullName

	if err := serviceCart.Create(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	c.JSON(200, serviceCart)
}

// Tạo mã đơn
func (_ CRestaurantOrder) CreateBill(c *gin.Context, prof models.CmsUser) {
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
	serviceCart.BillCode = "NONE"

	if err := serviceCart.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	serviceCart.BillCode = "OD-" + strconv.Itoa(int(body.BillId))
	serviceCart.BillStatus = constants.RES_STATUS_PROCESS

	if err := serviceCart.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	//find all item in bill
	restaurantItem := models.RestaurantItem{}
	restaurantItem.BillId = body.BillId

	list, err := restaurantItem.FindAll()

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	for _, item := range list {
		item.ItemStatus = constants.RES_STATUS_PROCESS

		if err := item.Update(); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	// createExportBillInventory(c, prof, serviceCart, serviceCart.BillCode)

	c.JSON(200, serviceCart)
}

func (_ CRestaurantOrder) GetListBill(c *gin.Context, prof models.CmsUser) {
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

	bookingDate, _ := time.Parse("2006-01-02", query.BookingDate)

	serviceCart := models.ServiceCart{}
	serviceCart.PartnerUid = query.PartnerUid
	serviceCart.CourseUid = query.CourseUid
	serviceCart.ServiceId = query.ServiceId
	serviceCart.BookingDate = datatypes.Date(bookingDate)
	serviceCart.BillStatus = query.BillStatus

	list, total, err := serviceCart.FindList(page)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	listData := make([]map[string]interface{}, len(list))

	for i, data := range list {
		//find all item in bill
		restaurantItem := models.RestaurantItem{}
		restaurantItem.BillId = data.Id

		listResItem, err := restaurantItem.FindAll()

		if err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		// Add infor to response
		listData[i]["bill_infor"] = data
		listData[i]["menu"] = listResItem
	}

	res := response.PageResponse{
		Total: total,
		Data:  listData,
	}

	c.JSON(200, res)
}

// Thêm sản phẩm vào hóa đơn
func (_ CRestaurantOrder) AddItemOrder(c *gin.Context, prof models.CmsUser) {
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
	if err := serviceCart.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate golf bag
	booking := model_booking.Booking{}
	booking.PartnerUid = body.PartnerUid
	booking.CourseUid = body.CourseUid
	booking.Bag = serviceCart.GolfBag
	booking.BookingDate = time.Now().Format("02/01/2006")
	if err := booking.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate kiosk
	kiosk := model_service.Kiosk{}
	kiosk.Id = serviceCart.ServiceId
	if err := kiosk.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// create cart item
	serviceCartItem := model_booking.BookingServiceItem{}

	// add res item
	restaurantItem := models.RestaurantItem{}

	// validate item code by group
	if body.Type == "COMBO" {

	} else {
		fb := model_service.FoodBeverage{}
		fb.PartnerUid = body.PartnerUid
		fb.CourseUid = body.CourseUid
		fb.FBCode = body.ItemCode

		if err := fb.FindFirstInKiosk(kiosk.Id); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}

		// add infor cart item
		serviceCartItem.GroupCode = fb.GroupCode
		serviceCartItem.Name = fb.VieName
		serviceCartItem.UnitPrice = int64(fb.Price)
		serviceCartItem.Unit = fb.Unit

		// add infor res item
		restaurantItem.Type = fb.Type
		restaurantItem.ItemName = fb.VieName
	}

	// update service cart
	serviceCart.Amount += (int64(body.Quantity) * serviceCartItem.UnitPrice)
	if err := serviceCart.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
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
	serviceCartItem.Order = body.ItemCode
	serviceCartItem.Quality = body.Quantity
	serviceCartItem.Amount = int64(body.Quantity) * serviceCartItem.UnitPrice
	serviceCartItem.UserAction = prof.UserName

	if err := serviceCartItem.Create(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// add infor restaurant item
	restaurantItem.PartnerUid = body.PartnerUid
	restaurantItem.CourseUid = body.CourseUid
	restaurantItem.ServiceId = serviceCart.ServiceId
	restaurantItem.OrderDate = time.Now().Format("02/01/2006")
	restaurantItem.BillId = serviceCart.Id
	restaurantItem.ItemId = serviceCartItem.Id
	restaurantItem.ItemCode = body.ItemCode
	restaurantItem.ItemStatus = constants.RES_STATUS_ORDER
	restaurantItem.Quatity = body.Quantity
	restaurantItem.QuatityProgress = body.Quantity

	// Đổi trạng thái món khi đã có bill code
	if serviceCart.BillCode != "NONE" {
		restaurantItem.ItemStatus = constants.RES_STATUS_PROCESS
	}

	if err := restaurantItem.Create(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	c.JSON(200, serviceCartItem)
}

// Update sản phẩm
func (_ CRestaurantOrder) UpdateItemOrder(c *gin.Context, prof models.CmsUser) {
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
	serviceCartItem.PartnerUid = prof.PartnerUid
	serviceCartItem.CourseUid = prof.CourseUid
	serviceCartItem.Id = body.ItemId

	if err := serviceCartItem.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate restaurant order
	serviceCart := models.ServiceCart{}
	serviceCart.PartnerUid = body.PartnerUid
	serviceCart.CourseUid = body.CourseUid
	serviceCart.Id = serviceCartItem.ServiceBill

	if err := serviceCart.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate res item
	restaurantItem := models.RestaurantItem{}
	restaurantItem.PartnerUid = body.PartnerUid
	restaurantItem.CourseUid = body.CourseUid
	restaurantItem.ServiceId = serviceCart.ServiceId
	restaurantItem.BillId = serviceCart.Id
	restaurantItem.ItemId = serviceCartItem.Id

	if err := restaurantItem.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// update service cart
	serviceCart.Amount += (int64(body.Quantity) * serviceCartItem.UnitPrice) - (int64(serviceCartItem.Quality) * serviceCartItem.UnitPrice)
	if err := serviceCart.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// update service item
	serviceCartItem.Quality = int(body.Quantity)
	serviceCartItem.Amount = int64(body.Quantity) * serviceCartItem.UnitPrice
	serviceCartItem.Input = body.Note

	if err := serviceCartItem.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// update res item
	restaurantItem.Quatity = body.Quantity
	restaurantItem.QuatityProgress = body.Quantity
	restaurantItem.ItemNote = body.Note

	if err := restaurantItem.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

// Delete sản phẩm
func (_ CRestaurantOrder) DeleteItemOrder(c *gin.Context, prof models.CmsUser) {
	idRequest := c.Param("id")
	id, errId := strconv.ParseInt(idRequest, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	// validate cart item
	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.Id = id

	if err := serviceCartItem.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate cart
	serviceCart := models.ServiceCart{}
	serviceCart.Id = serviceCartItem.ServiceBill

	if err := serviceCart.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate res item
	restaurantItem := models.RestaurantItem{}
	restaurantItem.BillId = serviceCart.Id
	restaurantItem.ItemId = serviceCartItem.Id

	if err := restaurantItem.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// update service cart
	serviceCart.Amount -= serviceCartItem.Amount
	if err := serviceCart.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// Delete Item
	if err := serviceCartItem.Delete(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Delete res item
	if err := restaurantItem.Delete(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}

// get list sản phẩm
func (_ CRestaurantOrder) GetListItemOrder(c *gin.Context, prof models.CmsUser) {
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

	// validate bill
	serviceCart := models.ServiceCart{}
	serviceCart.Id = query.BillId

	if err := serviceCart.FindFirst(); err != nil {
		res := response.PageResponse{
			Total: 0,
			Data:  nil,
		}

		c.JSON(200, res)
		return
	}

	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.ServiceBill = serviceCart.Id

	list, total, err := serviceCartItem.FindList(page)

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

// get list sản phẩm
func (_ CRestaurantOrder) UpdateResItem(c *gin.Context, prof models.CmsUser) {
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
	resItem.PartnerUid = body.PartnerUid
	resItem.CourseUid = body.CourseUid
	resItem.Id = body.ItemId

	if err := resItem.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Update quatity progress when finish
	if body.Action == constants.RES_STATUS_DONE {
		// Update trạng thái khi trả hết món
		if resItem.QuatityProgress-1 == 0 {
			resItem.ItemStatus = constants.RES_STATUS_DONE
		}

		resItem.QuatityProgress -= 1
	}

	// Update cancel res item
	if body.Action == constants.RES_STATUS_CANCEL {
		resItem.ItemStatus = constants.RES_STATUS_CANCEL

		// validate bill
		serviceCart := models.ServiceCart{}
		serviceCart.PartnerUid = body.PartnerUid
		serviceCart.CourseUid = body.CourseUid
		serviceCart.Id = resItem.BillId

		if err := serviceCart.FindFirst(); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}

		// validate bill
		serviceCartItem := model_booking.BookingServiceItem{}
		serviceCartItem.PartnerUid = body.PartnerUid
		serviceCartItem.CourseUid = body.CourseUid
		serviceCartItem.ServiceBill = resItem.BillId
		serviceCartItem.Order = resItem.ItemCode

		if err := serviceCartItem.FindFirst(); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}

		// update service cart
		serviceCart.Amount -= serviceCartItem.Amount
		if err := serviceCart.Update(); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}

		// Delete res item
		if err := serviceCartItem.Delete(); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
	}

	// update res item
	if err := resItem.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

// // get list sản phẩm
// func (_ CRestaurantOrder) UpdateResItem(c *gin.Context, prof models.CmsUser) {
// 	body := request.UpdateResItemBody{}
// 	if bindErr := c.ShouldBind(&body); bindErr != nil {
// 		response_message.BadRequest(c, bindErr.Error())
// 		return
// 	}

// 	// validate body
// 	validate := validator.New()

// 	if err := validate.Struct(body); err != nil {
// 		response_message.BadRequest(c, err.Error())
// 		return
// 	}

// 	okRes(c)
// }
