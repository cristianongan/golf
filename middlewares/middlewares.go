package middlewares

import (
	"log"
	"start/auth"
	"start/config"
	"start/constants"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

func T1AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("auth t1 middleware")
		c.Next()
	}
}

func T2AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("auth t2 middleware")
		c.Next()
	}
}

/*
 Acc VNPAY mới có quyền
*/
func RootPartnerMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := getUserJWTToken(c)

		user, err := auth.VerifyCmsJwtToken(token, config.GetJwtSecret())
		if err != nil {
			log.Println("cms user jwtauth err ", err.Error())
			response_message.UnAuthorized(c, err.Error())
			c.Abort()
			return
		}
		if user.PartnerUid != constants.ROOT_PARTNER_UID {
			response_message.PermissionDeny(c, "Not permission")
			c.Abort()
			return
		}
		c.Next()
	}
}
