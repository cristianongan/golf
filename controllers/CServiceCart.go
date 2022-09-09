package controllers

import (
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	model_booking "start/models/booking"
	kiosk_inventory "start/models/kiosk-inventory"
	model_service "start/models/service"
	"start/utils/response_message"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/datatypes"
)

type CServiceCart struct{}

// Thêm sản phẩm vào giỏ hàng
func (_ CServiceCart) AddItemServiceToCart(c *gin.Context, prof models.CmsUser) {
	var body request.AddItemServiceCartBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("AddItemToServiceCart BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate golf bag
	booking := model_booking.Booking{}
	booking.Bag = body.GolfBag
	booking.BookingDate = time.Now().Format("02/01/2006")
	if err := booking.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate kiosk
	kiosk := model_service.Kiosk{}
	kiosk.Id = body.ServiceId
	if err := kiosk.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// create cart item
	serviceCartItem := model_booking.BookingServiceItem{}

	// validate item code by group
	if body.GroupType == constants.GROUP_FB {
		fb := model_service.FoodBeverage{}
		fb.PartnerUid = prof.PartnerUid
		fb.CourseUid = prof.CourseUid
		fb.FBCode = body.ItemCode

		if err := fb.FindFirstInKiosk(kiosk.Id); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
		// add infor cart item
		serviceCartItem.GroupCode = fb.GroupCode
		serviceCartItem.Name = fb.VieName
		serviceCartItem.UnitPrice = int64(fb.Price)
	}

	if body.GroupType == constants.GROUP_PROSHOP {
		proshop := model_service.Proshop{}
		proshop.PartnerUid = prof.PartnerUid
		proshop.CourseUid = prof.CourseUid
		proshop.ProShopId = body.ItemCode

		if err := proshop.FindFirst(); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
		// add infor cart item
		serviceCartItem.GroupCode = proshop.GroupCode
		serviceCartItem.Name = proshop.VieName
		serviceCartItem.UnitPrice = int64(proshop.Price)
	}

	if body.GroupType == constants.GROUP_RENTAL {
		rental := model_service.Rental{}
		rental.PartnerUid = prof.PartnerUid
		rental.CourseUid = prof.CourseUid
		rental.RentalId = body.ItemCode

		if err := rental.FindFirst(); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
		// add infor cart item
		serviceCartItem.GroupCode = rental.GroupCode
		serviceCartItem.Name = rental.VieName
		serviceCartItem.UnitPrice = int64(rental.Price)
	}

	// validate quantity
	inventory := kiosk_inventory.InventoryItem{}
	inventory.PartnerUid = prof.PartnerUid
	inventory.CourseUid = prof.CourseUid
	inventory.ServiceId = body.ServiceId
	inventory.Code = body.ItemCode

	if err := inventory.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Kiểm tra số lượng hàng tồn trong kho
	if body.Quantity > inventory.Quantity {
		response_message.BadRequest(c, "The quantity of goods in stock is not enough")
		return
	}

	// Update số lượng hàng tồn trong kho
	inventory.Quantity -= body.Quantity
	if err := inventory.Update(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// check service cart
	serviceCart := models.ServiceCart{}
	serviceCart.PartnerUid = prof.PartnerUid
	serviceCart.CourseUid = prof.CourseUid
	serviceCart.GolfBag = body.GolfBag
	serviceCart.BookingDate = datatypes.Date(time.Now().UTC())
	serviceCart.ServiceId = body.ServiceId
	serviceCart.BillCode = "NONE"

	err := serviceCart.FindFirst()
	// no cart
	if err != nil {
		// create cart
		serviceCart.Amount = body.Quantity * serviceCartItem.UnitPrice
		if err := serviceCart.Create(); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	} else {
		// update tổng giá bill
		serviceCart.Amount += body.Quantity * serviceCartItem.UnitPrice
		if err := serviceCart.Update(); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	// add infor cart item
	serviceCartItem.PartnerUid = prof.PartnerUid
	serviceCartItem.CourseUid = prof.CourseUid
	serviceCartItem.Bag = booking.Bag
	serviceCartItem.BillCode = booking.BillCode
	serviceCartItem.BookingUid = booking.Uid
	serviceCartItem.PlayerName = booking.CustomerName
	serviceCartItem.ServiceId = strconv.Itoa(int(serviceCart.ServiceId))
	serviceCartItem.ServiceBill = serviceCart.Id
	serviceCartItem.Order = body.ItemCode
	serviceCartItem.Quality = int(body.Quantity)
	serviceCartItem.Amount = body.Quantity * serviceCartItem.UnitPrice
	serviceCartItem.UserAction = prof.UserName

	if err := serviceCartItem.Create(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ CServiceCart) AddDiscountToItem(c *gin.Context, prof models.CmsUser) {
	var body request.AddDiscountServiceItemBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("AddDiscountToItem BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate cart item
	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.PartnerUid = prof.PartnerUid
	serviceCartItem.CourseUid = prof.CourseUid
	serviceCartItem.Id = body.CartItemId

	if err := serviceCartItem.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate cart
	serviceCart := models.ServiceCart{}
	serviceCart.Id = serviceCartItem.ServiceBill
	serviceCart.PartnerUid = prof.PartnerUid
	serviceCart.CourseUid = prof.CourseUid
	serviceCart.BillCode = "NONE"

	if err := serviceCart.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	serviceCartItem.DiscountType = body.DiscountType
	serviceCartItem.DiscountValue = int64(body.DiscountPrice)
	serviceCartItem.DiscountReason = body.DiscountReason

	if err := serviceCartItem.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ CServiceCart) GetItemInCart(c *gin.Context, prof models.CmsUser) {
	query := request.GetItemServiceCartBody{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
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
	serviceCart.PartnerUid = prof.PartnerUid
	serviceCart.CourseUid = prof.CourseUid
	serviceCart.ServiceId = query.ServiceId
	serviceCart.GolfBag = query.GolfBag
	serviceCart.BookingDate = datatypes.Date(bookingDate)

	if err := serviceCart.FindFirst(); err != nil {
		res := response.PageResponse{
			Total: 0,
			Data:  nil,
		}

		c.JSON(200, res)
		return
	}

	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.PartnerUid = prof.PartnerUid
	serviceCartItem.CourseUid = prof.CourseUid
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

func (_ CServiceCart) GetBestItemInKiosk(c *gin.Context, prof models.CmsUser) {
	query := request.GetBestItemBody{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	page := models.Page{
		Limit:   query.PageRequest.Limit,
		Page:    query.PageRequest.Page,
		SortBy:  query.PageRequest.SortBy,
		SortDir: query.PageRequest.SortDir,
	}

	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.PartnerUid = prof.PartnerUid
	serviceCartItem.CourseUid = prof.CourseUid
	serviceCartItem.ServiceId = strconv.Itoa(int(query.ServiceId))
	serviceCartItem.GroupCode = query.GroupCode

	list, total, err := serviceCartItem.FindBestCartItem(page)

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
	query := request.GetServiceCartBody{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
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
	serviceCart.PartnerUid = prof.PartnerUid
	serviceCart.CourseUid = prof.CourseUid
	serviceCart.ServiceId = query.ServiceId
	serviceCart.BookingDate = datatypes.Date(bookingDate)

	list, total, err := serviceCart.FindList(page)

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
	var body request.UpdateServiceCartBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("UpdateQuantityToCart BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate cart item
	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.PartnerUid = prof.PartnerUid
	serviceCartItem.CourseUid = prof.CourseUid
	serviceCartItem.Id = body.CartItemId

	if err := serviceCartItem.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate cart
	serviceCart := models.ServiceCart{}
	serviceCart.PartnerUid = prof.PartnerUid
	serviceCart.CourseUid = prof.CourseUid
	serviceCart.Id = serviceCartItem.ServiceBill
	serviceCart.BillCode = "NONE"

	if err := serviceCart.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate quantity
	inventory := kiosk_inventory.InventoryItem{}
	inventory.PartnerUid = prof.PartnerUid
	inventory.CourseUid = prof.CourseUid
	inventory.ServiceId = serviceCart.ServiceId
	inventory.Code = serviceCartItem.Order

	if err := inventory.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Kiểm tra số lượng hàng tồn trong kho
	if body.Quantity > inventory.Quantity+int64(serviceCartItem.Quality) {
		response_message.BadRequest(c, "The quantity of goods in stock is not enough")
		return
	}

	// Update số lượng hàng tồn trong kho
	inventory.Quantity = inventory.Quantity + int64(serviceCartItem.Quality) - body.Quantity
	if err := inventory.Update(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// update service cart
	serviceCart.Amount += (body.Quantity * serviceCartItem.UnitPrice) - (int64(serviceCartItem.Quality) * serviceCartItem.UnitPrice)
	if err := serviceCart.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// update service item
	serviceCartItem.Quality = int(body.Quantity)
	serviceCartItem.Amount = body.Quantity * serviceCartItem.UnitPrice
	serviceCartItem.Input = body.Note

	if err := serviceCartItem.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ CServiceCart) DeleteItemInCart(c *gin.Context, prof models.CmsUser) {
	var body request.DeleteItemInKioskCartBody

	if err := c.BindJSON(&body); err != nil {
		log.Print("DeleteItemInCart BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	// validate cart item
	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.PartnerUid = prof.PartnerUid
	serviceCartItem.CourseUid = prof.CourseUid
	serviceCartItem.Id = body.CartItemId

	if err := serviceCartItem.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate cart
	serviceCart := models.ServiceCart{}
	serviceCart.PartnerUid = prof.PartnerUid
	serviceCart.CourseUid = prof.CourseUid
	serviceCart.Id = serviceCartItem.ServiceBill
	serviceCart.BillCode = "NONE"

	if err := serviceCart.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate quantity
	inventory := kiosk_inventory.InventoryItem{}
	inventory.PartnerUid = prof.PartnerUid
	inventory.CourseUid = prof.CourseUid
	inventory.ServiceId = serviceCart.ServiceId
	inventory.Code = serviceCartItem.Order

	if err := inventory.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Update số lượng hàng tồn trong kho
	inventory.Quantity += int64(serviceCartItem.Quality)
	if err := inventory.Update(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// update service cart
	serviceCart.Amount -= int64(serviceCartItem.Quality) * serviceCartItem.UnitPrice
	if err := serviceCart.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// Delete Item
	if err := serviceCartItem.Delete(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}

func (_ CServiceCart) CreateBilling(c *gin.Context, prof models.CmsUser) {
	var body request.CreateBillCodeBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("CreateBilling BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	serviceCart := models.ServiceCart{}
	serviceCart.PartnerUid = prof.PartnerUid
	serviceCart.CourseUid = prof.CourseUid
	serviceCart.ServiceId = body.ServiceId
	serviceCart.GolfBag = body.GolfBag
	serviceCart.BookingDate = datatypes.Date(time.Now().UTC())

	if err := serviceCart.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	serviceCart.BillCode = "KIOSK-BILLING-" + time.Now().Format("20060102150405")

	if err := serviceCart.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ CServiceCart) MoveItemToOtherCart(c *gin.Context, prof models.CmsUser) {
	var body request.MoveItemToOtherServiceCartBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("MoveItemToOtherCart BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate golf bag
	booking := model_booking.Booking{}
	booking.Bag = body.GolfBag
	booking.BookingDate = time.Now().Format("02/01/2006")
	if err := booking.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate cart code
	sourceServiceCart := models.ServiceCart{}
	sourceServiceCart.Id = body.ServiceCartId
	sourceServiceCart.PartnerUid = prof.PartnerUid
	sourceServiceCart.CourseUid = prof.CourseUid
	sourceServiceCart.BillCode = "NONE"

	if err := sourceServiceCart.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate cart by golf bag
	targetServiceCart := models.ServiceCart{}
	targetServiceCart.PartnerUid = prof.PartnerUid
	targetServiceCart.CourseUid = prof.CourseUid
	targetServiceCart.GolfBag = body.GolfBag
	targetServiceCart.BookingDate = datatypes.Date(time.Now().UTC())
	targetServiceCart.ServiceId = sourceServiceCart.ServiceId
	targetServiceCart.BillCode = "NONE"

	err := targetServiceCart.FindFirst()

	// no cart
	if err != nil {
		// create cart
		if err := targetServiceCart.Create(); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	hasError := false

	var errFor error

	for _, cartItemId := range body.CartItemIdList {
		serviceCartItem := model_booking.BookingServiceItem{}
		serviceCartItem.Id = cartItemId
		serviceCartItem.ServiceBill = sourceServiceCart.Id

		if err := serviceCartItem.FindFirst(); err != nil {
			continue
		}

		serviceCartItem.ServiceBill = targetServiceCart.Id

		if errFor = serviceCartItem.Update(); errFor != nil {
			hasError = true
			break
		}
	}

	if hasError {
		response_message.InternalServerError(c, errFor.Error())
		return
	}

	okRes(c)
}
