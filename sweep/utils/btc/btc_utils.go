package btc

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
)

/*
	6f6d6e69          // "omni" (4 bytes)
	0000              // version (2 bytes)
	0000              // type = Simple Send (2 bytes)
	xxxxxxxx          // property ID (4 bytes)
	yyyyyyyyyyyyyyyy  // amount (8 bytes, big-endian uint64)
*/

const (
	omniMarker = "6f6d6e6900000000" // omni + version0 + type0
	payloadLen = 40
)

type OmniData struct {
	TokenID     uint32
	TokenAmount string
}

func ParseOmniUSDTData(omniScriptHex string) (*OmniData, bool) {
	// 兼容带 OP_RETURN 前缀的情况（常见 6a14 + 40 hex）
	if len(omniScriptHex) > payloadLen {
		// 更稳健的方式：如果以 6a14 开头则去掉，否则尝试去掉前 4 字符
		if len(omniScriptHex) >= 4 && (omniScriptHex[:4] == "6a14" || len(omniScriptHex) == 44) {
			omniScriptHex = omniScriptHex[4:]
		}
	}

	if len(omniScriptHex) != payloadLen {
		return nil, false
	}

	if omniScriptHex[:16] != omniMarker {
		return nil, false
	}

	data, err := hex.DecodeString(omniScriptHex[16:])
	if err != nil || len(data) != 12 { // 4 + 8
		return nil, false
	}

	tokenID := binary.BigEndian.Uint32(data[:4])
	if tokenID != 31 {
		return nil, false
	}

	tokenAmount := binary.BigEndian.Uint64(data[4:])

	return &OmniData{
		TokenID:     tokenID,
		TokenAmount: strconv.FormatUint(tokenAmount, 10),
	}, true
}

// func ParseOmniUSDTData(omniScriptHex string) (map[string]int, bool) {
// 	if len(omniScriptHex) != 40 {
// 		omniScriptHex = omniScriptHex[4:]
// 	}

// 	if omniScriptHex[:16] == "6f6d6e6900000000" {
// 		dataHex := omniScriptHex[16:]

// 		tokenID, err := hex.DecodeString(dataHex[:8])
// 		if err != nil {
// 			return nil, false
// 		}

// 		tokenAmount, err := hex.DecodeString(dataHex[8:])
// 		if err != nil {
// 			return nil, false
// 		}

// 		omniData := map[string]int{
// 			"token_id":     int(tokenID[0])<<24 | int(tokenID[1])<<16 | int(tokenID[2])<<8 | int(tokenID[3]),
// 			"token_amount": int(tokenAmount[0])<<24 | int(tokenAmount[1])<<16 | int(tokenAmount[2])<<8 | int(tokenAmount[3]),
// 		}
// 		return omniData, true
// 	}

// 	return nil, false
// }

const (
	// "omni" + version(2B) + message type(2B)，00000000 = Simple Send
	omniMagicHex = "6f6d6e6900000000"
	// magic(8B) + propertyID(4B) + amount(8B) = 20 字节 => 40 个 hex 字符
	payloadHexLen = 40
)

// OmniSimpleSend 表示解析出的 Omni Layer Simple Send OP_RETURN 数据。
type OmniSimpleSend struct {
	PropertyID uint32
	Amount     uint64
}

// ParseOmniSimpleSend 解析 Omni Layer Simple Send 交易的 OP_RETURN 脚本(hex)。
// 可传入完整脚本(含 OP_RETURN + push 前缀)或纯 payload。
func ParseOmniSimpleSend(scriptHex string) (*OmniSimpleSend, error) {
	switch len(scriptHex) {
	case payloadHexLen + 4: // 6a14 + payload
		scriptHex = scriptHex[4:]
	case payloadHexLen:
		// 已经是纯 payload
	default:
		return nil, fmt.Errorf("omni: unexpected script length %d", len(scriptHex))
	}

	if scriptHex[:16] != omniMagicHex {
		return nil, errors.New("omni: not a simple-send payload")
	}

	payload, err := hex.DecodeString(scriptHex[16:])
	if err != nil {
		return nil, fmt.Errorf("omni: decode payload: %w", err)
	}

	if len(payload) != 12 { // propertyID(4B) + amount(8B)
		return nil, fmt.Errorf("omni: unexpected payload length %d", len(payload))
	}

	return &OmniSimpleSend{
		PropertyID: binary.BigEndian.Uint32(payload[0:4]),
		Amount:     binary.BigEndian.Uint64(payload[4:12]),
	}, nil
}
