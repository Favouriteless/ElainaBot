package discord

import (
	"encoding/json"
	"fmt"
)

// Command context as specified by https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-interaction-context-types
const (
	CmdContextGuild          = 0
	CmdContextBotDm          = 1
	CmdContextPrivateChannel = 2
)

// Command type as specified by https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-types
const (
	CmdTypeChatInput         = 1
	CmdTypeUserInput         = 2
	CmdTypeMessage           = 3
	CmdTypePrimaryEntryPoint = 4
)

// Command option type as specified by https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-option-type
const (
	CmdOptSubCommand      = 1
	CmdOptSubCommandGroup = 2
	CmdOptString          = 3
	CmdOptInt             = 4
	CmdOptBool            = 5
	CmdOptUser            = 6
	CmdOptChannel         = 7
	CmdOptRole            = 8
	CmdOptMentionable     = 9
	CmdOptFloat64         = 10
	CmdOptAttachment      = 11
)

var idToOptTypeName = map[int]string{
	CmdOptSubCommand:      "Subcommand",
	CmdOptSubCommandGroup: "Subcommand Group",
	CmdOptString:          "String",
	CmdOptInt:             "Integer",
	CmdOptBool:            "Boolean",
	CmdOptUser:            "User",
	CmdOptChannel:         "Channel",
	CmdOptRole:            "Role",
	CmdOptMentionable:     "Mentionable",
	CmdOptFloat64:         "Float64",
	CmdOptAttachment:      "Attachment",
}

var Commands = make([]*ApplicationCommand, 1)

func AddCommand(command *ApplicationCommand) {
	Commands = append(Commands, command)
}

type CommandHandler = func(data ApplicationCommandData, id Snowflake, token string) error

// ApplicationCommand represents https://discord.com/developers/docs/interactions/application-commands#application-command-object
// as well as its responses https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-application-command-data-structure
type ApplicationCommand struct {
	Name          string          `json:"name"`
	Type          int             `json:"type"` // 1 = CHAT_INPUT, 2 = USER, 3 = MESSAGE, 4 = PRIMARY_ENTRY_POINT.
	Id            Snowflake       `json:"id,omitempty"`
	ApplicationId Snowflake       `json:"application_id,omitempty"`
	GuildId       *Snowflake      `json:"guild_id,omitempty"`
	Description   string          `json:"description,omitempty"` // 1-100 characters, leave empty if not CHAT_INPUT
	Options       []CommandOption `json:"options,omitempty"`     // Optional, max 25 length. Do not access this directly, use the helpers instead

	Permissions int64          `json:"default_member_permissions,string,omitempty"` // Nullable (bit set). Annoyingly, discord sends this as a string.
	Nsfw        bool           `json:"nsfw,omitempty"`                              // Optional, default false
	Contexts    []int          `json:"contexts,omitempty"`                          // 0 = GUILD, 1 = BOT_DM, 2 = PRIVATE_CHANNEL
	Version     Snowflake      `json:"version,omitempty"`
	Handler     CommandHandler `json:"-"` // If true, the command will be consumed by this handler and not passed to others
}

// CommandOption represents https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-option-structure
// These are exposed by methods on OptionType and should not be created directly as generics can't express them properly
type CommandOption struct {
	Type        int                   `json:"type"` // https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-option-type
	Name        string                `json:"name"` // 1-32 characters
	Description string                `json:"description"`
	Required    bool                  `json:"required"`          // Optional, default false
	Choices     []CommandOptionChoice `json:"choices,omitempty"` // Optional, 25 max
	// Options     []commandOption           `json:"options,omitempty"` // TODO: Add support for subcommands
}

// CommandOptionChoice represents https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-option-choice-structure
type CommandOptionChoice struct {
	Name  string      `json:"name"`  // 1-100 characters
	Value interface{} `json:"value"` // This may be various types, matching the parent CommandOption.
}

// ApplicationCommandData represents an application command response sent by discord
// https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-application-command-data-structure
type ApplicationCommandData struct {
	Name         string              `json:"name"`
	Type         int                 `json:"type"` // 1 = CHAT_INPUT, 2 = USER, 3 = MESSAGE, 4 = PRIMARY_ENTRY_POINT.
	Id           Snowflake           `json:"id"`
	Options      []CommandOptionData `json:"options"`
	TargetId     *Snowflake          `json:"target_id"` // Only used for user & message commands
	GuildId      *Snowflake          `json:"guild_id"`
	ResolvedData *ResolvedData       `json:"resolved"`
}

// OptionByName iterates over all child options and returns the first one with a matching name. If no option is found,
// returns nil.
func (d ApplicationCommandData) OptionByName(name string) *CommandOptionData {
	for _, option := range d.Options {
		if option.Name == name {
			return &option
		}
	}
	return nil
}

// CommandOptionData represents https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-application-command-interaction-data-option-structure
type CommandOptionData struct {
	Name  string      `json:"name"`
	Type  int         `json:"type"`
	Value interface{} `json:"-"` // Gets decoded according to type.. do not read directly
	// Options []CommandOption `json:"options"` // TODO: Add support for groups & subcommands
}

func (o *CommandOptionData) UnmarshalJSON(data []byte) error {
	var partial struct {
		Name  string `json:"name"`
		Type  int    `json:"type"`
		Value json.RawMessage
	}
	if err := json.Unmarshal(data, &partial); err != nil {
		return err
	}
	o.Name = partial.Name
	o.Type = partial.Type

	switch partial.Type {
	case 4: // Int needs special handling because unmarshal defaults to a float64 & the interface cast would break
		var i int
		if err := json.Unmarshal(partial.Value, &i); err != nil {
			return err
		}
		o.Value = i
	case 11:
		var a Attachment
		if err := json.Unmarshal(partial.Value, &a); err != nil {
			return err
		}
		o.Value = a
	default:
		// Default unmarshal behaviour: Bool -> bool, Num -> float64, String -> string (Snowflake), Null -> nil
		if err := json.Unmarshal(partial.Value, &o.Value); err != nil {
			return err
		}
	}
	return nil
}

func (o *CommandOptionData) AsString() (string, error) {
	if err := o.assertType(CmdOptString); err != nil {
		return "", err
	}
	return o.Value.(string), nil
}

func (o *CommandOptionData) AsInt() (int, error) {
	if err := o.assertType(CmdOptInt); err != nil {
		return 0, err
	}
	return o.Value.(int), nil
}

func (o *CommandOptionData) AsBool() (bool, error) {
	if err := o.assertType(CmdOptBool); err != nil {
		return false, err
	}
	return o.Value.(bool), nil
}

func (o *CommandOptionData) AsSnowflake() (Snowflake, error) {
	if err := o.assertType(CmdOptUser); err != nil {
		return 0, err
	}
	return o.Value.(Snowflake), nil
}

func (o *CommandOptionData) AsFloat64() (float64, error) {
	if err := o.assertType(CmdOptFloat64); err != nil {
		return 0, err
	}
	return o.Value.(float64), nil
}

func (o *CommandOptionData) AsAttachment() (*Attachment, error) {
	if err := o.assertType(CmdOptAttachment); err != nil {
		return nil, err
	}
	return o.Value.(*Attachment), nil
}

func (o *CommandOptionData) assertType(expected int) error {
	if o.Type != expected {
		return fmt.Errorf("expected option of type %s but got %s", idToOptTypeName[expected], idToOptTypeName[o.Type])
	}
	return nil
}
