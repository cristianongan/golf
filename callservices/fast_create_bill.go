package callservices

import (
	"encoding/json"
	"log"
	"net/http"
	"start/config"
	"start/constants"
	"start/controllers/response"
	"start/utils"
	"time"

	"github.com/google/uuid"
)

func CreateFastBankCredit(body response.FastBillBody) (bool, response.FastBillRes) {
	data, err := json.Marshal(body)
	urlResult := config.GetGolfPartnerURL() + "fast/bank-credit"

	if err != nil {
		log.Println("Customer err: ", err)
		return false, response.FastBillRes{}
	}
	serverHeader := make(http.Header)
	serverHeader.Add("Content-Type", "application/json")
	resp, ok := writeOut(urlResult, data, serverHeader, "POST")

	customers := response.FastBillRes{}
	err1 := json.Unmarshal(resp, &customers)
	if err1 != nil {
		return false, customers
	}

	return ok, customers
}

func CreateFastCashVoucher(body response.FastBillBody) (bool, response.FastBillRes) {
	data, err := json.Marshal(body)
	urlResult := config.GetGolfPartnerURL() + "fast/cash-voucher"

	if err != nil {
		log.Println("Customer err: ", err)
		return false, response.FastBillRes{}
	}
	serverHeader := make(http.Header)
	serverHeader.Add("Content-Type", "application/json")
	resp, ok := writeOut(urlResult, data, serverHeader, "POST")

	customers := response.FastBillRes{}
	err1 := json.Unmarshal(resp, &customers)
	if err1 != nil {
		return false, customers
	}

	return ok, customers
}

func TransferFast(paymentType string, price int64, note, customerUid, customerName, billNo string) (bool, response.FastBillRes) {
	idOnes := "CL" + utils.HashCodeUuid(uuid.New().String())
	body := response.FastBillBody{
		IdOnes:       idOnes,
		MaDVCS:       "CTY",
		SoCT:         billNo,
		NgayCt:       time.Now().Format("2006-01-02T15:04:05.000Z"),
		MaNT:         "VND",
		TyGia:        1,
		MaKH:         customerUid,
		NguoiNopTien: customerName,
		DienGiai:     note,
		MaGD:         "2",
		TK:           "131111",
		Detail: []response.FastBillBodyItem{
			{
				TkCo: "131111",
				Tien: price,
			},
		},
	}

	check := false
	res := response.FastBillRes{}
	if paymentType == constants.PAYMENT_TYPE_CASH {
		check, res = CreateFastCashVoucher(body)
	} else {
		check, res = CreateFastBankCredit(body)
	}
	return check, res
}
