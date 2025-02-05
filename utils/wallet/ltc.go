package wallet

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"node/global"
	"node/utils"
	"strconv"

	"github.com/ltcsuite/ltcd/chaincfg"
	"github.com/ltcsuite/ltcd/chaincfg/chainhash"
	"github.com/ltcsuite/ltcd/ltcutil"
	"github.com/ltcsuite/ltcd/txscript"
	"github.com/ltcsuite/ltcd/wire"
)

func SendLtcTransfer(chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {

	sendValFloat, err := strconv.ParseFloat(sendVal, 64)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return "", err
	}

	satoshiValue := utils.FormatToSatoshiValue(sendValFloat)

	wif, err := ltcutil.DecodeWIF(pri)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	addrPubKey, err := ltcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeCompressed(), &chaincfg.TestNet4Params)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	global.NODE_LOG.Info(addrPubKey.EncodeAddress())

	pkScript := ""

	txid, balance, err := GetLtcUTXO(addrPubKey.EncodeAddress())
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	if balance < satoshiValue {
		return "", fmt.Errorf("the balance of the account is not sufficient")
	}

	destinationAddr, err := ltcutil.DecodeAddress(toAddress, &chaincfg.TestNet4Params)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	destinationAddrByte, err := txscript.PayToAddrScript(destinationAddr)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	redeemTx, err := LtcNewTx()
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	utxoHash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	outPoint := wire.NewOutPoint(utxoHash, 0)

	txIn := wire.NewTxIn(outPoint, nil, nil)
	txIn.Witness = nil
	redeemTx.AddTxIn(txIn)

	redeemTxOut := wire.NewTxOut(satoshiValue, destinationAddrByte)
	redeemTx.AddTxOut(redeemTxOut)

	finalRawTx, err := LtcSignTx(pri, pkScript, redeemTx)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	return finalRawTx, nil

}

func GetLtcUTXO(address string) (string, int64, error) {
	var previousTxid string = "166327ccddcf1428ab591681f9018ab1f2d78039efef0e14967850b68142f1cf"
	var balance int64 = 5000000
	return previousTxid, balance, nil
}

func LtcNewTx() (*wire.MsgTx, error) {
	return wire.NewMsgTx(wire.TxVersion), nil
}

func LtcSignTx(privKey string, pkScript string, redeemTx *wire.MsgTx) (string, error) {

	wif, err := ltcutil.DecodeWIF(privKey)
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

	redeemTx.TxIn[0].SignatureScript = signature
	redeemTx.TxIn[0].Witness = nil

	var signedTx bytes.Buffer
	redeemTx.Serialize(&signedTx)

	hexSignedTx := hex.EncodeToString(signedTx.Bytes())

	return hexSignedTx, nil
}
