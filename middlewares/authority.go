package middlewares

import (
	"regexp"
	"start/constants"
	"start/models"
	"start/utils/response_message"
	"strings"

	"github.com/gin-gonic/gin"
)

func keyMatch(key1 string, key2 string) bool {
	key2 = strings.Replace(key2, "/*", "/.*", -1)
	re := regexp.MustCompile(`:[^/]+`)
	key2 = re.ReplaceAllString(key2, "[^/]+")
	res, err := regexp.MatchString("^"+key2+"$", key1)
	if err != nil {
		panic(err)
	}
	return res
}

func AuthorityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		profileKey, exists := c.Get(constants.CMS_USER_PROFILE_KEY)
		if !exists {
			response_message.UnAuthorized(c, "cms_user_profile_key not found")
			c.Abort()
			return
		}

		cmsUserProfile, ok := profileKey.(models.CmsUserProfile)
		if !ok {
			response_message.UnAuthorized(c, "cms_user_profile error")
			c.Abort()
			return
		}

		cmsUser := models.CmsUser{}
		cmsUser.Uid = cmsUserProfile.Uid
		if err := cmsUser.FindFirst(); err != nil {
			response_message.UnAuthorized(c, err.Error())
			c.Abort()
			return
		}

		// roleIds := slices.Map(cmsUser.UserRoles, func(item models.AuthUserRole) int64 {
		// 	return int64(item.RoleID)
		// })

		// auth := models.Authority{}
		// permissions, err := auth.GetPermissions(roleIds)

		// if err != nil {
		// 	response_message.UnAuthorized(c, err.Error())
		// 	c.Abort()
		// 	return
		// }

		hasPermission := false

		// for _, item := range permissions {
		// 	hasPermission = keyMatch(c.Request.URL.Path, item.Path) && c.Request.Method == item.Method
		// 	if hasPermission {
		// 		break
		// 	}
		// }

		if !hasPermission {
			response_message.UnAuthorized(c, "Authorization required")
			c.Abort()
			return
		}

		c.Next()
	}
}
