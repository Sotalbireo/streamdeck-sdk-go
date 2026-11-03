# sdk-go

Golang library to ease writing Stream Deck plugins with Go.

Derived from `streamdeck-sdk` (a third-party Go port) and updated to follow the
[Stream Deck SDK v3.0.0 WebSocket API](https://docs.elgato.com/streamdeck/sdk/references/websocket/plugin).

## Requirements

Go 1.27 or later. The only dependency is [`github.com/coder/websocket`](https://github.com/coder/websocket).

## Changes from streamdeck-sdk

Refactoring:

- Go 1.27 (`go.mod`), `interface{}` replaced by `any`.
- `gorilla/websocket` (last release: 2024-06) replaced by `coder/websocket` (maintained by Coder, no dependency,
  `context` support, safe concurrent writes). The read limit is raised from 32 KiB to 16 MiB so that big settings
  and resources are received.
- `StartWithContext(ctx)` added next to `Start()`; both close the connection with a closing handshake.
- A malformed event is reported to the error handler instead of stopping the plugin; so is an event that cannot be encoded.
- The connection and the registration have a 10 seconds timeout, a write has a 10 seconds timeout.
- `.golangci.yml` migrated to the golangci-lint v2 format (not run: golangci-lint is not installed here).

New:

- Events `deviceDidChange` and `didReceiveResources`; commands `getResources` / `setResources` (Stream Deck 7.1+).
- Request `id` for `getSettings` / `getGlobalSettings` / `getResources` (`Get*WithID`), and `ReceivedEvent.ID`.
- All `DeviceType` values (0-13), the `Neo` controller, `resources` and `url` in the received payload,
  `Info.Colors.ButtonMouseOverBackgroundColor`.
- `SetTitle` / `SetImage` accept the `ForState(n)` and `ForTarget(t)` options.
- `ReceivedEvent.RawPayload` and `DecodePayload` to read any payload (notably `sendToPlugin`).
- `FeedbackItem` (with `BoolPtr` / `IntPtr` / `Float64Ptr`) as a typed value of the `SetFeedback` payload,
  next to plain strings and numbers.
- `SwitchProfileTo` with a typed `SendEventSwitchProfilePayload` (`Profile`, `Page`); `SwitchProfile` still takes a map.

Fixed:

- `SetState(ctx, 0)` sent an empty payload (`state` is now a `*uint8`).
- A `sendToPlugin` payload that is not a json object stopped the plugin reader.
  `ReceivedEvent.Payload` is now nil for `sendToPlugin`; use `RawPayload` / `DecodePayload`.
- `ValueLayout` is `$A1` (was `$A0`, the canvas layout).
- The key controller value is `Keypad`, as sent by Stream Deck (`KeyPad` is kept as a deprecated alias).
- `Info.DevicePixelRatio` is a `float64`.
- A failing handler no longer sends `showAlert` for events without a context.
- `Start` stops on `SIGTERM`/`Ctrl+C` even while waiting for a message (`os.Kill` cannot be caught).
- `SetSettings`/`SetGlobalSettings`/`SendToPropertyInspector` accept `nil` / any json value.
- `SetFeedback` with a nil payload and `SwitchProfile` with a nil map no longer send `"payload": null`.

`SendToPropertyInspector` now takes an `interface{}` payload.

## Not covered

The manifest, the CLI, Property Inspector pages, localization and packaging are outside of this library.

## Test

```
go test ./...
```
