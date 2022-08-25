package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/datatypes"
	"log"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	model_booking "start/models/booking"
	kiosk_cart "start/models/kiosk-cart"
	model_service "start/models/service"
	"start/utils/response_message"
	"time"
)

type CKioskCart struct {}

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
	kiosk.KioskCode = body.KioskCode
	if err := kiosk.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate item code
	fb := model_service.FoodBeverage{}
	fb.PartnerUid = prof.PartnerUid
	fb.CourseUid = prof.CourseUid
	fb.FBCode = body.ItemCode

	if err := fb.FindFirstInKiosk(kiosk.Id); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// check cart
	cart := kiosk_cart.Cart{}
	cart.PartnerUid = prof.PartnerUid
	cart.CourseUid = prof.CourseUid
	cart.GolfBag = body.GolfBag
	cart.BookingDate = datatypes.Date(time.Now())
	cart.KioskCode = body.KioskCode
	cart.BillingCode = ""

	err := cart.FindFirst()

	// no cart
	if err != nil {
		// create cart
		if err := cart.Create(); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	// add item
	cartItem := kiosk_cart.CartItem{}
	cartItem.PartnerUid = prof.PartnerUid
	cartItem.CourseUid = prof.CourseUid
	cartItem.KioskCode = body.KioskCode
	cartItem.KioskCartId = cart.Id
	cartItem.KioskCartCode = cart.Code
	cartItem.ItemCode = body.ItemCode
	cartItem.Quantity = body.Quantity
	cartItem.UnitPrice = fb.Price
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
	cart.BillingCode = ""

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

func (_ CKioskCart) UpdateQuantityToCart(c *gin.Context, prof models.CmsUser) {
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
	cart.BillingCode = ""

	if err := cart.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	cartItem.Quantity = body.Quantity

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

	// validate cart
	cart := kiosk_cart.Cart{}
	cart.PartnerUid = prof.PartnerUid
	cart.CourseUid = prof.CourseUid
	cart.Id = cartItem.KioskCartId
	cart.BillingCode = ""

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

	bookingDate, _ := time.Parse("2006-01-02", body.BookingDate)

	cart := kiosk_cart.Cart{}
	cart.PartnerUid = prof.PartnerUid
	cart.CourseUid = prof.CourseUid
	cart.KioskCode = body.KioskCode
	cart.GolfBag = body.GolfBag
	cart.BookingDate = datatypes.Date(bookingDate)

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

	// validate cart code
	sourceCart := kiosk_cart.Cart{}
	sourceCart.Code = body.CartCode
	sourceCart.PartnerUid = prof.PartnerUid
	sourceCart.CourseUid = prof.CourseUid
	sourceCart.BillingCode = ""

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
	targetCart.BillingCode = ""

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