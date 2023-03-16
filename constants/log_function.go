package constants

/*
Golf Action
*/
const (
	OP_LOG_ACTION_CREATE                           = "CREATE"
	OP_LOG_ACTION_UPDATE                           = "UPDATE"
	OP_LOG_ACTION_DELETE                           = "DELETE"
	OP_LOG_ACTION_CANCEL                           = "CANCEL"
	OP_LOG_ACTION_CANCEL_ALL                       = "CANCEL_ALL"
	OP_LOG_ACTION_INPUT_BAG_BOOKING                = "INPUT_BAG_BOOKING"
	OP_LOG_ACTION_BOOK_CADDIE                      = "BOOK_CADDIE"
	OP_LOG_ACTION_UPD_BOOK_CADDIE                  = "UPD_BOOK_CADDIE"
	OP_LOG_ACTION_MOVE                             = "MOVE"
	OP_LOG_ACTION_COPY                             = "COPY"
	OP_LOG_ACTION_ADD_MORE                         = "ADD_MORE"
	OP_LOG_ACTION_LOCK_TEE                         = "LOCK_TEE"
	OP_LOG_ACTION_UNLOCK_TEE                       = "UNLOCK_TEE"
	OP_LOG_ACTION_UNLOCK_TURN                      = "UNLOCK_TURN"
	OP_LOG_ACTION_PAYMENT                          = "PAYMENT"
	OP_LOG_ACTION_CHECK_IN                         = "CHECK_IN"
	OP_LOG_ACTION_LOCK_BAG                         = "LOCK_BAG"
	OP_LOG_ACTION_UN_LOCK_BAG                      = "UN_LOCK_BAG"
	OP_LOG_ACTION_ADD_ROUND                        = "ADD_ROUND"
	OP_LOG_ACTION_DEL_ROUND                        = "DEL_ROUND"
	OP_LOG_ACTION_SPLIT_ROUND                      = "SPLIT_ROUND"
	OP_LOG_ACTION_MERGE_ROUND                      = "MERGE_ROUND"
	OP_LOG_ACTION_CHANGE_GUEST_STYLE               = "CHANGE_GUEST_STYLE"
	OP_LOG_ACTION_CHECK_OUT                        = "CHECK_OUT"
	OP_LOG_ACTION_ADD_RENTAL                       = "ADD_RENTAL"
	OP_LOG_ACTION_ADD_DRIVING                      = "ADD_DRIVING"
	OP_LOG_ACTION_ADD_PROSHOP                      = "ADD_PROSHOP"
	OP_LOG_ACTION_ADD_RESTAURANT                   = "ADD_RESTAURANT"
	OP_LOG_ACTION_ADD_KIOSK                        = "ADD_KIOSK"
	OP_LOG_ACTION_ADD_MINI_B                       = "ADD_MINI_B"
	OP_LOG_ACTION_ADD_DISCOUNT                     = "ADD_DISCOUNT"
	OP_LOG_ACTION_UNDO_CHECK_IN                    = "UNDO_CHECK_IN"
	OP_LOG_ACTION_ADD_SUB_BAG                      = "ADD_SUB_BAG"
	OP_LOG_ACTION_CHANGE_TO_MAIN_BAG               = "CHANGE_TO_MAIN_BAG"
	OP_LOG_ACTION_RESET_CAD_SLOT                   = "RESET_CAD_SLOT"
	OP_LOG_ACTION_UPD_CAD_SLOT                     = "UPD_CAD_SLOT"
	OP_LOG_ACTION_COURSE_INFO_ATTACH               = "ATTACH"
	OP_LOG_ACTION_COURSE_INFO_CHANGE_ATTACH        = "CHANGE_ATTACH"
	OP_LOG_ACTION_COURSE_INFO_CREATE_FLIGHT        = "CREATE_FLIGHT"
	OP_LOG_ACTION_COURSE_INFO_OUT_ALL_FLIGHT       = "OUT_ALL_FLIGHT"
	OP_LOG_ACTION_COURSE_INFO_SIMPLE_OUT_FLIGHT    = "SIMPLE_OUT_FLIGHT"
	OP_LOG_ACTION_COURSE_INFO_UNDO_OUT_FLIGHT      = "UNDO_OUT_FLIGHT"
	OP_LOG_ACTION_COURSE_INFO_MOVE_FLIGHT          = "MOVE_FLIGHT"
	OP_LOG_ACTION_COURSE_INFO_ADD_BAG_TO_FLIGHT    = "ADD_BAG_TO_FLIGHT"
	OP_LOG_ACTION_COURSE_INFO_CHANGE_CADDIE        = "CHANGE_CADDIE"
	OP_LOG_ACTION_COURSE_INFO_CHANGE_BUGGY         = "CHANGE_BUGGY"
	OP_LOG_ACTION_COURSE_INFO_DELETE_ATTACH_FLIGHT = "DELETE_ATTACH_FLIGHT"
	OP_LOG_ACTION_CREATE_BAG                       = "CREATE_BAG"
	OP_LOG_ACTION_ADD_ITEM                         = "ADD_ITEM"
	OP_LOG_ACTION_UNDO_BILL                        = "UNDO_BILL"
	OP_LOG_ACTION_TRANSFER                         = "TRANSFER"
)

/*
Module golf
*/
const (
	OP_LOG_MODULE_RECEPTION = "RECEPTION"
	OP_LOG_MODULE_GO        = "GO"
	OP_LOG_MODULE_POS       = "POS"
	OP_LOG_MODULE_CADDIE    = "CADDIE"
	OP_LOG_MODULE_CUSTOMER  = "CUSTOMER"
	OP_LOG_MODULE_COMPANY   = "COMPANY"
)

/*
Function golf
*/
const (
	OP_LOG_FUNCTION_BOOKING                  = "BOOKING"
	OP_LOG_FUNCTION_WAITTING_LIST            = "WAITTING_LIST"
	OP_LOG_FUNCTION_CHECK_IN                 = "CHECK_IN"
	OP_LOG_FUNCTION_BOOKING_TEE_TIME         = "BOOKING_TEE_TIME"
	OP_LOG_FUNCTION_AGENCY_PAID              = "AGENCY_PAID"
	OP_LOG_FUNCTION_PAYMENT_SINGLE           = "PAYMENT_SINGLE"
	OP_LOG_FUNCTION_PAYMENT_AGENCY           = "PAYMENT_AGENCY"
	OP_LOG_FUNCTION_CADDIE_LIST              = "CADDIE_LIST"
	OP_LOG_FUNCTION_CADDIE_VACTION_CALENDAR  = "CADDIE_VACTION_CALENDAR"
	OP_LOG_FUNCTION_CADDIE_SLOT              = "CADDIE_SLOT"
	OP_LOG_FUNCTION_CADDIE_WORKING_SCHEDULE  = "CADDIE_WORKING_SCHEDULE"
	OP_LOG_FUNCTION_GROUP_MANAGEMENT         = "GROUP_MANAGEMENT"
	OP_LOG_FUNCTION_MEMBER_CARD              = "MEMBER_CARD"
	OP_LOG_FUNCTION_CUSTOMER_USER            = "CUSTOMER_USER"
	OP_LOG_FUNCTION_COMPANY_LIST             = "COMPANY_LIST"
	OP_LOG_FUNCTION_TRANSFER_CARD            = "TRANSFER_CARD"
	OP_LOG_FUNCTION_COURSE_INFO_WAITING_LIST = "COURSE_INFO_WAITING_LIST"
	OP_LOG_FUNCTION_COURSE_INFO_IN_COURSE    = "COURSE_INFO_IN_COURSE"
	OP_LOG_FUNCTION_COURSE_INFO_TIME_OUT     = "COURSE_INFO_TIME_OUT"
	OP_LOG_FUNCTION_GOLF_CLUB_RENTAL         = "GOLF_CLUB_RENTAL"
	OP_LOG_FUNCTION_KIOSK                    = "KIOSK"
	OP_LOG_FUNCTION_MINI_BAR                 = "MINI_BAR"
	OP_LOG_FUNCTION_PROSHOP                  = "PROSHOP"
	OP_LOG_FUNCTION_DRIVING                  = "DRIVING"
	OP_LOG_FUNCTION_RESTAURANT               = "RESTAURANT"
)
