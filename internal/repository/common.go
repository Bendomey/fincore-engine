package repository

import (
	"fmt"

	"github.com/Bendomey/fincore-engine/internal/lib"
	"gorm.io/gorm"
)

func DateRangeScope(tableName string, dateRange *lib.DateRangeType) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if dateRange == nil {
			return db
		}

		columnName := fmt.Sprintf("%s.created_at", tableName)
		return db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", columnName), dateRange.StartTime, dateRange.EndTime)
	}
}

func SearchScope(tableName string, search *lib.Search) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if search == nil || search.Query == "" {
			return db
		}

		// Group the ORs so the other scopes (notably the client filter) are
		// ANDed with the whole search, not just its first field.
		group := db.Session(&gorm.Session{NewDB: true})
		pattern := fmt.Sprintf("%%%s%%", search.Query)

		for index, singleQuery := range search.SearchFields {
			condition := fmt.Sprintf("%s.%s ILIKE ?", tableName, singleQuery)
			if index == 0 {
				group = group.Where(condition, pattern)
				continue
			}

			group = group.Or(condition, pattern)
		}

		return db.Where(group)
	}
}

func PaginationScope(page int, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page < 1 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func OrderScope(tableName string, orderBy string, order string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if orderBy == "" || order == "" {
			return db.Order(fmt.Sprintf("%s.created_at desc", tableName))
		}

		return db.Order(fmt.Sprintf("%s.%s %s", tableName, orderBy, order))
	}
}

func ClientFilterScope(tableName string, clientId string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if clientId == "" {
			return db
		}
		return db.Where(fmt.Sprintf("%s.client_id = ?", tableName), clientId)
	}
}
