package sdk

import (
	"encoding/json"
	"testing"
)

func decode(t *testing.T, s string) *ReceivedEvent {
	t.Helper()

	var e ReceivedEvent
	if err := json.Unmarshal([]byte(s), &e); err != nil {
		t.Fatalf("unmarshal %s: %v", s, err)
	}

	return &e
}

func TestReceivedEvent_SendToPluginAnyPayload(t *testing.T) {
	for name, payload := range map[string]string{
		"string": `"hello"`,
		"number": `42`,
		"array":  `[1,2,3]`,
		"null":   `null`,
		"object": `{"state":"not-a-number","key":"v"}`,
	} {
		t.Run(name, func(t *testing.T) {
			e := decode(t, `{"event":"sendToPlugin","action":"a","context":"c","payload":`+payload+`}`)
			if e.Event != SendToPlugin || e.Action != "a" || e.Context != "c" {
				t.Fatalf("unexpected event: %+v", e)
			}
			if e.Payload != nil {
				t.Fatalf("Payload must be nil for sendToPlugin, got %+v", e.Payload)
			}
			if string(e.RawPayload) != payload {
				t.Fatalf("RawPayload = %s, want %s", e.RawPayload, payload)
			}
		})
	}

	e := decode(t, `{"event":"sendToPlugin","context":"c","payload":{"key":"v"}}`)

	var got map[string]string
	if err := e.DecodePayload(&got); err != nil || got["key"] != "v" {
		t.Fatalf("DecodePayload = %v, %v", got, err)
	}
}

func TestReceivedEvent_KeyDown(t *testing.T) {
	e := decode(t, `{"event":"keyDown","action":"a","context":"c","device":"d","payload":{
		"controller":"Keypad","coordinates":{"column":1,"row":2},"isInMultiAction":true,
		"userDesiredState":1,"state":0,"settings":{"x":1},"resources":{"icon":"a.png"}}}`)

	p := e.Payload
	if p == nil {
		t.Fatal("Payload is nil")
	}
	if p.Controller != Keypad || p.Coordinates.Column != 1 || p.Coordinates.Row != 2 ||
		!p.IsInMultiAction || p.UserDesiredState != 1 || p.Resources["icon"] != "a.png" {
		t.Fatalf("unexpected payload: %+v", p)
	}
}

func TestReceivedEvent_NewEvents(t *testing.T) {
	e := decode(t, `{"event":"deviceDidChange","device":"d","deviceInfo":{"name":"n","type":13,"size":{"columns":9,"rows":4}}}`)
	if e.Event != DeviceDidChange || e.DeviceInfo.Type != KESDSDKDeviceTypeStreamDeckPlusXL || e.DeviceInfo.Size.Columns != 9 {
		t.Fatalf("unexpected event: %+v", e)
	}

	e = decode(t, `{"event":"didReceiveSettings","id":"req-1","action":"a","context":"c","payload":{"settings":{},"controller":"Neo"}}`)
	if e.ID != "req-1" || e.Payload.Controller != Neo {
		t.Fatalf("unexpected event: %+v", e)
	}

	e = decode(t, `{"event":"didReceiveDeepLink","payload":{"url":"/a/b?x=1"}}`)
	if e.Payload.URL != "/a/b?x=1" {
		t.Fatalf("unexpected URL: %q", e.Payload.URL)
	}

	e = decode(t, `{"event":"systemDidWakeUp"}`)
	if e.Payload != nil || len(e.RawPayload) != 0 {
		t.Fatalf("unexpected payload: %+v", e)
	}
}

func TestReceivedEvent_InvalidPayload(t *testing.T) {
	var e ReceivedEvent
	if err := json.Unmarshal([]byte(`{"event":"keyDown","payload":{"ticks":"x"}}`), &e); err == nil {
		t.Fatal("expected an error for a malformed Stream Deck payload")
	}
}

func TestConstants(t *testing.T) {
	if ValueLayout != "$A1" || CanvasLayout != "$A0" {
		t.Errorf("layouts: value=%s canvas=%s", ValueLayout, CanvasLayout)
	}
	if Keypad != "Keypad" || KeyPad != Keypad {
		t.Errorf("Keypad = %q", Keypad)
	}
	if KESDSDKDeviceTypeStreamDeckPlusXL != 13 || KESDSDKDeviceTypeVirtualStreamDeck != 11 || KESDSDKDeviceTypeStreamDeckPlus != 7 {
		t.Error("device type values do not match DeviceType")
	}
}
