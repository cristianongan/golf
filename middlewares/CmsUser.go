package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"start/auth"
	"start/config"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_role "start/models/role"
	"start/utils/response_message"

	. "github.com/vibrantbyte/go-antpath/antpath"

	"github.com/gin-gonic/gin"
)

var matcher PathMatcher

func init() {
	matcher = New()
}

// ================= JWT Token User Auth ================
func getUserJWTToken(c *gin.Context) string {
	token := c.Request.Header.Get("Authorization")
	if token != "" {
		return token
	}
	return c.Query("token")
}

func CmsUserJWTAuth(c *gin.Context) {
	token := getUserJWTToken(c)

	user, err := auth.VerifyCmsJwtToken(token, config.GetJwtSecret())
	if err != nil {
		log.Println("cms user jwtauth err ", err.Error())
		response_message.UnAuthorized(c, err.Error())
		c.Abort()
		return
	}

	// check cache
	jwtUserToken, errCache := datasources.GetCacheJwt(user.Uid)
	if errCache != nil {
		log.Println("Error cache: ", errCache)
		response_message.UnAuthorized(c, errCache.Error())
		c.Abort()
		return
	}

	if jwtUserToken != token {
		response_message.UnAuthorized(c, "jwtStore != token")
		c.Abort()
		return
	}

	c.Set(constants.CMS_USER_PROFILE_KEY, user)
	c.Next()
}

// =================================================
func AuthorizedCmsUserHandler(handler func(*gin.Context, models.CmsUser)) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, exists := c.Get(constants.CMS_USER_PROFILE_KEY)
		if !exists {
			response_message.UnAuthorized(c, "Not found profile")
			c.Abort()
			return
		}

		baseInfo, isUserProfile := value.(models.CmsUserProfile)
		if !isUserProfile {
			response_message.UnAuthorized(c, "Map to model error")
			c.Abort()
			return
		}

		user := models.CmsUser{
			Model: models.Model{
				Uid: baseInfo.Uid,
			},
		}
		errFind := user.FindFirst()
		if errFind != nil {
			response_message.UnAuthorized(c, errFind.Error())
			c.Abort()
			return
		}

		if user.Model.Status == constants.STATUS_DISABLE {
			response_message.UserLocked(c, "user be disable")
			c.Abort()
			return
		}

		// Với user partner Root là VNPay thì có quyền làm 1 số function cho partner khác
		// TODO: thêm định nghĩa 1 số function partner root có thể làm, hiện tại mở hết
		if user.PartnerUid != constants.ROOT_PARTNER_UID {
			body := request.CommonRequest{}
			partnerUidRequest := ""
			if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
				b := c.Request.URL.Query()
				partnerUidRequest = b.Get("partner_uid")
				body.PartnerUid = partnerUidRequest
			} else if c.Request.Method == "POST" || c.Request.Method == "PUT" {
				ByteBody, err := io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(ByteBody))
				if err != nil {
					log.Print(err.Error())
				}
				json.Unmarshal(ByteBody, &body)
			}

			if body.PartnerUid != "" && body.PartnerUid != user.PartnerUid {
				response_message.Forbidden(c, "forbidden")
				return
			}
		}

		/// OK
		handler(c, user)
	}
}

func PermissionHandler(handler func(*gin.Context, models.CmsUser)) func(*gin.Context, models.CmsUser) {
	return func(c *gin.Context, user models.CmsUser) {
		role := model_role.Role{}
		if user.RoleId > 0 {
			role.Id = user.RoleId
			errFR := role.FindFirst()
			if errFR == nil {
				rolePR := model_role.RolePermission{
					RoleId: role.Id,
				}
				listPermission, errRolePR := rolePR.FindAllPermission()
				if errRolePR == nil {
					accessible := isAccessible(listPermission, c.Request.Method, c.Request.URL.Path)
					if !accessible {
						response_message.Forbidden(c, "forbidden")
						return
					}
					handler(c, user)
				} else {
					response_message.BadRequest(c, errRolePR.Error())
					return
				}
			}
		} else {
			if user.RoleId == -1 {
				// Root Account
				handler(c, user)
			}
		}
	}
}

func isAccessible(permissions []model_role.Permission, method, reqPath string) bool {
	if len(permissions) == 0 {
		return true
	}
	for _, permission := range permissions {
		for _, httpAction := range permission.Resources {
			if (httpAction.Method == method || httpAction.Method == "ANY") &&
				matcher.Match(httpAction.Path, reqPath) {
				return true
			}
		}
	}
	return false
}
