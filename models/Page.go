package models

import "gorm.io/gorm"

type Page struct {
	Limit   int
	Page    int
	SortDir string
	SortBy  string
}

func (p *Page) Offset() int {
	return (p.Page - 1) * p.Limit
}

func (p *Page) Setup(db *gorm.DB) *gorm.DB {
	var setupDb = db
	setupDb = setupDb.Offset(p.Offset()).Limit(p.Limit)
	if p.SortBy != "" && p.SortDir != "" {
		setupDb = setupDb.Order(p.SortBy + " " + p.SortDir)
	}
	return setupDb
}
