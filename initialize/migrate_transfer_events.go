package initialize

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"node/global"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// migrateTransferEventSchema 将旧 uniqueIndex 迁移为事件幂等键，并回填 transfer_key。
func migrateTransferEventSchema(db *gorm.DB) error {
	if !db.Migrator().HasTable("node_own_transactions") {
		return nil
	}

	ensureColumn := func(name, ddl string) error {
		var count int64
		if err := db.Raw(`SELECT COUNT(*) FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'node_own_transactions' AND COLUMN_NAME = ?`, name).Scan(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := db.Exec(ddl).Error; err != nil {
				return fmt.Errorf("add column %s: %w", name, err)
			}
		}
		return nil
	}

	if err := ensureColumn("transfer_key", "ALTER TABLE node_own_transactions ADD COLUMN transfer_key varchar(64) NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := ensureColumn("transfer_index", "ALTER TABLE node_own_transactions ADD COLUMN transfer_index int NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumn("block_number", "ALTER TABLE node_own_transactions ADD COLUMN block_number bigint NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumn("confirmations", "ALTER TABLE node_own_transactions ADD COLUMN confirmations int NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumn("tx_status", "ALTER TABLE node_own_transactions ADD COLUMN tx_status varchar(16) NOT NULL DEFAULT 'detected'"); err != nil {
		return err
	}

	type row struct {
		ID           uint
		FromAddress  string
		ToAddress    string
		Amount       string
		TransactType string
		TransferKey  string
	}
	var rows []row
	if err := db.Table("node_own_transactions").
		Select("id, from_address, to_address, amount, transact_type, transfer_key").
		Where("transfer_key = '' OR transfer_key IS NULL").
		Find(&rows).Error; err != nil {
		return fmt.Errorf("load rows for transfer_key backfill: %w", err)
	}
	for _, r := range rows {
		raw := fmt.Sprintf("%s|%s|%s|%s",
			strings.ToLower(r.FromAddress),
			strings.ToLower(r.ToAddress),
			r.Amount,
			strings.ToLower(r.TransactType),
		)
		sum := sha256.Sum256([]byte(raw))
		key := "fp:" + hex.EncodeToString(sum[:16])
		if err := db.Exec("UPDATE node_own_transactions SET transfer_key = ? WHERE id = ?", key, r.ID).Error; err != nil {
			return fmt.Errorf("backfill transfer_key id=%d: %w", r.ID, err)
		}
	}

	for _, idx := range []string{"idx_own_tx_dedupe", "idx_transfer_event"} {
		var cnt int64
		_ = db.Raw(`SELECT COUNT(*) FROM information_schema.STATISTICS
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'node_own_transactions' AND INDEX_NAME = ?`, idx).Scan(&cnt).Error
		if cnt > 0 {
			if err := db.Exec("DROP INDEX `" + idx + "` ON node_own_transactions").Error; err != nil {
				global.NODE_LOG.Warn("drop index failed", zap.String("index", idx), zap.Error(err))
			}
		}
	}

	if err := db.Exec(`
DELETE t1 FROM node_own_transactions t1
INNER JOIN node_own_transactions t2
  ON t1.chain_id = t2.chain_id
 AND t1.hash = t2.hash
 AND t1.address = t2.address
 AND t1.transact_type = t2.transact_type
 AND t1.token = t2.token
 AND t1.transfer_key = t2.transfer_key
 AND t1.id > t2.id`).Error; err != nil {
		global.NODE_LOG.Warn("dedupe transfer events failed", zap.Error(err))
	}

	return nil
}
