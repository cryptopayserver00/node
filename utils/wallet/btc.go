package wallet

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"node/global"
	"node/utils"
	"strconv"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/okx/go-wallet-sdk/coins/bitcoin"
)

func SendBtcTransfer(chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {
	return "", nil
}

func SendBtcTransferByPsbt(chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {

	sendValFloat, err := strconv.ParseFloat(sendVal, 64)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return "", err
	}

	satoshiValue := utils.FormatToSatoshiValue(sendValFloat)

	wif, err := btcutil.DecodeWIF(pri)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	addrPubKey, err := btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeCompressed(), &chaincfg.TestNet3Params)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	// global.NODE_LOG.Info(addrPubKey.EncodeAddress())

	// psbt
	var inputs []*bitcoin.TxInput
	inputs = append(inputs, &bitcoin.TxInput{
		TxId:              "166327ccddcf1428ab591681f9018ab1f2d78039efef0e14967850b68142f1cf",
		VOut:              uint32(0),
		Sequence:          1,
		Amount:            satoshiValue,
		Address:           pub,
		PrivateKey:        pri,
		MasterFingerprint: 0xF23F9FD2,
		DerivationPath:    "m/84'/0'/0'/0/0",
		PublicKey:         addrPubKey.EncodeAddress(),
	})

	var outputs []*bitcoin.TxOutput
	outputs = append(outputs, &bitcoin.TxOutput{
		Address: toAddress,
		Amount:  int64(100000),
	})
	psbtHex, err := bitcoin.GenerateUnsignedPSBTHex(inputs, outputs, &chaincfg.TestNet3Params)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return "", err
	}

	fmt.Println(psbtHex)

	return "", nil
}

// func SendBtcTransfer(chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {

// 	sendValFloat, err := strconv.ParseFloat(sendVal, 64)
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return "", err
// 	}

// 	satoshiValue := utils.FormatToSatoshiValue(sendValFloat)

// 	txBuild := bitcoin.NewTxBuild(1, &chaincfg.TestNet3Params)
// 	txBuild.AddInput("166327ccddcf1428ab591681f9018ab1f2d78039efef0e14967850b68142f1cf", 0, "", "", "", 5000000)
// 	txBuild.AddOutput(toAddress, satoshiValue)
// 	pubKeyMap := make(map[int]string)
// 	pubKeyMap[0] = "001438dc6790cb66e6e8934af0eb6d41ea0f499d4c21"
// 	txHex, hashes, err := txBuild.UnSignedTx(pubKeyMap)
// 	if err != nil {
// 		return "", err
// 	}
// 	signatureMap := make(map[int]string)
// 	for i, h := range hashes {
// 		privateBytes, err := hex.DecodeString(pri)
// 		if err != nil {
// 			return "", err
// 		}
// 		prvKey, _ := btcec.PrivKeyFromBytes(privateBytes)
// 		sign := ecdsa.Sign(prvKey, util.RemoveZeroHex(h))
// 		signatureMap[i] = hex.EncodeToString(sign.Serialize())
// 	}
// 	txHex, err = bitcoin.SignTx(txHex, pubKeyMap, signatureMap)
// 	if err != nil {
// 		return "", err
// 	}
// 	fmt.Println(txHex)

// 	return "", nil
// }

// func SendBtcTransfer(chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {

// 	sendValFloat, err := strconv.ParseFloat(sendVal, 64)
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return "", err
// 	}

// 	satoshiValue := utils.FormatToSatoshiValue(sendValFloat)

// 	wif, err := btcutil.DecodeWIF(pri)
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return
// 	}

// 	addrPubKey, err := btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeCompressed(), &chaincfg.TestNet3Params)
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return
// 	}

// 	global.NODE_LOG.Info(addrPubKey.EncodeAddress())

// 	pkScript := "001438dc6790cb66e6e8934af0eb6d41ea0f499d4c21"

// 	txid, balance, err := GetUTXO(addrPubKey.EncodeAddress())
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return "", err
// 	}

// 	if balance < satoshiValue {
// 		return "", fmt.Errorf("the balance of the account is not sufficient")
// 	}

// 	destinationAddr, err := btcutil.DecodeAddress(toAddress, &chaincfg.TestNet3Params)
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return "", err
// 	}

// 	destinationAddrByte, err := txscript.PayToAddrScript(destinationAddr)
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return "", err
// 	}

// 	redeemTx, err := NewTx()
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return "", err
// 	}

// 	utxoHash, err := chainhash.NewHashFromStr(txid)
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return "", err
// 	}

// 	outPoint := wire.NewOutPoint(utxoHash, 0)

// 	var byteArray [][]byte
// 	txIn := wire.NewTxIn(outPoint, nil, nil)
// 	txIn.Witness = byteArray
// 	redeemTx.AddTxIn(txIn)

// 	redeemTxOut := wire.NewTxOut(satoshiValue, destinationAddrByte)
// 	redeemTx.AddTxOut(redeemTxOut)

// 	finalRawTx, err := SignTx(pri, pkScript, redeemTx)
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return "", err
// 	}

// 	return finalRawTx, nil
// }

func GetUTXO(address string) (string, int64, error) {
	var previousTxid string = "166327ccddcf1428ab591681f9018ab1f2d78039efef0e14967850b68142f1cf"
	var balance int64 = 5000000
	return previousTxid, balance, nil
}

func NewTx() (*wire.MsgTx, error) {
	return wire.NewMsgTx(wire.TxVersion), nil
}

func SignTx(privKey string, pkScript string, redeemTx *wire.MsgTx) (string, error) {

	wif, err := btcutil.DecodeWIF(privKey)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return "", err
	}

	sourcePKScript, err := hex.DecodeString(pkScript)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return "", err
	}

	signature, err := txscript.SignatureScript(redeemTx, 0, sourcePKScript, txscript.SigHashAll, wif.PrivKey, false)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return "", err
	}

	var byteArray [][]byte

	redeemTx.TxIn[0].SignatureScript = signature
	redeemTx.TxIn[0].Witness = byteArray

	var signedTx bytes.Buffer
	redeemTx.Serialize(&signedTx)

	hexSignedTx := hex.EncodeToString(signedTx.Bytes())

	return hexSignedTx, nil
}
