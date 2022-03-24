package server

import (
	"start/config"
	"start/controllers"
	"start/middlewares"
	"strings"

	"github.com/gin-gonic/gin"
	// _ "start/docs" // docs is generated by Swag CLI, you have to import it.
	// ginSwagger "github.com/swaggo/gin-swagger"
	// "github.com/swaggo/gin-swagger/swaggerFiles"
)

var versionCheck = "v1.1"

func healthcheck(c *gin.Context) {
	c.JSON(200, gin.H{"message": "success: " + versionCheck})
	c.Abort()
	return
}

func NewRouter() *gin.Engine {
	router := gin.New()

	moduleName := strings.Replace(config.GetModuleName(), "_", "-", -1)
	router.Group(moduleName).GET("/", healthcheck)

	if config.GetKibanaLog() {
		router.Use(middlewares.GinBodyLogMiddleware)
	}
	//router.Use(cors.AllowAll())

	/*
	 - Cấu trúc sub-group để custorm Middleware
	*/
	routerApi := router.Group(moduleName)
	{
		cHealCheck := new(controllers.HealCheck)
		//customer.GET("/", cHealCheck.HealCheck)
		routerApi.GET("/check-ip", cHealCheck.CheckIp)

		// ----------------------------------------------------------
		// *********************** CMS - Operation ***********************
		// ----------------------------------------------------------

		// Tạo 1 sub group để xử lý middleware
		// TODO: Thêm middleware cho group api
		groupApi := routerApi.Group("api")
		{
			/// =================== Auth =====================
			cCmsUser := new(controllers.CCmsUser)
			groupApi.POST("/user/login", cCmsUser.Login)
			groupApi.POST("/user", cCmsUser.CreateCmsUser)

			// ----------------------------------------------------------
			// ================== authorized api ===============================
			// ================== use Middleware check jwtToken ================
			cmsApiAuthorized := groupApi.Use(middlewares.CmsUserJWTAuth)

			/// =================== Partner =====================
			cPartner := new(controllers.CPartner)
			cmsApiAuthorized.POST("/partner", middlewares.AuthorizedCmsUserHandler(cPartner.CreatePartner))
			cmsApiAuthorized.GET("/partner/list", middlewares.AuthorizedCmsUserHandler(cPartner.GetListPartner))
			cmsApiAuthorized.PUT("/partner/:uid", middlewares.AuthorizedCmsUserHandler(cPartner.UpdatePartner))
			cmsApiAuthorized.DELETE("/partner/:uid", middlewares.AuthorizedCmsUserHandler(cPartner.DeletePartner))

			/// =================== Course =====================
			cCourse := new(controllers.CCourse)
			cmsApiAuthorized.POST("/course", middlewares.AuthorizedCmsUserHandler(cCourse.CreateCourse))
			cmsApiAuthorized.GET("/course/list", middlewares.AuthorizedCmsUserHandler(cCourse.GetListCourse))
			cmsApiAuthorized.PUT("/course/:uid", middlewares.AuthorizedCmsUserHandler(cCourse.UpdateCourse))
			cmsApiAuthorized.DELETE("/course/:uid", middlewares.AuthorizedCmsUserHandler(cCourse.DeleteCourse))

			/// =================== Member Card =====================
			cMemberCard := new(controllers.CMemberCard)
			cmsApiAuthorized.POST("/member-card", middlewares.AuthorizedCmsUserHandler(cMemberCard.CreateMemberCard))
			cmsApiAuthorized.GET("/member-card/list", middlewares.AuthorizedCmsUserHandler(cMemberCard.GetListMemberCard))
			cmsApiAuthorized.PUT("/member-card/:uid", middlewares.AuthorizedCmsUserHandler(cMemberCard.UpdateMemberCard))
			cmsApiAuthorized.DELETE("/member-card/:uid", middlewares.AuthorizedCmsUserHandler(cMemberCard.DeleteMemberCard))

			/// =================== Member Card Type =====================
			cMemberCardType := new(controllers.CMemberCardType)
			cmsApiAuthorized.POST("/member-card-type", middlewares.AuthorizedCmsUserHandler(cMemberCardType.CreateMemberCardType))
			cmsApiAuthorized.GET("/member-card-type/list", middlewares.AuthorizedCmsUserHandler(cMemberCardType.GetListMemberCardType))
			cmsApiAuthorized.PUT("/member-card-type/:id", middlewares.AuthorizedCmsUserHandler(cMemberCardType.UpdateMemberCardType))
			cmsApiAuthorized.DELETE("/member-card-type/:id", middlewares.AuthorizedCmsUserHandler(cMemberCardType.DeleteMemberCardType))

			/// =================== Customer Users =====================
			cCustomerUser := new(controllers.CCustomerUser)
			cmsApiAuthorized.POST("/customer-user", middlewares.AuthorizedCmsUserHandler(cCustomerUser.CreateCustomerUser))
			cmsApiAuthorized.GET("/customer-user/list", middlewares.AuthorizedCmsUserHandler(cCustomerUser.GetListCustomerUser))
			cmsApiAuthorized.PUT("/customer-user/:uid", middlewares.AuthorizedCmsUserHandler(cCustomerUser.UpdateCustomerUser))
			cmsApiAuthorized.DELETE("/customer-user/:uid", middlewares.AuthorizedCmsUserHandler(cCustomerUser.DeleteCustomerUser))
			cmsApiAuthorized.GET("/customer-user/:uid", middlewares.AuthorizedCmsUserHandler(cCustomerUser.GetCustomerUserDetail))

			/// =================== Buggy =====================
			cBuggy := new(controllers.CBuggy)
			cmsApiAuthorized.POST("/buggy", middlewares.AuthorizedCmsUserHandler(cBuggy.CreateBuggy))
			cmsApiAuthorized.GET("/buggy/list", middlewares.AuthorizedCmsUserHandler(cBuggy.GetBuggyList))
			cmsApiAuthorized.PUT("/buggy/:uid", middlewares.AuthorizedCmsUserHandler(cBuggy.UpdateBuggy))
			cmsApiAuthorized.DELETE("/buggy/:uid", middlewares.AuthorizedCmsUserHandler(cBuggy.DeleteBuggy))
		}

		// ----------------------------------------------------------
		// ====================== Application =======================
		// ----------------------------------------------------------

		// cronApi := customer.Group("cron-job").Use(middlewares.CronJobMiddleWare())
		// {
		// 	cCron := new(controllers.CCron)
		// 	cronApi.POST("/check-cron", cCron.CheckCron)
		// 	cronApi.POST("/backup-order", cCron.BackupOrder)
		// }
	}

	todoRouter := router.Group(moduleName)
	{
		cTodo := new(controllers.CTodo)

		todoRouter.POST("todo", cTodo.CreateTodo)
		todoRouter.GET("todo/list", cTodo.GetTodoList)
		todoRouter.PUT("todo/:uid", cTodo.UpdateTodo)
		todoRouter.DELETE("todo/:uid", cTodo.DeleteTodo)
	}

	return router
}
