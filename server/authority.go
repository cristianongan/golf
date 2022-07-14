package server

import (
	"github.com/gin-gonic/gin"
	"github.com/harranali/authority"
	"log"
	"start/datasources"
)

func createPermissions(auth *authority.Authority, routers gin.RoutesInfo) {
	for _, router := range routers {
		if err := auth.CreatePermission(router.Method + "|" + router.Path); err != nil {
			log.Println("[DEBUG] [AUTHORITY]", err.Error())
		}
	}
}

var auth *authority.Authority

func initAuthority(routers gin.RoutesInfo) {
	auth := authority.New(authority.Options{
		TablesPrefix: "auth_",
		DB:           datasources.GetDatabase(),
	})

	// create permissions
	createPermissions(auth, routers)
}
