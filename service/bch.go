package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"node/global"
	"node/global/constant"
	"node/model/node/request"
	"node/model/node/response"
	"node/model/node/response/blockstream"
	"node/utils/electrum"
	"sort"
	"strings"
	"time"
)

// trezor/blockbook style address txs (fallback when Esplora is unavailable)
type bchAddressTxResponse struct {
	Transactions []struct {
		TxID          string `json:"txid"`
		BlockTime     int64  `json:"blockTime"`
		Confirmations int64  `json:"confirmations"`
		Vin           []struct {
			Addresses []string `json:"addresses"`
			Value     string   `json:"value"`
		} `json:"vin"`
		Vout []struct {
			Addresses []string `json:"addresses"`
			Value     string   `json:"value"`
		} `json:"vout"`
	} `json:"transactions"`
}

func (n *NService) GetBchTransactions(ctx context.Context, req request.GetBchTransactions) ([]response.ClientTransaction, error) {
	txs, err := n.fetchBchTransactions(ctx, req.ChainId, req.Address)
	if err != nil {
		global.NODE_LOG.Error(fmt.Sprintf("bch txs failed: %v", err))
		return []response.ClientTransaction{}, nil
	}
	return filterAndSortEvmTxs(txs, req.Asset), nil
}

func bchBlockbookBases(chainId uint) []string {
	switch chainId {
	case constant.BCH_MAINNET:
		return []string{
			"https://bch1.trezor.io/api/v2",
			"https://bch2.trezor.io/api/v2",
		}
	case constant.BCH_TESTNET:
		return []string{
			"https://bchtest1.trezor.io/api/v2",
		}
	default:
		return nil
	}
}

func bchEsploraConfirmedURLs(chainId uint, address string) []string {
	switch chainId {
	case constant.BCH_MAINNET:
		return []string{
			fmt.Sprintf("https://bchexplorer.cash/api/address/%s/txs", address),
			fmt.Sprintf("https://bchexplorer.cash/api/address/%s/txs", url.PathEscape(address)),
		}
	default:
		return nil
	}
}

func bchEsploraMempoolURL(chainId uint, address string) string {
	switch chainId {
	case constant.BCH_MAINNET:
		return fmt.Sprintf("https://bchexplorer.cash/api/address/%s/txs/mempool", address)
	default:
		return ""
	}
}

func bchElectrumEndpoints(chainId uint) []electrum.Endpoint {
	switch chainId {
	case constant.BCH_TESTNET:
		// Public chipnet Fulcrum servers (BCH "testnet" in this stack is chipnet).
		return []electrum.Endpoint{
			{Host: "chipnet.imaginary.cash", Port: 50002, TLS: true},
			{Host: "chipnet.bch.ninja", Port: 50002, TLS: true},
			{Host: "blackie.c3-soft.com", Port: 64002, TLS: true},
		}
	case constant.BCH_MAINNET:
		return []electrum.Endpoint{
			{Host: "electrum.imaginary.cash", Port: 50002, TLS: true},
			{Host: "fulcrum.fountainhead.cash", Port: 50002, TLS: true},
		}
	default:
		return nil
	}
}

func (n *NService) fetchBchTransactions(ctx context.Context, chainId uint, address string) ([]response.ClientTransaction, error) {
	if address == "" {
		return nil, fmt.Errorf("empty bch address")
	}

	var lastErr error

	// Prefer Esplora (same JSON shape as BTC/LTC) when configured (mainnet today).
	if len(bchEsploraConfirmedURLs(chainId, address)) > 0 {
		if txs, err := n.fetchBchTransactionsEsplora(ctx, chainId, address); err == nil {
			return txs, nil
		} else {
			lastErr = err
			global.NODE_LOG.Error(fmt.Sprintf("bch esplora txs failed chain=%d: %v", chainId, err))
		}
	}

	// Chipnet/testnet public Blockbook is often down; Fulcrum/ElectrumX is the reliable path.
	if eps := bchElectrumEndpoints(chainId); len(eps) > 0 {
		if txs, err := n.fetchBchTransactionsElectrum(ctx, chainId, address, eps); err == nil {
			return txs, nil
		} else {
			lastErr = err
			global.NODE_LOG.Error(fmt.Sprintf("bch electrum txs failed chain=%d: %v", chainId, err))
		}
	}

	bases := bchBlockbookBases(chainId)
	for _, base := range bases {
		txs, err := n.fetchBchTransactionsBlockbook(ctx, chainId, address, base)
		if err != nil {
			lastErr = err
			global.NODE_LOG.Error(fmt.Sprintf("bch blockbook txs failed base=%s: %v", base, err))
			continue
		}
		return txs, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("unsupported bch chain %d", chainId)
}

func bchHTTPGetJSON(ctx context.Context, rawURL string, dest any) error {
	// Avoid utils/http.HTTPGet's forced Content-Type:application/json on GET —
	// Cloudflare fronting bchexplorer treats that as a bot and returns PoW HTML.
	timeout := 15 * time.Second
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/html,application/json,application/xhtml+xml,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://bchexplorer.cash/")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect the %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		snippet := string(body)
		if len(snippet) > 240 {
			snippet = snippet[:240]
		}
		return fmt.Errorf("unexpected status %d from %s, body: %s", resp.StatusCode, rawURL, snippet)
	}
	if len(body) > 0 && body[0] == '<' {
		return fmt.Errorf("non-json html response from %s", rawURL)
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("failed to unmarshal response from %s: %w", rawURL, err)
	}
	return nil
}

func (n *NService) fetchBchTransactionsEsplora(ctx context.Context, chainId uint, address string) ([]response.ClientTransaction, error) {
	confirmedURLs := bchEsploraConfirmedURLs(chainId, address)
	if len(confirmedURLs) == 0 {
		return nil, fmt.Errorf("no bch esplora url for chain %d", chainId)
	}

	var raw []blockstream.TransactionResponse
	var lastErr error
	for _, u := range confirmedURLs {
		var rows []blockstream.TransactionResponse
		if err := bchHTTPGetJSON(ctx, u, &rows); err != nil {
			lastErr = err
			continue
		}
		raw = append(raw, rows...)
		break
	}

	// Mempool is best-effort; never discard a successful confirmed list.
	if memURL := bchEsploraMempoolURL(chainId, address); memURL != "" {
		var pending []blockstream.TransactionResponse
		if err := bchHTTPGetJSON(ctx, memURL, &pending); err != nil {
			global.NODE_LOG.Error(fmt.Sprintf("bch esplora mempool optional: %v", err))
		} else {
			raw = append(raw, pending...)
		}
	}

	if len(raw) == 0 {
		if lastErr != nil {
			return nil, lastErr
		}
		return []response.ClientTransaction{}, nil
	}

	seen := make(map[string]response.ClientTransaction, len(raw))
	order := make([]string, 0, len(raw))
	for _, v := range raw {
		model, err := decodeBchEsploraTx(chainId, address, v)
		if err != nil || model.Hash == "" {
			continue
		}
		key := strings.ToLower(model.Hash)
		if existing, ok := seen[key]; ok {
			if existing.Status != "Success" && model.Status == "Success" {
				seen[key] = model
			}
			continue
		}
		seen[key] = model
		order = append(order, key)
	}

	out := make([]response.ClientTransaction, 0, len(order))
	for _, key := range order {
		out = append(out, seen[key])
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].BlockTimestamp > out[j].BlockTimestamp
	})
	return out, nil
}

func (n *NService) fetchBchTransactionsBlockbook(ctx context.Context, chainId uint, address, base string) ([]response.ClientTransaction, error) {
	var resp bchAddressTxResponse
	u := fmt.Sprintf("%s/address/%s?details=txs&pageSize=50", base, url.PathEscape(address))
	if err := bchHTTPGetJSON(ctx, u, &resp); err != nil {
		return nil, err
	}

	var out []response.ClientTransaction
	for _, tx := range resp.Transactions {
		sent := false
		received := false
		var sentVal, recvVal int64

		for _, vin := range tx.Vin {
			for _, a := range vin.Addresses {
				if bchAddressEqual(a, address) {
					sent = true
					sentVal += parseLooseInt64(vin.Value)
				}
			}
		}
		for _, vout := range tx.Vout {
			for _, a := range vout.Addresses {
				if bchAddressEqual(a, address) {
					received = true
					recvVal += parseLooseInt64(vout.Value)
				}
			}
		}

		model := response.ClientTransaction{
			ChainId:        chainId,
			Address:        address,
			Hash:           tx.TxID,
			Asset:          constant.BCH,
			Category:       "bch",
			Status:         "Success",
			BlockTimestamp: int(tx.BlockTime * 1000),
		}
		if tx.Confirmations == 0 {
			model.Status = "Pending"
		}

		switch {
		case sent && !received:
			model.Type = "Send"
			model.Amount = formatSatoshis(sentVal)
		case received && !sent:
			model.Type = "Received"
			model.Amount = formatSatoshis(recvVal)
		case sent && received:
			net := sentVal - recvVal
			if net < 0 {
				net = -net
				model.Type = "Received"
			} else {
				model.Type = "Send"
			}
			model.Amount = formatSatoshis(net)
		default:
			continue
		}
		if model.Amount == "" || model.Amount == "0" {
			continue
		}
		out = append(out, model)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].BlockTimestamp > out[j].BlockTimestamp
	})
	return out, nil
}

type bchElectrumHistoryItem struct {
	Height int64  `json:"height"`
	TxHash string `json:"tx_hash"`
}

type bchElectrumVerboseTx struct {
	TxID          string `json:"txid"`
	Hash          string `json:"hash"`
	BlockTime     int64  `json:"blocktime"`
	Time          int64  `json:"time"`
	Confirmations int64  `json:"confirmations"`
	Vin           []struct {
		TxID string `json:"txid"`
		Vout int    `json:"vout"`
	} `json:"vin"`
	Vout []struct {
		Value        float64 `json:"value"`
		ScriptPubKey struct {
			Addresses []string `json:"addresses"`
		} `json:"scriptPubKey"`
	} `json:"vout"`
}

func (n *NService) fetchBchTransactionsElectrum(ctx context.Context, chainId uint, address string, endpoints []electrum.Endpoint) ([]response.ClientTransaction, error) {
	cashAddr := ensureBchCashAddr(chainId, address)

	var history []bchElectrumHistoryItem
	ep, err := electrum.CallFirst(ctx, endpoints, "blockchain.address.get_history", []any{cashAddr}, &history)
	if err != nil {
		return nil, err
	}

	const maxTxs = 50
	if len(history) > maxTxs {
		history = history[len(history)-maxTxs:]
	}

	prevCache := make(map[string]*bchElectrumVerboseTx)
	getTx := func(txid string) (*bchElectrumVerboseTx, error) {
		if txid == "" {
			return nil, fmt.Errorf("empty txid")
		}
		if cached, ok := prevCache[txid]; ok {
			return cached, nil
		}
		var tx bchElectrumVerboseTx
		if err := electrum.Call(ctx, ep, "blockchain.transaction.get", []any{txid, true}, &tx); err != nil {
			return nil, err
		}
		prevCache[txid] = &tx
		return &tx, nil
	}

	out := make([]response.ClientTransaction, 0, len(history))
	for i := len(history) - 1; i >= 0; i-- {
		item := history[i]
		tx, err := getTx(item.TxHash)
		if err != nil || tx == nil {
			continue
		}
		model, err := decodeBchElectrumTx(chainId, cashAddr, tx, getTx)
		if err != nil || model.Hash == "" {
			continue
		}
		// Electrum history height 0 / negative means mempool.
		if item.Height <= 0 {
			model.Status = "Pending"
		}
		out = append(out, model)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].BlockTimestamp > out[j].BlockTimestamp
	})
	return out, nil
}

func decodeBchElectrumTx(
	chainId uint,
	address string,
	tx *bchElectrumVerboseTx,
	getTx func(string) (*bchElectrumVerboseTx, error),
) (response.ClientTransaction, error) {
	var model response.ClientTransaction
	txid := tx.TxID
	if txid == "" {
		txid = tx.Hash
	}
	if txid == "" {
		return model, fmt.Errorf("empty txid")
	}

	var sentSats, recvSats int64
	for _, vin := range tx.Vin {
		if vin.TxID == "" {
			continue
		}
		prev, err := getTx(vin.TxID)
		if err != nil || prev == nil || vin.Vout < 0 || vin.Vout >= len(prev.Vout) {
			continue
		}
		prevOut := prev.Vout[vin.Vout]
		for _, a := range prevOut.ScriptPubKey.Addresses {
			if bchAddressEqual(a, address) {
				sentSats += bchCoinsToSats(prevOut.Value)
				break
			}
		}
	}
	for _, vout := range tx.Vout {
		for _, a := range vout.ScriptPubKey.Addresses {
			if bchAddressEqual(a, address) {
				recvSats += bchCoinsToSats(vout.Value)
				break
			}
		}
	}

	model.ChainId = chainId
	model.Address = address
	model.Hash = txid
	model.Asset = constant.BCH
	model.Category = "bch"
	if tx.Confirmations > 0 {
		model.Status = "Success"
	} else {
		model.Status = "Pending"
	}
	ts := tx.BlockTime
	if ts == 0 {
		ts = tx.Time
	}
	if ts > 0 {
		model.BlockTimestamp = int(ts * 1000)
	}

	switch {
	case sentSats > 0 && recvSats == 0:
		model.Type = "Send"
		model.Amount = formatSatoshis(sentSats)
	case sentSats == 0 && recvSats > 0:
		model.Type = "Received"
		model.Amount = formatSatoshis(recvSats)
	case sentSats > 0 && recvSats > 0:
		net := sentSats - recvSats
		if net < 0 {
			model.Type = "Received"
			model.Amount = formatSatoshis(-net)
		} else {
			model.Type = "Send"
			model.Amount = formatSatoshis(net)
		}
	default:
		return model, fmt.Errorf("address not involved")
	}
	if model.Amount == "" || model.Amount == "0" {
		return model, fmt.Errorf("zero amount")
	}
	return model, nil
}

func ensureBchCashAddr(chainId uint, address string) string {
	s := strings.TrimSpace(address)
	if s == "" {
		return s
	}
	if strings.Contains(s, ":") {
		return s
	}
	prefix := "bitcoincash"
	if chainId == constant.BCH_TESTNET {
		prefix = "bchtest"
	}
	return prefix + ":" + s
}

func bchCoinsToSats(coins float64) int64 {
	if coins <= 0 {
		return 0
	}
	return int64(math.Round(coins * 1e8))
}

func decodeBchEsploraTx(chainId uint, address string, tx blockstream.TransactionResponse) (response.ClientTransaction, error) {
	var model response.ClientTransaction
	if tx.Txid == "" {
		return model, fmt.Errorf("empty txid")
	}

	sent := false
	var sentSats, recvSats int64
	for _, vin := range tx.Vin {
		if bchAddressEqual(vin.Prevout.ScriptpubkeyAddress, address) {
			sent = true
			sentSats += int64(vin.Prevout.Value)
		}
	}
	for _, vout := range tx.Vout {
		if bchAddressEqual(vout.ScriptpubkeyAddress, address) {
			recvSats += int64(vout.Value)
		}
	}

	model.ChainId = chainId
	model.Address = address
	model.Hash = tx.Txid
	model.Asset = constant.BCH
	model.Category = "bch"
	if tx.Status.Confirmed {
		model.Status = "Success"
		model.BlockTimestamp = int(tx.Status.BlockTime * 1000)
	} else {
		model.Status = "Pending"
		model.BlockTimestamp = 0
	}

	switch {
	case sent && recvSats == 0:
		// Approximate send as outputs not to self.
		var outSats int64
		for _, vout := range tx.Vout {
			if !bchAddressEqual(vout.ScriptpubkeyAddress, address) {
				outSats += int64(vout.Value)
			}
		}
		model.Type = "Send"
		model.Amount = formatSatoshis(outSats)
	case !sent && recvSats > 0:
		model.Type = "Received"
		model.Amount = formatSatoshis(recvSats)
	case sent && recvSats > 0:
		net := sentSats - recvSats
		if net < 0 {
			model.Type = "Received"
			model.Amount = formatSatoshis(-net)
		} else {
			model.Type = "Send"
			model.Amount = formatSatoshis(net)
		}
	default:
		return model, fmt.Errorf("address not involved")
	}

	if model.Amount == "" || model.Amount == "0" {
		return model, fmt.Errorf("zero amount")
	}
	return model, nil
}

// CashAddr may appear with or without bitcoincash:/bchtest: prefix.
func bchAddressEqual(a, b string) bool {
	return normalizeBchAddress(a) != "" && normalizeBchAddress(a) == normalizeBchAddress(b)
}

func normalizeBchAddress(addr string) string {
	s := strings.TrimSpace(strings.ToLower(addr))
	if s == "" {
		return ""
	}
	if i := strings.IndexByte(s, ':'); i >= 0 && i+1 < len(s) {
		s = s[i+1:]
	}
	return s
}

func parseLooseInt64(v string) int64 {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	var n int64
	for _, c := range v {
		if c < '0' || c > '9' {
			continue
		}
		n = n*10 + int64(c-'0')
	}
	return n
}

func formatSatoshis(sats int64) string {
	if sats <= 0 {
		return "0"
	}
	whole := sats / 1e8
	frac := sats % 1e8
	if frac == 0 {
		return fmt.Sprintf("%d", whole)
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%d.%08d", whole, frac), "0"), ".")
}
