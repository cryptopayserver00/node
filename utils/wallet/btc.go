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
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func SendBtcTransfer(chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {

	sendValFloat, err := strconv.ParseFloat(sendVal, 64)
	if err != nil {
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

	pkScript := "001438dc6790cb66e6e8934af0eb6d41ea0f499d4c21"

	txid, balance, err := GetUTXO(addrPubKey.EncodeAddress())
	if err != nil {
		return "", err
	}

	if balance < satoshiValue {
		return "", fmt.Errorf("the balance of the account is not sufficient")
	}

	destinationAddr, err := btcutil.DecodeAddress(toAddress, &chaincfg.TestNet3Params)
	if err != nil {
		return "", err
	}

	destinationAddrByte, err := txscript.PayToAddrScript(destinationAddr)
	if err != nil {
		return "", err
	}

	redeemTx, err := NewTx()
	if err != nil {
		return "", err
	}

	utxoHash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		return "", err
	}

	outPoint := wire.NewOutPoint(utxoHash, 0)

	var byteArray [][]byte
	txIn := wire.NewTxIn(outPoint, nil, nil)
	txIn.Witness = byteArray
	redeemTx.AddTxIn(txIn)

	redeemTxOut := wire.NewTxOut(satoshiValue, destinationAddrByte)
	redeemTx.AddTxOut(redeemTxOut)

	finalRawTx, err := SignTx(pri, pkScript, redeemTx)
	if err != nil {
		return "", err
	}

	return finalRawTx, nil
}

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
		return "", err
	}

	sourcePKScript, err := hex.DecodeString(pkScript)
	if err != nil {
		return "", err
	}

	signature, err := txscript.SignatureScript(redeemTx, 0, sourcePKScript, txscript.SigHashAll, wif.PrivKey, false)
	if err != nil {
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
