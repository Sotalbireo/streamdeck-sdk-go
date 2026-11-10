package sdk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coder/websocket"
)

// Language is a language supported by Stream Deck.
type Language string

const (
	// LanguageDE is German.
	LanguageDE Language = "de"

	// LanguageEN is English.
	LanguageEN Language = "en"

	// LanguageES is Spanish.
	LanguageES Language = "es"

	// LanguageFR is French.
	LanguageFR Language = "fr"

	// LanguageJA is Japanese.
	LanguageJA Language = "ja"

	// LanguageKO is Korean.
	LanguageKO Language = "ko"

	// LanguageZhCN is Simplified Chinese.
	LanguageZhCN Language = "zh_CN"

	// LanguageZhTW is Traditional Chinese.
	LanguageZhTW Language = "zh_TW"
)

// Platform is the operating system Stream Deck runs on.
type Platform string

const (
	// PlatformMac is macOS.
	PlatformMac Platform = "mac"

	// PlatformWindows is Windows.
	PlatformWindows Platform = "windows"
)

// RegistrationInfo is sent to the plugin on registration process.
// See [RegistrationInfo](https://docs.elgato.com/streamdeck/sdk/references/websocket/plugin#registrationinfo).
type RegistrationInfo struct {
	Application struct {
		Font            string   `json:"font"`
		Language        Language `json:"language"`
		Platform        Platform `json:"platform"`
		PlatformVersion string   `json:"platformVersion"`
		Version         string   `json:"version"`
	} `json:"application"`
	Colors struct {
		ButtonMouseOverBackgroundColor string `json:"buttonMouseOverBackgroundColor"`
		ButtonPressedBackgroundColor   string `json:"buttonPressedBackgroundColor"`
		ButtonPressedBorderColor       string `json:"buttonPressedBorderColor"`
		ButtonPressedTextColor         string `json:"buttonPressedTextColor"`
		HighlightColor                 string `json:"highlightColor"`
	} `json:"colors"`
	DevicePixelRatio float64 `json:"devicePixelRatio"`
	Devices          []struct {
		DeviceInfo

		ID string `json:"id"`
	} `json:"devices"`
	Plugin struct {
		UUID    string `json:"uuid"`
		Version string `json:"version"`
	} `json:"plugin"`
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
