package proto

import "fmt"

// HostToPlugin represents a message sent from the host to a plugin over the stream.
type HostToPlugin struct {
	PluginID string

	Hello    *HostHello
	Shutdown *HostShutdown
	Event    *EventEnvelope
}

func (m *HostToPlugin) Marshal() ([]byte, error) {
	if m == nil {
		return nil, fmt.Errorf("nil HostToPlugin")
	}
	buf := appendString(nil, 1, m.PluginID)
	if m.Hello != nil {
		var err error
		buf, err = appendMessage(buf, 10, m.Hello)
		if err != nil {
			return nil, err
		}
	}
	if m.Shutdown != nil {
		var err error
		buf, err = appendMessage(buf, 11, m.Shutdown)
		if err != nil {
			return nil, err
		}
	}
	if m.Event != nil {
		var err error
		buf, err = appendMessage(buf, 20, m.Event)
		if err != nil {
			return nil, err
		}
	}
	return buf, nil
}

type HostHello struct {
	APIVersion string
}

func (m *HostHello) Marshal() ([]byte, error) {
	return appendString(nil, 1, m.APIVersion), nil
}

type HostShutdown struct {
	Reason string
}

func (m *HostShutdown) Marshal() ([]byte, error) {
	return appendString(nil, 1, m.Reason), nil
}

type EventEnvelope struct {
	EventID string
	Type    string

	PlayerJoin *PlayerJoinEvent
	PlayerQuit *PlayerQuitEvent
	Chat       *ChatEvent
	Command    *CommandEvent
	BlockBreak *BlockBreakEvent
	WorldClose *WorldCloseEvent
}

func (m *EventEnvelope) Marshal() ([]byte, error) {
	buf := appendString(nil, 1, m.EventID)
	buf = appendString(buf, 2, m.Type)
	var err error
	if m.PlayerJoin != nil {
		buf, err = appendMessage(buf, 10, m.PlayerJoin)
		if err != nil {
			return nil, err
		}
	}
	if m.PlayerQuit != nil {
		buf, err = appendMessage(buf, 11, m.PlayerQuit)
		if err != nil {
			return nil, err
		}
	}
	if m.Chat != nil {
		buf, err = appendMessage(buf, 12, m.Chat)
		if err != nil {
			return nil, err
		}
	}
	if m.Command != nil {
		buf, err = appendMessage(buf, 13, m.Command)
		if err != nil {
			return nil, err
		}
	}
	if m.BlockBreak != nil {
		buf, err = appendMessage(buf, 14, m.BlockBreak)
		if err != nil {
			return nil, err
		}
	}
	if m.WorldClose != nil {
		buf, err = appendMessage(buf, 15, m.WorldClose)
		if err != nil {
			return nil, err
		}
	}
	return buf, nil
}

type PlayerJoinEvent struct {
	PlayerUUID string
	Name       string
}

func (m *PlayerJoinEvent) Marshal() ([]byte, error) {
	buf := appendString(nil, 1, m.PlayerUUID)
	buf = appendString(buf, 2, m.Name)
	return buf, nil
}

type PlayerQuitEvent struct {
	PlayerUUID string
	Name       string
}

func (m *PlayerQuitEvent) Marshal() ([]byte, error) {
	buf := appendString(nil, 1, m.PlayerUUID)
	buf = appendString(buf, 2, m.Name)
	return buf, nil
}

type ChatEvent struct {
	PlayerUUID string
	Name       string
	Message    string
}

func (m *ChatEvent) Marshal() ([]byte, error) {
	buf := appendString(nil, 1, m.PlayerUUID)
	buf = appendString(buf, 2, m.Name)
	buf = appendString(buf, 3, m.Message)
	return buf, nil
}

type CommandEvent struct {
	PlayerUUID string
	Name       string
	Raw        string
}

func (m *CommandEvent) Marshal() ([]byte, error) {
	buf := appendString(nil, 1, m.PlayerUUID)
	buf = appendString(buf, 2, m.Name)
	buf = appendString(buf, 3, m.Raw)
	return buf, nil
}

type BlockBreakEvent struct {
	PlayerUUID string
	Name       string
	World      string
	X          int32
	Y          int32
	Z          int32
}

func (m *BlockBreakEvent) Marshal() ([]byte, error) {
	buf := appendString(nil, 1, m.PlayerUUID)
	buf = appendString(buf, 2, m.Name)
	buf = appendString(buf, 3, m.World)
	buf = appendTag(buf, 4, wireVarint)
	buf = appendVarint(buf, uint64(uint32(m.X)))
	buf = appendTag(buf, 5, wireVarint)
	buf = appendVarint(buf, uint64(uint32(m.Y)))
	buf = appendTag(buf, 6, wireVarint)
	buf = appendVarint(buf, uint64(uint32(m.Z)))
	return buf, nil
}

type WorldCloseEvent struct{}

func (m *WorldCloseEvent) Marshal() ([]byte, error) { return nil, nil }

// PluginToHost and nested messages are decoded from bytes.
type PluginToHost struct {
	PluginID string

	Hello     *PluginHello
	Subscribe *EventSubscribe
	Actions   *ActionBatch
	Log       *LogMessage
}

func UnmarshalPluginToHost(data []byte) (*PluginToHost, error) {
	msg := &PluginToHost{}
	if err := msg.Unmarshal(data); err != nil {
		return nil, err
	}
	return msg, nil
}

func (m *PluginToHost) Unmarshal(data []byte) error {
	for len(data) > 0 {
		field, wire, n, err := readTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch field {
		case 1:
			if wire != wireLength {
				return fmt.Errorf("plugin_to_host: invalid wire type for plugin_id")
			}
			s, consumed, err := readString(data)
			if err != nil {
				return err
			}
			m.PluginID = s
			data = data[consumed:]
		case 10:
			if m.Hello == nil {
				m.Hello = &PluginHello{}
			}
			consumed, err := readMessage(data, m.Hello)
			if err != nil {
				return err
			}
			data = data[consumed:]
		case 11:
			if m.Subscribe == nil {
				m.Subscribe = &EventSubscribe{}
			}
			consumed, err := readMessage(data, m.Subscribe)
			if err != nil {
				return err
			}
			data = data[consumed:]
		case 20:
			if m.Actions == nil {
				m.Actions = &ActionBatch{}
			}
			consumed, err := readMessage(data, m.Actions)
			if err != nil {
				return err
			}
			data = data[consumed:]
		case 30:
			if m.Log == nil {
				m.Log = &LogMessage{}
			}
			consumed, err := readMessage(data, m.Log)
			if err != nil {
				return err
			}
			data = data[consumed:]
		default:
			consumed, err := skipField(data, wire)
			if err != nil {
				return err
			}
			data = data[consumed:]
		}
	}
	return nil
}

type PluginHello struct {
	Name       string
	Version    string
	APIVersion string
	Commands   []*CommandSpec
}

func (m *PluginHello) Unmarshal(data []byte) error {
	for len(data) > 0 {
		field, wire, n, err := readTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch field {
		case 1:
			if wire != wireLength {
				return fmt.Errorf("plugin_hello: invalid wire type for name")
			}
			s, consumed, err := readString(data)
			if err != nil {
				return err
			}
			m.Name = s
			data = data[consumed:]
		case 2:
			if wire != wireLength {
				return fmt.Errorf("plugin_hello: invalid wire type for version")
			}
			s, consumed, err := readString(data)
			if err != nil {
				return err
			}
			m.Version = s
			data = data[consumed:]
		case 3:
			if wire != wireLength {
				return fmt.Errorf("plugin_hello: invalid wire type for api_version")
			}
			s, consumed, err := readString(data)
			if err != nil {
				return err
			}
			m.APIVersion = s
			data = data[consumed:]
		case 4:
			spec := &CommandSpec{}
			consumed, err := readMessage(data, spec)
			if err != nil {
				return err
			}
			m.Commands = append(m.Commands, spec)
			data = data[consumed:]
		default:
			consumed, err := skipField(data, wire)
			if err != nil {
				return err
			}
			data = data[consumed:]
		}
	}
	return nil
}

type CommandSpec struct {
	Name        string
	Description string
}

func (m *CommandSpec) Marshal() ([]byte, error) {
	buf := appendString(nil, 1, m.Name)
	buf = appendString(buf, 2, m.Description)
	return buf, nil
}

func (m *CommandSpec) Unmarshal(data []byte) error {
	for len(data) > 0 {
		field, wire, n, err := readTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch field {
		case 1:
			if wire != wireLength {
				return fmt.Errorf("command_spec: invalid wire type for name")
			}
			s, consumed, err := readString(data)
			if err != nil {
				return err
			}
			m.Name = s
			data = data[consumed:]
		case 2:
			if wire != wireLength {
				return fmt.Errorf("command_spec: invalid wire type for description")
			}
			s, consumed, err := readString(data)
			if err != nil {
				return err
			}
			m.Description = s
			data = data[consumed:]
		default:
			consumed, err := skipField(data, wire)
			if err != nil {
				return err
			}
			data = data[consumed:]
		}
	}
	return nil
}

type EventSubscribe struct {
	Events []string
}

func (m *EventSubscribe) Marshal() ([]byte, error) {
	buf := make([]byte, 0)
	for _, e := range m.Events {
		buf = appendString(buf, 1, e)
	}
	return buf, nil
}

func (m *EventSubscribe) Unmarshal(data []byte) error {
	for len(data) > 0 {
		field, wire, n, err := readTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch field {
		case 1:
			if wire != wireLength {
				return fmt.Errorf("event_subscribe: invalid wire type for events")
			}
			s, consumed, err := readString(data)
			if err != nil {
				return err
			}
			m.Events = append(m.Events, s)
			data = data[consumed:]
		default:
			consumed, err := skipField(data, wire)
			if err != nil {
				return err
			}
			data = data[consumed:]
		}
	}
	return nil
}

type ActionBatch struct {
	Actions []*Action
}

func (m *ActionBatch) Unmarshal(data []byte) error {
	for len(data) > 0 {
		field, wire, n, err := readTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch field {
		case 1:
			act := &Action{}
			consumed, err := readMessage(data, act)
			if err != nil {
				return err
			}
			m.Actions = append(m.Actions, act)
			data = data[consumed:]
		default:
			consumed, err := skipField(data, wire)
			if err != nil {
				return err
			}
			data = data[consumed:]
		}
	}
	return nil
}

type Action struct {
	CorrelationID string

	SendChat *SendChatAction
	Teleport *TeleportAction
	Kick     *KickAction
}

func (m *Action) Unmarshal(data []byte) error {
	for len(data) > 0 {
		field, wire, n, err := readTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch field {
		case 1:
			if wire != wireLength {
				return fmt.Errorf("action: invalid wire type for correlation_id")
			}
			s, consumed, err := readString(data)
			if err != nil {
				return err
			}
			m.CorrelationID = s
			data = data[consumed:]
		case 10:
			if m.SendChat == nil {
				m.SendChat = &SendChatAction{}
			}
			consumed, err := readMessage(data, m.SendChat)
			if err != nil {
				return err
			}
			data = data[consumed:]
			m.Teleport, m.Kick = nil, nil
		case 11:
			if m.Teleport == nil {
				m.Teleport = &TeleportAction{}
			}
			consumed, err := readMessage(data, m.Teleport)
			if err != nil {
				return err
			}
			data = data[consumed:]
			m.SendChat, m.Kick = nil, nil
		case 12:
			if m.Kick == nil {
				m.Kick = &KickAction{}
			}
			consumed, err := readMessage(data, m.Kick)
			if err != nil {
				return err
			}
			data = data[consumed:]
			m.SendChat, m.Teleport = nil, nil
		default:
			consumed, err := skipField(data, wire)
			if err != nil {
				return err
			}
			data = data[consumed:]
		}
	}
	return nil
}

type SendChatAction struct {
	TargetUUID string
	Message    string
}

func (m *SendChatAction) Unmarshal(data []byte) error {
	for len(data) > 0 {
		field, wire, n, err := readTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch field {
		case 1:
			if wire != wireLength {
				return fmt.Errorf("send_chat: invalid wire type for target_uuid")
			}
			s, consumed, err := readString(data)
			if err != nil {
				return err
			}
			m.TargetUUID = s
			data = data[consumed:]
		case 2:
			if wire != wireLength {
				return fmt.Errorf("send_chat: invalid wire type for message")
			}
			s, consumed, err := readString(data)
			if err != nil {
				return err
			}
			m.Message = s
			data = data[consumed:]
		default:
			consumed, err := skipField(data, wire)
			if err != nil {
				return err
			}
			data = data[consumed:]
		}
	}
	return nil
}

type TeleportAction struct {
	PlayerUUID string
	X          float64
	Y          float64
	Z          float64
	Yaw        float32
	Pitch      float32
}

func (m *TeleportAction) Unmarshal(data []byte) error {
	for len(data) > 0 {
		field, wire, n, err := readTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch field {
		case 1:
			if wire != wireLength {
				return fmt.Errorf("teleport: invalid wire type for player_uuid")
			}
			s, consumed, err := readString(data)
			if err != nil {
				return err
			}
			m.PlayerUUID = s
			data = data[consumed:]
		case 2:
			if wire != wireFixed64 {
				return fmt.Errorf("teleport: invalid wire type for x")
			}
			v, consumed, err := readFloat64(data)
			if err != nil {
				return err
			}
			m.X = v
			data = data[consumed:]
		case 3:
			if wire != wireFixed64 {
				return fmt.Errorf("teleport: invalid wire type for y")
			}
			v, consumed, err := readFloat64(data)
			if err != nil {
				return err
			}
			m.Y = v
			data = data[consumed:]
		case 4:
			if wire != wireFixed64 {
				return fmt.Errorf("teleport: invalid wire type for z")
			}
			v, consumed, err := readFloat64(data)
			if err != nil {
				return err
			}
			m.Z = v
			data = data[consumed:]
		case 5:
			if wire != wireFixed32 {
				return fmt.Errorf("teleport: invalid wire type for yaw")
			}
			v, consumed, err := readFloat32(data)
			if err != nil {
				return err
			}
			m.Yaw = v
			data = data[consumed:]
		case 6:
			if wire != wireFixed32 {
				return fmt.Errorf("teleport: invalid wire type for pitch")
			}
			v, consumed, err := readFloat32(data)
			if err != nil {
				return err
			}
			m.Pitch = v
			data = data[consumed:]
		default:
			consumed, err := skipField(data, wire)
			if err != nil {
				return err
			}
			data = data[consumed:]
		}
	}
	return nil
}

type KickAction struct {
	PlayerUUID string
	Reason     string
}

func (m *KickAction) Unmarshal(data []byte) error {
	for len(data) > 0 {
		field, wire, n, err := readTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch field {
		case 1:
			if wire != wireLength {
				return fmt.Errorf("kick: invalid wire type for player_uuid")
			}
			s, consumed, err := readString(data)
			if err != nil {
				return err
			}
			m.PlayerUUID = s
			data = data[consumed:]
		case 2:
			if wire != wireLength {
				return fmt.Errorf("kick: invalid wire type for reason")
			}
			s, consumed, err := readString(data)
			if err != nil {
				return err
			}
			m.Reason = s
			data = data[consumed:]
		default:
			consumed, err := skipField(data, wire)
			if err != nil {
				return err
			}
			data = data[consumed:]
		}
	}
	return nil
}

type LogMessage struct {
	Level   string
	Message string
}

func (m *LogMessage) Unmarshal(data []byte) error {
	for len(data) > 0 {
		field, wire, n, err := readTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch field {
		case 1:
			if wire != wireLength {
				return fmt.Errorf("log: invalid wire type for level")
			}
			s, consumed, err := readString(data)
			if err != nil {
				return err
			}
			m.Level = s
			data = data[consumed:]
		case 2:
			if wire != wireLength {
				return fmt.Errorf("log: invalid wire type for message")
			}
			s, consumed, err := readString(data)
			if err != nil {
				return err
			}
			m.Message = s
			data = data[consumed:]
		default:
			consumed, err := skipField(data, wire)
			if err != nil {
				return err
			}
			data = data[consumed:]
		}
	}
	return nil
}
