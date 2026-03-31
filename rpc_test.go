package core

import (
	"encoding/json"
	"testing"
)

func newTestCore(t *testing.T) *Core {
	t.Helper()
	c, err := NewCore(t.TempDir())
	if err != nil {
		t.Fatalf("NewCore failed: %v", err)
	}
	t.Cleanup(c.Shutdown)
	return c
}

func TestHandleRPC_GetVersion(t *testing.T) {
	c := newTestCore(t)
	result := c.HandleRPC(`{"method":"getVersion"}`)
	var resp RPCResponse
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}
}

func TestHandleRPC_GetConfig(t *testing.T) {
	c := newTestCore(t)
	result := c.HandleRPC(`{"method":"getConfig"}`)
	var resp RPCResponse
	json.Unmarshal([]byte(result), &resp)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}
	if resp.Result == nil {
		t.Fatal("config should not be nil")
	}
}

func TestHandleRPC_ConvertFormats(t *testing.T) {
	c := newTestCore(t)
	result := c.HandleRPC(`{"method":"convert.formats"}`)
	var resp RPCResponse
	json.Unmarshal([]byte(result), &resp)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}
}

func TestHandleRPC_UnknownMethod(t *testing.T) {
	c := newTestCore(t)
	result := c.HandleRPC(`{"method":"nonexistent"}`)
	var resp RPCResponse
	json.Unmarshal([]byte(result), &resp)
	if resp.Error == nil {
		t.Fatal("should return error for unknown method")
	}
	if resp.Error.Code != "method_error" {
		t.Errorf("code = %s, want method_error", resp.Error.Code)
	}
}

func TestHandleRPC_InvalidJSON(t *testing.T) {
	c := newTestCore(t)
	result := c.HandleRPC(`not json`)
	var resp RPCResponse
	json.Unmarshal([]byte(result), &resp)
	if resp.Error == nil {
		t.Fatal("should return error for invalid JSON")
	}
}

func TestHandleRPC_QueueList(t *testing.T) {
	c := newTestCore(t)
	result := c.HandleRPC(`{"method":"queue.list"}`)
	var resp RPCResponse
	json.Unmarshal([]byte(result), &resp)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}
}

func TestHandleRPC_FetchContent(t *testing.T) {
	c := newTestCore(t)
	result := c.HandleRPC(`{"method":"fetchContent","params":{"url":"https://example.com/not-music"}}`)
	var resp RPCResponse
	json.Unmarshal([]byte(result), &resp)
	// Services may or may not resolve a bogus URL depending on network;
	// the RPC must return either a valid result or a structured error.
	if resp.Error == nil && resp.Result == nil {
		t.Fatal("should return either result or error")
	}
}

func TestHandleRPC_FetchContent_MissingURL(t *testing.T) {
	c := newTestCore(t)
	result := c.HandleRPC(`{"method":"fetchContent","params":{}}`)
	var resp RPCResponse
	json.Unmarshal([]byte(result), &resp)
	if resp.Error == nil {
		t.Fatal("expected error for missing URL")
	}
}

func TestHandleRPC_FetchContent_InvalidParams(t *testing.T) {
	c := newTestCore(t)
	result := c.HandleRPC(`{"method":"fetchContent","params":"bad"}`)
	var resp RPCResponse
	json.Unmarshal([]byte(result), &resp)
	if resp.Error == nil {
		t.Fatal("expected error for invalid params")
	}
}

func TestHandleRPC_HistoryList(t *testing.T) {
	c := newTestCore(t)
	result := c.HandleRPC(`{"method":"history.list"}`)
	var resp RPCResponse
	json.Unmarshal([]byte(result), &resp)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}
}
