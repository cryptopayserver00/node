package wallet

// "github.com/ltcsuite/ltcd/chaincfg"
// "github.com/ltcsuite/ltcd/chaincfg/chainhash"
// "github.com/ltcsuite/ltcd/ltcutil"
// "github.com/ltcsuite/ltcd/txscript"
// "github.com/ltcsuite/ltcd/wire"

func SendLtcTransfer(chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {

	return "", nil

	// sendValFloat, err := strconv.ParseFloat(sendVal, 64)
	// if err != nil {
	// 	global.NODE_LOG.Error(err.Error())
	// 	return
	// }

	// satoshiValue := utils.FormatToSatoshiValue(sendValFloat)

	// privateKey, err := hex.DecodeString(pri)
	// if err != nil {
	// 	global.NODE_LOG.Error(err.Error())
	// 	return
	// }

	// wif, err := ltcutil.DecodeWIF(pri)
	// if err != nil {
	// 	global.NODE_LOG.Error(err.Error())
	// 	return
	// }

	// addrPubKey, err := ltcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeCompressed(), &chaincfg.TestNet4Params)
	// if err != nil {
	// 	global.NODE_LOG.Error(err.Error())
	// 	return
	// }

	// global.NODE_LOG.Info(addrPubKey.EncodeAddress())

	// pkScript := ""

	// txid, balance, err := GetLtcUTXO(addrPubKey.EncodeAddress())
	// if err != nil {
	// 	global.NODE_LOG.Error(err.Error())
	// 	return
	// }

	// if balance < satoshiValue {
	// 	return "", fmt.Errorf("the balance of the account is not sufficient")
	// }

	// destinationAddr, err := ltcutil.DecodeAddress(toAddress, &chaincfg.TestNet4Params)
	// if err != nil {
	// 	global.NODE_LOG.Error(err.Error())
	// 	return
	// }

	// destinationAddrByte, err := txscript.PayToAddrScript(destinationAddr)
	// if err != nil {
	// 	global.NODE_LOG.Error(err.Error())
	// 	return
	// }

	// redeemTx, err := LtcNewTx()
	// if err != nil {
	// 	global.NODE_LOG.Error(err.Error())
	// 	return
	// }

	// utxoHash, err := chainhash.NewHashFromStr(txid)
	// if err != nil {
	// 	global.NODE_LOG.Error(err.Error())
	// 	return
	// }

	// outPoint := wire.NewOutPoint(utxoHash, 0)

	// txIn := wire.NewTxIn(outPoint, nil, nil)
	// txIn.Witness = nil
	// redeemTx.AddTxIn(txIn)

	// redeemTxOut := wire.NewTxOut(satoshiValue, destinationAddrByte)
	// redeemTx.AddTxOut(redeemTxOut)

	// finalRawTx, err := LtcSignTx(pri, pkScript, redeemTx)
	// if err != nil {
	// 	global.NODE_LOG.Error(err.Error())
	// 	return
	// }

	// return finalRawTx, nil
}

// func GetLtcUTXO(address string) (string, int64, error) {
// 	var previousTxid string = "fd02222cee28803d1f6a2f64599e5d16f61067db25ce11d0b41ba91dba65c72d"
// 	var balance int64 = 973
// 	return previousTxid, balance, nil
// }

// func LtcNewTx() (*wire.MsgTx, error) {
// 	return wire.NewMsgTx(wire.TxVersion), nil
// }

// func LtcSignTx(privKey string, pkScript string, redeemTx *wire.MsgTx) (string, error) {

// 	wif, err := ltcutil.DecodeWIF(privKey)
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return "", err
// 	}

// 	sourcePKScript, err := hex.DecodeString(pkScript)
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return "", err
// 	}

// 	signature, err := txscript.SignatureScript(redeemTx, 0, sourcePKScript, txscript.SigHashAll, wif.PrivKey, false)
// 	if err != nil {
// 		global.NODE_LOG.Error(err.Error())
// 		return "", err
// 	}

// 	redeemTx.TxIn[0].SignatureScript = signature
// 	redeemTx.TxIn[0].Witness = nil

// 	var signedTx bytes.Buffer
// 	redeemTx.Serialize(&signedTx)

// 	hexSignedTx := hex.EncodeToString(signedTx.Bytes())

// 	return hexSignedTx, nil
// }
