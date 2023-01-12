package model_booking

import (
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	"start/utils"

	"github.com/pkg/errors"
	"github.com/twharmon/slices"
	"gorm.io/gorm"
)

/*
	Lấy service item của main bag và sub bag nếu có
*/
func (item *Booking) FindServiceItems(db *gorm.DB) {
	//MainBag
	listServiceItems := ListBookingServiceItems{}
	serviceGolfs := BookingServiceItem{
		BillCode: item.BillCode,
	}

	listGolfService, _ := serviceGolfs.FindAll(db)
	if len(listGolfService) > 0 {
		for index, v := range listGolfService {
			// Check trạng thái bill
			if v.Location == constants.SERVICE_ITEM_ADD_BY_RECEPTION {
				// Add từ lễ tân thì k cần check
				listServiceItems = append(listServiceItems, v)
			} else {
				serviceCart := models.ServiceCart{}
				serviceCart.Id = v.ServiceBill

				errSC := serviceCart.FindFirst(db)
				if errSC != nil {
					log.Println("FindFristServiceCart errSC", errSC.Error())
					return
				}

				if serviceCart.BillStatus == constants.POS_BILL_STATUS_ACTIVE ||
					serviceCart.BillStatus == constants.RES_BILL_STATUS_PROCESS ||
					serviceCart.BillStatus == constants.RES_BILL_STATUS_FINISH ||
					serviceCart.BillStatus == constants.RES_BILL_STATUS_OUT {
					listServiceItems = append(listServiceItems, v)
				}
			}

			// Update lại bag cho service item thiếu bag
			if v.Bag == "" {
				listGolfService[index].Bag = item.Bag
				listGolfService[index].Update(db)
			}
		}
	}

	//Check Subbag
	listTemp := ListBookingServiceItems{}
	if item.SubBags != nil && len(item.SubBags) > 0 {
		for _, v := range item.SubBags {
			serviceGolfsTemp := BookingServiceItem{
				BillCode: v.BillCode,
			}
			listGolfServiceTemp, _ := serviceGolfsTemp.FindAll(db)

			RsubDetail := Booking{
				Bag: v.GolfBag,
			}

			subDetail, _ := RsubDetail.FindFirstByUId(db)
			isAgencyPaidBookingCaddie := subDetail.GetAgencyPaidBookingCaddie() > 0
			isAgencyPaidBuggy := subDetail.GetAgencyPaidBuggy() > 0

			if subDetail.CheckAgencyPaidAll() {
				break
			}

			for _, v1 := range listGolfServiceTemp {
				isCanAdd := false
				if item.MainBagPay != nil && len(item.MainBagPay) > 0 {
					for _, v2 := range item.MainBagPay {
						// Check trạng thái bill
						serviceCart := models.ServiceCart{}
						serviceCart.Id = v1.ServiceBill

						errSC := serviceCart.FindFirst(db)
						if errSC != nil {
							log.Println("FindFristServiceCart errSC", errSC.Error())
							return
						}

						// Check trong MainBag có trả mới add
						serviceTypV1 := v1.Type
						if serviceTypV1 == constants.MINI_B_SETTING || serviceTypV1 == constants.MINI_R_SETTING {
							serviceTypV1 = constants.GOLF_SERVICE_RESTAURANT
						}
						if serviceTypV1 == constants.DRIVING_SETTING {
							serviceTypV1 = constants.GOLF_SERVICE_RENTAL
						}
						if v2 == serviceTypV1 && v1.PaidBy != constants.PAID_BY_AGENCY {
							if v1.Location == constants.SERVICE_ITEM_ADD_BY_RECEPTION {
								isCanAdd = true
							} else {
								if serviceCart.BillStatus == constants.RES_BILL_STATUS_OUT ||
									serviceCart.BillStatus == constants.POS_BILL_STATUS_ACTIVE ||
									serviceCart.BillStatus == constants.RES_BILL_STATUS_PROCESS ||
									serviceCart.BillStatus == constants.RES_BILL_STATUS_FINISH {
									isCanAdd = true
								}
							}
						}
						if v1.ServiceType == constants.BUGGY_SETTING && isAgencyPaidBuggy {
							isCanAdd = false
						}
						if v1.ServiceType == constants.CADDIE_SETTING && isAgencyPaidBookingCaddie {
							isCanAdd = false
						}
					}
				}

				if item.CheckOutTime > 0 && v1.CreatedAt > item.CheckOutTime {
					isCanAdd = false
				}

				if isCanAdd {
					listTemp = append(listTemp, v1)
				}

			}
		}
	}

	listServiceItems = append(listServiceItems, listTemp...)

	item.ListServiceItems = listServiceItems
}

func (item *Booking) FindServiceItemsWithPaidInfo(db *gorm.DB) []BookingServiceItemWithPaidInfo {
	//MainBag
	listServiceItems := []BookingServiceItemWithPaidInfo{}
	serviceGolfs := BookingServiceItem{
		BillCode: item.BillCode,
	}

	mainPaidRental := false
	mainPaidProshop := false
	mainPaidRestaurant := false
	mainPaidKiosk := false
	mainPaidOtherFee := false
	mainCheckOutTime := int64(0)

	// Tính giá của khi có main bag
	if len(item.MainBags) > 0 {
		mainBook := Booking{
			CourseUid:   item.CourseUid,
			PartnerUid:  item.PartnerUid,
			Bag:         item.MainBags[0].GolfBag,
			BookingDate: item.BookingDate,
		}
		errFMB := mainBook.FindFirst(db)
		if errFMB != nil {
			log.Println("UpdateMushPay-"+item.Bag+"-Find Main Bag", errFMB.Error())
		}
		mainCheckOutTime = mainBook.CheckOutTime
		mainPaidRental = utils.ContainString(mainBook.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_RENTAL) > -1
		mainPaidProshop = utils.ContainString(mainBook.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_PROSHOP) > -1
		mainPaidRestaurant = utils.ContainString(mainBook.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_RESTAURANT) > -1
		mainPaidKiosk = utils.ContainString(mainBook.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_KIOSK) > -1
		mainPaidOtherFee = utils.ContainString(mainBook.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_OTHER_FEE) > -1
	}

	listGolfService, _ := serviceGolfs.FindAllWithPaidInfo(db)
	if len(listGolfService) > 0 {
		for index, v := range listGolfService {
			if mainCheckOutTime > 0 {

				if v.CreatedAt > mainCheckOutTime {
					v.IsPaid = false
				} else {
					if v.Type == constants.MAIN_BAG_FOR_PAY_SUB_RENTAL ||
						v.Type == constants.DRIVING_SETTING {
						if mainPaidRental {
							v.IsPaid = true
						}
						if v.ServiceType == constants.BUGGY_SETTING || v.ServiceType == constants.CADDIE_SETTING {
							if item.GetAgencyPaidBuggy() > 0 {
								v.IsPaid = true
							}
							if item.GetAgencyPaidBookingCaddie() > 0 {
								v.IsPaid = true
							}
						}
					} else if v.Type == constants.MAIN_BAG_FOR_PAY_SUB_PROSHOP && mainPaidProshop {
						v.IsPaid = true
					} else if (v.Type == constants.MAIN_BAG_FOR_PAY_SUB_RESTAURANT ||
						v.Type == constants.MINI_B_SETTING ||
						v.Type == constants.MINI_R_SETTING) && mainPaidRestaurant {
						v.IsPaid = true
					} else if v.Type == constants.MAIN_BAG_FOR_PAY_SUB_OTHER_FEE && mainPaidOtherFee {
						v.IsPaid = true
					} else if (v.Type == constants.GOLF_SERVICE_KIOSK) && mainPaidKiosk {
						v.IsPaid = true
					}
				}
			}

			// Check trạng thái bill
			if v.Location == constants.SERVICE_ITEM_ADD_BY_RECEPTION {
				// Add từ lễ tân thì k cần check
				listServiceItems = append(listServiceItems, v)
			} else {
				serviceCart := models.ServiceCart{}
				serviceCart.Id = v.ServiceBill

				errSC := serviceCart.FindFirst(db)
				if errSC != nil {
					log.Println("FindFristServiceCart errSC", errSC.Error())
					return []BookingServiceItemWithPaidInfo{}
				}

				if serviceCart.BillStatus == constants.POS_BILL_STATUS_ACTIVE ||
					serviceCart.BillStatus == constants.RES_BILL_STATUS_PROCESS ||
					serviceCart.BillStatus == constants.RES_BILL_STATUS_FINISH ||
					serviceCart.BillStatus == constants.RES_BILL_STATUS_OUT {
					listServiceItems = append(listServiceItems, v)
				}
			}

			// Update lại bag cho service item thiếu bag
			if v.Bag == "" {
				listGolfService[index].Bag = item.Bag
				listGolfService[index].Update(db)
			}
		}
	}

	//Check Subbag
	listTemp := []BookingServiceItemWithPaidInfo{}
	if item.SubBags != nil && len(item.SubBags) > 0 {
		for _, v := range item.SubBags {
			serviceGolfsTemp := BookingServiceItem{
				BillCode: v.BillCode,
			}
			listGolfServiceTemp, _ := serviceGolfsTemp.FindAllWithPaidInfo(db)

			RsubDetail := Booking{
				Bag: v.GolfBag,
			}
			subDetail, _ := RsubDetail.FindFirstByUId(db)
			isAgencyPaidBookingCaddie := subDetail.GetAgencyPaidBookingCaddie() > 0
			isAgencyPaidBuggy := subDetail.GetAgencyPaidBuggy() > 0

			// agency paid all
			if subDetail.CheckAgencyPaidAll() {
				break
			}

			for _, v1 := range listGolfServiceTemp {
				isCanAdd := false
				if item.MainBagPay != nil && len(item.MainBagPay) > 0 {
					for _, v2 := range item.MainBagPay {
						// Check trạng thái bill
						serviceCart := models.ServiceCart{}
						serviceCart.Id = v1.ServiceBill

						errSC := serviceCart.FindFirst(db)
						if errSC != nil {
							log.Println("FindFristServiceCart errSC", errSC.Error())
							return []BookingServiceItemWithPaidInfo{}
						}

						// Check trong MainBag có trả mới add
						serviceTypV1 := v1.Type
						if serviceTypV1 == constants.MINI_B_SETTING || serviceTypV1 == constants.MINI_R_SETTING {
							serviceTypV1 = constants.GOLF_SERVICE_RESTAURANT
						}
						if serviceTypV1 == constants.DRIVING_SETTING {
							serviceTypV1 = constants.GOLF_SERVICE_RENTAL
						}
						if v2 == serviceTypV1 {
							if v1.Location == constants.SERVICE_ITEM_ADD_BY_RECEPTION {
								isCanAdd = true
							} else {
								if serviceCart.BillStatus == constants.RES_BILL_STATUS_OUT ||
									serviceCart.BillStatus == constants.POS_BILL_STATUS_ACTIVE ||
									serviceCart.BillStatus == constants.RES_BILL_STATUS_PROCESS ||
									serviceCart.BillStatus == constants.RES_BILL_STATUS_FINISH {
									isCanAdd = true
								}
							}
						}

						if v1.ServiceType == constants.BUGGY_SETTING && isAgencyPaidBuggy {
							isCanAdd = false
						}
						if v1.ServiceType == constants.CADDIE_SETTING && isAgencyPaidBookingCaddie {
							isCanAdd = false
						}
					}
				}

				if item.CheckOutTime > 0 && v1.CreatedAt > item.CheckOutTime {
					isCanAdd = false
				}

				if isCanAdd {
					listTemp = append(listTemp, v1)
				}

			}
		}
	}

	listServiceItems = append(listServiceItems, listTemp...)

	return listServiceItems
}

func (item *Booking) GetCurrentBagGolfFee() BookingGolfFee {
	golfFee := BookingGolfFee{}
	if item.ListGolfFee == nil {
		return golfFee
	}
	if len(item.ListGolfFee) <= 0 {
		return golfFee
	}

	return item.ListGolfFee[0]
}

func (item *Booking) GetTotalServicesFee() int64 {
	total := int64(0)
	if item.ListServiceItems != nil {
		for _, v := range item.ListServiceItems {
			temp := v.Amount
			total += temp
		}
	}

	return total
}

func (item *Booking) GetTotalGolfFee() int64 {
	total := int64(0)
	if item.ListGolfFee != nil {
		for _, v := range item.ListGolfFee {
			golfFeeTemp := v.BuggyFee + v.CaddieFee + v.GreenFee
			total += golfFeeTemp
		}
	}

	return total
}

func (item *Booking) UpdateBagGolfFee() {
	if len(item.ListGolfFee) > 0 {
		item.ListGolfFee[0].Bag = item.Bag
	}
}

// Udp MushPay
func (item *Booking) UpdateMushPay(db *gorm.DB) {
	mushPay := BookingMushPay{}
	listRoundGolfFee := []models.Round{}

	roundToFindList := models.Round{BillCode: item.BillCode}
	listRoundOfCurrentBag, _ := roundToFindList.FindAll(db)
	mainPaidRound1 := false
	mainPaidRound2 := false
	mainPaidRental := false
	mainPaidProshop := false
	mainPaidRestaurant := false
	mainPaidKiosk := false
	mainPaidOtherFee := false
	mainCheckOutTime := int64(0)

	subBagFee := int64(0)     // Giá của sub bag
	feePaid := int64(0)       // Giá đã được trả bởi agency or main bag
	agencyPaidAll := int64(0) // Agency trả all

	buggyCaddieAgencyPaid := int64(0)
	feePaid += item.GetAgencyService()
	buggyCaddieAgencyPaid += item.GetAgencyService()

	// Tính giá của khi có main bag
	if len(item.MainBags) > 0 {
		mainBook := Booking{
			CourseUid:   item.CourseUid,
			PartnerUid:  item.PartnerUid,
			Bag:         item.MainBags[0].GolfBag,
			BookingDate: item.BookingDate,
		}
		errFMB := mainBook.FindFirst(db)
		if errFMB != nil {
			log.Println("UpdateMushPay-"+item.Bag+"-Find Main Bag", errFMB.Error())
		}
		mainCheckOutTime = mainBook.CheckOutTime
		mainPaidRound1 = utils.ContainString(mainBook.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND) > -1
		mainPaidRound2 = utils.ContainString(mainBook.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS) > -1
		mainPaidRental = utils.ContainString(mainBook.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_RENTAL) > -1
		mainPaidProshop = utils.ContainString(mainBook.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_PROSHOP) > -1
		mainPaidRestaurant = utils.ContainString(mainBook.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_RESTAURANT) > -1
		mainPaidKiosk = utils.ContainString(mainBook.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_KIOSK) > -1
		mainPaidOtherFee = utils.ContainString(mainBook.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_OTHER_FEE) > -1
	}

	for _, round := range listRoundOfCurrentBag {
		if item.CheckAgencyPaidAll() {
			agencyPaidAll += round.GetAmountGolfFee()
		}

		if round.Index == 1 {
			if !item.CheckAgencyPaidRound1() {
				// Nếu agency không trả thì xet tiếp
				if !mainPaidRound1 {
					// Nếu main bag ko trả round 1 thì add
					listRoundGolfFee = append(listRoundGolfFee, round)
				} else {
					// main bag đã check out đi về, sub bag chơi tiếp thì vẫn tính tiền
					if mainCheckOutTime > 0 && round.CreatedAt > mainCheckOutTime {
						listRoundGolfFee = append(listRoundGolfFee, round)
					} else {
						// main bag trả cho round1
						feePaid += round.GetAmountGolfFee()
					}
				}
			} else {
				// Agency trả cho round1
				feePaid += round.GetAmountGolfFee()
			}

		} else if round.Index == 2 {
			if !mainPaidRound2 {
				// Nếu main bag ko trả round 2 thì add
				listRoundGolfFee = append(listRoundGolfFee, round)
			} else {
				// main bag đã check out đi về, sub bag chơi tiếp thì vẫn tính tiền
				if mainCheckOutTime > 0 && round.CreatedAt > mainCheckOutTime {
					listRoundGolfFee = append(listRoundGolfFee, round)
				} else {
					// main bag trả cho round2
					feePaid += round.GetAmountGolfFee()
				}
			}
		} else {
			listRoundGolfFee = append(listRoundGolfFee, round)
		}
	}

	// Tính giá golf fee của main khi có sub bag
	if len(item.SubBags) > 0 {
		checkIsFirstRound := utils.ContainString(item.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND)
		checkIsNextRound := utils.ContainString(item.MainBagPay, constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS)
		for _, sub := range item.SubBags {
			roundToFindList := models.Round{BillCode: sub.BillCode}
			listSubRound, _ := roundToFindList.FindAll(db)

			for _, round := range listSubRound {

				// if len(sub.AgencyPaid) > 1 && sub.AgencyPaid[1].Fee > 0 {
				// 	buggyCaddieAgencyPaid += sub.AgencyPaid[1].Fee
				// }

				// if len(sub.AgencyPaid) > 2 && sub.AgencyPaid[2].Fee > 0 {
				// 	buggyCaddieAgencyPaid += sub.AgencyPaid[2].Fee
				// }

				if round.Index == 1 {
					if !(len(sub.AgencyPaid) > 0 && sub.AgencyPaid[0].Fee > 0) && checkIsFirstRound > -1 {
						listRoundGolfFee = append(listRoundGolfFee, round)
						subBagFee += round.GetAmountGolfFee()
					}
				}
				if round.Index == 2 && checkIsNextRound > -1 {
					listRoundGolfFee = append(listRoundGolfFee, round)
					subBagFee += round.GetAmountGolfFee()
				}
			}
		}
	}

	bookingCaddieFee := slices.Reduce(listRoundGolfFee, func(prev int64, item models.Round) int64 {
		return prev + item.CaddieFee
	})

	bookingBuggyFee := slices.Reduce(listRoundGolfFee, func(prev int64, item models.Round) int64 {
		return prev + item.BuggyFee
	})

	bookingGreenFee := slices.Reduce(listRoundGolfFee, func(prev int64, item models.Round) int64 {
		return prev + item.GreenFee
	})

	totalGolfFeeOfSubBag := bookingCaddieFee + bookingBuggyFee + bookingGreenFee
	mushPay.TotalGolfFee = totalGolfFeeOfSubBag

	// SubBag

	// Sub Service Item của current Bag
	// Get item for current Bag
	// update lại lấy service items mới
	buggyCaddieRentalFee := int64(0)
	item.FindServiceItems(db)
	for _, v := range item.ListServiceItems {
		isNeedPay := false
		isBuggyCaddieRental := false

		if v.ServiceType == constants.BUGGY_SETTING || v.ServiceType == constants.CADDIE_SETTING {
			isBuggyCaddieRental = true
			buggyCaddieRentalFee += v.Amount
		}

		if len(item.MainBags) > 0 {
			// Nếu có main bag
			if mainCheckOutTime > 0 && v.CreatedAt > mainCheckOutTime {
				// main bag đã check out đi về, sub bag dùng tiếp service thì phải trả v
				isNeedPay = true
			} else {
				if (v.Type == constants.MAIN_BAG_FOR_PAY_SUB_RENTAL ||
					v.Type == constants.DRIVING_SETTING) && !mainPaidRental {
					isNeedPay = true
				} else if v.Type == constants.MAIN_BAG_FOR_PAY_SUB_PROSHOP && !mainPaidProshop {
					isNeedPay = true
				} else if (v.Type == constants.MAIN_BAG_FOR_PAY_SUB_RESTAURANT ||
					v.Type == constants.MINI_B_SETTING ||
					v.Type == constants.MINI_R_SETTING) && !mainPaidRestaurant {
					isNeedPay = true
				} else if v.Type == constants.MAIN_BAG_FOR_PAY_SUB_OTHER_FEE && !mainPaidOtherFee {
					isNeedPay = true
				} else if (v.Type == constants.GOLF_SERVICE_KIOSK) && !mainPaidKiosk {
					isNeedPay = true
				}
			}

			if !isNeedPay && !isBuggyCaddieRental {
				feePaid += v.Amount
			}

		} else {
			if item.CheckAgencyPaidAll() {
				agencyPaidAll += v.Amount
			}

			if v.Bag != item.Bag {
				// Tính giá service của sub
				subBagFee += v.Amount
			}
			isNeedPay = true
		}

		if isNeedPay && !isBuggyCaddieRental {
			mushPay.TotalServiceItem += v.Amount
		}
	}

	buggyCaddieRentalMushPay := buggyCaddieRentalFee - buggyCaddieAgencyPaid

	if mainPaidRental {
		feePaid += buggyCaddieRentalMushPay
		buggyCaddieRentalMushPay = 0
	}

	if buggyCaddieRentalMushPay < 0 {
		buggyCaddieRentalMushPay = 0
	}

	if item.CustomerType == constants.BOOKING_CUSTOMER_TYPE_FOC {
		mushPay.MushPay = subBagFee
	}

	if item.CheckAgencyPaidAll() {
		mushPay.MushPay = subBagFee
		if item.GetAgencyPaid() != agencyPaidAll-subBagFee {
			item.AgencyPaid = utils.ListBookingAgencyPayForBagData{}
			item.AgencyPaid = append(item.AgencyPaid, utils.BookingAgencyPayForBagData{
				Type: constants.BOOKING_AGENCY_PAID_ALL,
				Fee:  agencyPaidAll - subBagFee,
			})
		}
		item.CurrentBagPrice.MainBagPaid = agencyPaidAll
	} else {
		total := mushPay.TotalGolfFee + mushPay.TotalServiceItem + buggyCaddieRentalMushPay
		if total < 0 {
			mushPay.MushPay = 0
		} else {
			mushPay.MushPay = total
		}

		item.MushPayInfo.Amount = total
		item.CurrentBagPrice.MainBagPaid = feePaid
	}

	item.MushPayInfo = mushPay

	// Update date lại giá USD
	currencyPaidGet := models.CurrencyPaid{
		Currency: "usd",
	}
	if err := currencyPaidGet.FindFirst(); err == nil {
		item.CurrentBagPrice.AmountUsd = mushPay.MushPay / currencyPaidGet.Rate
	}
}

// Udp lại giá cho Booking
// Udp sub bag price
func (item *Booking) UpdatePriceDetailCurrentBag(db *gorm.DB) {
	priceDetail := BookingCurrentBagPriceDetail{}

	roundToFindList := models.Round{BillCode: item.BillCode}
	listRound, _ := roundToFindList.FindAll(db)

	bookingCaddieFee := slices.Reduce(listRound, func(prev int64, item models.Round) int64 {
		return prev + item.CaddieFee
	})

	bookingBuggyFee := slices.Reduce(listRound, func(prev int64, item models.Round) int64 {
		return prev + item.BuggyFee
	})

	bookingGreenFee := slices.Reduce(listRound, func(prev int64, item models.Round) int64 {
		return prev + item.GreenFee
	})

	if len(item.ListGolfFee) == 0 {
		return
	}

	bookingGolfFee := item.ListGolfFee[0]
	bookingGolfFee.BookingUid = item.Uid
	bookingGolfFee.CaddieFee = bookingCaddieFee
	bookingGolfFee.BuggyFee = bookingBuggyFee
	bookingGolfFee.GreenFee = bookingGreenFee
	item.ListGolfFee[0] = bookingGolfFee

	if len(item.ListGolfFee) > 0 {
		priceDetail.GolfFee = item.ListGolfFee[0].BuggyFee + item.ListGolfFee[0].CaddieFee + item.ListGolfFee[0].GreenFee
	}

	item.FindServiceItems(db)
	for _, serviceItem := range item.ListServiceItems {
		if serviceItem.BillCode == item.BillCode {
			// Udp service detail cho booking uid
			if serviceItem.Type == constants.GOLF_SERVICE_RENTAL ||
				serviceItem.Type == constants.DRIVING_SETTING ||
				serviceItem.Type == constants.BUGGY_SETTING {
				priceDetail.Rental += serviceItem.Amount
			}
			if serviceItem.Type == constants.GOLF_SERVICE_PROSHOP {
				priceDetail.Proshop += serviceItem.Amount
			}
			if serviceItem.Type == constants.GOLF_SERVICE_RESTAURANT ||
				serviceItem.Type == constants.MINI_B_SETTING ||
				serviceItem.Type == constants.MINI_R_SETTING {
				priceDetail.Restaurant += serviceItem.Amount
			}
			if serviceItem.Type == constants.GOLF_SERVICE_KIOSK {
				priceDetail.Kiosk += serviceItem.Amount
			}
		}
	}

	priceDetail.UpdateAmount()

	item.CurrentBagPrice = priceDetail
}

// Check Duplicated
func (item *Booking) IsDuplicated(db *gorm.DB, checkTeeTime, checkBag bool) (bool, error) {
	//Check Bag đã tồn tại trước
	if checkBag {
		if item.Bag != "" {
			booking := Booking{
				PartnerUid:  item.PartnerUid,
				CourseUid:   item.CourseUid,
				BookingDate: item.BookingDate,
				Bag:         item.Bag,
			}
			errBagFind := booking.FindFirst(db)
			if errBagFind == nil || booking.Uid != "" {
				return true, errors.New("Duplicated Bag")
			}
		}
	}

	if item.TeeTime == "" {
		return false, nil
	}
	//Check turn time đã tồn tại
	if checkTeeTime {
		booking := Booking{
			PartnerUid:  item.PartnerUid,
			CourseUid:   item.CourseUid,
			TeeTime:     item.TeeTime,
			TurnTime:    item.TurnTime,
			BookingDate: item.BookingDate,
			RowIndex:    item.RowIndex,
			TeeType:     item.TeeType,
			CourseType:  item.CourseType,
		}

		errFind := booking.FindFirstNotCancel(db)
		if errFind == nil || booking.Uid != "" {
			return true, errors.New("Duplicated TeeTime")
		}
	}

	return false, nil
}

func (item *Booking) CheckAgencyPaidRound1() bool {
	return len(item.AgencyPaid) > 0 && item.AgencyPaid[0].Fee > 0
}

func (item *Booking) GetAgencyPaid() int64 {
	totalAgencyPaid := int64(0)
	for _, v := range item.AgencyPaid {
		totalAgencyPaid += v.Fee
	}
	return totalAgencyPaid
}

func (item *Booking) GetAgencyService() int64 {
	totalAgencyPaid := int64(0)
	for _, v := range item.AgencyPaid {
		if v.Type != constants.BOOKING_AGENCY_GOLF_FEE {
			totalAgencyPaid += v.Fee
		}
	}
	return totalAgencyPaid
}

func (item *Booking) GetAgencyPaidBuggy() int64 {
	totalAgencyPaid := int64(0)
	for _, v := range item.AgencyPaid {
		if v.Type == constants.BOOKING_AGENCY_BUGGY_FEE {
			totalAgencyPaid += v.Fee
		}
	}
	return totalAgencyPaid
}

func (item *Booking) GetAgencyPaidBookingCaddie() int64 {
	totalAgencyPaid := int64(0)
	for _, v := range item.AgencyPaid {
		if v.Type == constants.BOOKING_AGENCY_BOOKING_CADDIE_FEE {
			totalAgencyPaid += v.Fee
		}
	}
	return totalAgencyPaid
}

func (item *Booking) CheckAgencyPaidAll() bool {
	return item.AgencyPaidAll != nil && *item.AgencyPaidAll
}

func (item *Booking) NumberOfRound() int {
	db := datasources.GetDatabaseWithPartner(item.PartnerUid)
	roundToFindList := models.Round{BillCode: item.BillCode}
	listRound, _ := roundToFindList.FindAll(db)

	return len(listRound)
}
