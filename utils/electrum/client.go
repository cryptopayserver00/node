package electrum

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// Endpoint is a Fulcrum/ElectrumX server (newline-delimited JSON-RPC).
type Endpoint struct {
	Host string
	Port int
	TLS  bool
}

type rpcRequest struct {
	ID     int    `json:"id"`
	Method string `json:"method"`
	Params []any  `json:"params"`
}

type rpcResponse struct {
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Call dials the endpoint once, sends a single JSON-RPC request, and closes.
func Call(ctx context.Context, ep Endpoint, method string, params []any, dest any) error {
	if ep.Host == "" || ep.Port == 0 {
		return fmt.Errorf("invalid electrum endpoint")
	}

	dialer := net.Dialer{Timeout: 10 * time.Second}
	addr := fmt.Sprintf("%s:%d", ep.Host, ep.Port)

	var conn net.Conn
	var err error
	if ep.TLS {
		tlsCfg := &tls.Config{ServerName: ep.Host, MinVersion: tls.VersionTLS12}
		conn, err = tls.DialWithDialer(&dialer, "tcp", addr, tlsCfg)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("electrum dial %s: %w", addr, err)
	}
	defer conn.Close()

	deadline := time.Now().Add(20 * time.Second)
	if dl, ok := ctx.Deadline(); ok && dl.Before(deadline) {
		deadline = dl
	}
	_ = conn.SetDeadline(deadline)

	payload, err := json.Marshal(rpcRequest{ID: 1, Method: method, Params: params})
	if err != nil {
		return err
	}
	if _, err := conn.Write(append(payload, '\n')); err != nil {
		return fmt.Errorf("electrum write %s: %w", method, err)
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("electrum read %s: %w", method, err)
	}

	var resp rpcResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		return fmt.Errorf("electrum decode %s: %w", method, err)
	}
	if resp.Error != nil {
		return fmt.Errorf("electrum %s: %s", method, resp.Error.Message)
	}
	if dest == nil {
		return nil
	}
	if len(resp.Result) == 0 || string(resp.Result) == "null" {
		return fmt.Errorf("electrum %s: empty result", method)
	}
	if err := json.Unmarshal(resp.Result, dest); err != nil {
		return fmt.Errorf("electrum unmarshal %s: %w", method, err)
	}
	return nil
}

// CallFirst tries endpoints in order until one succeeds.
func CallFirst(ctx context.Context, endpoints []Endpoint, method string, params []any, dest any) (Endpoint, error) {
	var last error
	for _, ep := range endpoints {
		if err := ctx.Err(); err != nil {
			return Endpoint{}, err
		}
		if err := Call(ctx, ep, method, params, dest); err != nil {
			last = err
			continue
		}
		return ep, nil
	}
	if last != nil {
		return Endpoint{}, last
	}
	return Endpoint{}, fmt.Errorf("no electrum endpoints")
}
