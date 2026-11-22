package setup

import (
	"context"
	"fmt"
	"node/config"
	"node/global"
	"node/global/constant"
	"node/model"
	"node/outbox"
)

// RequiredConfirmations 各链入账确认数（P0 默认值，后续可配置化）。
func RequiredConfirmations(chainId uint) int {
	switch chainId {
	case constant.BTC_MAINNET, constant.BTC_TESTNET,
		constant.LTC_MAINNET, constant.LTC_TESTNET,
		constant.BCH_MAINNET, constant.BCH_TESTNET:
		return 3
	case constant.ETH_MAINNET, constant.ETH_SEPOLIA,
		constant.BSC_MAINNET, constant.BSC_TESTNET,
		constant.OP_MAINNET, constant.OP_SEPOLIA,
		constant.ARBITRUM_ONE, constant.ARBITRUM_NOVA, constant.ARBITRUM_SEPOLIA,
		constant.POL_MAINNET, constant.POL_TESTNET,
		constant.AVAX_MAINNET, constant.AVAX_TESTNET,
		constant.BASE_MAINNET, constant.BASE_SEPOLIA:
		return 12
	case constant.TRON_MAINNET, constant.TRON_NILE:
		return 20
	case constant.SOL_MAINNET, constant.SOL_DEVNET:
		return 32
	case constant.XRP_MAINNET, constant.XRP_TESTNET:
		return 1
	default:
		return 12
	}
}

// UpdateConfirmations 按 tip 刷新 detected 事件的确认数与状态。
func UpdateConfirmations(ctx context.Context, chainId uint, tipHeight int64) {
	if global.NODE_DB == nil || tipHeight <= 0 {
		return
	}
	required := RequiredConfirmations(chainId)

	var readyIDs []uint
	if global.NODE_CONFIG.Blockchain.EffectiveNotifyMode() == config.NotifyModeConfirmed {
		_ = global.NODE_DB.WithContext(ctx).Model(&model.OwnTransaction{}).
			Where("chain_id = ? AND block_number > 0 AND tx_status = ? AND deleted_at IS NULL AND (? - block_number + 1) >= ?",
				chainId, model.TxStatusDetected, tipHeight, required).
			Pluck("id", &readyIDs).Error
	}

	if err := global.NODE_DB.WithContext(ctx).Exec(`
UPDATE node_own_transactions
SET confirmations = GREATEST(0, ? - block_number + 1),
    tx_status = CASE
      WHEN block_number > 0 AND (? - block_number + 1) >= ? THEN ?
      ELSE ?
    END,
    updated_at = NOW(3)
WHERE chain_id = ?
  AND block_number > 0
  AND tx_status = ?
  AND deleted_at IS NULL
`, tipHeight, tipHeight, required, model.TxStatusConfirmed, model.TxStatusDetected, chainId, model.TxStatusDetected).Error; err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	for _, id := range readyIDs {
		if err := outbox.EnqueueOwnTx(ctx, id); err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("enqueue confirmed own tx failed, id=%d: %s", id, err.Error()))
		}
	}
}
