package common

import (
	"encoding/json"
	"strconv"
)

// Macro represents a text macro where a given trigger string sends a response message in the chat.
type Macro struct {
	Guild    Snowflake `json:"guild"`
	Key      string    `json:"key"`
	Response string    `json:"response"`
}

// GuildSettings represents the config of a single guild.
type GuildSettings struct {
	HoneypotChannel *Snowflake `json:"honeypot_channel"`
	HelloEnabled    bool       `json:"hello_enabled"`
}

func DefaultGuildSettings() GuildSettings {
	return GuildSettings{
		HoneypotChannel: nil,
		HelloEnabled:    false,
	}
}

// Nullable represents a serializable primitive which can also be Null. For example, representing a Null string
type Nullable[T any] struct {
	Value T
	Null  bool
}

func (n *Nullable[T]) MarshalJSON() ([]byte, error) {
	if n.Null {
		return json.Marshal(nil)
	}
	return json.Marshal(n.Value)
}

// StringInt64 is used to cleanly represent the 64-bit ints sent by discord API as they are serialized as strings
type StringInt64 uint64

func (s *StringInt64) String() string {
	return strconv.FormatUint(uint64(*s), 10)
}

func (s *StringInt64) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	id, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return err
	}
	*s = StringInt64(id)
	return nil
}

func (s *StringInt64) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
