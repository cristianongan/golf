package middlewares

import (
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

func getOSCodeFromHeader(c *gin.Context) string {
	osCode := c.Request.Header.Get("os_code")
	if osCode != "" {
		return osCode
	}
	return ""
}

func getRequestTimeFromHeader(c *gin.Context) string {
	requestTime := c.Request.Header.Get("request_time")
	if requestTime != "" {
		return requestTime
	}
	return ""
}

func getSignatureFromHeader(c *gin.Context) string {
	signature := c.Request.Header.Get("signature")
	if signature != "" {
		return signature
	}
	return ""
}

func AppApiHeaderMiddleware(c *gin.Context) {
	//header data
	osCode := getOSCodeFromHeader(c)
	requestTime := getRequestTimeFromHeader(c)
	signature := getSignatureFromHeader(c)
	if osCode == "" || requestTime == "" || signature == "" {
		response_message.PermissionDeny(c, "input header miss data")
		c.Abort()
		return
	}

	/* Disable Check window time */
	// requestTimeInt64, errParse := strconv.ParseInt(requestTime, 10, 64)
	// if errParse != nil {
	// 	response_message.PermissionDeny(c, "rtime parse error")
	// 	c.Abort()
	// 	return
	// }

	// if config.GetEnviromentName() == constants.ENV_PROD {
	// 	currentTime := time.Now().Unix()

	// 	if currentTime-requestTimeInt64 > int64(config.GetTimeAppMiddleware()) || requestTimeInt64-currentTime > int64(config.GetTimeAppMiddleware()) {
	// 		response_message.PermissionDeny(c, "rtime error")
	// 		c.Abort()
	// 		return
	// 	}
	// }

	//Get OrderSource Api Key
	// orderSource := models.OrderSource{
	// 	Code: osCode,
	// }
	// errF := orderSource.FindFromCache()
	// if errF != nil || orderSource.Code == "" {
	// 	response_message.PermissionDeny(c, errF.Error())
	// 	c.Abort()
	// 	return
	// }

	// signatureBE := utils.GetMD5Hash(requestTime + "_" + orderSource.ApiKey)
	// if signature != signatureBE {
	// 	response_message.PermissionDeny(c, "Wrong signature")
	// 	c.Abort()
	// 	return
	// }

	c.Next()

}
