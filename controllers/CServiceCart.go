package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	kiosk_inventory "start/models/kiosk-inventory"
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
	booking := model_booking.Booking{}
	booking.PartnerUid = body.PartnerUid
	booking.CourseUid = body.CourseUid
	booking.Bag = body.GolfBag
	booking.BookingDate = time.Now().Format("02/01/2006")
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Booking "+err.Error())
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
		response_message.BadRequest(c, "Kiosk "+err.Error())
		return
	}

	// validate quantity
	inventory := kiosk_inventory.InventoryItem{}
	inventory.PartnerUid = body.PartnerUid
	inventory.CourseUid = body.CourseUid
	inventory.ServiceId = body.ServiceId
	inventory.Code = body.ItemCode

	if err := inventory.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Inventory "+err.Error())
		return
	}

	// Kiểm tra số lượng hàng tồn trong kho
	if body.Quantity > inventory.Quantity {
		response_message.BadRequest(c, "The quantity of goods in stock is not enough")
		return
	}

	// Update số lượng hàng tồn trong kho
	inventory.Quantity -= body.Quantity
	if err := inventory.Update(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

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
		serviceCartItem.Type = kiosk.KioskType
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
		serviceCartItem.Type = kiosk.KioskType
		serviceCartItem.GroupCode = proshop.GroupCode
		serviceCartItem.Name = proshop.VieName
		serviceCartItem.EngName = proshop.EnglishName
		serviceCartItem.UnitPrice = int64(proshop.Price)
		serviceCartItem.Unit = proshop.Unit
	}

	if kiosk.ServiceType == constants.GROUP_RENTAL {
		rental := model_service.Rental{}
		rental.PartnerUid = body.PartnerUid
		rental.CourseUid = body.CourseUid
		rental.RentalId = body.ItemCode

		if err := rental.FindFirst(db); err != nil {
			response_message.BadRequest(c, "Rental "+err.Error())
			return
		}
		// add infor cart item
		serviceCartItem.Type = kiosk.KioskType
		serviceCartItem.GroupCode = rental.GroupCode
		serviceCartItem.Name = rental.VieName
		serviceCartItem.EngName = rental.EnglishName
		serviceCartItem.UnitPrice = int64(rental.Price)
		serviceCartItem.Unit = rental.Unit
	}

	// check service cart
	serviceCart := models.ServiceCart{}
	serviceCart.PartnerUid = body.PartnerUid
	serviceCart.CourseUid = body.CourseUid
	serviceCart.GolfBag = body.GolfBag
	serviceCart.BookingUid = booking.Uid
	serviceCart.BookingDate = datatypes.Date(time.Now().UTC())
	serviceCart.ServiceId = body.ServiceId

	if body.BillId != 0 {
		serviceCart.Id = body.BillId
	} else {
		serviceCart.BillCode = constants.BILL_NONE
		serviceCart.StaffOrder = prof.UserName
		serviceCart.BillStatus = constants.POS_BILL_STATUS_PENDING
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
		// Kiểm tra trạng thái bill
		if serviceCart.BillStatus != constants.POS_BILL_STATUS_PENDING {
			response_message.BadRequest(c, "Bill status invalid")
			return
		}
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
	serviceCartItem.Id = body.CartItemId

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

	serviceCartItem.DiscountType = body.DiscountType
	serviceCartItem.DiscountValue = body.DiscountPrice
	serviceCartItem.DiscountReason = body.DiscountReason

	if err := serviceCartItem.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
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

	// validate cart item
	serviceCartItem := model_booking.BookingServiceItem{}
	serviceCartItem.Id = body.CartItemId

	if err := serviceCartItem.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate golf bag
	booking := model_booking.Booking{}
	booking.PartnerUid = body.PartnerUid
	booking.CourseUid = body.CourseUid
	booking.Bag = serviceCartItem.Bag
	booking.BookingDate = time.Now().Format("02/01/2006")
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Booking "+err.Error())
		return
	}

	// Check bag status
	if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		response_message.BadRequest(c, "Bag check out")
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
	if serviceCart.BillStatus != constants.POS_BILL_STATUS_PENDING {
		response_message.BadRequest(c, "Bill status invalid")
		return
	}

	// validate quantity
	inventory := kiosk_inventory.InventoryItem{}
	inventory.PartnerUid = body.PartnerUid
	inventory.CourseUid = body.CourseUid
	inventory.ServiceId = serviceCart.ServiceId
	inventory.Code = serviceCartItem.ItemCode

	if err := inventory.FindFirst(db); err != nil {
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
	if err := inventory.Update(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// update service cart
	serviceCart.Amount += (body.Quantity * serviceCartItem.UnitPrice) - (int64(serviceCartItem.Quality) * serviceCartItem.UnitPrice)
	if err := serviceCart.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// update service item
	serviceCartItem.Quality = int(body.Quantity)
	serviceCartItem.Amount = body.Quantity * serviceCartItem.UnitPrice
	serviceCartItem.Input = body.Note

	if err := serviceCartItem.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
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

	if err := serviceCartItem.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate golf bag
	booking := model_booking.Booking{}
	booking.PartnerUid = serviceCartItem.PartnerUid
	booking.CourseUid = serviceCartItem.CourseUid
	booking.Bag = serviceCartItem.Bag
	booking.BookingDate = time.Now().Format("02/01/2006")
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Booking "+err.Error())
		return
	}

	// validate cart
	serviceCart := models.ServiceCart{}
	serviceCart.Id = serviceCartItem.ServiceBill

	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate quantity
	inventory := kiosk_inventory.InventoryItem{}
	inventory.PartnerUid = serviceCartItem.PartnerUid
	inventory.CourseUid = serviceCartItem.CourseUid
	inventory.ServiceId = serviceCart.ServiceId
	inventory.Code = serviceCartItem.ItemCode

	if err := inventory.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Update số lượng hàng tồn trong kho
	inventory.Quantity += int64(serviceCartItem.Quality)
	if err := inventory.Update(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

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

	okRes(c)
}

func (_ CServiceCart) CreateBilling(c *gin.Context, prof models.CmsUser) {
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
	booking := model_booking.Booking{}
	booking.PartnerUid = body.PartnerUid
	booking.CourseUid = body.CourseUid
	booking.Bag = body.GolfBag
	booking.BookingDate = time.Now().Format("02/01/2006")
	if err := booking.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	serviceCart := models.ServiceCart{}
	serviceCart.PartnerUid = body.PartnerUid
	serviceCart.CourseUid = body.CourseUid
	serviceCart.ServiceId = body.ServiceId
	serviceCart.GolfBag = body.GolfBag
	serviceCart.BookingDate = datatypes.Date(time.Now().UTC())
	serviceCart.BillCode = constants.BILL_NONE

	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	serviceCart.BillCode = time.Now().Format("20060102150405")

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
	booking := model_booking.Booking{}
	booking.Bag = body.GolfBag
	booking.BookingDate = time.Now().Format("02/01/2006")
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
	bookingSourse := model_booking.Booking{}
	bookingSourse.Bag = sourceServiceCart.GolfBag
	bookingSourse.BookingDate = time.Now().Format("02/01/2006")
	if err := bookingSourse.FindFirst(db); err != nil {
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
	targetServiceCart.BillCode = constants.BILL_NONE
	targetServiceCart.BillStatus = constants.POS_BILL_STATUS_PENDING

	err := targetServiceCart.FindFirst(db)

	// no cart
	if err != nil {
		// create cart
		targetServiceCart.BookingUid = booking.Uid
		targetServiceCart.StaffOrder = prof.UserName

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
		totalAmount += (serviceCartItem.Amount - serviceCartItem.DiscountValue)

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
	if err := targetServiceCart.Update(db); err != nil {
		response_message.InternalServerError(c, "Update target cart "+err.Error())
		return
	}

	// Update amount target bill
	sourceServiceCart.Amount = sourceServiceCart.Amount - totalAmount

	if err := sourceServiceCart.Update(db); err != nil {
		response_message.InternalServerError(c, "Update target cart "+err.Error())
		return
	}

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

	serviceCart.BillStatus = constants.POS_BILL_STATUS_OUT

	if err := serviceCart.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

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
		cKioskOutputInventory.MethodOutputBill(c, prof, bodyOutputBill, constants.KIOSK_BILL_INVENTORY_SELL, code)
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
		booking := model_booking.Booking{}
		booking.Bag = bagClone
		booking.BookingDate = time.Now().Format("02/01/2006")
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

	// Create booking
	booking := model_booking.Booking{
		PartnerUid:   body.PartnerUid,
		CourseUid:    body.CourseUid,
		Bag:          bagClone,
		BookingDate:  time.Now().Format("02/01/2006"),
		BagStatus:    constants.BAG_STATUS_WAITING,
		InitType:     constants.BOOKING_INIT_TYPE_CHECKIN,
		CheckInTime:  time.Now().Unix(),
		CustomerName: body.GuestName,
		CustomerType: constants.CUSTOMER_TYPE_NONE_GOLF,
	}

	errC := booking.Create(db, bUid)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

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
	if err := serviceCart.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Find service Cart "+err.Error())
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

	// Update trạng thái
	serviceCart.BillStatus = constants.POS_BILL_STATUS_ACTIVE
	if err := serviceCart.Update(db); err != nil {
		response_message.BadRequest(c, "Update service Cart "+err.Error())
		return
	}

	//Update lại giá trong booking
	updatePriceWithServiceItem(booking, prof)

	okRes(c)
}
