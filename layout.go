package sdk

import (
	"encoding/json"
	"fmt"
)

// LayoutItemType is the type of an item of a Layout. It can't be changed at runtime.
type LayoutItemType string

const (
	// LayoutItemBar is a bar item.
	LayoutItemBar LayoutItemType = "bar"

	// LayoutItemGBar is a gradient bar item.
	LayoutItemGBar LayoutItemType = "gbar"

	// LayoutItemPixmap is an image item.
	LayoutItemPixmap LayoutItemType = "pixmap"

	// LayoutItemText is a text item.
	LayoutItemText LayoutItemType = "text"
)

// Rect is the position and the size of a layout item: x, y, width and height, in pixels.
type Rect [4]int

// Layout is a custom layout of the Stream Deck + touch display, to be written as a JSON file
// that the manifest refers to, and then used with SetFeedbackLayout.
// See [Layouts](https://docs.elgato.com/streamdeck/sdk/references/layouts).
type Layout struct {
	// ID is the unique identifier of the layout.
	ID string `json:"id"`

	// Controller is the controller the layout is for: Encoder or Neo.
	Controller Controller `json:"controller,omitempty"`

	// Items are the items of the layout.
	Items []LayoutItem `json:"items"`
}

// JSON returns the layout encoded as indented JSON, ready to be written to a layout file.
func (l *Layout) JSON() ([]byte, error) {
	return json.MarshalIndent(l, "", "  ")
}

// LayoutItem is an item of a Layout. Key, Type and Rect are required and can't be changed at runtime
// (see SetFeedback to change the other properties); the other properties are optional,
// and only the ones that apply to the Type are meaningful.
// The properties of Bar, GBar, Pixmap and Text are the same as the ones of the values of a FeedbackPayload.
type LayoutItem struct {
	// Key identifies the item, it is the key of the FeedbackPayload to change it.
	// An item with the key "title" is overridden by the title settings of the property inspector.
	Key string `json:"key"`

	// Type of the item.
	Type LayoutItemType `json:"type"`

	// Rect is the position and size of the item.
	Rect Rect `json:"rect"`

	// Common properties.

	Background string   `json:"background,omitempty"`
	Enabled    *bool    `json:"enabled,omitempty"`
	Opacity    *float64 `json:"opacity,omitempty"`

	// ZOrder is the stacking order of the item, from 0 to 700.
	ZOrder *int `json:"zOrder,omitempty"`

	// Value is the initial value of the item. Its type depends on Type:
	//
	//   - [LayoutItemBar], [LayoutItemGBar]: a number (float64 or int), within Range (0 to 100 by default).
	//   - [LayoutItemText]: a string, the text to display.
	//   - [LayoutItemPixmap]: a string, which is one of a path to an image file relative to the plugin,
	//     a base64 encoded image with its MIME type (data URI), or an SVG string.
	//
	// Any other type (bool, slice, ...) is not valid for Stream Deck. The compiler does not check it
	// because the field is an any, see [LayoutItem.Validate]. A nil Value is omitted.
	Value any `json:"value,omitempty"`

	// Bar and GBar properties.

	Range              *Range      `json:"range,omitempty"`
	Subtype            *BarSubType `json:"subtype,omitempty"`
	BarBackgroundColor string      `json:"bar_bg_c,omitempty"`
	BarBorderColor     string      `json:"bar_border_c,omitempty"`
	BarFillColor       string      `json:"bar_fill_c,omitempty"`
	BorderWidth        *int        `json:"border_w,omitempty"`

	// BarHeight is the height of the indicator of a GBar (bar_h).
	BarHeight *int `json:"bar_h,omitempty"`

	// Text properties.

	Alignment    Alignment    `json:"alignment,omitempty"`
	Color        string       `json:"color,omitempty"`
	Font         *Font        `json:"font,omitempty"`
	TextOverflow TextOverflow `json:"text-overflow,omitempty"`
}

// Validate checks the type of Value against Type, and returns an error when they do not match.
func (i *LayoutItem) Validate() error {
	switch i.Type {
	case LayoutItemBar, LayoutItemGBar:
		switch i.Value.(type) {
		case nil, float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return nil
		}
	case LayoutItemText, LayoutItemPixmap:
		switch i.Value.(type) {
		case nil, string:
			return nil
		}
	default:
		return fmt.Errorf("layout item [%s]: unknown type %q", i.Key, i.Type)
	}

	return fmt.Errorf("layout item [%s]: value of type %T is not valid for type %q", i.Key, i.Value, i.Type)
}
