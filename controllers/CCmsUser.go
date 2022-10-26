package controllers

import (
	"errors"
	"log"
	"start/auth"
	"start/config"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_role "start/models/role"
	"start/utils"
	"start/utils/response_message"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type CCmsUser struct{}

func (_ *CCmsUser) Test1(c *gin.Context) {
	okResponse(c, gin.H{"message": "success"})
}

func (_ *CCmsUser) Test(c *gin.Context, prof models.CmsUser) {
	log.Println("test")

	okResponse(c, gin.H{"message": "success"})
}

func (_ *CCmsUser) Login(c *gin.Context) {
	body := request.LoginBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Ttl <= 0 {
		body.Ttl = 604800 // 1 Tuần
	}

	user := models.CmsUser{
		UserName: body.UserName,
	}

	errFind := user.FindFirst()
	if errFind != nil {
		response_message.InternalServerError(c, errFind.Error())
		return
	}

	if user.Status != constants.STATUS_ENABLE {
		response_message.BadRequestDynamicKey(c, "USER_BE_LOCKED", errors.New("account be locked").Error())
		return
	}

	if user.LoggedIn {
		errCheck := utils.ComparePassword(user.Password, body.Password)
		if errCheck != nil {
			response_message.BadRequest(c, errFind.Error())
			return
		}
	} else {
		//TODO: check password
		if user.Password != "" && user.Password != body.Password {
			response_message.BadRequest(c, errors.New("wrong info").Error())
			return
		}
		hashPass, errHash := utils.GeneratePassword(body.Password)
		if errHash != nil {
			response_message.BadRequest(c, errHash.Error())
			return
		}
		user.Password = hashPass
		user.LoggedIn = true
		errUpdate := user.Update()
		if errUpdate != nil {
			response_message.InternalServerError(c, errUpdate.Error())
			return
		}
	}

	partner := models.Partner{}
	partner.Uid = user.PartnerUid
	errFind = partner.FindFirst()
	if errFind != nil {
		response_message.InternalServerError(c, errFind.Error())
		return
	}

	// // create jwt
	prof := models.CmsUserProfile{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * time.Duration(body.Ttl)).Unix(),
		},
	}
	prof.Uid = user.Uid
	prof.PartnerUid = user.PartnerUid
	prof.CourseUid = user.CourseUid
	prof.UserName = user.UserName
	prof.Status = user.Status

	jwt, errJwt := auth.CreateToken(prof, config.GetJwtSecret())
	if errJwt != nil {
		log.Println("cms login errJwt ", errJwt.Error())
		response_message.InternalServerError(c, errFind.Error())
		return
	}

	datasources.SetCacheJwt(user.Model.Uid, jwt, int64(body.Ttl))

	userToken := models.CmsUserToken{
		UserUid:    user.Uid,
		UserName:   user.UserName,
		PartnerUid: user.PartnerUid,
		CourseUid:  user.CourseUid,
		Token:      jwt,
	}
	errCreate := userToken.Create()
	if errCreate != nil {
		log.Println("cmsUserToken.Create: ", errCreate)
	}

	courseInfo := models.Course{}
	courseInfo.Uid = user.CourseUid
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	errFindCourse := courseInfo.FindFirst(db)
	if errFindCourse != nil {
		response_message.BadRequest(c, errFindCourse.Error())
		return
	}

	userDataRes := map[string]interface{}{
		"user_name":   user.UserName,
		"phone":       user.Phone,
		"partner_uid": user.PartnerUid,
		"course_uid":  user.CourseUid,
		"course_info": courseInfo,
	}

	okResponse(c, gin.H{"token": jwt, "data": userDataRes})
}

func (_ *CCmsUser) GetList(c *gin.Context, prof models.CmsUser) {
	form := request.GetListCmsUserForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	cmsUserR := models.CmsUser{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		UserName:   form.UserName,
	}
	list, total, err := cmsUserR.FindList(page)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total": total,
		"data":  list,
	}

	okResponse(c, res)
}

func (_ *CCmsUser) CreateCmsUser(c *gin.Context) {
	body := request.CreateCmsUserBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	partner := models.Partner{}
	partner.Uid = body.PartnerUid
	errFind := partner.FindFirst()
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	//Find Role
	if body.RoleId > 0 {
		role := model_role.Role{}
		role.Id = body.RoleId
		errFR := role.FindFirst()
		if errFR != nil {
			response_message.BadRequest(c, errFR.Error())
			return
		}
	}

	cmsUser := models.CmsUser{
		UserName:   body.UserName,
		FullName:   body.FullName,
		Email:      body.Email,
		Phone:      body.Phone,
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		RoleId:     body.RoleId,
	}

	errCreate := cmsUser.Create()
	if errCreate != nil {
		response_message.InternalServerError(c, errCreate.Error())
		return
	}

	okResponse(c, cmsUser)
}
