package lib

import (
	"net/url"
	"strings"
	"time"
)

// FilterQuery type to help generate filter for queries
type FilterQuery struct {
	Page      int            `json:"page" validate:"gte=0"`
	PageSize  int            `json:"page_size" validate:"gte=0"`
	Order     string         `json:"order" validate:"omitempty,oneof=asc desc"`
	OrderBy   string         `json:"order_by" validate:"omitempty"`
	Search    *Search        `json:"search" validate:"omitempty,dive"`
	DateRange *DateRangeType `json:"date_range" validate:"omitempty,dive"`
	Populate  *[]string      `json:"populate" validate:"omitempty,dive"`
}

// DateRangeType
type DateRangeType struct {
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required,gtfield=StartTime"`
}

// Search
type Search struct {
	Criteria     string   `json:"criteria" validate:"required"`
	SearchFields []string `json:"search_fields" validate:"required,min=1,dive,required"`
}

// GenerateQuery takes a loook at what is coming from client and then generates a sieve
func GenerateQuery(argument url.Values) (*FilterQuery, error) {
	filterResult := FilterQuery{
		Page:      1,
		PageSize:  10,
		Order:     "desc",
		OrderBy:   "created_at",
		Search:    nil,
		DateRange: nil,
		Populate:  nil,
	}

	page := argument.Get("page")
	if page != "" {
		page, err := ConvertStringToInt(page)
		if err != nil {
			return nil, err
		}

		filterResult.Page = page
	}

	pageSize := argument.Get("page_size")
	if pageSize != "" {
		pageSize, err := ConvertStringToInt(pageSize)
		if err != nil {
			return nil, err
		}
		filterResult.PageSize = pageSize
	}

	//order
	order := argument.Get("order")
	if order != "" {
		filterResult.Order = order
	}

	//orderBy
	orderBy := argument.Get("order_by")
	if orderBy != "" {
		filterResult.OrderBy = orderBy
	}

	//dateRange
	startDate := argument.Get("start_date")
	endDate := argument.Get("end_date")

	if startDate != "" && endDate != "" {
		start, startErr := time.Parse(time.RFC3339, startDate)
		if startErr != nil {
			return nil, startErr
		}

		end, endErr := time.Parse(time.RFC3339, endDate)
		if endErr != nil {
			return nil, endErr
		}

		filterResult.DateRange = &DateRangeType{
			StartTime: start,
			EndTime:   end,
		}
	}

	query := argument.Get("query")
	searchFields := argument.Get("search_fields")

	if query != "" && searchFields != "" {
		fields := strings.Split(searchFields, ",")

		filterResult.Search = &Search{
			Criteria:     query,
			SearchFields: fields,
		}
	}

	populate := argument.Get("populate")

	if populate != "" {
		fields := strings.Split(populate, ",")
		filterResult.Populate = &fields
	}

	return &filterResult, nil
}
