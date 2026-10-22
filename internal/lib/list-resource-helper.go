package lib

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"
)

// FilterQuery type to help generate filter for queries
type FilterQuery struct {
	Page      int            `json:"page"       validate:"gte=0"`
	PageSize  int            `json:"page_size"  validate:"gte=0"`
	Order     string         `json:"order"      validate:"omitempty,oneof=asc desc"`
	OrderBy   string         `json:"order_by"   validate:"omitempty"`
	Search    *Search        `json:"search"     validate:"omitempty"`
	DateRange *DateRangeType `json:"date_range" validate:"omitempty"`
	Populate  *[]string      `json:"populate"   validate:"omitempty"`
}

// DateRangeType
type DateRangeType struct {
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time"   validate:"required,gtfield=StartTime"`
}

// Search
type Search struct {
	Query        string   `json:"query"         validate:"required"`
	SearchFields []string `json:"search_fields" validate:"required,min=1"`
}

// GenerateQuery takes a loook at what is coming from client and then generates a sieve
func GenerateQuery(argument url.Values) (*FilterQuery, error) {
	filterResult := GenerateEmptyQuery()

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

		if pageSize > 0 {
			filterResult.PageSize = pageSize
		}
	}

	// order
	order := argument.Get("order")
	if order != "" {
		filterResult.Order = order
	}

	// orderBy
	orderBy := argument.Get("order_by")
	if orderBy != "" {
		filterResult.OrderBy = orderBy
	}

	// dateRange
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
			Query:        query,
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

func GenerateEmptyQuery() FilterQuery {
	return FilterQuery{
		Page:      1,
		PageSize:  10,
		Order:     "desc",
		OrderBy:   "created_at",
		Search:    nil,
		DateRange: nil,
		Populate:  nil,
	}
}

// ListOptions is the allowlist a list endpoint accepts for sorting, searching
// and eager-loading. Populate maps the API name to the GORM preload path.
type ListOptions struct {
	OrderBy      []string
	SearchFields []string
	Populate     map[string]string
}

// Restrict rejects any order_by, search field or populate value outside opts
// and rewrites Populate to preload paths. These values end up in raw SQL, so
// they must never pass through unchecked.
func (q *FilterQuery) Restrict(opts ListOptions) error {
	if q.OrderBy != "" && !slices.Contains(opts.OrderBy, q.OrderBy) {
		return fmt.Errorf("order_by must be one of: %s", strings.Join(opts.OrderBy, ", "))
	}

	if q.Search != nil {
		for _, field := range q.Search.SearchFields {
			if !slices.Contains(opts.SearchFields, field) {
				return fmt.Errorf("search_fields must be from: %s", strings.Join(opts.SearchFields, ", "))
			}
		}
	}

	populate, err := RestrictPopulate(q.Populate, opts.Populate)
	if err != nil {
		return err
	}
	q.Populate = populate

	return nil
}

// RestrictPopulate maps API populate names to preload paths, rejecting unknown names.
func RestrictPopulate(populate *[]string, allowed map[string]string) (*[]string, error) {
	if populate == nil {
		return nil, nil
	}

	paths := make([]string, 0, len(*populate))
	for _, field := range *populate {
		path, ok := allowed[field]
		if !ok {
			names := make([]string, 0, len(allowed))
			for name := range allowed {
				names = append(names, name)
			}
			slices.Sort(names)
			return nil, fmt.Errorf("populate must be from: %s", strings.Join(names, ", "))
		}
		paths = append(paths, path)
	}

	return &paths, nil
}

// ParseDateParam accepts a date (2006-01-02) or an RFC3339 timestamp. With
// endOfDay, a date-only value returns the start of the next day (UTC) so it
// can be used as an exclusive upper bound covering the whole day.
func ParseDateParam(value string, endOfDay bool) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}

	t, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q: use YYYY-MM-DD or RFC3339", value)
	}

	if endOfDay {
		return t.AddDate(0, 0, 1), nil
	}

	return t, nil
}
