package sdk

// PayloadOption customizes the payload of SetTitle and SetImage.
type PayloadOption func(*SendEventPayload)

// ForState restricts the change to the given 0-based state of an action supporting multiple states.
// When omitted, the change applies to all states.
func ForState(state uint8) PayloadOption {
	return func(p *SendEventPayload) {
		p.State = &state
	}
}

// ForTarget specifies where the change is displayed (hardware, software, or both).
func ForTarget(target Target) PayloadOption {
	return func(p *SendEventPayload) {
		p.Target = target
	}
}

// Alert sends an alert on the StreamDeck of action with given context.
func (s *StreamDeck) Alert(context string) {
	s.send(&SendEvent{Event: ShowAlert, Context: context})
}

// OpenURL tell the Stream Deck application to open an URL in the default browser.
func (s *StreamDeck) OpenURL(u string) {
	s.send(&SendEvent{Event: OpenURL, Payload: &SendEventPayload{URL: u}})
}

// SetTitle tell StreamDeck to dynamically change the title of action with given context.
// The title is sent as is: an empty title hides the title, see ResetTitle to restore the title set by the user.
// Use ForState to change the title of a single state.
func (s *StreamDeck) SetTitle(context string, title string, target Target, opts ...PayloadOption) {
	s.setTitle(context, &title, target, opts...)
}

// ResetTitle tell StreamDeck to restore the title set by the user, of action with given context.
// Use ForState to reset the title of a single state.
func (s *StreamDeck) ResetTitle(context string, target Target, opts ...PayloadOption) {
	s.setTitle(context, nil, target, opts...)
}

func (s *StreamDeck) setTitle(context string, title *string, target Target, opts ...PayloadOption) {
	payload := &SendEventPayload{Title: title, Target: target}
	for _, opt := range opts {
		opt(payload)
	}

	s.send(&SendEvent{
		Event:   SetTitle,
		Context: context,
		Payload: payload,
	})
}

// ShowOK temporarily show an OK checkmark icon of action with given context.
func (s *StreamDeck) ShowOK(context string) {
	s.send(&SendEvent{Event: ShowOk, Context: context})
}

// SetState change the state of an action supporting multiple states.
func (s *StreamDeck) SetState(context string, state uint8) {
	s.send(&SendEvent{
		Event:   SetState,
		Context: context,
		Payload: &SendEventPayload{State: &state},
	})
}

// SetImage change the image of an action with given context.
// An empty image resets the image to the default one from the manifest.
// Use ForState and ForTarget to restrict the change to a state or a target.
func (s *StreamDeck) SetImage(context string, image string, opts ...PayloadOption) {
	payload := &SendEventPayload{Image: image}
	for _, opt := range opts {
		opt(payload)
	}

	s.send(&SendEvent{
		Event:   SetImage,
		Context: context,
		Payload: payload,
	})
}

// SetTriggerDescription change the trigger description of an action with given context.
func (s *StreamDeck) SetTriggerDescription(context string, payload *SendEventSetTriggerDescriptionPayload) {
	if payload == nil {
		payload = &SendEventSetTriggerDescriptionPayload{}
	}

	s.send(&SendEvent{
		Event:   SetTriggerDescription,
		Context: context,
		Payload: payload,
	})
}

// SetFeedback change the feedback of an action with given context.
// The payload maps the key of a layout item to its new value: a string, a number, or a Bar, GBar, Pixmap or Text.
func (s *StreamDeck) SetFeedback(context string, payload *FeedbackPayload) {
	if payload == nil {
		payload = &FeedbackPayload{}
	}

	s.send(&SendEvent{
		Event:   SetFeedback,
		Context: context,
		Payload: payload,
	})
}

// SetFeedbackLayout change the feedback layout of an action with given context.
func (s *StreamDeck) SetFeedbackLayout(context string, layout string) {
	s.send(&SendEvent{
		Event:   SetFeedbackLayout,
		Context: context,
		Payload: &SendEventSetFeedbackLayoutPayload{Layout: layout},
	})
}

// SendToPropertyInspector sends a payload to the property inspector for an action instance.
// The payload can be any value that can be encoded to json.
func (s *StreamDeck) SendToPropertyInspector(context string, payload any) {
	s.send(&SendEvent{
		Event:   SendToPropertyInspector,
		Context: context,
		Payload: payload,
	})
}

// SetSettings change the settings of an action with given context.
func (s *StreamDeck) SetSettings(context string, settings map[string]any) {
	s.send(&SendEvent{
		Event:   SetSettings,
		Context: context,
		Payload: nonNilSettings(settings),
	})
}

// GetSettings get the settings of an action with given context.
// Settings will be sent back to the plugin as a ReceivedEvent with the event name DidReceiveSettings.
func (s *StreamDeck) GetSettings(context string) {
	s.GetSettingsWithID(context, "")
}

// GetSettingsWithID is like GetSettings, the given id is sent back in ReceivedEvent.ID of the response
// so that it can be told apart from settings changed in the property inspector (Stream Deck 7.1+).
func (s *StreamDeck) GetSettingsWithID(context string, id string) {
	s.send(&SendEvent{Event: GetSettings, Context: context, ID: id})
}

// SetGlobalSettings change the global settings of the plugin.
//
// Stream Deck expects the UUID of the plugin as the context: an empty context means StreamDeck.UUID.
func (s *StreamDeck) SetGlobalSettings(context string, settings map[string]any) {
	s.send(&SendEvent{
		Event:   SetGlobalSettings,
		Context: s.pluginContext(context),
		Payload: nonNilSettings(settings),
	})
}

// GetGlobalSettings get the global settings of the plugin.
//
// Stream Deck expects the UUID of the plugin as the context: an empty context means StreamDeck.UUID.
// Global settings will be sent back to the plugin as a ReceivedEvent with the event name DidReceiveGlobalSettings.
func (s *StreamDeck) GetGlobalSettings(context string) {
	s.GetGlobalSettingsWithID(context, "")
}

// GetGlobalSettingsWithID is like GetGlobalSettings, the given id is sent back in ReceivedEvent.ID of the response
// so that it can be told apart from settings changed in the property inspector (Stream Deck 7.1+).
func (s *StreamDeck) GetGlobalSettingsWithID(context string, id string) {
	s.send(&SendEvent{Event: GetGlobalSettings, Context: s.pluginContext(context), ID: id})
}

// SetResources set the resources of an action with given context (Stream Deck 7.1+).
// Resources map a name to a file path, they are embedded when the action is exported.
func (s *StreamDeck) SetResources(context string, resources map[string]string) {
	if resources == nil {
		resources = map[string]string{}
	}

	s.send(&SendEvent{
		Event:   SetResources,
		Context: context,
		Payload: resources,
	})
}

// GetResources get the resources of an action with given context (Stream Deck 7.1+).
// Resources will be sent back to the plugin as a ReceivedEvent with the event name DidReceiveResources.
func (s *StreamDeck) GetResources(context string) {
	s.GetResourcesWithID(context, "")
}

// GetResourcesWithID is like GetResources, the given id is sent back in ReceivedEvent.ID of the response.
func (s *StreamDeck) GetResourcesWithID(context string, id string) {
	s.send(&SendEvent{Event: GetResources, Context: context, ID: id})
}

// SwitchProfile switches to a pre-defined profile distributed with the plugin on the specified device.
// The payload should contain "profile" (string) and optionally "page" (int, 0-indexed).
// When profile is omitted, Stream Deck switches back to the previous profile.
//
// Stream Deck expects the UUID of the plugin as the context: an empty context means StreamDeck.UUID.
//
// See SwitchProfileTo for a typed payload.
func (s *StreamDeck) SwitchProfile(context, device string, payload map[string]any) {
	event := &SendEvent{
		Event:   SwitchToProfile,
		Context: s.pluginContext(context),
		Device:  device,
	}

	// A nil map in the interface would be encoded as null, not omitted.
	if payload != nil {
		event.Payload = payload
	}

	s.send(event)
}

// nonNilSettings makes sure settings are encoded as a json object, not as null.
func nonNilSettings(settings map[string]any) map[string]any {
	if settings == nil {
		return map[string]any{}
	}

	return settings
}

// pluginContext returns the context for the commands that are not related to an action instance
// (global settings, profiles): Stream Deck expects the UUID of the plugin.
func (s *StreamDeck) pluginContext(context string) string {
	if context == "" {
		return s.UUID
	}

	return context
}
