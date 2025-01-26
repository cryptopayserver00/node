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

	pkScript := "76a9148010460ef38bacdaf332c0354126d42b1e5781a588ac"

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

	// creating a new bitcoin transaction, different sections of the tx, including
	// input list (contain UTXOs) and outputlist (contain destination address and usually our address)
	// in next steps, sections will be field and pass to sign
	redeemTx, err := NewTx()
	if err != nil {
		return "", err
	}

	utxoHash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		return "", err
	}

	// the second argument is vout or Tx-index, which is the index
	// of spending UTXO in the transaction that Txid referred to
	// in this case is 1, but can vary different numbers
	outPoint := wire.NewOutPoint(utxoHash, 0)

	// making the input, and adding it to transaction
	txIn := wire.NewTxIn(outPoint, nil, nil)
	// txIn.Witness = nil
	redeemTx.AddTxIn(txIn)

	// adding the destination address and the amount to
	// the transaction as output
	redeemTxOut := wire.NewTxOut(satoshiValue, destinationAddrByte)
	redeemTx.AddTxOut(redeemTxOut)

	// now sign the transaction
	finalRawTx, err := SignTx(pri, pkScript, redeemTx)
	if err != nil {
		return "", err
	}

	return finalRawTx, nil
}

func GetUTXO(address string) (string, int64, error) {
	// Provide your url to get UTXOs, read the response
	// unmarshal it, and extract necessary data

	var previousTxid string = "062409e24ac7f7c89ec11508d9ec8a4a5e0f8f5b0a8e0700a414df292ddffd12"
	var balance int64 = 590504
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

	// since there is only one input in our transactio
	// we use 0 as second argument, if the transaction
	// has more args, should pass related index
	signature, err := txscript.SignatureScript(redeemTx, 0, sourcePKScript, txscript.SigHashAll, wif.PrivKey, false)
	if err != nil {
		return "", err
	}

	// since there is only one input, and want to add
	// signature to it use 0 as index
	redeemTx.TxIn[0].SignatureScript = signature
	// redeemTx.TxIn[0].Witness = nil

	var signedTx bytes.Buffer
	redeemTx.Serialize(&signedTx)

	hexSignedTx := hex.EncodeToString(signedTx.Bytes())

	return hexSignedTx, nil
}
