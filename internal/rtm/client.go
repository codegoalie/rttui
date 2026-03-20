package rtm

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

const baseURL = "https://api.rememberthemilk.com/services/rest/"

// Client holds API credentials and an HTTP client.
type Client struct {
	apiKey       string
	sharedSecret string
	http         *http.Client
}

// NewClient creates a new RTM API client.
func NewClient(apiKey, sharedSecret string) *Client {
	return &Client{
		apiKey:       apiKey,
		sharedSecret: sharedSecret,
		http:         &http.Client{},
	}
}

// sign returns the MD5 api_sig for the given params.
func sign(sharedSecret string, params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(sharedSecret)
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(params[k])
	}

	return fmt.Sprintf("%x", md5.Sum([]byte(sb.String())))
}

// Call performs a signed GET request to the RTM API and returns the response body.
func (c *Client) Call(method string, params map[string]string) ([]byte, error) {
	allParams := make(map[string]string, len(params)+3)
	for k, v := range params {
		allParams[k] = v
	}
	allParams["method"] = method
	allParams["api_key"] = c.apiKey
	allParams["format"] = "json"

	allParams["api_sig"] = sign(c.sharedSecret, allParams)

	q := url.Values{}
	for k, v := range allParams {
		q.Set(k, v)
	}

	resp, err := c.http.Get(baseURL + "?" + q.Encode())
	if err != nil {
		return nil, fmt.Errorf("rtm call %s: %w", method, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("rtm read body %s: %w", method, err)
	}
	return body, nil
}
