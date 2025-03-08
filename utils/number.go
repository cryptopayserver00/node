package utils

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

func CalculateBalance(transactionValue *big.Int, decimals int) string {
	base := big.NewInt(10)
	exponent := big.NewInt(int64(decimals))
	exponentValue := new(big.Int).Exp(base, exponent, nil)

	transactionValueFloat := new(big.Float).SetInt(transactionValue)
	exponentValueFloat := new(big.Float).SetInt(exponentValue)

	resultFloat := new(big.Float).Quo(transactionValueFloat, exponentValueFloat)

	resultString := fmt.Sprintf("%.*f", decimals, resultFloat)

	resultString = strings.TrimRight(resultString, "0")

	resultString = strings.TrimRight(resultString, ".")

	return resultString
}

func HexStringToUint64(hexString string) (uint64, error) {
	if hexString == "" {
		return 0, errors.New("hexString can not be empty")
	}
	intValue, err := strconv.ParseUint(hexString, 0, 64)
	if err != nil {
		return 0, err
	}
	return intValue, nil
}

func HexStringToBigInt(hexString string) (*big.Int, error) {
	if hexString == "" {
		return &big.Int{}, errors.New("hexString can not be empty")
	}

	value, good := new(big.Int).SetString(hexString, 0)
	if !good {
		return &big.Int{}, errors.New("no support")
	}

	return value, nil
}

func CalSubForBtcValue(incoming, outgoing string) (float64, error) {

	incomingFloatValue, err := strconv.ParseFloat(incoming, 64)
	if err != nil {
		return 0, err
	}
	outgoingFloatValue, err := strconv.ParseFloat(outgoing, 64)
	if err != nil {
		return 0, err
	}

	formattedBalance := fmt.Sprintf("%.8f", incomingFloatValue-outgoingFloatValue)

	balance, err := strconv.ParseFloat(formattedBalance, 64)
	if err != nil {
		return 0, err
	}

	return balance, nil
}

func FormatToBtcValue(value int64) (float64, error) {

	formattedBalance := CalculateBalance(big.NewInt(value), 8)

	balance, err := strconv.ParseFloat(formattedBalance, 64)
	if err != nil {
		return 0, err
	}

	return balance, nil
}

func FormatToSatoshiValue(value float64) int64 {

	formattedBalance := value * 100000000

	return int64(formattedBalance)
}

func FormatToEtherValue(value float64) int64 {

	formattedBalance := value * 1000000000000000000

	return int64(formattedBalance)
}

func FormatToOriginalValue(value string, decimals int) (*big.Int, error) {
	parts := strings.Split(value, ".")
	if len(parts) > 2 {
		return nil, errors.New("invalid value format")
	}

	intPart := strings.TrimLeft(parts[0], "0")
	if intPart == "" {
		intPart = "0"
	}

	var fracPart string
	if len(parts) == 2 {
		fracPart = parts[1]
	} else {
		fracPart = ""
	}

	if len(fracPart) < decimals {
		fracPart += strings.Repeat("0", decimals-len(fracPart))
	} else if len(fracPart) > decimals {
		fracPart = fracPart[:decimals]
	}

	combined := intPart + fracPart

	result := new(big.Int)
	_, ok := result.SetString(combined, 10)
	if !ok {
		return nil, errors.New("error converting to big.Int")
	}

	return result, nil
}
