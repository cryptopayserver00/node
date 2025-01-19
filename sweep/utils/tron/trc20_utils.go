package tron

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"node/global"
	"strings"
)

func TronDecodeMethod(data string) (string, string, error) {
	if len(data) < 8 {
		return "", "", errors.New("length too short for tron decode method")
	}
	return data[:8], data[8:], nil
}

var (
	Transfer     = "transfer"
	TransferFrom = "transferFrom"

	BatchTransferFrom  = "batchTransferFrom"
	Collect            = "collect"
	SenderTransferFrom = "senderTransferFrom"
	Withdraw           = "withdraw"

	TransferContract     = "TransferContract"
	TriggerSmartContract = "TriggerSmartContract"

	KnownMethods = map[string]string{
		"a9059cbb": Transfer,
		"23b872dd": TransferFrom,

		"b818f9e4": BatchTransferFrom,
		"1e13eee1": Collect,
		"b19385f7": SenderTransferFrom,
		"f7ece0cf": Withdraw,
	}
)

var base58Alphabets = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func ToHexAddress(address string) string {
	return hex.EncodeToString(base58Decode([]byte(address)))
}

func FromHexAddress(hexAddress string) (string, error) {
	addrByte, err := hex.DecodeString(hexAddress)
	if err != nil {
		return "", err
	}

	sha := sha256.New()
	sha.Write(addrByte)
	shaStr := sha.Sum(nil)

	sha2 := sha256.New()
	sha2.Write(shaStr)
	shaStr2 := sha2.Sum(nil)

	addrByte = append(addrByte, shaStr2[:4]...)
	return string(base58Encode(addrByte)), nil
}

func base58Encode(input []byte) []byte {
	x := big.NewInt(0).SetBytes(input)
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := &big.Int{}
	var result []byte
	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, base58Alphabets[mod.Int64()])
	}
	reverseBytes(result)
	return result
}

func base58Decode(input []byte) []byte {
	result := big.NewInt(0)
	for _, b := range input {
		charIndex := bytes.IndexByte(base58Alphabets, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}
	decoded := result.Bytes()
	if input[0] == base58Alphabets[0] {
		decoded = append([]byte{0x00}, decoded...)
	}
	return decoded[:len(decoded)-4]
}

func reverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

func IsHandleTransaction(chainId uint, hash, contractType, fromAddress, toAddress, monitorAddress, data string) bool {
	if fromAddress == "" || toAddress == "" {
		return false
	}

	switch contractType {
	case TransferContract:
		return handleTransferContract(fromAddress, toAddress, monitorAddress)
	case TriggerSmartContract:
		return handleTriggerSmartContract(chainId, hash, fromAddress, toAddress, monitorAddress, data)
	}

	return false
}

func handleTransferContract(fromAddress, toAddress, monitorAddress string) bool {
	sendAddress, err := FromHexAddress(fromAddress)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return false
	}
	receiveAddress, err := FromHexAddress(toAddress)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return false
	}
	if strings.EqualFold(sendAddress, monitorAddress) || strings.EqualFold(receiveAddress, monitorAddress) {
		return true
	}
	return false
}

func handleTriggerSmartContract(chainId uint, hash, fromAddress, toAddress, monitorAddress, data string) bool {

	methodID, _, _ := TronDecodeMethod(data)

	method := KnownMethods[methodID]
	if method == "" {
		return false
	}

	switch method {
	case Transfer:
		sendAddress, err := FromHexAddress(fromAddress)
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			return false
		}
		receiveAddress, err := FromHexAddress("41" + data[32:72])
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			return false
		}

		if strings.EqualFold(sendAddress, monitorAddress) || strings.EqualFold(receiveAddress, monitorAddress) {
			return true
		}
	case TransferFrom:
		sendAddress, err := FromHexAddress("41" + data[32:72])
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			return false
		}
		receiveAddress, err := FromHexAddress("41" + data[96:136])
		if err != nil {
			global.NODE_LOG.Error(err.Error())
			return false
		}

		if strings.EqualFold(sendAddress, monitorAddress) || strings.EqualFold(receiveAddress, monitorAddress) {
			return true
		}
	}

	return false
}
