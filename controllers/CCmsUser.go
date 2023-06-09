package controllers

import (
	"encoding/json"
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
	"strconv"
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

func CheckLoginFailedManyTime(user *models.CmsUser) {
	redisLoginKey := datasources.GetRedisKeyUserLogin(user.UserName)
	countLogin, errRedis := datasources.GetCache(redisLoginKey)

	if errRedis != nil {
		datasources.SetCache(redisLoginKey, "1", 10*60)
	} else {
		i, err := strconv.Atoi(countLogin)
		if err != nil {
			// panic(err)
			print(err)
		} else {
			i++
			datasources.SetCache(redisLoginKey, strconv.Itoa(i), 10*60)
			if i >= 5 {
				user.Status = constants.STATUS_DISABLE
				user.Update()
			}
		}
	}
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

	redisLoginKey := datasources.GetRedisKeyUserLogin(body.UserName)
	countLogin, errRedis := datasources.GetCache(redisLoginKey)
	if errRedis == nil {
		i, err := strconv.Atoi(countLogin)
		if err == nil && i >= 5 {
			response_message.BadRequestDynamicKey(c, "USER_BE_LOCKED", errors.New("account be locked").Error())
			return
		}
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
		passw, errDec := utils.DecryptAES([]byte(config.GetPassSecretKey()), body.Password)
		if errDec != nil {
			response_message.BadRequest(c, errDec.Error())
			return
		}
		errCheck := utils.ComparePassword(user.Password, passw)
		if errCheck != nil {
			CheckLoginFailedManyTime(&user)
			response_message.BadRequest(c, "login failse")
			return
		}
	} else {
		//TODO: check password
		if user.Password != "" && user.Password != body.Password {
			CheckLoginFailedManyTime(&user)
			response_message.BadRequest(c, errors.New("wrong info").Error())
			return
		}

		passw, errDec := utils.DecryptAES([]byte(config.GetPassSecretKey()), body.Password)
		if errDec != nil {
			response_message.BadRequest(c, errDec.Error())
			return
		}

		hashPass, errHash := utils.GeneratePassword(passw)
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
			ExpiresAt: utils.GetTimeNow().Add(time.Second * time.Duration(body.Ttl)).Unix(),
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

	key := datasources.GetRedisKeyUserLogin(user.UserName)
	datasources.DelCacheByKey(key)

	courseInfo := models.Course{}
	if user.CourseUid != "" {
		courseInfo.Uid = user.CourseUid
		errFindCourse := courseInfo.FindFirst()
		if errFindCourse != nil {
			response_message.BadRequest(c, errFindCourse.Error())
			return
		}
	}

	// Find Role
	// Find Permission
	listPerMis := utils.ListString{}
	role := model_role.Role{}
	if user.RoleId > 0 {
		role.Id = user.RoleId

		key := user.GetKeyRedisPermission()
		listPer, _ := datasources.GetCache(key)

		if !config.GetUseRedisPermissionRole() {
			listPer = ""
		}

		if len(listPer) > 0 {
			_ = json.Unmarshal([]byte(listPer), &listPerMis)
		} else {
			errFR := role.FindFirst()
			if errFR == nil {
				rolePR := model_role.RolePermission{
					RoleId: role.Id,
				}
				listPermission, errRolePR := rolePR.FindAll()
				if errRolePR == nil {
					for _, v := range listPermission {
						listPerMis = append(listPerMis, v.PermissionUid)
					}
				}
			}
			user.SaveKeyRedisPermission(listPerMis)
		}

	} else {
		if user.RoleId == -1 {
			// Root Account
			role.Id = user.RoleId
			errFR := role.FindFirst()
			if errFR == nil {
				permis := model_role.Permission{}
				listP, errLP := permis.FindAll()
				if errLP == nil {
					for _, v := range listP {
						listPerMis = append(listPerMis, v.Uid)
					}
				}
			}
		}
	}

	if user.CaddieId > 0 {
		caddie := models.Caddie{}
		caddie.Id = user.CaddieId
		db := datasources.GetDatabaseWithPartner(user.PartnerUid)
		errFCd := caddie.FindFirst(db)
		if errFCd == nil {
			userDataRes := map[string]interface{}{
				"user_name":   user.UserName,
				"phone":       user.Phone,
				"partner_uid": user.PartnerUid,
				"course_uid":  user.CourseUid,
				"course_info": courseInfo,
				"role_name":   role.Name,
				"role_id":     user.RoleId,
				"permissions": listPerMis,
				"caddie_info": caddie,
			}

			okResponse(c, gin.H{"token": jwt, "data": userDataRes})
			return
		}
	}

	userDataRes := map[string]interface{}{
		"user_name":   user.UserName,
		"full_name":   user.FullName,
		"phone":       user.Phone,
		"partner_uid": user.PartnerUid,
		"course_uid":  user.CourseUid,
		"course_info": courseInfo,
		"role_name":   role.Name,
		"role_id":     user.RoleId,
		"permissions": listPerMis,
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
		Type:       form.Type,
		RoleId:     form.RoleId,
	}

	subRoles, err := model_role.GetAllSubRoleUids(int(prof.RoleId))
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	isRootUser := false
	if prof.RoleId == -1 {
		isRootUser = true
	}

	list, total, err := cmsUserR.FindList(page, form.Search, subRoles, isRootUser)
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

func (_ *CCmsUser) CreateCmsUser(c *gin.Context, prof models.CmsUser) {
	body := request.CreateCmsUserBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	passw, errDec := utils.DecryptAES([]byte(config.GetPassSecretKey()), body.Password)

	if errDec != nil {
		response_message.BadRequest(c, errDec.Error())
		return
	}

	log.Println("CreateCmsUser descypt pass", passw)

	//verify password
	eightOrMore, number, upper, special := utils.VerifyPassword(passw)
	if !eightOrMore || !number || !upper || !special {
		response_message.BadRequestDynamicKey(c, "USER_VALIDATE_PASSWORD_POLICY", "")
		return
	}

	if checkStringInArray(config.GetBlacklistPass(), body.Password) {
		response_message.BadRequestDynamicKey(c, "USER_VALIDATE_PASSWORD_WEEK", "")
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
	roleType := constants.ROLE_TYPE_CMS
	if body.RoleId > 0 {
		role := model_role.Role{}
		role.Id = body.RoleId
		errFR := role.FindFirst()
		if errFR != nil {
			response_message.BadRequest(c, errFR.Error())
			return
		} else {
			roleType = role.Type
		}
	}

	//Find Caddie
	if body.CaddieId > 0 {
		db := datasources.GetDatabaseWithPartner(body.PartnerUid)
		caddie := models.Caddie{}
		caddie.Id = body.CaddieId
		errFCAD := caddie.FindFirst(db)
		if errFCAD != nil {
			response_message.BadRequest(c, errFCAD.Error())
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
		CaddieId:   body.CaddieId,
		Type:       roleType,
	}

	hashPass, errHash := utils.GeneratePassword(passw)
	if errHash != nil {
		response_message.BadRequest(c, errHash.Error())
		return
	}
	cmsUser.Password = hashPass
	cmsUser.LoggedIn = true

	errCreate := cmsUser.Create()
	if errCreate != nil {
		response_message.InternalServerError(c, errCreate.Error())
		return
	}

	okResponse(c, cmsUser)
}

/*
Update Cms User
*/
func (_ *CCmsUser) UpdateCmsUser(c *gin.Context, prof models.CmsUser) {
	userUidStr := c.Param("uid")
	if userUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	cmsUser := models.CmsUser{}
	cmsUser.Uid = userUidStr
	errF := cmsUser.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := request.UdpCmsUserBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.FullName != "" {
		cmsUser.FullName = body.FullName
	}
	if body.Phone != "" {
		cmsUser.Phone = body.Phone
	}
	if body.Email != "" {
		cmsUser.Email = body.Email
	}
	if body.RoleId > 0 {
		//Find Role
		roleType := constants.ROLE_TYPE_CMS
		if body.RoleId > 0 {
			role := model_role.Role{}
			role.Id = body.RoleId
			errFR := role.FindFirst()
			if errFR != nil {
				response_message.BadRequest(c, errFR.Error())
				return
			} else {
				roleType = role.Type
			}
		}
		cmsUser.RoleId = body.RoleId
		cmsUser.Type = roleType
	}
	if body.Status != "" {
		cmsUser.Status = body.Status
	}

	errUdp := cmsUser.Update()

	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	if body.RoleId > 0 {
		role := model_role.Role{}
		role.Id = cmsUser.RoleId
		errFR := role.FindFirst()
		listPerMis := utils.ListString{}
		if errFR == nil {
			rolePR := model_role.RolePermission{
				RoleId: role.Id,
			}
			listPermission, errRolePR := rolePR.FindAll()
			if errRolePR == nil {
				for _, v := range listPermission {
					listPerMis = append(listPerMis, v.PermissionUid)
				}
			}
		}
		cmsUser.SaveKeyRedisPermission(listPerMis)
	}

	okResponse(c, cmsUser)
}

/*
Delete Role
*/
func (_ *CCmsUser) DeleteCmsUser(c *gin.Context, prof models.CmsUser) {
	userUidStr := c.Param("uid")
	if userUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	cmsUser := models.CmsUser{}
	cmsUser.Uid = userUidStr
	errF := cmsUser.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	key := cmsUser.GetKeyRedisPermission()
	datasources.DelCacheByKey(key)

	errDel := cmsUser.Delete()
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

func (_ *CCmsUser) EnableCmsUser(c *gin.Context) {
	user := models.CmsUser{}
	listUserLocked, _, _ := user.FindUserLocked()
	for _, v := range listUserLocked {
		redisLoginKey := datasources.GetRedisKeyUserLogin(v.UserName)
		datasources.DelCacheByKey(redisLoginKey)
		v.Status = constants.STATUS_ENABLE
		v.Update()
	}
	okRes(c)
}

/*
Log out
*/
func (_ *CCmsUser) LogOut(c *gin.Context, prof models.CmsUser) {

	datasources.DelCacheJwt(prof.Uid)

	okRes(c)
}

/*
Change pass Cms User
*/
func (_ *CCmsUser) ChangePassCmsUser(c *gin.Context, prof models.CmsUser) {
	body := request.ChangePassCmsUserBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	user := models.CmsUser{}
	user.Uid = prof.Uid
	errF := user.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	if body.OldPass != "" {
		passw, errDec := utils.DecryptAES([]byte(config.GetPassSecretKey()), body.OldPass)
		if errDec != nil {
			response_message.BadRequest(c, errDec.Error())
			return
		}
		errCheck := utils.ComparePassword(user.Password, passw)
		if errCheck != nil {
			response_message.BadRequest(c, "old pass is wrong")
			return
		}
	} else {
		response_message.BadRequest(c, "old pass is empty")
		return
	}

	if body.NewPass != "" {
		passw, errDec := utils.DecryptAES([]byte(config.GetPassSecretKey()), body.NewPass)
		if errDec != nil {
			response_message.BadRequest(c, errDec.Error())
			return
		}

		hashPass, errHash := utils.GeneratePassword(passw)
		if errHash != nil {
			response_message.BadRequest(c, errHash.Error())
			return
		}
		user.Password = hashPass
	} else {
		response_message.BadRequest(c, "new pass is empty")
		return
	}

	errUdp := user.Update()

	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, user)
}

/*
Reset pass Cms User
*/
func (_ *CCmsUser) ResetPassCmsUser(c *gin.Context, prof models.CmsUser) {
	body := request.ResetPassCmsUserBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	user := models.CmsUser{}
	user.Uid = body.UserUid
	errF := user.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	passw, errDec := utils.DecryptAES([]byte(config.GetPassSecretKey()), config.GetResetP())
	if errDec != nil {
		response_message.BadRequest(c, errDec.Error())
		return
	}

	hashPass, errHash := utils.GeneratePassword(passw)
	if errHash != nil {
		response_message.BadRequest(c, errHash.Error())
		return
	}
	user.Password = hashPass

	errUdp := user.Update()

	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, user)
}

/*
Get permission user
*/
func (_ *CCmsUser) GetPermissionCmsUser(c *gin.Context, prof models.CmsUser) {

	user := models.CmsUser{}
	user.Uid = prof.Uid
	errF := user.FindFirst()
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	// Find Role
	// Find Permission
	listPerMis := utils.ListString{}
	role := model_role.Role{}
	if user.RoleId > 0 {
		role.Id = user.RoleId

		key := user.GetKeyRedisPermission()
		listPer, _ := datasources.GetCache(key)

		if !config.GetUseRedisPermissionRole() {
			listPer = ""
		}

		if len(listPer) > 0 {
			_ = json.Unmarshal([]byte(listPer), &listPerMis)
		} else {
			errFR := role.FindFirst()
			if errFR == nil {
				rolePR := model_role.RolePermission{
					RoleId: role.Id,
				}
				listPermission, errRolePR := rolePR.FindAll()
				if errRolePR == nil {
					for _, v := range listPermission {
						listPerMis = append(listPerMis, v.PermissionUid)
					}
				}
			}
			user.SaveKeyRedisPermission(listPerMis)
		}
	} else {
		if user.RoleId == -1 {
			// Root Account
			role.Id = user.RoleId
			errFR := role.FindFirst()
			if errFR == nil {
				permis := model_role.Permission{}
				listP, errLP := permis.FindAll()
				if errLP == nil {
					for _, v := range listP {
						listPerMis = append(listPerMis, v.Uid)
					}
				}
			}
		}
	}

	res := map[string]interface{}{
		"permissions": listPerMis,
	}

	okResponse(c, res)
}

/*
Find list permissions
*/
func findListPermissionForCmsUser(user models.CmsUser) utils.ListString {
	listPerMis := utils.ListString{}
	role := model_role.Role{}
	if user.RoleId > 0 {
		role.Id = user.RoleId

		key := user.GetKeyRedisPermission()
		listPer, _ := datasources.GetCache(key)

		if !config.GetUseRedisPermissionRole() {
			listPer = ""
		}

		if len(listPer) > 0 {
			_ = json.Unmarshal([]byte(listPer), &listPerMis)
		} else {
			errFR := role.FindFirst()
			if errFR == nil {
				rolePR := model_role.RolePermission{
					RoleId: role.Id,
				}
				listPermission, errRolePR := rolePR.FindAll()
				if errRolePR == nil {
					for _, v := range listPermission {
						listPerMis = append(listPerMis, v.PermissionUid)
					}
				}
			}
			// Hàm này dùng nhiều nên bỏ đoạn lưu
			// user.SaveKeyRedisPermission(listPerMis)
		}
	}
	return listPerMis
}
