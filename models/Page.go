package models

import (
	"regexp"
	"strings"

	"gorm.io/gorm"
)

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
		if strings.ToUpper(p.SortDir) != "ASC" && strings.ToUpper(p.SortDir) != "DESC" {
			p.SortBy = "desc"
		}
		pattern := "[^a-zA-Z0-9_]"
		output := regexp.MustCompile(pattern).ReplaceAllString(p.SortBy, "")
		p.SortBy = output
		setupDb = setupDb.Order(p.SortBy + " " + p.SortDir)
	}
	return setupDb
}

func GetPageDefault() Page {
	return Page{
		Limit:   20,
		Page:    1,
		SortDir: "desc",
		SortBy:  "created_at",
	}
}
