package rtm

import (
	"encoding/json"
	"fmt"
)

// GetTimeline creates a new RTM timeline required for write operations.
func (c *Client) GetTimeline(token string) (string, error) {
	data, err := c.Call("rtm.timelines.create", map[string]string{
		"auth_token": token,
	})
	if err != nil {
		return "", err
	}
	var resp struct {
		Rsp struct {
			Stat     string `json:"stat"`
			Timeline string `json:"timeline"`
		} `json:"rsp"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("GetTimeline: %w", err)
	}
	if resp.Rsp.Stat != "ok" {
		return "", fmt.Errorf("GetTimeline: API stat=%q", resp.Rsp.Stat)
	}
	return resp.Rsp.Timeline, nil
}

// CompleteTask marks a task as complete on RTM servers.
func (c *Client) CompleteTask(token, timeline, listID, taskseriesID, taskID string) error {
	data, err := c.Call("rtm.tasks.complete", map[string]string{
		"auth_token":    token,
		"timeline":      timeline,
		"list_id":       listID,
		"taskseries_id": taskseriesID,
		"task_id":       taskID,
	})
	if err != nil {
		return err
	}
	var resp struct {
		Rsp struct {
			Stat string `json:"stat"`
		} `json:"rsp"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("CompleteTask: %w", err)
	}
	if resp.Rsp.Stat != "ok" {
		return fmt.Errorf("CompleteTask: API stat=%q", resp.Rsp.Stat)
	}
	return nil
}

// AddTask adds a new task using RTM's Smart Add syntax (smart=1).
func (c *Client) AddTask(token, timeline, smartAddString string) error {
	data, err := c.Call("rtm.tasks.add", map[string]string{
		"auth_token": token,
		"timeline":   timeline,
		"name":       smartAddString,
		"parse":      "1",
	})
	if err != nil {
		return err
	}
	var resp struct {
		Rsp struct {
			Stat string `json:"stat"`
		} `json:"rsp"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("AddTask: %w", err)
	}
	if resp.Rsp.Stat != "ok" {
		return fmt.Errorf("AddTask: API stat=%q", resp.Rsp.Stat)
	}
	return nil
}
