package request

import (
	"start/models"
	"strings"
)

type PageRequest struct {
	Limit   int    `form:"limit" json:"limit" binding:"omitempty,min=1"`
	Page    int    `form:"page" json:"page" binding:"omitempty,min=1"`
	SortBy  string `form:"sort_by" json:"sort_by"`
	SortDir string `form:"sort_dir" json:"sort_dir" binding:"omitempty,eq=asc|eq=desc"`
}

type AdvanceSearchPageRequest struct {
	PageRequest
	Search string `form:"search"`
}

type GeneralPageRequest struct {
	PageRequest
	Search     string `form:"search"`
	CourseUid  string `form:"course_uid"`
	PartnerUid string `form:"partner_uid"`
}

func (p *PageRequest) ToPage() models.Page {
	page := models.Page{SortBy: p.SortBy}
	if p.Limit > 0 {
		page.Limit = p.Limit
	} else {
		page.Limit = 100
	}

	if p.Page > 0 {
		page.Page = p.Page
	} else {
		page.Page = 1
	}

	dir := strings.ToUpper(p.SortDir)
	if dir == "DESC" {
		page.SortDir = dir
	} else {
		page.SortDir = "ASC"
	}

	return page
}
