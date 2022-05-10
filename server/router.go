package server

import (
	"start/config"
	"start/controllers"
	"start/middlewares"
	"strings"

	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
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
	router.Use(cors.AllowAll()) // Để login từ localhost

	/*
	 - Cấu trúc sub-group để custom Middleware
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

			/// =================== System ===================
			cSystem := new(controllers.CSystem)
			cmsApiAuthorized.GET("/system/customer-type", middlewares.AuthorizedCmsUserHandler(cSystem.GetListCategoryType))

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
			cmsApiAuthorized.GET("/member-card/:uid", middlewares.AuthorizedCmsUserHandler(cMemberCard.GetDetail))
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

			/// =================== Table Prices =====================
			cTablePrice := new(controllers.CTablePrice)
			cmsApiAuthorized.POST("/table-price", middlewares.AuthorizedCmsUserHandler(cTablePrice.CreateTablePrice))
			cmsApiAuthorized.GET("/table-price/list", middlewares.AuthorizedCmsUserHandler(cTablePrice.GetListTablePrice))
			cmsApiAuthorized.PUT("/table-price/:id", middlewares.AuthorizedCmsUserHandler(cTablePrice.UpdateTablePrice))
			cmsApiAuthorized.DELETE("/table-price/:id", middlewares.AuthorizedCmsUserHandler(cTablePrice.DeleteTablePrice))

			/// =================== Golf Fee =====================
			cGolfFee := new(controllers.CGolfFee)
			cmsApiAuthorized.POST("/golf-fee", middlewares.AuthorizedCmsUserHandler(cGolfFee.CreateGolfFee))
			cmsApiAuthorized.GET("/golf-fee/list", middlewares.AuthorizedCmsUserHandler(cGolfFee.GetListGolfFee))
			cmsApiAuthorized.PUT("/golf-fee/:id", middlewares.AuthorizedCmsUserHandler(cGolfFee.UpdateGolfFee))
			cmsApiAuthorized.DELETE("/golf-fee/:id", middlewares.AuthorizedCmsUserHandler(cGolfFee.DeleteGolfFee))

			/// =================== Annual Fee =====================
			cAnnualFee := new(controllers.CAnnualFee)
			cmsApiAuthorized.POST("/annual-fee", middlewares.AuthorizedCmsUserHandler(cAnnualFee.CreateAnnualFee))
			cmsApiAuthorized.GET("/annual-fee/list", middlewares.AuthorizedCmsUserHandler(cAnnualFee.GetListAnnualFee))
			cmsApiAuthorized.PUT("/annual-fee/:id", middlewares.AuthorizedCmsUserHandler(cAnnualFee.UpdateAnnualFee))
			cmsApiAuthorized.DELETE("/annual-fee/:id", middlewares.AuthorizedCmsUserHandler(cAnnualFee.DeleteAnnualFee))

			/// =================== Group Fee =====================
			/// Tạo sửa cùng Golf Fee
			cGroupFee := new(controllers.CGroupFee)
			cmsApiAuthorized.GET("/group-fee/list", middlewares.AuthorizedCmsUserHandler(cGroupFee.GetListGroupFee))

			/// =================== Hole Price Formula =====================
			cHolePriceFormula := new(controllers.CHolePriceFormula)
			cmsApiAuthorized.POST("/hole-price-formula", middlewares.AuthorizedCmsUserHandler(cHolePriceFormula.CreateHolePriceFormula))
			cmsApiAuthorized.GET("/hole-price-formula/list", middlewares.AuthorizedCmsUserHandler(cHolePriceFormula.GetListHolePriceFormula))
			cmsApiAuthorized.PUT("/hole-price-formula/:id", middlewares.AuthorizedCmsUserHandler(cHolePriceFormula.UpdateHolePriceFormula))
			cmsApiAuthorized.DELETE("/hole-price-formula/:id", middlewares.AuthorizedCmsUserHandler(cHolePriceFormula.DeleteHolePriceFormula))

			/// =================== Booking Setting ===================
			cBookingSetting := new(controllers.CBookingSetting)
			cmsApiAuthorized.POST("/booking/setting/group", middlewares.AuthorizedCmsUserHandler(cBookingSetting.CreateBookingSettingGroup))
			cmsApiAuthorized.GET("/booking/setting/group/list", middlewares.AuthorizedCmsUserHandler(cBookingSetting.GetListBookingSettingGroup))
			cmsApiAuthorized.PUT("/booking/setting/group/:id", middlewares.AuthorizedCmsUserHandler(cBookingSetting.UpdateBookingSettingGroup))
			cmsApiAuthorized.DELETE("/booking/setting/group/:id", middlewares.AuthorizedCmsUserHandler(cBookingSetting.DeleteBookingSettingGroup))

			cmsApiAuthorized.POST("/booking/setting", middlewares.AuthorizedCmsUserHandler(cBookingSetting.CreateBookingSetting))
			cmsApiAuthorized.GET("/booking/setting/list", middlewares.AuthorizedCmsUserHandler(cBookingSetting.GetListBookingSetting))
			cmsApiAuthorized.PUT("/booking/setting/:id", middlewares.AuthorizedCmsUserHandler(cBookingSetting.UpdateBookingSetting))
			cmsApiAuthorized.DELETE("/booking/setting/:id", middlewares.AuthorizedCmsUserHandler(cBookingSetting.DeleteBookingSetting))

			/// =================== Booking ===================
			cBooking := new(controllers.CBooking)
			cmsApiAuthorized.POST("/booking", middlewares.AuthorizedCmsUserHandler(cBooking.CreateBooking))
			cmsApiAuthorized.POST("/booking-with-checkin", middlewares.AuthorizedCmsUserHandler(cBooking.CreateBookingCheckIn))
			cmsApiAuthorized.POST("/booking/check-in", middlewares.AuthorizedCmsUserHandler(cBooking.CheckIn))
			cmsApiAuthorized.GET("/booking/list", middlewares.AuthorizedCmsUserHandler(cBooking.GetListBooking))
			cmsApiAuthorized.GET("/booking/:uid", middlewares.AuthorizedCmsUserHandler(cBooking.GetBookingDetail))           // Get Booking detail by uid
			cmsApiAuthorized.GET("/booking/by-bag", middlewares.AuthorizedCmsUserHandler(cBooking.GetBookingByBag))          // Get booking detail by Bag
			cmsApiAuthorized.PUT("/booking/:uid", middlewares.AuthorizedCmsUserHandler(cBooking.UpdateBooking))              // Thêm Info..., rental, kiosk, ...
			cmsApiAuthorized.POST("/booking/sub-bag/add", middlewares.AuthorizedCmsUserHandler(cBooking.AddSubBagToBooking)) // Add SubBag
			cmsApiAuthorized.POST("/booking/round/add", middlewares.AuthorizedCmsUserHandler(cBooking.AddRound))             // Add Round

			/// =================== Buggy =====================
			cBuggy := new(controllers.CBuggy)
			cmsApiAuthorized.POST("/buggy", middlewares.AuthorizedCmsUserHandler(cBuggy.CreateBuggy))
			cmsApiAuthorized.GET("/buggy/list", middlewares.AuthorizedCmsUserHandler(cBuggy.GetBuggyList))
			cmsApiAuthorized.PUT("/buggy/:id", middlewares.AuthorizedCmsUserHandler(cBuggy.UpdateBuggy))
			cmsApiAuthorized.DELETE("/buggy/:id", middlewares.AuthorizedCmsUserHandler(cBuggy.DeleteBuggy))

			/// =================== Caddie =====================
			cCaddie := new(controllers.CCaddie)
			cmsApiAuthorized.POST("/caddie", middlewares.AuthorizedCmsUserHandler(cCaddie.CreateCaddie))
			cmsApiAuthorized.POST("/caddie-batch", middlewares.AuthorizedCmsUserHandler(cCaddie.CreateCaddieBatch))
			cmsApiAuthorized.GET("/caddie/list", middlewares.AuthorizedCmsUserHandler(cCaddie.GetCaddieList))
			cmsApiAuthorized.PUT("/caddie/:id", middlewares.AuthorizedCmsUserHandler(cCaddie.UpdateCaddie))
			cmsApiAuthorized.DELETE("/caddie/:id", middlewares.AuthorizedCmsUserHandler(cCaddie.DeleteCaddie))

			/// =================== Caddie Note =====================
			cCaddieNote := new(controllers.CCaddieNote)
			cmsApiAuthorized.POST("/caddie-note", middlewares.AuthorizedCmsUserHandler(cCaddieNote.CreateCaddieNote))
			cmsApiAuthorized.GET("/caddie-note/list", middlewares.AuthorizedCmsUserHandler(cCaddieNote.GetCaddieNoteList))
			cmsApiAuthorized.PUT("/caddie-note/:id", middlewares.AuthorizedCmsUserHandler(cCaddieNote.UpdateCaddieNote))
			cmsApiAuthorized.DELETE("/caddie-note/:id", middlewares.AuthorizedCmsUserHandler(cCaddieNote.DeleteCaddieNote))
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
		todoRouter.POST("todo/batch", cTodo.CreateTodoBatch)
		todoRouter.GET("todo/list", cTodo.GetTodoList)
		todoRouter.PUT("todo/:uid", cTodo.UpdateTodo)
		todoRouter.DELETE("todo/:uid", cTodo.DeleteTodo)
	}

	return router
}
