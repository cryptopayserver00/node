package wallet

import "context"

func SendPolTransfer(ctx context.Context, chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {
	return SendEthTransfer(ctx, chainId, pri, pub, toAddress, sendVal)
}

func SendPolTokenTransfer(ctx context.Context, chainId uint, pri, pub, toAddress string, coin, sendVal string) (hash string, err error) {
	return SendEthTokenTransfer(ctx, chainId, pri, pub, toAddress, coin, sendVal)
}
