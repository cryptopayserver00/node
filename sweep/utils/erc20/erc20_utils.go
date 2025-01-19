package erc20

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"node/global"
	"node/global/constant"
	"node/model/node/response"
	sweepUtils "node/sweep/utils"
	"node/utils"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	Transfer     = "transfer"
	TransferFrom = "transferFrom"
	Approve      = "approve"

	BatchTransferFrom  = "batchTransferFrom"
	Collect            = "collect"
	SenderTransferFrom = "senderTransferFrom"
	Withdraw           = "withdraw"

	knownMethods = map[string]string{
		"0xa9059cbb": Transfer,
		"0x23b872dd": TransferFrom,

		"0xb818f9e4": BatchTransferFrom,
		"0x1e13eee1": Collect,
		"0xb19385f7": SenderTransferFrom,
		"0xf7ece0cf": Withdraw,
	}
)

func IsHandleTokenTransaction(chainId uint, hash, contractName, fromAddress, contractAddress, monitorAddress, data string) bool {
	if !sweepUtils.IsChainJoinSweep(chainId) {
		return false
	}

	switch contractName {
	case constant.SWAP:
		break
	default:
		return handleERC20Transaction(chainId, hash, fromAddress, contractAddress, monitorAddress, data)
	}

	return false
}

func GetAllAddressByTransaction(chainId uint, from, hash, data string) ([]string, error) {
	_, decodeFromAddress, decodeToAddress, _, err := DecodeERC20TransactionInputData(chainId, hash, data)
	if err != nil {
		return nil, err
	}

	if decodeFromAddress == "" {
		decodeFromAddress = from
	}

	return []string{decodeFromAddress, decodeToAddress}, nil
}

func GetAllAddressByTransactionTwo(chainId uint, contractName, from, to, hash, data string) ([]string, error) {
	allAddress := make([]string, 0)
	switch contractName {
	default:
		_, decodeFromAddress, decodeToAddress, _, err := DecodeERC20TransactionInputData(chainId, hash, data)
		if err != nil {
			return nil, err
		}
		if decodeFromAddress == "" {
			decodeFromAddress = from
		}

		allAddress = append(allAddress, decodeFromAddress, decodeToAddress)
	}

	return allAddress, nil
}

func handleERC20Transaction(chainId uint, hash, fromAddress, contractAddress, monitorAddress, data string) bool {
	methodName, decodeFromAddress, decodeToAddress, _, err := DecodeERC20TransactionInputData(chainId, hash, data)

	if err != nil {
		return false
	}

	switch methodName {
	case Transfer:
		if utils.HexToAddress(monitorAddress) == utils.HexToAddress(fromAddress) {
			return true
		}

		if utils.HexToAddress(monitorAddress) == utils.HexToAddress(decodeToAddress) {
			return true
		}
	case TransferFrom:
		if utils.HexToAddress(monitorAddress) == utils.HexToAddress(decodeFromAddress) {
			return true
		}

		if utils.HexToAddress(monitorAddress) == utils.HexToAddress(decodeToAddress) {
			return true
		}
	}

	return false
}

func DecodeERC20TransactionInputData(chainId uint, hash, data string) (methodName, decodeFromAddress, decodeToAddress string, amount *big.Int, err error) {

	if len(data) < 138 {
		err = errors.New("insufficient data length")
		return
	}

	methodSigData := data[:10]
	inputSigData := data[10:]

	decodedData, err := hex.DecodeString(inputSigData)
	if err != nil {
		err = errors.New("can not decode input sig data")
		return
	}

	file, err := os.Open("json/ERC20.json")
	if err != nil {
		err = errors.New("can not open erc20 file")
		global.NODE_LOG.Error(fmt.Sprintf("%s -> hash:%s, message:%s", constant.GetChainName(chainId), hash, err.Error()))
		return
	}
	defer file.Close()

	contractABI, err := abi.JSON(file)
	if err != nil {
		err = errors.New("can not from file to json of abi")
		global.NODE_LOG.Error(fmt.Sprintf("%s -> hash:%s, message:%s", constant.GetChainName(chainId), hash, err.Error()))
		return
	}

	methodName, isKnownMethod := knownMethods[methodSigData]
	if !isKnownMethod || (methodName != Transfer && methodName != TransferFrom) {
		err = errors.New("not a valid transfer method")
		return
	}

	method, isAbiMethod := contractABI.Methods[methodName]
	if !isAbiMethod {
		err = errors.New("transfer method not found in ABI")
		return
	}

	inputsMap := make(map[string]interface{})

	if err = method.Inputs.UnpackIntoMap(inputsMap, decodedData); err != nil {
		err = errors.New("can not decode: Unpack Into Map")
		return
	}

	switch method.Name {
	case Transfer:
		address, ok := inputsMap["_to"].(common.Address)
		if !ok {
			err = errors.New("can not get the value of _to")
			return
		}

		value, ok := inputsMap["_value"].(*big.Int)
		if !ok {
			err = errors.New("can not get the value of _value")
			return
		}

		return method.Name, "", address.Hex(), value, nil
	case TransferFrom:
		fromAddress, ok := inputsMap["_from"].(common.Address)
		if !ok {
			err = errors.New("can not get the value of _from")
			return
		}
		toAddress, ok := inputsMap["_to"].(common.Address)
		if !ok {
			err = errors.New("can not get the value of _to")
			return
		}
		value, ok := inputsMap["_value"].(*big.Int)
		if !ok {
			err = errors.New("can not get the value of _value")
			return
		}

		return method.Name, fromAddress.Hex(), toAddress.Hex(), value, nil
	}

	err = errors.New("not found method")
	return
}

func IsHandleReceiptLogTransaction(chainId uint, log response.RPCReceiptLogs, monitorAddress string) bool {
	decodeFromAddress, decodeToAddress, _, _, _, err := DecodeERC20TransactionReceiptLog(chainId, log)

	if err != nil {
		return false
	}

	if utils.HexToAddress(monitorAddress) == utils.HexToAddress(decodeFromAddress) {
		return true
	}

	if utils.HexToAddress(monitorAddress) == utils.HexToAddress(decodeToAddress) {
		return true
	}

	return false
}

func DecodeERC20TransactionReceiptLog(chainId uint, log response.RPCReceiptLogs) (decodeFromAddress, decodeToAddress, contractName string, decimals int, amount *big.Int, err error) {
	if len(log.Topics) != 3 || log.Topics[0] != "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" || log.Topics[1] == "" || log.Topics[2] == "" || log.Data == "" {
		err = errors.New("no support")
		return
	}

	// Transfer
	isSupportContract, contractName, _, decimals := sweepUtils.GetContractInfo(chainId, log.Address)
	if !isSupportContract {
		err = errors.New("no support the contract")
		return
	}

	// from
	fromTopic := log.Topics[1]
	if len(fromTopic) > 2 && fromTopic[:2] == "0x" {
		fromTopic = fromTopic[2:]
	}

	fromBytes, err := hex.DecodeString(fromTopic)
	if err != nil {
		return
	}

	if len(fromBytes) < 20 {
		err = errors.New("no support the fromBytes")
		return
	}

	fromAddressBytes := fromBytes[len(fromBytes)-20:]
	decodeFromAddress = "0x" + hex.EncodeToString(fromAddressBytes)

	// if utils.HexToAddress(monitorAddress) == utils.HexToAddress(fromAddress) {
	// 	return true
	// }

	// to
	toTopic := log.Topics[2]
	if len(toTopic) > 2 && toTopic[:2] == "0x" {
		toTopic = toTopic[2:]
	}

	toBytes, err := hex.DecodeString(toTopic)
	if err != nil {
		return
	}

	if len(toBytes) < 20 {
		err = errors.New("no support the toBytes")
		return
	}

	toAddressBytes := toBytes[len(toBytes)-20:]
	decodeToAddress = "0x" + hex.EncodeToString(toAddressBytes)

	// if utils.HexToAddress(monitorAddress) == utils.HexToAddress(toAddress) {
	// 	return true
	// }

	data := log.Data
	if len(data) > 2 && data[:2] == "0x" {
		data = data[2:]
	}

	dataBytes, err := hex.DecodeString(data)
	if err != nil {
		return
	}

	amount = new(big.Int).SetBytes(dataBytes)

	return
}
