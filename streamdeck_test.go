package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

const (
	testInfo    = `{"application":{"version":"7.1"},"devicePixelRatio":1.5,"colors":{"buttonMouseOverBackgroundColor":"#fff"},"devices":[],"plugin":{"uuid":"p","version":"1"}}`
	testTimeout = 3 * time.Second
)

// fakeDeck is a Stream Deck WebSocket server.
type fakeDeck struct {
	url  string
	conn chan *websocket.Conn
}

func newFakeDeck(t *testing.T) *fakeDeck {
	t.Helper()

	d := &fakeDeck{conn: make(chan *websocket.Conn, 1)}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Errorf("accept: %v", err)
			return
		}
		c.SetReadLimit(readLimit)
		d.conn <- c
	}))
	d.url = "ws" + strings.TrimPrefix(srv.URL, "http")

	t.Cleanup(srv.Close)

	return d
}

func (d *fakeDeck) accept(t *testing.T) *websocket.Conn {
	t.Helper()

	select {
	case c := <-d.conn:
		t.Cleanup(func() { _ = c.CloseNow() })
		return c
	case <-time.After(testTimeout):
		t.Fatal("no connection")
		return nil
	}
}

func readRaw(t *testing.T, c *websocket.Conn) []byte {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	_, data, err := c.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	return data
}

func readMessage(t *testing.T, c *websocket.Conn) map[string]any {
	t.Helper()

	var m map[string]any
	if err := json.Unmarshal(readRaw(t, c), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	return m
}

func writeMessage(t *testing.T, c *websocket.Conn, data string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	if err := c.Write(ctx, websocket.MessageText, []byte(data)); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func startPlugin(t *testing.T, opts ...Option) (*StreamDeck, *websocket.Conn) {
	t.Helper()

	d := newFakeDeck(t)

	type result struct {
		s   *StreamDeck
		err error
	}
	ch := make(chan result, 1)
	go func() {
		s, err := connect(d.url, "plugin-uuid", "registerPlugin", testInfo, opts...)
		ch <- result{s, err}
	}()

	server := d.accept(t)

	reg := readMessage(t, server)
	if reg["event"] != "registerPlugin" || reg["uuid"] != "plugin-uuid" {
		t.Fatalf("unexpected registration: %v", reg)
	}

	r := <-ch
	if r.err != nil {
		t.Fatal(r.err)
	}

	t.Cleanup(func() { _ = r.s.conn.CloseNow() })

	if r.s.Info.DevicePixelRatio != 1.5 || r.s.Info.Colors.ButtonMouseOverBackgroundColor != "#fff" {
		t.Fatalf("unexpected info: %+v", r.s.Info)
	}

	return r.s, server
}

func TestConnect_Validation(t *testing.T) {
	for name, tc := range map[string]struct {
		uuid, event, info string
		want              error
	}{
		"uuid":  {"", "e", "{}", ErrMissingUUID},
		"event": {"u", "", "{}", ErrMissingRegisterEvent},
		"empty": {"u", "e", "", ErrMissingOrInvalidInfo},
		"bad":   {"u", "e", "{", ErrMissingOrInvalidInfo},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := connect("ws://127.0.0.1:1", tc.uuid, tc.event, tc.info); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestConnect_Unreachable(t *testing.T) {
	if _, err := connect("ws://127.0.0.1:1", "u", "e", "{}"); err == nil {
		t.Fatal("expected an error")
	}
}

func jsonEqual(t *testing.T, a, b string) bool {
	t.Helper()

	var x, y any
	if err := json.Unmarshal([]byte(a), &x); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(b), &y); err != nil {
		t.Fatal(err)
	}

	xb, _ := json.Marshal(x)
	yb, _ := json.Marshal(y)

	return string(xb) == string(yb)
}

// TestSentCommands checks the messages written on the wire.
func TestSentCommands(t *testing.T) {
	s, server := startPlugin(t)

	go s.writer(t.Context())

	for _, tc := range []struct {
		name string
		call func()
		want string
	}{
		{"setState 0", func() { s.SetState("c", 0) }, `{"event":"setState","context":"c","payload":{"state":0}}`},
		{"setState 1", func() { s.SetState("c", 1) }, `{"event":"setState","context":"c","payload":{"state":1}}`},
		{"setTitle", func() { s.SetTitle("c", "t", Hardware) }, `{"event":"setTitle","context":"c","payload":{"title":"t","target":1}}`},
		{"setTitle state 0", func() { s.SetTitle("c", "t", HardwareAndSoftware, ForState(0)) }, `{"event":"setTitle","context":"c","payload":{"title":"t","state":0}}`},
		{"setTitle empty hides the title", func() { s.SetTitle("c", "", HardwareAndSoftware) }, `{"event":"setTitle","context":"c","payload":{"title":""}}`},
		{"resetTitle", func() { s.ResetTitle("c", Software, ForState(1)) }, `{"event":"setTitle","context":"c","payload":{"target":2,"state":1}}`},
		{"setImage", func() { s.SetImage("c", "img", ForTarget(Software), ForState(1)) }, `{"event":"setImage","context":"c","payload":{"target":2,"state":1,"image":"img"}}`},
		{"getSettings", func() { s.GetSettings("c") }, `{"event":"getSettings","context":"c"}`},
		{"getSettings id", func() { s.GetSettingsWithID("c", "i") }, `{"event":"getSettings","context":"c","id":"i"}`},
		{"getGlobalSettings id", func() { s.GetGlobalSettingsWithID("c", "i") }, `{"event":"getGlobalSettings","context":"c","id":"i"}`},
		{"getResources", func() { s.GetResources("c") }, `{"event":"getResources","context":"c"}`},
		{"setResources", func() { s.SetResources("c", map[string]string{"a": "b"}) }, `{"event":"setResources","context":"c","payload":{"a":"b"}}`},
		{"setSettings nil", func() { s.SetSettings("c", nil) }, `{"event":"setSettings","context":"c","payload":{}}`},
		{"sendToPropertyInspector", func() { s.SendToPropertyInspector("c", "str") }, `{"event":"sendToPropertyInspector","context":"c","payload":"str"}`},
		{"setFeedback typed", func() {
			s.SetFeedback("c", &FeedbackPayload{
				"title":  "Volume",
				"value":  42,
				"bar":    Bar{Value: Float64Ptr(42), Range: &Range{Min: 0, Max: 100}, Enabled: BoolPtr(false), ZOrder: IntPtr(0), Opacity: Float64Ptr(0)},
				"text":   Text{Value: StringPtr(""), Font: &Font{Size: 16}, TextOverflow: TextOverflowEllipsis},
				"canvas": Pixmap{Background: "#000"},
			})
		}, `{"event":"setFeedback","context":"c","payload":{"title":"Volume","value":42,
			"bar":{"value":42,"range":{"min":0,"max":100},"enabled":false,"zOrder":0,"opacity":0},
			"text":{"value":"","font":{"size":16},"text-overflow":"ellipsis"},
			"canvas":{"background":"#000"}}}`},
		{"setFeedbackLayout", func() { s.SetFeedbackLayout("c", IndicatorLayout) }, `{"event":"setFeedbackLayout","context":"c","payload":{"layout":"$B1"}}`},
		{"setTriggerDescription", func() {
			s.SetTriggerDescription("c", &SendEventSetTriggerDescriptionPayload{Rotate: "Volume", Push: "Mute"})
		}, `{"event":"setTriggerDescription","context":"c","payload":{"push":"Mute","rotate":"Volume"}}`},
		{"setTriggerDescription nil", func() { s.SetTriggerDescription("c", nil) }, `{"event":"setTriggerDescription","context":"c","payload":{}}`},
		{"setFeedback nil", func() { s.SetFeedback("c", nil) }, `{"event":"setFeedback","context":"c","payload":{}}`},
		{"switchProfile typed", func() {
			s.SwitchProfileTo("c", "d", &SendEventSwitchProfilePayload{Profile: "p", Page: new(0)})
		}, `{"event":"switchToProfile","context":"c","device":"d","payload":{"profile":"p","page":0}}`},
		{"switchProfile previous", func() { s.SwitchProfileTo("c", "d", nil) }, `{"event":"switchToProfile","context":"c","device":"d"}`},
		{"switchProfile nil map", func() { s.SwitchProfile("c", "d", nil) }, `{"event":"switchToProfile","context":"c","device":"d"}`},
		{"switchProfile map", func() {
			s.SwitchProfile("c", "d", map[string]any{"profile": "p", "page": 2})
		}, `{"event":"switchToProfile","context":"c","device":"d","payload":{"profile":"p","page":2}}`},
		{"openUrl", func() { s.OpenURL("https://example.com") }, `{"event":"openUrl","payload":{"url":"https://example.com"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			go tc.call()

			if raw := readRaw(t, server); !jsonEqual(t, string(raw), tc.want) {
				t.Fatalf("got %s, want %s", raw, tc.want)
			}
		})
	}
}

// TestStart_SurvivesNonObjectPayload checks a sendToPlugin string payload does not stop the plugin.
func TestStart_SurvivesNonObjectPayload(t *testing.T) {
	errs := make(chan error, 4)
	s, server := startPlugin(t, WithErrorHandler(func(err error) { errs <- err }))

	got := make(chan *ReceivedEvent, 2)
	s.HandlerFunc(func(e *ReceivedEvent) error {
		got <- e
		return nil
	})

	go s.StartWithContext(t.Context())

	writeMessage(t, server, `{"event":"sendToPlugin","action":"a","context":"c","payload":"a string"}`)
	writeMessage(t, server, `{"event":"keyDown","action":"a","context":"c","device":"d","payload":{"controller":"Keypad"}}`)

	for _, want := range []EventName{SendToPlugin, KeyDown} {
		select {
		case e := <-got:
			if e.Event != want {
				t.Fatalf("event = %s, want %s", e.Event, want)
			}
		case err := <-errs:
			t.Fatalf("unexpected error: %v", err)
		case <-time.After(testTimeout):
			t.Fatalf("event %s not received", want)
		}
	}
}

// TestStart_SurvivesMalformedEvent checks a malformed event is reported and the next ones are still handled.
func TestStart_SurvivesMalformedEvent(t *testing.T) {
	errs := make(chan error, 4)
	s, server := startPlugin(t, WithErrorHandler(func(err error) { errs <- err }))

	got := make(chan *ReceivedEvent, 1)
	s.HandlerFunc(func(e *ReceivedEvent) error {
		got <- e
		return nil
	})

	go s.StartWithContext(t.Context())

	writeMessage(t, server, `not json`)
	writeMessage(t, server, `{"event":"keyDown","payload":{"ticks":"x"}}`)
	writeMessage(t, server, `{"event":"systemDidWakeUp"}`)

	for range 2 {
		select {
		case <-errs:
		case <-time.After(testTimeout):
			t.Fatal("malformed event not reported")
		}
	}

	select {
	case e := <-got:
		if e.Event != SystemDidWakeUp {
			t.Fatalf("event = %s", e.Event)
		}
	case <-time.After(testTimeout):
		t.Fatal("event after the malformed ones not received")
	}
}

// TestStart_LargeEvent checks events bigger than the default read limit of the websocket package are received.
func TestStart_LargeEvent(t *testing.T) {
	errs := make(chan error, 4)
	s, server := startPlugin(t, WithErrorHandler(func(err error) { errs <- err }))

	got := make(chan *ReceivedEvent, 1)
	s.HandlerFunc(func(e *ReceivedEvent) error {
		got <- e
		return nil
	})

	go s.StartWithContext(t.Context())

	big := strings.Repeat("x", 1<<20)
	writeMessage(t, server, `{"event":"didReceiveGlobalSettings","payload":{"settings":{"big":"`+big+`"}}}`)

	select {
	case e := <-got:
		if e.Payload == nil || e.Payload.Settings["big"] != big {
			t.Fatal("settings not received intact")
		}
	case err := <-errs:
		t.Fatalf("unexpected error: %v", err)
	case <-time.After(testTimeout):
		t.Fatal("event not received")
	}
}

// TestStart_ClosesGracefully checks the connection is closed with a closing handshake once the context is done.
func TestStart_ClosesGracefully(t *testing.T) {
	s, server := startPlugin(t)

	ctx, cancel := context.WithCancel(t.Context())
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		s.StartWithContext(ctx)
	}()

	cancel()

	readCtx, readCancel := context.WithTimeout(context.Background(), testTimeout)
	defer readCancel()

	_, _, err := server.Read(readCtx)
	if got := websocket.CloseStatus(err); got != websocket.StatusNormalClosure {
		t.Fatalf("close status = %d (%v), want %d", got, err, websocket.StatusNormalClosure)
	}

	select {
	case <-stopped:
	case <-time.After(testTimeout):
		t.Fatal("StartWithContext did not return")
	}
}

// TestStart_ReturnsWhenDeckCloses checks the plugin stops when Stream Deck closes the connection.
func TestStart_ReturnsWhenDeckCloses(t *testing.T) {
	s, server := startPlugin(t)

	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		s.StartWithContext(t.Context())
	}()

	if err := server.Close(websocket.StatusNormalClosure, ""); err != nil {
		t.Logf("close: %v", err)
	}

	select {
	case <-stopped:
	case <-time.After(testTimeout):
		t.Fatal("StartWithContext did not return")
	}
}

// TestHandlerError_AlertOnlyWithContext checks a failing handler alerts only when there is a context.
func TestHandlerError_AlertOnlyWithContext(t *testing.T) {
	s, server := startPlugin(t, WithErrorHandler(func(error) {}))

	s.HandlerFunc(func(*ReceivedEvent) error { return errors.New("boom") })

	go s.StartWithContext(t.Context())

	writeMessage(t, server, `{"event":"deviceDidConnect","device":"d"}`)
	writeMessage(t, server, `{"event":"keyDown","action":"a","context":"ctx","payload":{}}`)

	msg := readMessage(t, server)
	if msg["event"] != "showAlert" || msg["context"] != "ctx" {
		t.Fatalf("first message = %v, want showAlert for ctx", msg)
	}
}

// TestStart_DebugDoesNotBlockWriter checks that commands are still sent when the debug mode is on:
// the writer used to deadlock by sending its own "writer started" log to the channel it consumes.
func TestStart_DebugDoesNotBlockWriter(t *testing.T) {
	s, server := startPlugin(t, WithDebug(true))

	s.HandlerFunc(func(e *ReceivedEvent) error {
		s.SetState(e.Context, 1)
		return nil
	})

	go s.StartWithContext(t.Context())

	writeMessage(t, server, `{"event":"keyDown","action":"a","context":"c","device":"d","payload":{"controller":"Keypad"}}`)

	// The debug logs and the command may come in any order.
	for {
		m := readMessage(t, server)
		if m["event"] == "setState" {
			if m["context"] != "c" {
				t.Fatalf("unexpected setState: %v", m)
			}

			return
		}

		if m["event"] != "logMessage" {
			t.Fatalf("unexpected message: %v", m)
		}
	}
}

// TestGlobalCommandsUsePluginUUID checks the commands that are not related to an action use the UUID of the plugin
// as the context, like the official SDK does.
func TestGlobalCommandsUsePluginUUID(t *testing.T) {
	s, server := startPlugin(t)
	go s.StartWithContext(t.Context())

	s.SetGlobalSettings("", map[string]any{"k": "v"})
	s.GetGlobalSettings("")
	s.SwitchProfileTo("", "dev", &SendEventSwitchProfilePayload{Profile: "p"})
	s.SwitchProfile("", "dev", nil)
	s.GetGlobalSettingsWithID("explicit", "1")

	for _, want := range []string{"plugin-uuid", "plugin-uuid", "plugin-uuid", "plugin-uuid", "explicit"} {
		if m := readMessage(t, server); m["context"] != want {
			t.Fatalf("context = %v, want %v (%v)", m["context"], want, m)
		}
	}
}

// TestStart_IgnoresMessageWithoutEvent checks a message that is not an event is reported, not dispatched.
func TestStart_IgnoresMessageWithoutEvent(t *testing.T) {
	errs := make(chan error, 2)
	s, server := startPlugin(t, WithErrorHandler(func(err error) { errs <- err }))

	got := make(chan *ReceivedEvent, 2)
	s.HandlerFunc(func(e *ReceivedEvent) error {
		got <- e
		return nil
	})

	go s.StartWithContext(t.Context())

	writeMessage(t, server, `{"hello":"world"}`)
	writeMessage(t, server, `{"event":"systemDidWakeUp"}`)

	select {
	case e := <-got:
		if e.Event != SystemDidWakeUp {
			t.Fatalf("event = %q, want systemDidWakeUp", e.Event)
		}
	case <-time.After(testTimeout):
		t.Fatal("event not received")
	}

	select {
	case <-errs:
	case <-time.After(testTimeout):
		t.Fatal("unknown message not reported")
	}
}
