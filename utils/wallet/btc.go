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
		global.NODE_LOG.Error(err.Error())
		return
	}

	satoshiValue := utils.FormatToSatoshiValue(sendValFloat)

	// privateKey, err := hex.DecodeString(pri)
	// if err != nil {
	// 	global.NODE_LOG.Error(err.Error())
	// 	return
	// }

	wif, err := btcutil.DecodeWIF(pri)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return
	}

	// use TestNet3Params for interacting with bitcoin testnet
	// if we want to interact with main net should use MainNetParams
	addrPubKey, err := btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeUncompressed(), &chaincfg.TestNet3Params)
	if err != nil {
		return "", err
	}

	txid, balance, err := GetUTXO(addrPubKey.EncodeAddress())
	if err != nil {
		return "", err
	}

	/*
	 * 1 or unit-amount in Bitcoin is equal to 1 satoshi and 1 Bitcoin = 100000000 satoshi
	 */

	// checking for sufficiency of account
	if balance < satoshiValue {
		return "", fmt.Errorf("the balance of the account is not sufficient")
	}

	// extracting destination address as []byte from function argument (destination string)
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
	redeemTx, err := BtcNewTx()
	if err != nil {
		return "", err
	}

	utxoHash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		return "", err
	}

	// the second argument is vout or Tx-index, which is the index
	// of spending UTXO in the transaction that Txid referred to
	// in this case is 0, but can vary different numbers
	outPoint := wire.NewOutPoint(utxoHash, 0)

	// making the input, and adding it to transaction
	txIn := wire.NewTxIn(outPoint, nil, nil)
	redeemTx.AddTxIn(txIn)

	// adding the destination address and the amount to
	// the transaction as output
	redeemTxOut := wire.NewTxOut(satoshiValue, destinationAddrByte)
	redeemTx.AddTxOut(redeemTxOut)

	pkScript := "001438dc6790cb66e6e8934af0eb6d41ea0f499d4c21"

	// now sign the transaction
	finalRawTx, err := BtcSignTx(pri, pkScript, redeemTx)
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

func BtcNewTx() (*wire.MsgTx, error) {
	return wire.NewMsgTx(2), nil
}

func BtcSignTx(privKey string, pkScript string, redeemTx *wire.MsgTx) (string, error) {

	wif, err := btcutil.DecodeWIF(privKey)
	if err != nil {
		return "", err
	}

	sourcePKScript, err := hex.DecodeString(pkScript)
	if err != nil {
		return "", nil
	}

	// // txSigHashes := txscript.NewTxSigHashes(redeemTx, txscript.SigHashAll)
	// // since there is only one input in our transaction
	// // we use 0 as second argument, if the transaction
	// // has more args, should pass related index
	// // signature, err := txscript.SignatureScript(redeemTx, 0, sourcePKScript, txscript.SigHashAll, wif.PrivKey, false)
	// signature, err := txscript.WitnessSignature(redeemTx, &txscript.TxSigHashes{}, 0, 200, sourcePKScript, txscript.SigHashAll, wif.PrivKey, false)
	// if err != nil {
	// 	return "", nil
	// }

	// // since there is only one input, and want to add
	// // signature to it use 0 as index
	// // redeemTx.TxIn[0].SignatureScript = signature
	// redeemTx.TxIn[0].Witness = signature

	// Generating the signature
	signature, err := txscript.SignatureScript(redeemTx, 0, sourcePKScript, txscript.SigHashAll, wif.PrivKey, false)
	if err != nil {
		return "", err
	}

	// Adding the witness data (for SegWit)
	pubKey := wif.PrivKey.PubKey().SerializeCompressed() // Compressed pubKey
	witnessData := [][]byte{
		[]byte{0x00}, // Witness version, usually 0 for P2WPKH
		pubKey,       // PubKey (compressed)
		signature,    // Signature
	}

	// Assign the witness data to the input
	redeemTx.TxIn[0].Witness = witnessData

	var signedTx bytes.Buffer
	redeemTx.Serialize(&signedTx)

	hexSignedTx := hex.EncodeToString(signedTx.Bytes())

	return hexSignedTx, nil
}
