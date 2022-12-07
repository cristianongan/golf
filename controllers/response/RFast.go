package response

import (
	"start/controllers/request"
	"time"
)

type CustomerRes struct {
	Status  string               `json:"status"`
	Message string               `json:"message"`
	Item    request.CustomerBody `json:"item"`
}

type FastBillRes struct {
	Status  string       `json:"status"`
	Message string       `json:"message"`
	Item    FastBillBody `json:"item"`
}

type FastBillBody struct {
	IdOnes       string             `json:"id_ones" binding:"required"` // ID của ONE-S khi tạo phiếu
	MaDVCS       string             `json:"ma_dvcs" binding:"required"` // Mã đơn vị cơ sở của Fast
	SoCT         string             `json:"so_ct" binding:"required"`   // Số phiếu báo có
	NgayCt       time.Time          `json:"ngay_ct" binding:"required"` // Ngày chứng từ
	MaNT         string             `json:"ma_nt" binding:"required"`   // Mã đồng tiền hạch toán (VND, USD…)
	TyGia        int                `json:"ty_gia" binding:"required"`  // Tỷ giá theo mã đồng tiền hạch toán
	MaKH         string             `json:"ma_kh" binding:"required"`   // Mã khách hàng nộp tiền
	NguoiNopTien string             `json:"nguoi_nop_tien"`             // Người nộp tiền
	DienGiai     string             `json:"dien_giai"`                  // Diễn giải nộp tiền
	MaGD         string             `json:"ma_gd" binding:"required"`   // Loại chứng từ
	TK           string             `json:"tk" binding:"required"`      // Tài khoản tiền thu
	Detail       []FastBillBodyItem `json:"details" binding:"required"` //
}

type FastBillBodyItem struct {
	TkCo  string `json:"tk_co" binding:"required"` // Tài khoản đối ứng
	Tien  int64  `json:"tien" binding:"required"`  // Phát sinh tiền
	MaKhI string `json:"ma_kh_i"`                  // Mã khách hàng chi tiết
	MaVV  string `json:"ma_vv"`                    // Mã vụ việc phát sinh
	MaPhi string `json:"ma_phi"`                   // Mã phí phát sinh
	MaBP  string `json:"ma_bp"`                    // Mã bộ phận phát sinh
	MaHD  string `json:"ma_hd"`                    // Mã hợp đồng
	MaKU  string `json:"ma_ku"`                    // Mã khế ước
}
