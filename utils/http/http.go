package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	URL     string
	Headers map[string]string
	client  http.Client
	Timeout time.Duration
}

// doRequest 是三个导出方法共用的核心逻辑：构造 header、发送请求、
// 校验状态码并返回响应体字节。所有的错误信息都会尽量带上响应体内容，
// 便于排查后端返回的错误详情。
func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	response, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect the %s: %w", c.URL, err)
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", c.URL, err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s, body: %s", response.StatusCode, c.URL, string(data))
	}

	return data, nil
}

// HTTPGet 发送 GET 请求，并将响应体反序列化为 JSON 写入 dest（dest 必须是指针）。
func (c *Client) HTTPGet(ctx context.Context, dest any) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeoutOrDefault())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create get request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	data, err := c.doRequest(req)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal response body from %s: %w, body: %s", c.URL, err, string(data))
	}

	return nil
}

// HTTPGetUnique 发送 GET 请求，并将响应体（非 JSON，纯文本/数字）转换写入 dest。
// dest 支持 *string、*int、*int64、*float64、*bool。
func (c *Client) HTTPGetUnique(ctx context.Context, dest any) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeoutOrDefault())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create get request: %w", err)
	}

	data, err := c.doRequest(req)
	if err != nil {
		return err
	}

	text := strings.TrimSpace(string(data))
	switch v := dest.(type) {
	case *string:
		*v = text
	case *int:
		n, err := strconv.Atoi(text)
		if err != nil {
			return fmt.Errorf("failed to convert response body to int: %w", err)
		}
		*v = n
	case *int64:
		n, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert response body to int64: %w", err)
		}
		*v = n
	case *float64:
		n, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return fmt.Errorf("failed to convert response body to float64: %w", err)
		}
		*v = n
	case *bool:
		b, err := strconv.ParseBool(text)
		if err != nil {
			return fmt.Errorf("failed to convert response body to bool: %w", err)
		}
		*v = b
	default:
		return fmt.Errorf("unsupported type: %T", dest)
	}
	return nil
}

// HTTPPost 发送 POST 请求，请求体为 source 序列化后的 JSON，
// 响应体反序列化为 JSON 写入 dest（dest 必须是指针）。
func (c *Client) HTTPPost(ctx context.Context, source any, dest any) error {
	body, err := json.Marshal(source)
	if err != nil {
		return fmt.Errorf("failed to encode source: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeoutOrDefault())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to create post request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	data, err := c.doRequest(req)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to decode response body from %s: %w, body: %s", c.URL, err, string(data))
	}
	return nil
}

func (c *Client) timeoutOrDefault() time.Duration {
	if c.Timeout <= 0 {
		return 30 * time.Second
	}
	return c.Timeout
}

// HTTPPostText 以 text/plain 原样发送 body，并原样返回响应体（不做 JSON 编解码）。
// 用于 Esplora 的 POST /tx：请求体是 hex 字符串，响应体是纯文本 txid。
func (c *Client) HTTPPostText(ctx context.Context, body string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeoutOrDefault())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create post request: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	return c.doRequest(req)
}
