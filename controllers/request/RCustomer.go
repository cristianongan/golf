package request

type CustomerBody struct {
	MaKh      string `json:"ma_kh"`
	TenKh     string `json:"ten_kh"`
	MaSoThue  string `json:"ma_so_thue"`
	DiaChi    string `json:"dia_chi"`
	Tk        string `json:"tk"`
	DienThoai string `json:"dien_thoai"`
	Fax       string `json:"fax"`
	EMail     string `json:"e_mail"`
	DoiTac    string `json:"doi_tac"`
	NganHang  string `json:"ngan_hang"`
	TkNh      string `json:"tk_nh"`
}
