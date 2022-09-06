package controllers

import (
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	model_booking "start/models/booking"
	kiosk_cart "start/models/kiosk-cart"
	kiosk_inventory "start/models/kiosk-inventory"
	model_service "start/models/service"
	"start/utils/response_message"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/datatypes"
)

type CKioskCart struct{}

// Thêm sản phẩm vào giỏ hàng
func (_ CKioskCart) AddItemToCart(c *gin.Context, prof models.CmsUser) {
	var body request.AddItemToKioskCartBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("AddItemToCart BindJSON error")
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
	kiosk.Id = body.KioskCode
	if err := kiosk.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// check cart
	cart := kiosk_cart.Cart{}
	cart.PartnerUid = prof.PartnerUid
	cart.CourseUid = prof.CourseUid
	cart.GolfBag = body.GolfBag
	cart.BookingDate = datatypes.Date(time.Now().UTC())
	cart.KioskCode = body.KioskCode
	cart.BillingCode = "NONE"

	err := cart.FindFirst()

	// no cart
	if err != nil {
		// create cart
		if err := cart.Create(); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	// create cart item
	cartItem := kiosk_cart.CartItem{}

	// validate item code by group
	if body.KioskType == constants.GROUP_FB {
		fb := model_service.FoodBeverage{}
		fb.PartnerUid = prof.PartnerUid
		fb.CourseUid = prof.CourseUid
		fb.FBCode = body.ItemCode

		if err := fb.FindFirstInKiosk(kiosk.Id); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
		// add infor cart item
		cartItem.ItemGroupId = fb.GroupCode
		cartItem.ItemName = fb.VieName
		cartItem.UnitPrice = fb.Price
	}

	if body.KioskType == constants.GROUP_PROSHOP {
		proshop := model_service.Proshop{}
		proshop.PartnerUid = prof.PartnerUid
		proshop.CourseUid = prof.CourseUid
		proshop.ProShopId = body.ItemCode

		if err := proshop.FindFirst(); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
		// add infor cart item
		cartItem.ItemGroupId = proshop.GroupCode
		cartItem.ItemName = proshop.VieName
		cartItem.UnitPrice = proshop.Price
	}

	if body.KioskType == constants.GROUP_RENTAL {
		rental := model_service.Rental{}
		rental.PartnerUid = prof.PartnerUid
		rental.CourseUid = prof.CourseUid
		rental.RentalId = body.ItemCode

		if err := rental.FindFirst(); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
		// add infor cart item
		cartItem.ItemGroupId = rental.GroupCode
		cartItem.ItemName = rental.VieName
		cartItem.UnitPrice = rental.Price
	}

	// validate quantity
	inventory := kiosk_inventory.InventoryItem{}
	inventory.PartnerUid = prof.PartnerUid
	inventory.CourseUid = prof.CourseUid
	inventory.KioskCode = strconv.Itoa(int(body.KioskCode))
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

	// add infor cart item
	cartItem.PartnerUid = prof.PartnerUid
	cartItem.CourseUid = prof.CourseUid
	cartItem.KioskCode = body.KioskCode
	cartItem.KioskCartId = cart.Id
	cartItem.KioskCartCode = cart.Code
	cartItem.ItemCode = body.ItemCode
	cartItem.Quantity = body.Quantity
	cartItem.TotalPrice = float64(cartItem.Quantity) * cartItem.UnitPrice
	cartItem.ActionBy = prof.Uid
	if err := cartItem.Create(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ CKioskCart) AddDiscountToItem(c *gin.Context, prof models.CmsUser) {
	var body request.AddDiscountToKioskItemBody
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
	cartItem := kiosk_cart.CartItem{}
	cartItem.PartnerUid = prof.PartnerUid
	cartItem.CourseUid = prof.CourseUid
	cartItem.Id = body.CartItemId

	if err := cartItem.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate cart
	cart := kiosk_cart.Cart{}
	cart.Id = cartItem.KioskCartId
	cart.PartnerUid = prof.PartnerUid
	cart.CourseUid = prof.CourseUid
	cart.BillingCode = "NONE"

	if err := cart.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	cartItem.DiscountType = body.DiscountType
	cartItem.DiscountPrice = body.DiscountPrice
	cartItem.DiscountReason = body.DiscountReason

	if err := cartItem.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ CKioskCart) GetItemInCart(c *gin.Context, prof models.CmsUser) {
	query := request.GetItemInKioskCartBody{}
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

	cart := kiosk_cart.Cart{}
	cart.PartnerUid = prof.PartnerUid
	cart.CourseUid = prof.CourseUid
	cart.KioskCode = query.KioskCode
	cart.GolfBag = query.GolfBag
	cart.BookingDate = datatypes.Date(bookingDate)

	if err := cart.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	cartItem := kiosk_cart.CartItem{}
	cartItem.PartnerUid = prof.PartnerUid
	cartItem.CourseUid = prof.CourseUid
	cartItem.KioskCartId = cart.Id

	list, total, err := cartItem.FindList(page)

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

func (_ CKioskCart) GetBestItemInKiosk(c *gin.Context, prof models.CmsUser) {
	query := request.GetBestItemInKioskBody{}
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

	cartItem := kiosk_cart.CartItem{}
	cartItem.PartnerUid = prof.PartnerUid
	cartItem.CourseUid = prof.CourseUid
	cartItem.KioskCode = query.KioskCode
	cartItem.ItemGroupId = query.GroupCode

	list, total, err := cartItem.FindBestCartItem(page)

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

func (_ CKioskCart) GetListCart(c *gin.Context, prof models.CmsUser) {
	query := request.GetCartInKioskBody{}
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

	cart := kiosk_cart.Cart{}
	cart.PartnerUid = prof.PartnerUid
	cart.CourseUid = prof.CourseUid
	cart.KioskCode = query.KioskCode
	cart.BookingDate = datatypes.Date(bookingDate)

	list, total, err := cart.FindList(page)

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

func (_ CKioskCart) UpdateItemCart(c *gin.Context, prof models.CmsUser) {
	var body request.UpdateQuantityToKioskCartBody
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
	cartItem := kiosk_cart.CartItem{}
	cartItem.PartnerUid = prof.PartnerUid
	cartItem.CourseUid = prof.CourseUid
	cartItem.Id = body.CartItemId

	if err := cartItem.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate cart
	cart := kiosk_cart.Cart{}
	cart.PartnerUid = prof.PartnerUid
	cart.CourseUid = prof.CourseUid
	cart.Id = cartItem.KioskCartId
	cart.BillingCode = "NONE"

	if err := cart.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate quantity
	inventory := kiosk_inventory.InventoryItem{}
	inventory.PartnerUid = prof.PartnerUid
	inventory.CourseUid = prof.CourseUid
	inventory.KioskCode = strconv.Itoa(int(cartItem.KioskCode))
	inventory.Code = cartItem.ItemCode

	if err := inventory.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Kiểm tra số lượng hàng tồn trong kho
	if body.Quantity > inventory.Quantity+cartItem.Quantity {
		response_message.BadRequest(c, "The quantity of goods in stock is not enough")
		return
	}

	// Update số lượng hàng tồn trong kho
	inventory.Quantity = inventory.Quantity + cartItem.Quantity - body.Quantity
	if err := inventory.Update(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	cartItem.Quantity = body.Quantity
	cartItem.Note = body.Note

	if err := cartItem.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ CKioskCart) DeleteItemInCart(c *gin.Context, prof models.CmsUser) {
	var body request.DeleteItemInKioskCartBody

	if err := c.BindJSON(&body); err != nil {
		log.Print("DeleteItemInCart BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	// validate cart item
	cartItem := kiosk_cart.CartItem{}
	cartItem.PartnerUid = prof.PartnerUid
	cartItem.CourseUid = prof.CourseUid
	cartItem.Id = body.CartItemId

	if err := cartItem.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate quantity
	inventory := kiosk_inventory.InventoryItem{}
	inventory.PartnerUid = prof.PartnerUid
	inventory.CourseUid = prof.CourseUid
	inventory.KioskCode = strconv.Itoa(int(cartItem.KioskCode))
	inventory.Code = cartItem.ItemCode

	if err := inventory.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Update số lượng hàng tồn trong kho
	inventory.Quantity += cartItem.Quantity
	if err := inventory.Update(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate cart
	cart := kiosk_cart.Cart{}
	cart.PartnerUid = prof.PartnerUid
	cart.CourseUid = prof.CourseUid
	cart.Id = cartItem.KioskCartId
	cart.BillingCode = "NONE"

	if err := cart.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	if err := cartItem.Delete(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}

func (_ CKioskCart) CreateBilling(c *gin.Context, prof models.CmsUser) {
	var body request.CreateKioskBillingBody
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

	cart := kiosk_cart.Cart{}
	cart.PartnerUid = prof.PartnerUid
	cart.CourseUid = prof.CourseUid
	cart.KioskCode = body.KioskCode
	cart.GolfBag = body.GolfBag
	cart.BookingDate = datatypes.Date(time.Now().UTC())

	if err := cart.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	cart.BillingCode = "KIOSK-BILLING-" + time.Now().Format("20060102150405")

	if err := cart.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ CKioskCart) MoveItemToOtherCart(c *gin.Context, prof models.CmsUser) {
	var body request.MoveItemToOtherKioskCartBody
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
	sourceCart := kiosk_cart.Cart{}
	sourceCart.Code = body.CartCode
	sourceCart.PartnerUid = prof.PartnerUid
	sourceCart.CourseUid = prof.CourseUid
	sourceCart.BillingCode = "NONE"

	if err := sourceCart.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate cart by golf bag
	targetCart := kiosk_cart.Cart{}
	targetCart.PartnerUid = prof.PartnerUid
	targetCart.CourseUid = prof.CourseUid
	targetCart.GolfBag = body.GolfBag
	targetCart.BookingDate = sourceCart.BookingDate
	targetCart.KioskCode = sourceCart.KioskCode
	targetCart.BillingCode = "NONE"

	err := targetCart.FindFirst()

	// no cart
	if err != nil {
		// create cart
		if err := targetCart.Create(); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	hasError := false

	var errFor error

	for _, cartItemId := range body.CartItemIdList {
		cartItemTemp := kiosk_cart.CartItem{}
		cartItemTemp.Id = cartItemId
		cartItemTemp.KioskCartId = sourceCart.Id
		cartItemTemp.KioskCartCode = sourceCart.Code

		if err := cartItemTemp.FindFirst(); err != nil {
			continue
		}

		cartItemTemp.KioskCartId = targetCart.Id
		cartItemTemp.KioskCartCode = targetCart.Code

		if errFor = cartItemTemp.Update(); errFor != nil {
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
