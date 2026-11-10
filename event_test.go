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
		!p.IsInMultiAction || p.UserDesiredState == nil || *p.UserDesiredState != 1 || p.Resources["icon"] != "a.png" {
		t.Fatalf("unexpected payload: %+v", p)
	}
}

func TestReceivedEvent_NewEvents(t *testing.T) {
	e := decode(t, `{"event":"deviceDidChange","device":"d","deviceInfo":{"name":"n","type":13,"size":{"columns":9,"rows":4}}}`)
	if e.Event != DeviceDidChange || e.DeviceInfo.Type != DeviceTypeStreamDeckPlusXL || e.DeviceInfo.Size.Columns != 9 {
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
	if Keypad != "Keypad" {
		t.Errorf("Keypad = %q", Keypad)
	}
	if DeviceTypeStreamDeckPlusXL != 13 || DeviceTypeVirtualStreamDeck != 11 || DeviceTypeStreamDeckPlus != 7 {
		t.Error("device type values do not match DeviceType")
	}
}

func TestRegistrationInfoDevices(t *testing.T) {
	const raw = `{"application":{"language":"ja","platform":"windows"},
		"devices":[{"id":"D1","name":"XL","size":{"columns":8,"rows":4},"type":2}]}`

	var info RegistrationInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		t.Fatal(err)
	}

	d := info.Devices[0]
	if d.ID != "D1" || d.Type != DeviceTypeStreamDeckXL || d.Size != (Size{Columns: 8, Rows: 4}) || d.Name != "XL" {
		t.Errorf("unexpected device: %+v", d)
	}

	if info.Application.Language != LanguageJA || info.Application.Platform != PlatformWindows {
		t.Errorf("unexpected application: %+v", info.Application)
	}
}

func TestTitleParametersTypes(t *testing.T) {
	var e ReceivedEvent
	raw := `{"event":"titleParametersDidChange","action":"a","context":"c","device":"d","payload":{
		"controller":"Keypad","coordinates":{"column":0,"row":0},"settings":{},"title":"T",
		"titleParameters":{"fontStyle":"Bold Italic","titleAlignment":"bottom"}}}`

	if err := json.Unmarshal([]byte(raw), &e); err != nil {
		t.Fatal(err)
	}

	tp := e.Payload.TitleParameters
	if tp.FontStyle != FontStyleBoldItalic || tp.TitleAlignment != TitleAlignmentBottom {
		t.Errorf("unexpected title parameters: %+v", tp)
	}
}

func TestBarSubTypeValues(t *testing.T) {
	if BarSubTypeRectangle != 0 || BarSubTypeDoubleRectangle != 1 || BarSubTypeTrapezoid != 2 ||
		BarSubTypeDoubleTrapezoid != 3 || BarSubTypeGroove != 4 {
		t.Error("bar sub type values do not match BarSubType")
	}
}

func TestPayloadStateIsOptional(t *testing.T) {
	for name, tc := range map[string]struct {
		payload     string
		state, user *uint8
	}{
		"absent":        {`{"controller":"Keypad"}`, nil, nil},
		"zero":          {`{"controller":"Keypad","state":0,"userDesiredState":0}`, new(uint8(0)), new(uint8(0))},
		"first and one": {`{"controller":"Keypad","state":1,"userDesiredState":1}`, new(uint8(1)), new(uint8(1))},
	} {
		t.Run(name, func(t *testing.T) {
			e := decode(t, `{"event":"keyDown","action":"a","context":"c","device":"d","payload":`+tc.payload+`}`)

			check := func(field string, got, want *uint8) {
				t.Helper()

				if (got == nil) != (want == nil) || got != nil && *got != *want {
					t.Errorf("%s = %v, want %v", field, got, want)
				}
			}

			check("State", e.Payload.State, tc.state)
			check("UserDesiredState", e.Payload.UserDesiredState, tc.user)
		})
	}
}

func TestReceivedEvent_TouchAndDial(t *testing.T) {
	e := decode(t, `{"event":"touchTap","action":"a","context":"c","device":"d","payload":{
		"controller":"Encoder","coordinates":{"column":2,"row":0},"hold":true,"tapPos":[120,35],"settings":{}}}`)
	if e.Event != TouchTap || e.Payload.Controller != Encoder || !e.Payload.Hold ||
		e.Payload.TapPos != [2]int{120, 35} || e.Payload.Coordinates.Column != 2 {
		t.Fatalf("unexpected touchTap: %+v", e.Payload)
	}

	for _, name := range []EventName{DialDown, DialUp} {
		e = decode(t, `{"event":"`+string(name)+`","action":"a","context":"c","device":"d","payload":{
			"controller":"Encoder","coordinates":{"column":1,"row":0},"settings":{}}}`)
		if e.Event != name || e.Payload.Controller != Encoder || e.Payload.Coordinates.Column != 1 {
			t.Fatalf("unexpected %s: %+v", name, e.Payload)
		}
	}

	e = decode(t, `{"event":"dialRotate","action":"a","context":"c","device":"d","payload":{
		"controller":"Encoder","coordinates":{"column":3,"row":0},"pressed":true,"ticks":-2,"settings":{}}}`)
	if e.Event != DialRotate || e.Payload.Ticks != -2 || !e.Payload.Pressed {
		t.Fatalf("unexpected dialRotate: %+v", e.Payload)
	}
}

func TestLayout_JSON(t *testing.T) {
	l := &Layout{
		ID:         "com.example.layout",
		Controller: Encoder,
		Items: []LayoutItem{
			{Key: "icon", Type: LayoutItemPixmap, Rect: Rect{16, 10, 40, 40}, Value: "icon.svg"},
			{Key: "title", Type: LayoutItemText, Rect: Rect{0, 50, 100, 20}, Alignment: AlignmentCenter, Font: &Font{Size: 16, Weight: 700}},
			{Key: "bar", Type: LayoutItemGBar, Rect: Rect{8, 80, 84, 16}, Value: 50.0, Range: &Range{Min: 0, Max: 100},
				Subtype: BarSubTypePtr(BarSubTypeGroove), BarHeight: IntPtr(10), ZOrder: IntPtr(0)},
		},
	}

	b, err := l.JSON()
	if err != nil {
		t.Fatal(err)
	}

	want := `{"id":"com.example.layout","controller":"Encoder","items":[
		{"key":"icon","type":"pixmap","rect":[16,10,40,40],"value":"icon.svg"},
		{"key":"title","type":"text","rect":[0,50,100,20],"alignment":"center","font":{"size":16,"weight":700}},
		{"key":"bar","type":"gbar","rect":[8,80,84,16],"zOrder":0,"value":50,"range":{"min":0,"max":100},"subtype":4,"bar_h":10}]}`
	if !jsonEqual(t, string(b), want) {
		t.Fatalf("got %s", b)
	}
}

func TestLayoutItem_Validate(t *testing.T) {
	for name, tc := range map[string]struct {
		item LayoutItem
		ok   bool
	}{
		"bar number":    {LayoutItem{Key: "k", Type: LayoutItemBar, Value: 50}, true},
		"gbar float":    {LayoutItem{Key: "k", Type: LayoutItemGBar, Value: 50.5}, true},
		"bar string":    {LayoutItem{Key: "k", Type: LayoutItemBar, Value: "50"}, false},
		"text string":   {LayoutItem{Key: "k", Type: LayoutItemText, Value: "x"}, true},
		"text number":   {LayoutItem{Key: "k", Type: LayoutItemText, Value: 1}, false},
		"pixmap string": {LayoutItem{Key: "k", Type: LayoutItemPixmap, Value: "a.svg"}, true},
		"pixmap bool":   {LayoutItem{Key: "k", Type: LayoutItemPixmap, Value: true}, false},
		"nil value":     {LayoutItem{Key: "k", Type: LayoutItemText}, true},
		"unknown type":  {LayoutItem{Key: "k", Type: "foo"}, false},
	} {
		t.Run(name, func(t *testing.T) {
			if err := tc.item.Validate(); (err == nil) != tc.ok {
				t.Fatalf("Validate() = %v, want ok=%v", err, tc.ok)
			}
		})
	}
}
