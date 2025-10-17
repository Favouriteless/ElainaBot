package discord

import "strconv"

// var OptSubCommand = unimplemented
// var OptSubCommandGroup = unimplemented

var OptString = OptionType[string]{3}
var OptInt = OptionType[int]{4}
var OptBool = OptionType[bool]{5}
var OptUser = OptionType[Snowflake]{6}
var OptChannel = OptionType[Snowflake]{7}
var OptRole = OptionType[Snowflake]{8}
var OptMentionable = OptionType[Snowflake]{9}
var OptFloat64 = OptionType[float64]{10}
var OptAttachment = OptionType[Attachment]{11}

const (
	ContextGuild          = 0
	ContextBotDm          = 1
	ContextPrivateChannel = 2
)

type OptionType[T any] struct {
	id int
}

func (o OptionType[T]) Create(name string, description string, required bool, choices ...commandOptionChoice[T]) *CommandOption {
	return &CommandOption{o.id, name, description, required, castOptionChoices(choices)}
}

// CreateGlobalSlashCommand returns a command builder for a global slash command
func CreateGlobalSlashCommand(name string, description string, contexts ...int) *ApplicationCommandBuilder {
	typeId := 1 // This is dumb, but you can't initialise an int pointer directly
	return &ApplicationCommandBuilder{
		&ApplicationCommand{
			Type:        &typeId,
			Name:        name,
			Description: description,
			Contexts:    contexts,
		},
	}
}

type ApplicationCommandBuilder struct {
	command *ApplicationCommand
}

func (builder *ApplicationCommandBuilder) Option(options *CommandOption) *ApplicationCommandBuilder {
	builder.command.Options = append(builder.command.Options, options)
	return builder
}

func (builder *ApplicationCommandBuilder) Nsfw() *ApplicationCommandBuilder {
	builder.command.Nsfw = true
	return builder
}

func (builder *ApplicationCommandBuilder) Permissions(permissions int) *ApplicationCommandBuilder {
	builder.command.Permissions = strconv.Itoa(permissions)
	return builder
}

func (builder *ApplicationCommandBuilder) Build() *ApplicationCommand {
	return builder.command
}

// ApplicationCommand represents https://discord.com/developers/docs/interactions/application-commands#application-command-object
// You should not directly be creating these, use CreateGlobalSlashCommand instead.
type ApplicationCommand struct {
	Id            Snowflake        `json:"id,omitempty"`
	Type          *int             `json:"type,omitempty"` // 1 = CHAT_INPUT, 2 = USER, 3 = MESSAGE, 4 = PRIMARY_ENTRY_POINT. Optional, default 1
	ApplicationId Snowflake        `json:"application_id,omitempty"`
	GuildId       *Snowflake       `json:"guild_id,omitempty"`
	Name          string           `json:"name"`
	Description   string           `json:"description,omitempty"`                // 1-100 characters, leave empty if not CHAT_INPUT
	Options       []*CommandOption `json:"options,omitempty"`                    // Optional, max 25 length. Do not access this directly, use the helpers instead
	Permissions   string           `json:"default_member_permissions,omitempty"` // Nullable (bit set). Annoyingly, discord sends this as a string.
	Nsfw          bool             `json:"nsfw,omitempty"`                       // Optional, default false
	Contexts      []int            `json:"contexts,omitempty"`                   // 0 = GUILD, 1 = BOT_DM, 2 = PRIVATE_CHANNEL
	Version       Snowflake        `json:"version,omitempty"`
	// Handler       *int             `json:"handler,omitempty"` // 1 = APP_HANDLER, 2 = DISCORD_LAUNCH_ACTIVITY, optional // Not supported yet
}

func (com *ApplicationCommand) AddOption(options *CommandOption) {
	com.Options = append(com.Options, options)
}

// CommandOption represents https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-option-structure
// These are exposed by methods on OptionType and should not be created directly as generics can't express them properly
type CommandOption struct {
	Type        int                        `json:"type"` // https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-option-type
	Name        string                     `json:"name"` // 1-32 characters
	Description string                     `json:"description"`
	Required    bool                       `json:"required"`          // Optional, default false
	Choices     []commandOptionChoice[any] `json:"choices,omitempty"` // Optional, 25 max
	// Options     []commandOption           `json:"options,omitempty"` // TODO: Add support for subcommands
}

// commandOptionChoice represents https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-option-choice-structure
type commandOptionChoice[T any] struct {
	Name  string `json:"name"` // 1-100 characters
	Value T      `json:"value"`
}

// castOptionChoices converts a typed []commandOptionChoice into a "wildcard" any version for flattening & marshaling.
func castOptionChoices[T any](choices []commandOptionChoice[T]) []commandOptionChoice[any] {
	if len(choices) == 0 {
		return nil
	} else {
		out := make([]commandOptionChoice[any], len(choices))
		for i, choice := range choices {
			out[i] = commandOptionChoice[any]{choice.Name, choice.Value}
		}
		return out
	}
}
