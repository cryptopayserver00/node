package wallet

func SendTrxTransfer(chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {
	return "", nil
}

func SendTronTokenTransfer(chainId uint, pri, pub, toAddress, coin string, sendVal string) (hash string, err error) {
	return "", nil
}

// func SendTrxTransfer(chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {

// 	url := constant.GetHttpUrlByNetwork(chainId)
// 	if url == "" {
// 		err = errors.New("chain not support")
// 		return
// 	}

// 	c := client.NewGrpcClient(url)
// 	err = c.Start(grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		return
// 	}

// 	sendValue, err := utils.FormatToOriginalValue(sendVal, 6)
// 	if err != nil {
// 		return "", err
// 	}

// 	tx, err := c.Transfer(pub, toAddress, sendValue.Int64())
// 	if err != nil {
// 		return
// 	}

// 	rawData, err := proto.Marshal(tx.Transaction.GetRawData())
// 	if err != nil {
// 		return
// 	}

// 	h256h := sha256.New()
// 	h256h.Write(rawData)
// 	txHash := h256h.Sum(nil)

// 	privateKeyBytes, err := hex.DecodeString(pri)
// 	if err != nil {
// 		return
// 	}

// 	sk, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

// 	signature, err := crypto.Sign(txHash, sk.ToECDSA())
// 	if err != nil {
// 		return
// 	}

// 	tx.Transaction.Signature = append(tx.Transaction.Signature, signature)

// 	result, err := c.Broadcast(tx.Transaction)
// 	if err != nil {
// 		return
// 	}

// 	global.NODE_LOG.Info("SendTrxTransfer: " + string(result.Message))

// 	return "", nil
// }

// func SendTronTokenTransfer(chainId uint, pri, pub, toAddress, coin string, sendVal string) (hash string, err error) {

// 	url := constant.GetHttpUrlByNetwork(chainId)
// 	if url == "" {
// 		err = errors.New("chain not support")
// 		return
// 	}

// 	c := client.NewGrpcClient(url)
// 	err = c.Start(grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		return
// 	}

// 	isSupport, _, _, decimals := sweepUtils.GetContractInfoByChainIdAndSymbol(chainId, coin)
// 	if !isSupport {
// 		return "", errors.New("contract address not found")
// 	}

// 	sendValue, err := utils.FormatToOriginalValue(sendVal, decimals)
// 	if err != nil {
// 		return "", err
// 	}

// 	tx, err := c.TransferAsset(pub, toAddress, coin, sendValue.Int64())
// 	if err != nil {
// 		return
// 	}

// 	rawData, err := proto.Marshal(tx.Transaction.GetRawData())
// 	if err != nil {
// 		return
// 	}

// 	h256h := sha256.New()
// 	h256h.Write(rawData)
// 	txHash := h256h.Sum(nil)

// 	privateKeyBytes, err := hex.DecodeString(pri)
// 	if err != nil {
// 		return
// 	}

// 	sk, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

// 	signature, err := crypto.Sign(txHash, sk.ToECDSA())
// 	if err != nil {
// 		return
// 	}

// 	tx.Transaction.Signature = append(tx.Transaction.Signature, signature)

// 	result, err := c.Broadcast(tx.Transaction)
// 	if err != nil {
// 		return
// 	}

// 	global.NODE_LOG.Info("SendTronTokenTransfer: " + string(result.Message))

// 	return "", nil
// }
