package wallet

import (
	"context"
	"node/utils"

	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/system"
	"github.com/blocto/solana-go-sdk/rpc"
	"github.com/blocto/solana-go-sdk/types"
)

func SendSolTransfer(chainId uint, pri, pub, toAddress string, sendVal string) (hash string, err error) {
	client := rpc.NewRpcClient(rpc.DevnetRPCEndpoint)

	sendValue, err := utils.FormatToOriginalValue(sendVal, 9)
	if err != nil {
		return "", err
	}

	// feePayer := types.Account{
	// 	PublicKey: common.PublicKeyFromString(pub),
	// }

	feePayer, err := types.AccountFromSeed([]byte(pri))
	if err != nil {
		return
	}

	// feePayer, err := types.AccountFromHex(pri)
	// if err != nil {
	// 	return
	// }

	recentBlockhashResponse, err := client.GetLatestBlockhash(context.Background())
	if err != nil {
		return
	}

	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{feePayer},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        feePayer.PublicKey,
			RecentBlockhash: recentBlockhashResponse.Result.Value.Blockhash,
			Instructions: []types.Instruction{
				system.Transfer(system.TransferParam{
					From:   feePayer.PublicKey,
					To:     common.PublicKeyFromString(toAddress),
					Amount: sendValue.Uint64(),
				}),
			},
		}),
	})

	if err != nil {
		return
	}

	serializedTx, err := tx.Serialize()
	if err != nil {
		return
	}

	txHash, err := client.SendTransaction(context.Background(), string(serializedTx))
	if err != nil {
		return
	}

	return txHash.Result, nil
}

func SendSolTokenTransfer(chainId uint, pri, pub, toAddress, coin string, sendVal string) (hash string, err error) {
	return "", nil
}
