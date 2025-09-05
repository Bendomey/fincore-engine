package transformations

import (
	"github.com/Bendomey/fincore-engine/internal/models"
)

// DBAccountToRestAccount transforms account db input to rest type
func DBAccountToRestAccount(i *models.Account, populate *[]string) interface{} {
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
		"is_group":          i.IsGroup,
		"parent_account_id": i.ParentAccountID,
		"created_at":        i.CreatedAt,
		"updated_at":        i.UpdatedAt,
	}

	if populate != nil {
		for _, field := range *populate {
			switch field {
			case "ParentAccount":
				data["parent_account"] = DBAccountToRestAccount(i.ParentAccount, populate)
			}
		}
	}

	return data
}
