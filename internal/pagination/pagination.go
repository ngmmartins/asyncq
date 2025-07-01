package pagination

import (
	"slices"
	"strings"

	"github.com/ngmmartins/asyncq/internal/validator"
)

var JobSortSafelist = []string{"id", "task", "run_at", "status", "created_at", "-id", "-task", "-run_at", "-status", "-created_at"}

type Params struct {
	Page         int
	PageSize     int
	SortBy       string
	SortSafelist []string
}

func (p Params) SortColumn() string {
	if slices.Contains(p.SortSafelist, p.SortBy) {
		return strings.TrimPrefix(p.SortBy, "-")
	}

	panic("unsafe sort parameter: " + p.SortBy)
}

func (p Params) SortDirection() string {
	if strings.HasPrefix(p.SortBy, "-") {
		return "DESC"
	}
	return "ASC"
}

func (p Params) Limit() int {
	return p.PageSize
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func Validate(v *validator.Validator, p *Params, applyDefaults bool) {
	if applyDefaults {
		if p.Page == 0 {
			p.Page = 1
		}
		if p.PageSize == 0 {
			p.PageSize = 20
		}
		if p.SortBy == "" {
			p.SortBy = "-created_at"
		}
		p.SortSafelist = JobSortSafelist
	}

	v.Check(p.Page > 0, "page", "must be greater than zero")
	v.Check(p.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(p.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(p.PageSize <= 100, "page_size", "must be a maximum of 100")
	v.Check(slices.Contains(p.SortSafelist, p.SortBy), "sortBy", "invalid sort value")
}

type Metadata struct {
	CurrentPage  int `json:"current_page,omitzero"`
	PageSize     int `json:"page_size,omitzero"`
	FirstPage    int `json:"first_page,omitzero"`
	LastPage     int `json:"last_page,omitzero"`
	TotalRecords int `json:"total_records,omitzero"`
}

func NewMetadata(totalRecords, page, pageSize int) *Metadata {
	if totalRecords == 0 {
		return &Metadata{}
	}

	return &Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     (totalRecords + pageSize - 1) / pageSize,
		TotalRecords: totalRecords,
	}
}
