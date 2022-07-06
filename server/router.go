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

			/// =================== Upload Image =====================
			cUpload := new(controllers.CUpload)
			groupApi.POST("/upload/image", cUpload.UploadImage)

			// ----------------------------------------------------------
			// ================== authorized api ===============================
			// ================== use Middleware check jwtToken ================
			cmsApiAuthorized := groupApi.Use(middlewares.CmsUserJWTAuth)

			/// =================== Config ====================
			cConfig := new(controllers.CConfig)
			cmsApiAuthorized.GET("/config", middlewares.AuthorizedCmsUserHandler(cConfig.GetConfig))

			/// =================== System ===================
			cSystem := new(controllers.CSystem)
			cmsApiAuthorized.GET("/system/customer-type", middlewares.AuthorizedCmsUserHandler(cSystem.GetListCategoryType))
			cmsApiAuthorized.GET("/system/nationality", middlewares.AuthorizedCmsUserHandler(cSystem.GetListNationality))

			// ----- Job -----
			cmsApiAuthorized.POST("/system/job", middlewares.AuthorizedCmsUserHandler(cSystem.CreateJob))
			cmsApiAuthorized.GET("/system/job/list", middlewares.AuthorizedCmsUserHandler(cSystem.GetListJob))
			cmsApiAuthorized.PUT("/system/job/:id", middlewares.AuthorizedCmsUserHandler(cSystem.UpdateJob))
			cmsApiAuthorized.DELETE("/system/job/:id", middlewares.AuthorizedCmsUserHandler(cSystem.DeleteJob))
			// ----- position -----
			cmsApiAuthorized.POST("/system/position", middlewares.AuthorizedCmsUserHandler(cSystem.CreatePosition))
			cmsApiAuthorized.GET("/system/position/list", middlewares.AuthorizedCmsUserHandler(cSystem.GetListPosition))
			cmsApiAuthorized.PUT("/system/position/:id", middlewares.AuthorizedCmsUserHandler(cSystem.UpdatePosition))
			cmsApiAuthorized.DELETE("/system/position/:id", middlewares.AuthorizedCmsUserHandler(cSystem.DeletePosition))
			// ----- CompanyType -----
			cmsApiAuthorized.POST("/system/company-type", middlewares.AuthorizedCmsUserHandler(cSystem.CreateCompanyType))
			cmsApiAuthorized.GET("/system/company-type/list", middlewares.AuthorizedCmsUserHandler(cSystem.GetListCompanyType))
			cmsApiAuthorized.PUT("/system/company-type/:id", middlewares.AuthorizedCmsUserHandler(cSystem.UpdateCompanyType))
			cmsApiAuthorized.DELETE("/system/company-type/:id", middlewares.AuthorizedCmsUserHandler(cSystem.DeleteCompanyType))

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
			cmsApiAuthorized.GET("/course/:uid", middlewares.AuthorizedCmsUserHandler(cCourse.GetCourseDetail))

			/// =================== Member Card =====================
			cMemberCard := new(controllers.CMemberCard)
			cmsApiAuthorized.POST("/member-card", middlewares.AuthorizedCmsUserHandler(cMemberCard.CreateMemberCard))
			cmsApiAuthorized.POST("/member-card/unactive", middlewares.AuthorizedCmsUserHandler(cMemberCard.UnactiveMemberCard))
			cmsApiAuthorized.GET("/member-card/list", middlewares.AuthorizedCmsUserHandler(cMemberCard.GetListMemberCard))
			cmsApiAuthorized.GET("/member-card/:uid", middlewares.AuthorizedCmsUserHandler(cMemberCard.GetDetail))
			cmsApiAuthorized.PUT("/member-card/:uid", middlewares.AuthorizedCmsUserHandler(cMemberCard.UpdateMemberCard))
			cmsApiAuthorized.DELETE("/member-card/:uid", middlewares.AuthorizedCmsUserHandler(cMemberCard.DeleteMemberCard))

			/// =================== Member Card Type =====================
			cMemberCardType := new(controllers.CMemberCardType)
			cmsApiAuthorized.POST("/member-card-type", middlewares.AuthorizedCmsUserHandler(cMemberCardType.CreateMemberCardType))
			cmsApiAuthorized.GET("/member-card-type/list", middlewares.AuthorizedCmsUserHandler(cMemberCardType.GetListMemberCardType))
			cmsApiAuthorized.GET("/member-card-type/get-fee-by-hole", middlewares.AuthorizedCmsUserHandler(cMemberCardType.GetFeeByHole))
			cmsApiAuthorized.PUT("/member-card-type/:id", middlewares.AuthorizedCmsUserHandler(cMemberCardType.UpdateMemberCardType))
			cmsApiAuthorized.DELETE("/member-card-type/:id", middlewares.AuthorizedCmsUserHandler(cMemberCardType.DeleteMemberCardType))

			/// =================== Customer Users =====================
			cCustomerUser := new(controllers.CCustomerUser)
			cmsApiAuthorized.POST("/customer-user", middlewares.AuthorizedCmsUserHandler(cCustomerUser.CreateCustomerUser))
			cmsApiAuthorized.GET("/customer-user/list", middlewares.AuthorizedCmsUserHandler(cCustomerUser.GetListCustomerUser))
			cmsApiAuthorized.PUT("/customer-user/:uid", middlewares.AuthorizedCmsUserHandler(cCustomerUser.UpdateCustomerUser))
			cmsApiAuthorized.DELETE("/customer-user/:uid", middlewares.AuthorizedCmsUserHandler(cCustomerUser.DeleteCustomerUser))
			cmsApiAuthorized.GET("/customer-user/:uid", middlewares.AuthorizedCmsUserHandler(cCustomerUser.GetCustomerUserDetail))
			cmsApiAuthorized.POST("/customer-user/agency-delete", middlewares.AuthorizedCmsUserHandler(cCustomerUser.DeleteAgencyCustomerUser))

			/// =================== Birthday Management ===================
			cmsApiAuthorized.GET("/birthday-management", middlewares.AuthorizedCmsUserHandler(cCustomerUser.GetBirthday))

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
			cmsApiAuthorized.GET("/golf-fee/list/guest-style", middlewares.AuthorizedCmsUserHandler(cGolfFee.GetListGuestStyle))
			cmsApiAuthorized.PUT("/golf-fee/:id", middlewares.AuthorizedCmsUserHandler(cGolfFee.UpdateGolfFee))
			cmsApiAuthorized.DELETE("/golf-fee/:id", middlewares.AuthorizedCmsUserHandler(cGolfFee.DeleteGolfFee))

			/// =================== Annual Fee =====================
			cAnnualFee := new(controllers.CAnnualFee)
			cmsApiAuthorized.POST("/annual-fee", middlewares.AuthorizedCmsUserHandler(cAnnualFee.CreateAnnualFee))
			cmsApiAuthorized.GET("/annual-fee/list", middlewares.AuthorizedCmsUserHandler(cAnnualFee.GetListAnnualFee))
			cmsApiAuthorized.GET("/annual-fee/member-card/list", middlewares.AuthorizedCmsUserHandler(cAnnualFee.GetListAnnualFeeWithGroupMemberCard))
			cmsApiAuthorized.PUT("/annual-fee/:id", middlewares.AuthorizedCmsUserHandler(cAnnualFee.UpdateAnnualFee))
			cmsApiAuthorized.DELETE("/annual-fee/:id", middlewares.AuthorizedCmsUserHandler(cAnnualFee.DeleteAnnualFee))

			cAnnualFeePay := new(controllers.CAnnualFeePay)
			cmsApiAuthorized.POST("/annual-fee-pay", middlewares.AuthorizedCmsUserHandler(cAnnualFeePay.CreateAnnualFeePay))
			cmsApiAuthorized.GET("/annual-fee-pay/list", middlewares.AuthorizedCmsUserHandler(cAnnualFeePay.GetListAnnualFeePay))

			/// =================== Group Fee =====================
			/// Tạo sửa cùng
			cGroupFee := new(controllers.CGroupFee)
			cmsApiAuthorized.GET("/group-fee/list", middlewares.AuthorizedCmsUserHandler(cGroupFee.GetListGroupFee))
			cmsApiAuthorized.POST("/group-fee", middlewares.AuthorizedCmsUserHandler(cGroupFee.CreateGroupFee))
			cmsApiAuthorized.PUT("/group-fee/:id", middlewares.AuthorizedCmsUserHandler(cGroupFee.UpdateGroupFee))
			cmsApiAuthorized.DELETE("/group-fee/:id", middlewares.AuthorizedCmsUserHandler(cGroupFee.DeleteGroupFee))

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
			cmsApiAuthorized.GET("/booking/setting/on-date", middlewares.AuthorizedCmsUserHandler(cBookingSetting.GetListBookingSettingOnDate))
			cmsApiAuthorized.PUT("/booking/setting/:id", middlewares.AuthorizedCmsUserHandler(cBookingSetting.UpdateBookingSetting))
			cmsApiAuthorized.DELETE("/booking/setting/:id", middlewares.AuthorizedCmsUserHandler(cBookingSetting.DeleteBookingSetting))

			/// =================== Booking ===================
			cBooking := new(controllers.CBooking)
			cmsApiAuthorized.POST("/booking", middlewares.AuthorizedCmsUserHandler(cBooking.CreateBooking)) // Tạo booking or tạo booking check in luôn
			cmsApiAuthorized.POST("/booking/check-in", middlewares.AuthorizedCmsUserHandler(cBooking.CheckIn))
			cmsApiAuthorized.GET("/booking/list", middlewares.AuthorizedCmsUserHandler(cBooking.GetListBooking))
			cmsApiAuthorized.GET("/booking/:uid", middlewares.AuthorizedCmsUserHandler(cBooking.GetBookingDetail))                       // Get Booking detail by uid
			cmsApiAuthorized.GET("/booking/by-bag", middlewares.AuthorizedCmsUserHandler(cBooking.GetBookingByBag))                      // Get booking detail by Bag
			cmsApiAuthorized.PUT("/booking/:uid", middlewares.AuthorizedCmsUserHandler(cBooking.UpdateBooking))                          // Thêm Info..., rental, kiosk, ...
			cmsApiAuthorized.POST("/booking/sub-bag/add", middlewares.AuthorizedCmsUserHandler(cBooking.AddSubBagToBooking))             // Add SubBag
			cmsApiAuthorized.POST("/booking/sub-bag/edit", middlewares.AuthorizedCmsUserHandler(cBooking.EditSubBagToBooking))           // Edit SubBag
			cmsApiAuthorized.POST("/booking/round/add", middlewares.AuthorizedCmsUserHandler(cBooking.AddRound))                         // Add Round
			cmsApiAuthorized.GET("/booking/list/add-sub-bag", middlewares.AuthorizedCmsUserHandler(cBooking.GetListBookingForAddSubBag)) // List booking for add sub bag
			cmsApiAuthorized.GET("/booking/sub-bag-detail/:uid", middlewares.AuthorizedCmsUserHandler(cBooking.GetSubBagDetail))         // Get Sub bag detail
			cmsApiAuthorized.POST("/booking/other-paid/add", middlewares.AuthorizedCmsUserHandler(cBooking.AddOtherPaid))                // Add Other Paid
			cmsApiAuthorized.POST("/booking/cancel", middlewares.AuthorizedCmsUserHandler(cBooking.CancelBooking))                       // Cancel booking
			cmsApiAuthorized.POST("/booking/moving", middlewares.AuthorizedCmsUserHandler(cBooking.MovingBooking))                       // Moving booking

			/// =================== BagsNote ===================
			cBagsNote := new(controllers.CBagsNote)
			cmsApiAuthorized.GET("/bags-note/list", middlewares.AuthorizedCmsUserHandler(cBagsNote.GetListBagsNote))

			/// =================== Company =====================
			cCompany := new(controllers.CCompany)
			cmsApiAuthorized.POST("/company", middlewares.AuthorizedCmsUserHandler(cCompany.CreateCompany))
			cmsApiAuthorized.GET("/company/list", middlewares.AuthorizedCmsUserHandler(cCompany.GetListCompany))
			cmsApiAuthorized.PUT("/company/:id", middlewares.AuthorizedCmsUserHandler(cCompany.UpdateCompany))
			cmsApiAuthorized.DELETE("/company/:id", middlewares.AuthorizedCmsUserHandler(cCompany.DeleteCompany))

			/// =================== Agency =====================
			cAgency := new(controllers.CAgency)
			cmsApiAuthorized.POST("/agency", middlewares.AuthorizedCmsUserHandler(cAgency.CreateAgency))
			cmsApiAuthorized.GET("/agency/list", middlewares.AuthorizedCmsUserHandler(cAgency.GetListAgency))
			cmsApiAuthorized.PUT("/agency/:id", middlewares.AuthorizedCmsUserHandler(cAgency.UpdateAgency))
			cmsApiAuthorized.DELETE("/agency/:id", middlewares.AuthorizedCmsUserHandler(cAgency.DeleteAgency))
			cmsApiAuthorized.GET("/agency/:id", middlewares.AuthorizedCmsUserHandler(cAgency.GetAgencyDetail))

			cAgencySpecialPrice := new(controllers.CAgencySpecialPrice)
			cmsApiAuthorized.POST("/agency-special-price", middlewares.AuthorizedCmsUserHandler(cAgencySpecialPrice.CreateAgencySpecialPrice))
			cmsApiAuthorized.GET("/agency-special-price/list", middlewares.AuthorizedCmsUserHandler(cAgencySpecialPrice.GetListAgencySpecialPrice))
			cmsApiAuthorized.PUT("/agency-special-price/:id", middlewares.AuthorizedCmsUserHandler(cAgencySpecialPrice.UpdateAgencySpecialPrice))
			cmsApiAuthorized.DELETE("/agency-special-price/:id", middlewares.AuthorizedCmsUserHandler(cAgencySpecialPrice.DeleteAgencySpecialPrice))

			/// **************** GO ****************

			/// =================== Course Operating ====================
			cCourseOperating := new(controllers.CCourseOperating)
			cmsApiAuthorized.GET("/course-operating/booking/list-for-caddie", middlewares.AuthorizedCmsUserHandler(cCourseOperating.GetListBookingCaddieOnCourse))
			cmsApiAuthorized.POST("/course-operating/booking/add-caddie-buggy", middlewares.AuthorizedCmsUserHandler(cCourseOperating.AddCaddieBuggyToBooking))
			cmsApiAuthorized.POST("/course-operating/flight/create", middlewares.AuthorizedCmsUserHandler(cCourseOperating.CreateFlight))
			cmsApiAuthorized.POST("/course-operating/caddie/out", middlewares.AuthorizedCmsUserHandler(cCourseOperating.OutCaddie))
			cmsApiAuthorized.POST("/course-operating/caddie/undo", middlewares.AuthorizedCmsUserHandler(cCourseOperating.UndoOutCaddie))
			cmsApiAuthorized.POST("/course-operating/caddie/out-all-in-flight", middlewares.AuthorizedCmsUserHandler(cCourseOperating.OutAllInFlight))
			cmsApiAuthorized.POST("/course-operating/caddie/need-more", middlewares.AuthorizedCmsUserHandler(cCourseOperating.NeedMoreCaddie))         // Đổi caddie
			cmsApiAuthorized.POST("/course-operating/caddie/delete-attach", middlewares.AuthorizedCmsUserHandler(cCourseOperating.DeleteAttachCaddie)) // Xoá caddie, buggy, flight
			cmsApiAuthorized.GET("/course-operating/starting-sheet", middlewares.AuthorizedCmsUserHandler(cCourseOperating.GetStartingSheet))          // Get for starting sheet

			/// =================== + More Course Operating ===================
			cmsApiAuthorized.POST("/course-operating/change-caddie", middlewares.AuthorizedCmsUserHandler(cCourseOperating.ChangeCaddie))
			cmsApiAuthorized.POST("/course-operating/change-buggy", middlewares.AuthorizedCmsUserHandler(cCourseOperating.ChangeBuggy))
			cmsApiAuthorized.POST("/course-operating/edit-holes-of-caddie", middlewares.AuthorizedCmsUserHandler(cCourseOperating.EditHolesOfCaddie))
			cmsApiAuthorized.POST("/course-operating/add-bag-to-flight", middlewares.AuthorizedCmsUserHandler(cCourseOperating.AddBagToFlight))

			/// =================== + More Course Operating ===================
			cmsApiAuthorized.GET("/course-operating/flight/list", middlewares.AuthorizedCmsUserHandler(cCourseOperating.GetFlight))
			cmsApiAuthorized.POST("/course-operating/move-bag-to-flight", middlewares.AuthorizedCmsUserHandler(cCourseOperating.MoveBagToFlight))

			/// =================== Golf Bag ===================
			cGolfBag := new(controllers.CGolfBag)
			cmsApiAuthorized.GET("/golf-bag/list", middlewares.AuthorizedCmsUserHandler(cGolfBag.GetGolfBag))

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
			cmsApiAuthorized.GET("/caddie/:id", middlewares.AuthorizedCmsUserHandler(cCaddie.GetCaddieDetail))
			cmsApiAuthorized.PUT("/caddie/:id", middlewares.AuthorizedCmsUserHandler(cCaddie.UpdateCaddie))
			cmsApiAuthorized.DELETE("/caddie/:id", middlewares.AuthorizedCmsUserHandler(cCaddie.DeleteCaddie))

			/// =================== Caddie Note =====================
			cCaddieNote := new(controllers.CCaddieNote)
			cmsApiAuthorized.POST("/caddie-note", middlewares.AuthorizedCmsUserHandler(cCaddieNote.CreateCaddieNote))
			cmsApiAuthorized.GET("/caddie-note/list", middlewares.AuthorizedCmsUserHandler(cCaddieNote.GetCaddieNoteList))
			cmsApiAuthorized.PUT("/caddie-note/:id", middlewares.AuthorizedCmsUserHandler(cCaddieNote.UpdateCaddieNote))
			cmsApiAuthorized.DELETE("/caddie-note/:id", middlewares.AuthorizedCmsUserHandler(cCaddieNote.DeleteCaddieNote))

			/// =================== Caddie Working Time =====================
			cCaddieWorkingTime := new(controllers.CCaddieWorkingTime)
			cmsApiAuthorized.POST("/caddie-working-time/checkin", middlewares.AuthorizedCmsUserHandler(cCaddieWorkingTime.CaddieCheckInWorkingTime))
			cmsApiAuthorized.POST("/caddie-working-time/checkout", middlewares.AuthorizedCmsUserHandler(cCaddieWorkingTime.CaddieCheckOutWorkingTime))
			cmsApiAuthorized.GET("/caddie-working-time/list", middlewares.AuthorizedCmsUserHandler(cCaddieWorkingTime.GetCaddieWorkingTimeDetail))
			cmsApiAuthorized.PUT("/caddie-working-time/:id", middlewares.AuthorizedCmsUserHandler(cCaddieWorkingTime.UpdateCaddieWorkingTime))
			cmsApiAuthorized.DELETE("/caddie-working-time/:id", middlewares.AuthorizedCmsUserHandler(cCaddieWorkingTime.DeleteCaddieWorkingTime))

			/// =================== Caddie Evaluation ===================
			cCaddieEvaluation := new(controllers.CCaddieEvaluation)
			cmsApiAuthorized.POST("/caddie-evaluation", middlewares.AuthorizedCmsUserHandler(cCaddieEvaluation.CreateCaddieEvaluation))
			cmsApiAuthorized.GET("/caddie-evaluation", middlewares.AuthorizedCmsUserHandler(cCaddieEvaluation.GetCaddieEvaluationList))
			cmsApiAuthorized.PUT("/caddie-evaluation/:id", middlewares.AuthorizedCmsUserHandler(cCaddieEvaluation.UpdateCaddieEvaluation))

			/// =================== Recent Caddie Booking ===================
			cCaddieBookingList := new(controllers.CCaddieBookingList)
			cmsApiAuthorized.GET("/caddie-booking", middlewares.AuthorizedCmsUserHandler(cCaddieBookingList.GetCaddieBookingList))
			cmsApiAuthorized.GET("/caddie-booking/agency", middlewares.AuthorizedCmsUserHandler(cCaddieBookingList.GetAgencyBookingList))
			cmsApiAuthorized.GET("/caddie-booking/cancel", middlewares.AuthorizedCmsUserHandler(cCaddieBookingList.GetCancelBookingList))

			/// =================== Caddie Calendar ===================
			cCaddieCalendar := new(controllers.CCaddieCalendar)
			cmsApiAuthorized.POST("/caddie-calendar", middlewares.AuthorizedCmsUserHandler(cCaddieCalendar.CreateCaddieCalendar))
			cmsApiAuthorized.GET("/caddie-calendar", middlewares.AuthorizedCmsUserHandler(cCaddieCalendar.GetCaddieCalendarList))
			cmsApiAuthorized.PUT("/caddie-calendar/:id", middlewares.AuthorizedCmsUserHandler(cCaddieCalendar.UpdateCaddieCalendar))

			/// =================== Caddie Working Calendar ===================
			cCaddieWorkingCalendar := new(controllers.CCaddieWorkingCalendar)
			cmsApiAuthorized.POST("/caddie-working-calendar", middlewares.AuthorizedCmsUserHandler(cCaddieWorkingCalendar.CreateCaddieWorkingCalendar))
			cmsApiAuthorized.GET("/caddie-working-calendar", middlewares.AuthorizedCmsUserHandler(cCaddieWorkingCalendar.GetCaddieWorkingCalendarList))
			cmsApiAuthorized.PUT("/caddie-working-calendar/:id", middlewares.AuthorizedCmsUserHandler(cCaddieWorkingCalendar.UpdateCaddieWorkingCalendar))

			/// =================== Buggy Used Statistic ===================
			cBuggyUsedList := new(controllers.CBuggyUsedList)
			cmsApiAuthorized.GET("/buggy-used-list", middlewares.AuthorizedCmsUserHandler(cBuggyUsedList.GetBuggyUsedList))

			// =================== Buggy Statistic ===================
			cBuggyList := new(controllers.CBuggyList)
			cmsApiAuthorized.GET("/buggy-list", middlewares.AuthorizedCmsUserHandler(cBuggyList.GetBuggyList))

			// =================== Buggy In Course ===================
			cBuggyInCourse := new(controllers.CBuggyInCourse)
			cmsApiAuthorized.GET("/buggy-in-course", middlewares.AuthorizedCmsUserHandler(cBuggyInCourse.GetBuggyInCourse))

			// =================== Buggy Calendar ===================
			cBuggyCalendar := new(controllers.CBuggyCalendar)
			cmsApiAuthorized.GET("/buggy-calendar", middlewares.AuthorizedCmsUserHandler(cBuggyCalendar.GetBuggyCalendar))

			///
			/// =================== CGolf Service: Rental, Proshop, Restaurent, Kiosk =====================
			cGolfService := new(controllers.CGolfService)
			cmsApiAuthorized.GET("/golf-service/list/reception", middlewares.AuthorizedCmsUserHandler(cGolfService.GetGolfServiceForReception))
			/// =================== Rental =====================
			cRental := new(controllers.CRental)
			cmsApiAuthorized.POST("/rental", middlewares.AuthorizedCmsUserHandler(cRental.CreateRental))
			cmsApiAuthorized.GET("/rental/list", middlewares.AuthorizedCmsUserHandler(cRental.GetListRental))
			cmsApiAuthorized.PUT("/rental/:id", middlewares.AuthorizedCmsUserHandler(cRental.UpdateRental))
			cmsApiAuthorized.DELETE("/rental/:id", middlewares.AuthorizedCmsUserHandler(cRental.DeleteRental))
			/// =================== F&B =====================
			cFoodBeverage := new(controllers.CFoodBeverage)
			cmsApiAuthorized.POST("/f&b", middlewares.AuthorizedCmsUserHandler(cFoodBeverage.CreateFoodBeverage))
			cmsApiAuthorized.GET("/f&b/list", middlewares.AuthorizedCmsUserHandler(cFoodBeverage.GetListFoodBeverage))
			cmsApiAuthorized.PUT("/f&b/:id", middlewares.AuthorizedCmsUserHandler(cFoodBeverage.UpdateFoodBeverage))
			cmsApiAuthorized.DELETE("/f&b/:id", middlewares.AuthorizedCmsUserHandler(cFoodBeverage.DeleteFoodBeverage))
			/// =================== Proshop =====================
			cProshop := new(controllers.CProshop)
			cmsApiAuthorized.POST("/proshop", middlewares.AuthorizedCmsUserHandler(cProshop.CreateProshop))
			cmsApiAuthorized.GET("/proshop/list", middlewares.AuthorizedCmsUserHandler(cProshop.GetListProshop))
			cmsApiAuthorized.PUT("/proshop/:id", middlewares.AuthorizedCmsUserHandler(cProshop.UpdateProshop))
			cmsApiAuthorized.DELETE("/proshop/:id", middlewares.AuthorizedCmsUserHandler(cProshop.DeleteProshop))
			/// =================== Group Services =====================
			cGroupServices := new(controllers.CGroupServices)
			cmsApiAuthorized.POST("/group-services", middlewares.AuthorizedCmsUserHandler(cGroupServices.CreateGroupServices))
			cmsApiAuthorized.GET("/group-services/list", middlewares.AuthorizedCmsUserHandler(cGroupServices.GetGroupServicesList))
			cmsApiAuthorized.DELETE("/group-services/:id", middlewares.AuthorizedCmsUserHandler(cGroupServices.DeleteServices))
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
