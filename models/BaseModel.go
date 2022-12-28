package models

import (
	"database/sql/driver"
	"encoding/json"
)

type Model struct {
	Uid       string `gorm:"primary_key" sql:"not null;" json:"uid"`
	CreatedAt int64  `json:"created_at" gorm:"index"`
	UpdatedAt int64  `json:"updated_at"`
	Status    string `json:"status" gorm:"index;type:varchar(50)"` //ENABLE, DISABLE, TESTING
}

type ModelId struct {
	Id        int64  `gorm:"AUTO_INCREMENT:yes" sql:"bigint;not null;primary_key"  json:"id"`
	CreatedAt int64  `json:"created_at" gorm:"index"`
	UpdatedAt int64  `json:"updated_at"`
	Status    string `json:"status"  gorm:"type:varchar(50)"` //ENABLE, DISABLE, TESTING, DELETED
}

type ModelLog struct {
	Id        int64 `gorm:"AUTO_INCREMENT:yes" sql:"bigint;not null;primary_key"  json:"id"`
	CreatedAt int64 `json:"created_at" gorm:"index"`
}

type ListInt64 []int64

func (item *ListInt64) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListInt64) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type ListInt []int

func (item *ListInt) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item *ListInt) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ==================================================
type ListString []string

func (item *ListString) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListString) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ================= YtElmInfo ======================
type YtElmInfo struct {
	YtElmId      string `json:"yt_eml_id"`
	YtElmKeyword string `json:"yt_eml_keyword"`
	Duration     string `json:"duration"`
}

func (item *YtElmInfo) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item YtElmInfo) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ================= ListYtElmInfo ======================
type ListYtElmInfo []YtElmInfo

func (item *ListYtElmInfo) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListYtElmInfo) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ====================================================
