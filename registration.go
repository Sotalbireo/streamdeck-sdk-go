package sdk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coder/websocket"
)

// Info is sent to the plugin on registration process.
type Info struct {
	Application struct {
		Font            string `json:"font"`
		Language        string `json:"language"`
		Platform        string `json:"platform"`
		PlatformVersion string `json:"platformVersion"`
		Version         string `json:"version"`
	} `json:"application"`
	Plugin struct {
		UUID    string `json:"uuid"`
		Version string `json:"version"`
	} `json:"plugin"`
	DevicePixelRatio float64 `json:"devicePixelRatio"`
	Colors           struct {
		ButtonMouseOverBackgroundColor string `json:"buttonMouseOverBackgroundColor"`
		ButtonPressedBackgroundColor   string `json:"buttonPressedBackgroundColor"`
		ButtonPressedBorderColor       string `json:"buttonPressedBorderColor"`
		ButtonPressedTextColor         string `json:"buttonPressedTextColor"`
		HighlightColor                 string `json:"highlightColor"`

		// Deprecated: no longer part of the registration info.
		DisabledColor string `json:"disabledColor"`

		// Deprecated: no longer part of the registration info.
		MouseDownColor string `json:"mouseDownColor"`
	} `json:"colors"`
	Devices []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Size struct {
			Columns int `json:"columns"`
			Rows    int `json:"rows"`
		} `json:"size"`
		Type int `json:"type"`
	} `json:"devices"`
}

// register send a registration event to StreamDeck SDK.
func (s *StreamDeck) register(ctx context.Context, inRegisterEvent string) error {
	data, err := json.Marshal(&SendEvent{
		Event: EventName(inRegisterEvent),
		UUID:  s.UUID,
	})
	if err != nil {
		return fmt.Errorf("encode registration: %w", err)
	}

	return s.conn.Write(ctx, websocket.MessageText, data)
}
