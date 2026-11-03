package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/coder/websocket"
)

const (
	// dialTimeout is the time given to open the WebSocket and to register the plugin.
	dialTimeout = 10 * time.Second

	// writeTimeout is the time given to send a single event.
	writeTimeout = 10 * time.Second

	// readLimit is the size of the biggest event that can be received.
	// The default of the websocket package (32 KiB) is too small for settings and resources.
	readLimit = 16 << 20
)

// StreamDeck will handle events to send/received to/from the StreamDeck application.
type StreamDeck struct {
	// UUID is a unique identifier string that should be used to register the plugin once the WebSocket is opened
	UUID string

	// Info containing the Stream Deck application information and devices information.
	Info *Info

	conn    *websocket.Conn
	readCh  chan *ReceivedEvent
	writeCh chan *SendEvent
	done    chan struct{}

	// handlers will process incoming events
	handlers []HandlerFunc

	debug   bool
	onError func(error)
}

var (
	port          = flag.Int("port", 0, "The port that should be used to create the WebSocket")
	pluginUUID    = flag.String("pluginUUID", "", "A unique identifier string that should be used to register the plugin once the WebSocket is opened")
	registerEvent = flag.String("registerEvent", "", "The event type that should be used to register the plugin once the WebSocket is opened")
	info          = flag.String("info", "", "A stringified json containing the Stream Deck application information and devices information.")

	// ErrMissingPort is returned when the port is not set.
	ErrMissingPort = errors.New("missing -port")

	// ErrMissingUUID is returned when the UUID is not set.
	ErrMissingUUID = errors.New("missing -pluginUUID")

	// ErrMissingRegisterEvent is returned when the registerEvent is not set.
	ErrMissingRegisterEvent = errors.New("missing -registerEvent")

	// ErrMissingOrInvalidInfo is returned when the info is not set or is not a valid json.
	ErrMissingOrInvalidInfo = errors.New("missing or invalid -info")
)

// New create our plugin, listen to websocket events and register handlers.
func New(opts ...Option) (*StreamDeck, error) {
	flag.Parse()

	// port
	if *port == 0 {
		return nil, ErrMissingPort
	}

	return connect(fmt.Sprintf("ws://localhost:%d", *port), *pluginUUID, *registerEvent, *info, opts...)
}

// connect validates the registration parameters, opens the WebSocket at url and registers the plugin.
func connect(url, uuid, regEvent, infoJSON string, opts ...Option) (*StreamDeck, error) {
	// string values
	if uuid == "" {
		return nil, ErrMissingUUID
	}

	if regEvent == "" {
		return nil, ErrMissingRegisterEvent
	}

	if infoJSON == "" {
		return nil, ErrMissingOrInvalidInfo
	}

	// info json object
	var r Info
	if err := json.Unmarshal([]byte(infoJSON), &r); err != nil {
		return nil, ErrMissingOrInvalidInfo
	}

	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot init websocket connection: %w", err)
	}

	conn.SetReadLimit(readLimit)

	streamdeck := &StreamDeck{
		UUID:     uuid,
		Info:     &r,
		conn:     conn,
		readCh:   make(chan *ReceivedEvent),
		writeCh:  make(chan *SendEvent),
		done:     make(chan struct{}),
		handlers: make([]HandlerFunc, 0),
		debug:    false,
	}

	for _, opt := range opts {
		opt(streamdeck)
	}

	if err := streamdeck.register(ctx, regEvent); err != nil {
		_ = conn.CloseNow()
		return nil, fmt.Errorf("cannot register plugin: %w", err)
	}

	return streamdeck, nil
}

// Start to serve the plugin until it is interrupted (Ctrl+C, SIGTERM) or the connection is closed.
// Ensure you call Handler before.
func (s *StreamDeck) Start() {
	// os.Kill cannot be caught, SIGTERM is what a process manager sends to ask for a graceful stop.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	s.StartWithContext(ctx)
}

// StartWithContext serves the plugin until ctx is done or the connection is closed.
// Ensure you call Handler before.
func (s *StreamDeck) StartWithContext(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	wg.Go(func() {
		s.reader(ctx) // read incoming events
	})

	go s.writer(ctx) // send events
	s.process(ctx)   // will block until ctx is closed or the connection is lost

	cancel()
	wg.Wait() // the connection is closed once the reader is done
}

// reader listen on incoming messages and send them to dedicated channel.
func (s *StreamDeck) reader(ctx context.Context) {
	defer close(s.readCh)

	// Reading is not tied to ctx: cancelling a read context makes the websocket package drop the connection,
	// whereas closing it here ends the reading with a proper closing handshake.
	closed := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		defer close(closed)
		_ = s.conn.Close(websocket.StatusNormalClosure, "")
	})
	defer func() {
		if stop() {
			// ctx is not done: the connection is lost or closed by Stream Deck.
			_ = s.conn.CloseNow()
			return
		}

		<-closed // wait for the closing handshake started because of ctx
	}()

	if s.debug {
		s.Log("[DEBUG] reader started")
	}

	for {
		_, data, err := s.conn.Read(context.Background())
		if err != nil {
			if ctx.Err() != nil {
				return
			}

			switch websocket.CloseStatus(err) {
			case websocket.StatusNormalClosure, websocket.StatusGoingAway:
				// Stream Deck is closing the connection.
			case -1:
				s.reportError(fmt.Errorf("read message: %w", err))
			default:
				s.reportError(fmt.Errorf("unexpected close connection: %w", err))
			}

			return
		}

		var event ReceivedEvent
		if err := json.Unmarshal(data, &event); err != nil {
			// A single malformed event must not stop the plugin.
			s.reportError(fmt.Errorf("decode message: %w", err))
			continue
		}

		select {
		case s.readCh <- &event:
		case <-ctx.Done():
			return
		}
	}
}

// writer listen on write channel and send messages.
func (s *StreamDeck) writer(ctx context.Context) {
	var writeErr error
	defer func() {
		close(s.done)
		if writeErr != nil {
			s.reportError(writeErr)
		}
	}()

	if s.debug {
		s.Log("[DEBUG] writer started")
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-s.writeCh:
			data, err := json.Marshal(event)
			if err != nil {
				// An event that cannot be encoded must not stop the plugin.
				s.reportError(fmt.Errorf("encode event [%s] for action [%s]: %w", event.Event, event.Action, err))
				continue
			}

			if err := s.write(data); err != nil {
				writeErr = fmt.Errorf("write event [%s] for action [%s]: %w", event.Event, event.Action, err)
				return
			}
		}
	}
}

// write sends an already encoded message.
func (s *StreamDeck) write(data []byte) error {
	// Not tied to the context of the plugin: a cancelled context makes the websocket package drop the connection.
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	return s.conn.Write(ctx, websocket.MessageText, data)
}

// process will listen to incoming events and process them.
func (s *StreamDeck) process(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-s.readCh:
			if !ok {
				// The channel is closed
				return
			}

			// Send event to all to registered handlers
			if s.debug {
				s.Logf("[DEBUG] received event [%s] for action [%s]", event.Event, event.Action)
			}

			for _, h := range s.handlers {
				if err := h(event); err != nil {
					s.reportError(fmt.Errorf("event [%s] action [%s]: %w", event.Event, event.Action, err))

					// Only the events related to an action instance can show an alert.
					if event.Context != "" {
						s.Alert(event.Context)
					}
				}
			}
		}
	}
}

func (s *StreamDeck) reportError(err error) {
	if s.onError != nil {
		s.onError(err)
		return
	}
	s.Logf("[ERROR] %v", err)
}

// send queues event for the writer and drops it once the writer has stopped.
func (s *StreamDeck) send(event *SendEvent) {
	select {
	case s.writeCh <- event:
	case <-s.done:
	}
}
