package model_service

import (
	"errors"
	"start/constants"
	"start/models"
	"start/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

// FbPromotionSet
type FbPromotionSet struct {
	models.ModelId
	PartnerUid string           `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string           `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	GroupCode  string           `json:"group_code" gorm:"type:varchar(100);index"`
	SetName    string           `json:"set_name"`
	Code       string           `json:"code"` // Mã Set
	Discount   int64            `json:"discount"`
	Note       string           `json:"note"`
	FBList     utils.ListString `json:"fb_list,omitempty" gorm:"type:json"` // ds món, liên kết qua fb code
	InputUser  string           `json:"input_user" gorm:"type:varchar(100)"`
	Price      float64          `json:"price"` // Giá Set
}

type FbPromotionSetRequest struct {
	FbPromotionSet
	CodeOrName string `form:"code_or_name"`
}
type FBPromotionSetResponse struct {
	FbPromotionSet
	GroupName string `json:"group_name"`
}

func (item *FbPromotionSet) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}
	return db.Create(item).Error
}

func (item *FbPromotionSet) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *FbPromotionSet) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *FbPromotionSet) Count(database *gorm.DB) (int64, error) {
	db := database.Model(FbPromotionSet{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *FbPromotionSetRequest) FindList(database *gorm.DB, page models.Page) ([]FBPromotionSetResponse, int64, error) {
	db := database.Model(FbPromotionSet{})
	list := []FBPromotionSetResponse{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""

	if status != "" {
		db = db.Where("fb_promotion_sets.status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("fb_promotion_sets.partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("fb_promotion_sets.course_uid = ?", item.CourseUid)
	}
	if item.SetName != "" {
		db = db.Where("fb_promotion_sets.set_name LIKE ?", "%"+item.SetName+"%")
	}
	if item.GroupCode != "" {
		db = db.Where("fb_promotion_sets.group_code = ?", item.GroupCode)
	}
	if item.CodeOrName != "" {
		db = db.Where("fb_promotion_sets.code = ?", item.CodeOrName).Or("fb_promotion_sets.set_name COLLATE utf8mb4_general_ci LIKE ?", "%"+item.CodeOrName+"%")
	}

	db = db.Joins("JOIN group_services ON fb_promotion_sets.group_code = group_services.group_code AND " +
		"fb_promotion_sets.partner_uid = group_services.partner_uid AND " +
		"fb_promotion_sets.course_uid = group_services.course_uid")
	db = db.Select("fb_promotion_sets.*, group_services.group_name")
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

func (item *FbPromotionSet) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
