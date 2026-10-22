package jobs

import (
	"github.com/go-gormigrate/gormigrate/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// PerClientCodesAndUniqueReferences makes account codes unique per client
// instead of globally, and journal entry references unique per client.
//
// Existing codes are already globally unique, so the per-client index can't
// fail and no code is renumbered. Duplicate live references are renamed,
// keeping the oldest entry's reference; each renamed entry keeps the original
// in metadata.original_reference.
//
// Run this before deploying the code that generates codes per client: that
// code would collide with the old global index.
func PerClientCodesAndUniqueReferences() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20261021_per_client_codes_unique_references",
		Migrate: func(db *gorm.DB) error {
			return db.Transaction(func(tx *gorm.DB) error {
				if err := tx.Exec(`DROP INDEX IF EXISTS idx_accounts_code`).Error; err != nil {
					return err
				}

				if err := tx.Exec(
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_client_code ON accounts (client_id, code)`,
				).Error; err != nil {
					return err
				}

				renamed := tx.Exec(`
WITH ranked AS (
  SELECT id, reference,
         ROW_NUMBER() OVER (PARTITION BY client_id, reference ORDER BY created_at, id) AS rn
  FROM journal_entries
  WHERE deleted_at IS NULL
)
UPDATE journal_entries je
SET reference  = ranked.reference || '-dup-' || LEFT(je.id::text, 8),
    metadata   = COALESCE(je.metadata::jsonb, '{}'::jsonb) || jsonb_build_object('original_reference', ranked.reference),
    updated_at = NOW()
FROM ranked
WHERE je.id = ranked.id AND ranked.rn > 1`)
				if renamed.Error != nil {
					return renamed.Error
				}
				log.Infof("[Migration] renamed %d duplicate journal entry references", renamed.RowsAffected)

				return tx.Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS idx_journal_entries_client_reference
ON journal_entries (client_id, reference) WHERE deleted_at IS NULL`).Error
			})
		},
		Rollback: func(db *gorm.DB) error {
			// Renamed references stay renamed; the originals remain in metadata.
			return db.Transaction(func(tx *gorm.DB) error {
				if err := tx.Exec(`DROP INDEX IF EXISTS idx_journal_entries_client_reference`).Error; err != nil {
					return err
				}

				if err := tx.Exec(`DROP INDEX IF EXISTS idx_accounts_client_code`).Error; err != nil {
					return err
				}

				return tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_code ON accounts (code)`).Error
			})
		},
	}
}
