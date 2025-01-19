package btc

import "encoding/hex"

func ParseOmniUSDTData(omniScriptHex string) (map[string]int, bool) {
	if len(omniScriptHex) != 40 {
		omniScriptHex = omniScriptHex[4:]
	}

	if omniScriptHex[:16] == "6f6d6e6900000000" {
		dataHex := omniScriptHex[16:]

		tokenID, err := hex.DecodeString(dataHex[:8])
		if err != nil {
			return nil, false
		}

		tokenAmount, err := hex.DecodeString(dataHex[8:])
		if err != nil {
			return nil, false
		}

		omniData := map[string]int{
			"token_id":     int(tokenID[0])<<24 | int(tokenID[1])<<16 | int(tokenID[2])<<8 | int(tokenID[3]),
			"token_amount": int(tokenAmount[0])<<24 | int(tokenAmount[1])<<16 | int(tokenAmount[2])<<8 | int(tokenAmount[3]),
		}
		return omniData, true
	}

	return nil, false
}
