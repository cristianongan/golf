package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	"start/utils/response_message"
	"strconv"
	"time"
)

type CDeposit struct{}

func (_ *CDeposit) CreateDeposit(c *gin.Context, prof models.CmsUser) {
	var body request.CreateDepositBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "")
		return
	}

	validate := validator.New()
	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	var customer models.CustomerUser

	if body.CustomerUid != "" {
		// validate customer_uid
		customer = models.CustomerUser{}
		customer.Uid = body.CustomerUid
		if err := customer.FindFirst(); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}
	} else {
		customer = models.CustomerUser{}
		customer.PartnerUid = prof.PartnerUid
		customer.CourseUid = prof.CourseUid
		customer.Phone = body.CustomerPhone
		if err := customer.FindFirst(); err == nil {
			response_message.BadRequest(c, errors.New("phone number is exist").Error())
			return
		} else {
			customer.Name = body.CustomerName
			customer.Identify = body.CustomerIdentity
			customer.Type = "VISITOR"
			if err := customer.Create(); err != nil {
				response_message.BadRequest(c, errors.New("phone number is exist").Error())
				return
			}
		}
	}

	inputDate, _ := time.Parse("2006-01-02", body.InputDate)

	deposit := models.Deposit{}
	deposit.PartnerUid = prof.PartnerUid
	deposit.CourseUid = prof.CourseUid
	deposit.InputDate = datatypes.Date(inputDate)
	deposit.CustomerUid = customer.Uid
	deposit.CustomerName = customer.Name
	deposit.CustomerIdentity = customer.Identify
	deposit.CustomerPhone = customer.Phone
	deposit.CustomerType = customer.Type
	deposit.PaymentType = body.PaymentType
	deposit.TmCurrency = body.TmCurrency
	deposit.TmRate = body.TmRate
	deposit.TmOriginAmount = body.TmOriginAmount
	deposit.TmTotalAmount = body.TmRate * float64(body.TmOriginAmount)
	deposit.TcCurrency = body.TcCurrency
	deposit.TcRate = body.TcRate
	deposit.TcOriginAmount = body.TcOriginAmount
	deposit.TcTotalAmount = body.TcRate * float64(body.TcOriginAmount)
	deposit.TotalAmount = deposit.TmTotalAmount + deposit.TcTotalAmount
	deposit.Note = body.Note

	if err := deposit.Create(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	c.JSON(200, deposit)
}

func (_ *CDeposit) GetDeposit(c *gin.Context, prof models.CmsUser) {
	query := request.GetDepositList{}
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

	deposit := models.DepositList{}
	deposit.CustomerIdentity = query.CustomerIdentity
	deposit.CustomerPhone = query.CustomerPhone
	deposit.CustomerStyle = query.CustomerStyle
	deposit.InputDate = query.InputDate

	list, total, err := deposit.FindList(page)

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

func (_ *CDeposit) UpdateDeposit(c *gin.Context, prof models.CmsUser) {
	var body request.UpdateDepositBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "")
		return
	}

	validate := validator.New()
	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate deposit_uid
	deposit := models.Deposit{}
	deposit.CustomerUid = body.CustomerUid
	deposit.Id, _ = strconv.ParseInt(c.Param("id"), 10, 64)

	if err := deposit.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	deposit.PaymentType = body.PaymentType
	deposit.TmCurrency = body.TmCurrency
	deposit.TmRate = body.TmRate
	deposit.TmOriginAmount = body.TmOriginAmount
	deposit.TmTotalAmount = body.TmRate * float64(body.TmOriginAmount)
	deposit.TcCurrency = body.TcCurrency
	deposit.TcRate = body.TcRate
	deposit.TcOriginAmount = body.TcOriginAmount
	deposit.TcTotalAmount = body.TcRate * float64(body.TcOriginAmount)
	deposit.TotalAmount = deposit.TmTotalAmount + deposit.TcTotalAmount
	deposit.Note = body.Note

	if err := deposit.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	c.JSON(200, deposit)
}
