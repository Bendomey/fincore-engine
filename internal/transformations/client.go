package transformations

import (
	"github.com/Bendomey/fincore-engine/internal/models"
)

// DBClientToRestClient transforms client db input to rest type
func DBClientToRestClient(i *models.Client, secret *string) interface{} {
	if i == nil {
		return nil
	}

	data := map[string]interface{}{
		"id":         i.ID.String(),
		"name":       i.Name,
		"email":      i.Email,
		"client_id":  i.ClientId,
		"created_at": i.CreatedAt,
		"updated_at": i.UpdatedAt,
	}

	// don't send this as output if it doesn't exist.
	if secret != nil {
		data["clientSecret"] = secret
	}

	return data
}
