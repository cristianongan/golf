package middlewares

import (
	"start/config"
	"start/constants"
	"start/datasources"

	"github.com/gin-gonic/gin"
)

func CronJobMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		key := constants.CRONJOB_PREFIX + config.GetEnvironmentName() + "-" + c.Request.URL.String()
		counter, _ := datasources.IncreaseFlagCounter(key)
		if counter > 1 {
			if counter > 10 {
				datasources.DelCacheByKey(key)
			}
			c.JSON(400, gin.H{"message": "Old data has been handling"})
			c.Abort()
			return
		}

		mapBoxKey := config.GetCronJobSecretKey()
		if mapBoxKey == token {
			c.Next()
			datasources.DelCacheByKey(key)
		} else {
			datasources.DelCacheByKey(key)
			c.JSON(401, gin.H{"message": "Key is not matched"})
			c.Abort()
			return
		}
	}
}
