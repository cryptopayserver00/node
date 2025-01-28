package wallet

import (
	"node/global"
	"node/utils"
	"strconv"

	"github.com/ltcsuite/ltcd/chaincfg"
	"github.com/ltcsuite/ltcd/ltcutil"
)

func SendLtcTransfer(chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {

	sendValFloat, err := strconv.ParseFloat(sendVal, 64)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return "", err
	}

	// satoshiValue := utils.FormatToSatoshiValue(sendValFloat)
	_ = utils.FormatToSatoshiValue(sendValFloat)

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

	// pkScript := ""

	return "", nil
}
