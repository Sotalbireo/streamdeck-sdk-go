package sdk

// FeedbackItem describes the properties of an item of a layout of the Stream Deck + touch display,
// for use as a value of SendEventSetFeedbackPayload.
// See [Layouts](https://docs.elgato.com/streamdeck/sdk/references/layouts).
//
// All the properties are optional and only the ones that are set are sent,
// so that the other properties of the item are left unchanged.
// Properties that are meaningful when zero (Enabled, Opacity, ZOrder, BorderWidth, Range...)
// are pointers, see BoolPtr, IntPtr and Float64Ptr.
//
// The Value is either a string or a number, depending on the type of the item:
// text of a text item, base64 image or path of a pixmap item, current value of a bar item.
type FeedbackItem struct {
	// Value of the item. It is not sent when nil, an empty string is sent.
	Value any `json:"value,omitempty"`

	// Enabled shows or hides the item.
	Enabled *bool `json:"enabled,omitempty"`

	// Opacity of the item, from 0 to 1 in steps of 0.1.
	Opacity *float64 `json:"opacity,omitempty"`

	// ZOrder is the stacking order of the item.
	ZOrder *int `json:"zOrder,omitempty"`

	// Background is the background of the item: a color, or a gradient.
	Background string `json:"background,omitempty"`

	// Text items.

	// Alignment of the text: "left", "center" or "right".
	Alignment string `json:"alignment,omitempty"`

	// Color of the text.
	Color string `json:"color,omitempty"`

	// Font of the text.
	Font *FeedbackFont `json:"font,omitempty"`

	// TextOverflow is how the text is shortened: "clip", "ellipsis" or "fade".
	TextOverflow string `json:"text-overflow,omitempty"`

	// Bar items.

	// BarBackgroundColor is the background color of the bar.
	BarBackgroundColor string `json:"bar_bg_c,omitempty"`

	// BarBorderColor is the border color of the bar.
	BarBorderColor string `json:"bar_border_c,omitempty"`

	// BarFillColor is the fill color of the bar.
	BarFillColor string `json:"bar_fill_c,omitempty"`

	// BarHeight is the height of the bar.
	BarHeight *int `json:"bar_h,omitempty"`

	// BorderWidth is the width of the border of the bar.
	BorderWidth *int `json:"border_w,omitempty"`

	// Range is the minimum and the maximum value of the bar.
	Range *FeedbackRange `json:"range,omitempty"`

	// Subtype is the shape of the bar, from 0 to 4.
	Subtype *int `json:"subtype,omitempty"`
}

// FeedbackFont describes the font of a text item.
type FeedbackFont struct {
	// Size of the font.
	Size int `json:"size,omitempty"`

	// Weight of the font.
	Weight int `json:"weight,omitempty"`
}

// FeedbackRange is the range of the values of a bar item.
type FeedbackRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// BoolPtr returns a pointer to b.
func BoolPtr(b bool) *bool { return new(b) }

// IntPtr returns a pointer to i.
func IntPtr(i int) *int { return new(i) }

// Float64Ptr returns a pointer to f.
func Float64Ptr(f float64) *float64 { return new(f) }

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
func (s *StreamDeck) SwitchProfileTo(context, device string, payload *SendEventSwitchProfilePayload) {
	event := &SendEvent{
		Event:   SwitchToProfile,
		Context: context,
		Device:  device,
	}

	// A nil pointer in the interface would be encoded as null, not omitted.
	if payload != nil {
		event.Payload = payload
	}

	s.send(event)
}
