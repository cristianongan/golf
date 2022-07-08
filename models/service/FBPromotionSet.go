package model_service

import (
	"errors"
	"start/constants"
	"start/datasources"
	"start/models"
	"strings"
	"time"
)

// FbPromotionSet
type FbPromotionSet struct {
	models.ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	GroupCode  string `json:"group_code" gorm:"type:varchar(100);index"`
	SetName    string `json:"set_name"`
	Discount   int64  `json:"discount"`
	Note       string `json:"note"`
	FBList     string `json:"fb_list"`
	InputUser  string `json:"input_user" gorm:"type:varchar(100)"`
}
type FBPromotionSetResponse struct {
	FbPromotionSet
	GroupName string `json:"group_name"`
}
type FBPromotionSetResponseFE struct {
	Id         int64    `json:"id"`
	Status     string   `json:"status"`
	PartnerUid string   `json:"partner_uid"` // Hang Golf
	CourseUid  string   `json:"course_uid"`  // San Golf
	GroupCode  string   `json:"group_code"`
	SetName    string   `json:"set_name"`
	Discount   int64    `json:"discount"`
	Note       string   `json:"note"`
	FBList     []string `json:"fb_list"`
	InputUser  string   `json:"input_user"`
	GroupName  string   `json:"group_name"`
}

func (item *FbPromotionSet) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *FbPromotionSet) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *FbPromotionSet) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *FbPromotionSet) Count() (int64, error) {
	db := datasources.GetDatabase().Model(FbPromotionSet{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *FbPromotionSet) FindList(page models.Page) ([]FBPromotionSetResponseFE, int64, error) {
	db := datasources.GetDatabase().Model(FbPromotionSet{})
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

	db = db.Joins("JOIN group_services ON fb_promotion_sets.group_code = group_services.group_code AND " +
		"fb_promotion_sets.partner_uid = group_services.partner_uid AND " +
		"fb_promotion_sets.course_uid = group_services.course_uid")
	db = db.Select("fb_promotion_sets.*, group_services.group_name")
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	listPromotion := []FBPromotionSetResponseFE{}
	for _, v := range list {
		pro := FBPromotionSetResponseFE{
			Id:         v.Id,
			Status:     v.Status,
			PartnerUid: v.PartnerUid,
			CourseUid:  v.CourseUid,
			SetName:    v.SetName,
			GroupCode:  v.GroupCode,
			Discount:   v.Discount,
			Note:       v.Note,
			InputUser:  v.InputUser,
			GroupName:  v.GroupName,
			FBList:     strings.Split(v.FBList, ","),
		}
		listPromotion = append(listPromotion, pro)
	}
	return listPromotion, total, db.Error
}

func (item *FbPromotionSet) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
