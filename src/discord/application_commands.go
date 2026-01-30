package discord

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
)

// CommandContext as specified by https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-interaction-context-types
type CommandContext int

const (
	CmdContextGuild CommandContext = iota
	CmdContextBotDm
	CmdContextPrivateChannel
)

// CommandType as specified by https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-types
type CommandType int

const (
	CmdTypeChatInput CommandType = iota + 1
	CmdTypeUser
	CmdTypeMessage
	CmdTypePrimaryEntryPoint
)

// CommandOptionType as specified by https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-option-type
type CommandOptionType int

const (
	CmdOptSubcommand CommandOptionType = iota + 1
	CmdOptSubcommandGroup
	CmdOptString
	CmdOptInt
	CmdOptBool
	CmdOptUser
	CmdOptChannel
	CmdOptRole
	CmdOptMentionable
	CmdOptFloat64
	CmdOptAttachment
)

var idToOptTypeName = map[CommandOptionType]string{
	CmdOptSubcommand:      "Subcommand",
	CmdOptSubcommandGroup: "Subcommand Group",
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

type CommandHandler = func(params CommandParams) error

// ApplicationCommand represents https://discord.com/developers/docs/interactions/application-commands#application-command-object
// as well as its responses https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-application-command-data-structure
type ApplicationCommand struct {
	Name          string          `json:"name"`
	Type          CommandType     `json:"type"`
	Id            Snowflake       `json:"id,omitempty"`
	ApplicationId Snowflake       `json:"application_id,omitempty"`
	GuildId       *Snowflake      `json:"guild_id,omitempty"`
	Description   string          `json:"description,omitempty"` // 1-100 characters, leave empty if not CHAT_INPUT
	Options       []CommandOption `json:"options,omitempty"`     // Optional, max 25 length. Do not access this directly, use the helpers instead

	Permissions int64            `json:"default_member_permissions,string,omitempty"` // Nullable (bit set). Annoyingly, discord sends this as a string.
	Nsfw        bool             `json:"nsfw,omitempty"`                              // Optional, default false
	Contexts    []CommandContext `json:"contexts,omitempty"`
	Version     Snowflake        `json:"version,omitempty"`
	Handler     CommandHandler   `json:"-"` // If true, the command will be consumed by this handler and not passed to others
}

// Dispatch attempts to execute the command given an input ApplicationCommandData from discord. This will only be called if the names match.
func (c *ApplicationCommand) Dispatch(guild Snowflake, interactionId Snowflake, interactionToken string, data ApplicationCommandData) error {
	params := CommandParams{
		GuildId:          guild,
		InteractionId:    interactionId,
		InteractionToken: interactionToken,
		Options:          nil,
		Resolved:         data.ResolvedData,
	}

	if c.Handler != nil {
		slog.Info("[Command] Dispatching application command: " + c.Name)

		params.Options = &data.Options
		if err := c.Handler(params); err != nil {
			return err
		}
		return nil
	}
	// If handler is nil, assume subcommands or subcommand groups are present
	var subcommand *CommandOption
	var err error

	for _, option := range data.Options { // Linear search for the subcommand in question
		if option.Type == CmdOptSubcommand {
			if subcommand, err = c.GetSubcommand(option.Name); err != nil {
				return err
			}
			params.Options = &option.Options
			break
		} else if option.Type == CmdOptSubcommandGroup {
			group, err := c.GetSubcommandGroup(option.Name)
			if err != nil {
				return err
			} else if group == nil {
				return fmt.Errorf("subcommand group %s does not exist", option.Name)
			}

			for _, suboption := range option.Options {
				if subcommand, err = group.GetSubcommand(suboption.Name); err != nil {
					return err
				}
				params.Options = &suboption.Options
			}

			break
		}
	}

	if subcommand == nil {
		return errors.New("subcommand does not exist")
	} else if subcommand.Handler == nil {
		return fmt.Errorf("subcommand %s does not have a handler", subcommand.Name)
	}

	return subcommand.Handler(params)
}

// GetSubcommand returns the CommandOption matching name or throws an error if CommandOption.Type != CmdOptSubcommand
func (c *ApplicationCommand) GetSubcommand(name string) (*CommandOption, error) {
	return findOption(name, CmdOptSubcommand, c.Options)
}

// GetSubcommandGroup returns the CommandOption matching name or throws an error if CommandOption.Type != CmdOptSubcommandGroup
func (c *ApplicationCommand) GetSubcommandGroup(name string) (*CommandOption, error) {
	return findOption(name, CmdOptSubcommandGroup, c.Options)
}

// CommandOption represents https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-option-structure
// These are exposed by methods on OptionType and should not be created directly as generics can't express them properly
type CommandOption struct {
	Type        CommandOptionType     `json:"type"` // https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-option-type
	Name        string                `json:"name"` // 1-32 characters
	Description string                `json:"description"`
	Required    bool                  `json:"required,omitempty"`   // Optional, default false
	Choices     []CommandOptionChoice `json:"choices,omitempty"`    // Optional, 25 max
	MinValue    float64               `json:"min_value,omitempty"`  // MUST match with the correct CommandOptionType
	MaxValue    float64               `json:"max_value,omitempty"`  // MUST match with the correct CommandOptionType
	MinLength   int                   `json:"min_length,omitempty"` // Only applicable for CmdOptString
	MaxLength   int                   `json:"max_length,omitempty"` // Only applicable for CmdOptString
	Options     []CommandOption       `json:"options,omitempty"`    // Only applicable for CmdOptSubcommand and CmdOptSubcommandGroup
	Handler     CommandHandler        `json:"-"`                    // Only applicable for CmdOptSubcommand
}

// GetSubcommand returns the CommandOption matching name or throws an error if CommandOption.Type != CmdOptSubcommand
func (c *CommandOption) GetSubcommand(name string) (*CommandOption, error) {
	return findOption(name, CmdOptSubcommand, c.Options)
}

func findOption(name string, optType CommandOptionType, options []CommandOption) (*CommandOption, error) {
	for _, opt := range options {
		if opt.Name == name {
			if opt.Type != optType {
				return nil, fmt.Errorf("expected command option of type %s but got %s", idToOptTypeName[optType], idToOptTypeName[opt.Type])
			}
			return &opt, nil
		}
	}
	return nil, nil
}

// CommandOptionChoice represents https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-option-choice-structure
type CommandOptionChoice struct {
	Name  string      `json:"name"`  // 1-100 characters
	Value interface{} `json:"value"` // This may be various types, matching the parent CommandOption.
}

type CommandParams struct {
	GuildId          Snowflake
	InteractionId    Snowflake
	InteractionToken string
	Options          *[]CommandOptionData
	Resolved         *ResolvedData
}

// GetOption iterates over all child options and returns the first one with a matching name. If no option is found,
// returns nil.
func (p CommandParams) GetOption(name string) *CommandOptionData {
	for _, option := range *p.Options {
		if option.Name == name {
			return &option
		}
	}
	return nil
}

// ApplicationCommandData represents an application command response sent by discord
// https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-application-command-data-structure
type ApplicationCommandData struct {
	Name         string              `json:"name"`
	Type         CommandType         `json:"type"`
	Id           Snowflake           `json:"id"`
	Options      []CommandOptionData `json:"options"`
	TargetId     *Snowflake          `json:"target_id"` // Only used for user & message commands
	GuildId      *Snowflake          `json:"guild_id"`
	ResolvedData *ResolvedData       `json:"resolved"`
}

// CommandOptionData represents https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-application-command-interaction-data-option-structure
type CommandOptionData struct {
	Name    string              `json:"name"`
	Type    CommandOptionType   `json:"type"`
	Options []CommandOptionData `json:"options"`
	Value   interface{}         `json:"-"` // Value gets decoded according to Type. do not read directly
}

func (o *CommandOptionData) UnmarshalJSON(data []byte) error {
	var p struct { // Sadly can't compose this, or it'll recursively unmarshal itself
		Name     string              `json:"name"`
		Type     CommandOptionType   `json:"type"`
		Options  []CommandOptionData `json:"options"`
		RawValue json.RawMessage     `json:"value"`
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}

	o.Name = p.Name
	o.Type = p.Type
	o.Options = p.Options

	if p.RawValue == nil {
		return nil
	}

	if o.Type == 4 { // Int needs special handling because unmarshal defaults to a float64 & the interface cast would break
		var i int
		if err := json.Unmarshal(p.RawValue, &i); err != nil {
			return err
		}
		o.Value = i
	} else if o.Type == 11 {
		var a Attachment
		if err := json.Unmarshal(p.RawValue, &a); err != nil {
			return err
		}
		o.Value = a
	} else if o.Type == 6 || o.Type == 7 || o.Type == 8 || o.Type == 9 {
		var s Snowflake
		if err := json.Unmarshal(p.RawValue, &s); err != nil {
			return err
		}
		o.Value = s
	} else if err := json.Unmarshal(p.RawValue, &o.Value); err != nil {
		return err // Default unmarshal behaviour: Bool -> bool, Num -> float64, String -> string (Snowflake), Null -> nil
	}

	return nil
}

func (o *CommandOptionData) AsString() string {
	o.assertType(CmdOptString)
	return o.Value.(string)
}

func (o *CommandOptionData) AsInt() int {
	o.assertType(CmdOptInt)
	return o.Value.(int)
}

func (o *CommandOptionData) AsBool() bool {
	o.assertType(CmdOptBool)
	return o.Value.(bool)
}

func (o *CommandOptionData) AsSnowflake() Snowflake {
	o.assertSnowflake()
	return o.Value.(Snowflake)
}

func (o *CommandOptionData) AsFloat64() float64 {
	o.assertType(CmdOptFloat64)
	return o.Value.(float64)
}

func (o *CommandOptionData) AsAttachment() *Attachment {
	o.assertType(CmdOptAttachment)
	return o.Value.(*Attachment)
}

func (o *CommandOptionData) assertType(expected CommandOptionType) {
	if o.Type != expected {
		panic(fmt.Errorf("expected option of type %s but got %s", idToOptTypeName[expected], idToOptTypeName[o.Type]))
	}
}

func (o *CommandOptionData) assertSnowflake() {
	if o.Type != CmdOptUser && o.Type != CmdOptChannel && o.Type != CmdOptRole && o.Type != CmdOptMentionable {
		panic(errors.New("expected option of type snowflake but got " + idToOptTypeName[o.Type]))
	}
}
