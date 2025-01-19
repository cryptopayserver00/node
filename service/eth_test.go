package service

import (
	"node/model/node/response"
	"testing"
)

func TestInnerTxForEth(t *testing.T) {

	var infos response.RPCInnerTxInfo
	client.URL = ""

	payload := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "debug_traceTransaction",
		"params": []interface{}{
			"0xc4138d43102c50dd41fa323e44d30ac983271c943c6b0a7e03d67eeb5475cc9a",
			map[string]interface{}{
				"tracer": "callTracer",
			},
		},
	}
	err := client.HTTPPost(payload, &infos)
	if err != nil {
		t.Log(err.Error())
	}
}
