package transformations

import (
	"github.com/Bendomey/fincore-engine/internal/models"
)

// DBAccountToRestAccount transforms account db input to rest type
func DBAccountToRestAccount(i *models.Account) interface{} {
	if i == nil {
		return nil
	}

	data := map[string]interface{}{
		"id":                i.ID.String(),
		"code":              i.Code,
		"name":              i.Name,
		"description":       i.Description,
		"type":              i.Type,
		"is_contra":         i.IsContra,
		"parent_account_id": i.ParentAccountID,
		"parent_account":    DBAccountToRestAccount(i.ParentAccount),
		"created_at":        i.CreatedAt,
		"updated_at":        i.UpdatedAt,
	}

	return data
}
