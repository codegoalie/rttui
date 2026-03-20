package rtm

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

const authURL = "https://www.rememberthemilk.com/services/auth/"

func tokenPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "rttui", "token"), nil
}

// LoadToken reads the persisted auth token from disk.
func LoadToken() (string, error) {
	p, err := tokenPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// SaveToken writes the auth token to disk (dir: 0700, file: 0600).
func SaveToken(token string) error {
	p, err := tokenPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(token), 0600)
}

func verifyToken(c *Client, token string) error {
	data, err := c.Call("rtm.auth.checkToken", map[string]string{"auth_token": token})
	if err != nil {
		return err
	}
	var resp struct {
		Rsp struct {
			Stat string `json:"stat"`
		} `json:"rsp"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}
	if resp.Rsp.Stat != "ok" {
		return errors.New("token invalid")
	}
	return nil
}

func getFrob(c *Client) (string, error) {
	data, err := c.Call("rtm.auth.getFrob", nil)
	if err != nil {
		return "", err
	}
	var resp struct {
		Rsp struct {
			Stat string `json:"stat"`
			Frob string `json:"frob"`
		} `json:"rsp"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	if resp.Rsp.Stat != "ok" {
		return "", fmt.Errorf("getFrob failed: %s", resp.Rsp.Stat)
	}
	return resp.Rsp.Frob, nil
}

func getToken(c *Client, frob string) (string, error) {
	data, err := c.Call("rtm.auth.getToken", map[string]string{"frob": frob})
	if err != nil {
		return "", err
	}
	var resp struct {
		Rsp struct {
			Stat  string `json:"stat"`
			Auth  struct {
				Token string `json:"token"`
			} `json:"auth"`
		} `json:"rsp"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	if resp.Rsp.Stat != "ok" {
		return "", fmt.Errorf("getToken failed: %s", resp.Rsp.Stat)
	}
	return resp.Rsp.Auth.Token, nil
}

// EnsureAuthenticated loads or runs the frob flow to obtain a valid token.
func EnsureAuthenticated(c *Client) (string, error) {
	token, err := LoadToken()
	if err != nil {
		return "", err
	}
	if token != "" {
		if err := verifyToken(c, token); err == nil {
			return token, nil
		}
	}

	// Run desktop frob flow.
	frob, err := getFrob(c)
	if err != nil {
		return "", fmt.Errorf("getFrob: %w", err)
	}

	params := map[string]string{
		"api_key": c.apiKey,
		"perms":   "read",
		"frob":    frob,
	}
	sig := sign(c.sharedSecret, params)
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	q.Set("api_sig", sig)
	fmt.Printf("Open this URL in your browser to authorize rttui:\n\n  %s?%s\n\n", authURL, q.Encode())
	fmt.Print("Press Enter after authorizing...")
	bufio.NewReader(os.Stdin).ReadString('\n') //nolint:errcheck

	token, err = getToken(c, frob)
	if err != nil {
		return "", fmt.Errorf("getToken: %w", err)
	}
	if err := SaveToken(token); err != nil {
		return "", fmt.Errorf("saveToken: %w", err)
	}
	return token, nil
}
