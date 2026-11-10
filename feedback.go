package sdk

// FeedbackPayload is the payload of SetFeedback: it maps the key of an item of the layout
// to its new value, which is a string, a number, or one of Bar, GBar, Pixmap and Text.
// See [SetFeedback](https://docs.elgato.com/streamdeck/sdk/references/websocket/plugin#setfeedback).
type FeedbackPayload map[string]any

// BarSubType is the shape of a Bar or a GBar.
type BarSubType uint8

const (
	// BarSubTypeRectangle is a rectangle.
	BarSubTypeRectangle BarSubType = iota

	// BarSubTypeDoubleRectangle is a double rectangle.
	BarSubTypeDoubleRectangle

	// BarSubTypeTrapezoid is a trapezoid.
	BarSubTypeTrapezoid

	// BarSubTypeDoubleTrapezoid is a double trapezoid.
	BarSubTypeDoubleTrapezoid

	// BarSubTypeGroove is a groove.
	BarSubTypeGroove
)

// Alignment is the alignment of the text of a Text item.
type Alignment string

const (
	// AlignmentCenter centers the text.
	AlignmentCenter Alignment = "center"

	// AlignmentLeft aligns the text to the left.
	AlignmentLeft Alignment = "left"

	// AlignmentRight aligns the text to the right.
	AlignmentRight Alignment = "right"
)

// TextOverflow is how the text of a Text item is shortened when it does not fit.
type TextOverflow string

const (
	// TextOverflowClip clips the text.
	TextOverflowClip TextOverflow = "clip"

	// TextOverflowEllipsis ends the text with an ellipsis.
	TextOverflowEllipsis TextOverflow = "ellipsis"

	// TextOverflowFade fades out the end of the text.
	TextOverflowFade TextOverflow = "fade"
)

// Range is the minimum and the maximum value of a Bar or a GBar.
type Range struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// Font is the font of a Text item.
type Font struct {
	// Size of the font.
	Size int `json:"size,omitempty"`

	// Weight of the font.
	Weight int `json:"weight,omitempty"`
}

// The layout items below are the values of a FeedbackPayload, see
// [Layouts](https://docs.elgato.com/streamdeck/sdk/references/layouts).
//
// All the properties are optional and only the ones that are set are sent,
// so that the other properties of the item are left unchanged.
// Properties that are meaningful when zero (Enabled, Opacity, ZOrder, Value...) are pointers,
// see BoolPtr, IntPtr, Float64Ptr and StringPtr.

// Bar is a bar item.
type Bar struct {
	// Background of the item: a color, or a gradient.
	Background string `json:"background,omitempty"`

	// Enabled shows or hides the item.
	Enabled *bool `json:"enabled,omitempty"`

	// Opacity of the item, from 0 to 1 in steps of 0.1.
	Opacity *float64 `json:"opacity,omitempty"`

	// ZOrder is the stacking order of the item.
	ZOrder *int `json:"zOrder,omitempty"`

	// BarBackgroundColor is the background color of the bar (bar_bg_c).
	BarBackgroundColor string `json:"bar_bg_c,omitempty"`

	// BarBorderColor is the border color of the bar (bar_border_c).
	BarBorderColor string `json:"bar_border_c,omitempty"`

	// BarFillColor is the fill color of the bar (bar_fill_c).
	BarFillColor string `json:"bar_fill_c,omitempty"`

	// BorderWidth is the width of the border of the bar (border_w).
	BorderWidth *int `json:"border_w,omitempty"`

	// Range is the minimum and the maximum value of the bar.
	Range *Range `json:"range,omitempty"`

	// Subtype is the shape of the bar.
	Subtype *BarSubType `json:"subtype,omitempty"`

	// Value is the current value of the bar.
	Value *float64 `json:"value,omitempty"`
}

// GBar is a gradient bar item: a Bar with a height.
type GBar struct {
	Background         string      `json:"background,omitempty"`
	Enabled            *bool       `json:"enabled,omitempty"`
	Opacity            *float64    `json:"opacity,omitempty"`
	ZOrder             *int        `json:"zOrder,omitempty"`
	BarBackgroundColor string      `json:"bar_bg_c,omitempty"`
	BarBorderColor     string      `json:"bar_border_c,omitempty"`
	BarFillColor       string      `json:"bar_fill_c,omitempty"`
	BorderWidth        *int        `json:"border_w,omitempty"`
	Range              *Range      `json:"range,omitempty"`
	Subtype            *BarSubType `json:"subtype,omitempty"`
	Value              *float64    `json:"value,omitempty"`

	// BarHeight is the height of the bar (bar_h).
	BarHeight *int `json:"bar_h,omitempty"`
}

// Pixmap is an image item.
type Pixmap struct {
	Background string   `json:"background,omitempty"`
	Enabled    *bool    `json:"enabled,omitempty"`
	Opacity    *float64 `json:"opacity,omitempty"`
	ZOrder     *int     `json:"zOrder,omitempty"`

	// Value is the image: a base64 encoded image, or a path to an image file.
	Value *string `json:"value,omitempty"`
}

// Text is a text item.
type Text struct {
	Background string   `json:"background,omitempty"`
	Enabled    *bool    `json:"enabled,omitempty"`
	Opacity    *float64 `json:"opacity,omitempty"`
	ZOrder     *int     `json:"zOrder,omitempty"`

	// Alignment of the text.
	Alignment Alignment `json:"alignment,omitempty"`

	// Color of the text.
	Color string `json:"color,omitempty"`

	// Font of the text.
	Font *Font `json:"font,omitempty"`

	// TextOverflow is how the text is shortened (text-overflow).
	TextOverflow TextOverflow `json:"text-overflow,omitempty"`

	// Value is the text. It is a pointer so that an empty text can be sent.
	Value *string `json:"value,omitempty"`
}

// BoolPtr returns a pointer to b.
func BoolPtr(b bool) *bool { return new(b) }

// IntPtr returns a pointer to i.
func IntPtr(i int) *int { return new(i) }

// Float64Ptr returns a pointer to f.
func Float64Ptr(f float64) *float64 { return new(f) }

// StringPtr returns a pointer to s.
func StringPtr(s string) *string { return new(s) }

// BarSubTypePtr returns a pointer to t.
func BarSubTypePtr(t BarSubType) *BarSubType { return new(t) }

// SendEventSwitchProfilePayload describes a payload for switchToProfile event to send to StreamDeck SDK.
type SendEventSwitchProfilePayload struct {
	// The name of the profile to switch to, identical to the name provided in the manifest.json file.
	// When empty, Stream Deck switches back to the previous profile.
	Profile string `json:"profile,omitempty"`

	// The 0-indexed page of the profile to switch to.
	// It is a pointer so that the first page (0) can be sent explicitly.
	Page *int `json:"page,omitempty"`
}

// SwitchProfileTo switches to a pre-defined profile distributed with the plugin on the specified device.
// A nil payload switches back to the previous profile.
// Stream Deck expects the UUID of the plugin as the context: an empty context means StreamDeck.UUID.
func (s *StreamDeck) SwitchProfileTo(context, device string, payload *SendEventSwitchProfilePayload) {
	event := &SendEvent{
		Event:   SwitchToProfile,
		Context: s.pluginContext(context),
		Device:  device,
	}

	// A nil pointer in the interface would be encoded as null, not omitted.
	if payload != nil {
		event.Payload = payload
	}

	s.send(event)
}
