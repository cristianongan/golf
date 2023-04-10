package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"start/constants"
	"start/datasources"
	"start/utils"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type CmsUser struct {
	Model

	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hãng Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	UserName   string `json:"user_name" gorm:"type:varchar(100);uniqueIndex"`

	FullName string `json:"full_name" gorm:"type:varchar(256);index"`
	Password string `json:"-" gorm:"type:varchar(256)"`
	LoggedIn bool   `json:"logged_in"`

	Email      string `json:"email" gorm:"type:varchar(100)"`
	Phone      string `json:"phone" gorm:"type:varchar(20)"`
	Dob        int64  `json:"dob"`
	Position   string `json:"position" gorm:"type:varchar(100)"`   // chức vụ
	Sex        int    `json:"sex"`                                 // gioi tinh
	Department string `json:"department" gorm:"type:varchar(100)"` // Đơn vị
	RoleId     int64  `json:"role_id" gorm:"index"`                // Quyền hạn
	CaddieId   int64  `json:"caddie_id" gorm:"index"`              // Gan voi Caddie (Neu co)
}

type CmsUserDetail struct {
	CmsUser
	RoleName    string           `json:"role_name"`
	Permissions utils.ListString `json:"permissions"`
}

type CmsUserResponse struct {
	Model

	UserName string `json:"user_name"`
	Phone    string `json:"phone"`
	FullName string `json:"full_name"`
}

func (item *CmsUserResponse) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item CmsUserResponse) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type CmsUserBaseInfo struct {
	Uid        string `json:"uid"`
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	UserName   string `json:"user_name"`
	Status     string `json:"status"`
}

type CmsUserProfile struct {
	CmsUserBaseInfo
	jwt.StandardClaims
}

func (item *CmsUser) Create() error {
	uid := uuid.New()
	now := utils.GetTimeNow()
	item.Model.Uid = uid.String()
	item.Model.CreatedAt = now.Unix()
	item.Model.UpdatedAt = now.Unix()
	item.Model.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabaseAuth()
	return db.Create(item).Error
}

func (item *CmsUser) Update() error {
	mydb := datasources.GetDatabaseAuth()
	item.Model.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CmsUser) FindFirst() error {
	db := datasources.GetDatabaseAuth()
	return db.Where(item).First(item).Error
}

func (item *CmsUser) Count() (int64, error) {
	db := datasources.GetDatabaseAuth().Model(CmsUser{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CmsUser) FindList(page Page, search string, subRoles []int, isRootUser bool) ([]CmsUser, int64, error) {
	db := datasources.GetDatabaseAuth().Model(CmsUser{})
	list := []CmsUser{}
	total := int64(0)
	status := item.Model.Status
	item.Model.Status = ""
	db = db.Where(item)

	if !isRootUser {
		db = db.Where("role_id IN (?)", subRoles)
	}
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if search != "" {
		db = db.Where("(user_name LIKE ? OR full_name LIKE ?)", "%"+search+"%", "%"+search+"%")
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *CmsUser) FindListWithRole() ([]CmsUser, int64, error) {
	db := datasources.GetDatabaseAuth().Model(CmsUser{})
	list := []CmsUser{}
	total := int64(0)

	db = db.Where("role_id = ?", item.RoleId)
	db.Count(&total)
	db = db.Find(&list)

	return list, total, db.Error
}

func (item *CmsUser) Delete() error {
	if item.Model.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabaseAuth().Delete(item).Error
}

func (item *CmsUser) FindUserLocked() ([]CmsUser, int64, error) {
	var list []CmsUser
	total := int64(0)

	db := datasources.GetDatabaseAuth().Model(CmsUser{})

	db = db.Where("status = ?", constants.STATUS_DISABLE)
	db.Count(&total)

	db = db.Find(&list)
	return list, total, db.Error
}

func (item *CmsUser) GetKeyRedisPermission() string {
	return datasources.GetPrefixRedisKeyUserRolePermission() + item.PartnerUid + "_" + item.CourseUid + "_" + item.Uid
}

func (item *CmsUser) SaveKeyRedisPermission(lisPermission []string) {
	json, _ := json.Marshal(lisPermission)
	datasources.SetCache(item.GetKeyRedisPermission(), string(json), constants.TIME_REDIS_PERMISION)
}
