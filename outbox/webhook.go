package outbox

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"node/global"
	"node/model"
	"strconv"
	"time"
)

type eventPayload struct {
	ID             uint   `json:"id"`
	ChainId        uint   `json:"chain_id"`
	Hash           string `json:"hash"`
	Address        string `json:"address"`
	FromAddress    string `json:"from_address"`
	ToAddress      string `json:"to_address"`
	Token          string `json:"token"`
	TransactType   string `json:"transact_type"`
	TransferKey    string `json:"transfer_key"`
	TransferIndex  int    `json:"transfer_index"`
	Amount         string `json:"amount"`
	BlockNumber    int64  `json:"block_number"`
	BlockTimestamp int    `json:"block_timestamp"`
	Confirmations  int    `json:"confirmations"`
	TxStatus       string `json:"tx_status"`
}

func deliverWebhook(ctx context.Context, row *model.NotificationOutbox) error {
	if row.Target == "" {
		return fmt.Errorf("webhook target empty")
	}
	var tx model.OwnTransaction
	if err := global.NODE_DB.WithContext(ctx).Where("id = ?", row.OwnTxID).First(&tx).Error; err != nil {
		return fmt.Errorf("load own tx %d: %w", row.OwnTxID, err)
	}

	body, err := json.Marshal(eventPayload{
		ID:             tx.ID,
		ChainId:        tx.ChainId,
		Hash:           tx.Hash,
		Address:        tx.Address,
		FromAddress:    tx.FromAddress,
		ToAddress:      tx.ToAddress,
		Token:          tx.Token,
		TransactType:   tx.TransactType,
		TransferKey:    tx.TransferKey,
		TransferIndex:  tx.TransferIndex,
		Amount:         tx.Amount,
		BlockNumber:    tx.BlockNumber,
		BlockTimestamp: tx.BlockTimestamp,
		Confirmations:  tx.Confirmations,
		TxStatus:       tx.TxStatus,
	})
	if err != nil {
		return err
	}

	timeout := 10 * time.Second
	if sec := global.NODE_CONFIG.Webhook.TimeoutSeconds; sec > 0 {
		timeout = time.Duration(sec) * time.Second
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, row.Target, bytes.NewReader(body))
	if err != nil {
		return err
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CryptoPay-Timestamp", ts)
	req.Header.Set("X-CryptoPay-Event-Id", strconv.FormatUint(uint64(tx.ID), 10))
	req.Header.Set("X-CryptoPay-Delivery-Id", strconv.FormatUint(uint64(row.ID), 10))
	if secret := global.NODE_CONFIG.Webhook.Secret; secret != "" {
		req.Header.Set("X-CryptoPay-Signature", signBody(secret, ts, body))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook status %d", resp.StatusCode)
	}
	return nil
}

func signBody(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp))
	_, _ = mac.Write([]byte{'.'})
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature 供接收方校验 Webhook 签名。
// signature 形如 sha256=<hex>；timestamp 为请求头 X-CryptoPay-Timestamp。
func VerifySignature(secret, timestamp, signature string, body []byte) bool {
	if secret == "" || timestamp == "" || signature == "" {
		return false
	}
	expected := signBody(secret, timestamp, body)
	return hmac.Equal([]byte(expected), []byte(signature))
}
