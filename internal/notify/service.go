package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type OrderHandoff struct {
	OrderID  string `json:"order_id"`
	BuyerID  string `json:"buyer_id"`
	SellerID string `json:"seller_id"`
	Status   string `json:"status"`
}

type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func (c *Client) request(ctx context.Context, method, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("decode envelope: %w", err)
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected: %s", string(env.Error))
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := time.Duration(1<<attempt) * 100 * time.Millisecond
			if v, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil && v > 0 {
				delay = time.Duration(v) * time.Second
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				continue
			}
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("infrai transport status %d", res.StatusCode)
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return errors.New("retry budget exhausted")
}

func (c *Client) EnsureChannel(ctx context.Context, channel string) error {
	return c.request(ctx, http.MethodPost, "/v1/realtime/channel/create", map[string]any{"channel": channel, "type": "private"}, nil)
}

func (c *Client) Publish(ctx context.Context, n OrderHandoff) error {
	data, _ := json.Marshal(n)
	return c.request(ctx, http.MethodPost, "/v1/realtime/publish", map[string]any{"channel": "buyer-" + n.BuyerID, "event": "order.handoff", "data": json.RawMessage(data), "account_id": n.SellerID}, nil)
}

func (c *Client) IssueToken(ctx context.Context, clientID string) (map[string]any, error) {
	var token map[string]any
	err := c.request(ctx, http.MethodPost, "/v1/realtime/token/issue", map[string]any{"client_id": clientID, "channels": []string{"buyer-" + clientID}, "capabilities": []string{"subscribe"}, "ttl_seconds": 3600}, &token)
	return token, err
}
