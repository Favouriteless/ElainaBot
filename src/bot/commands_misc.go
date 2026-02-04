package main

import (
	. "elaina-common"
	"log/slog"
)

var echoCommand = ApplicationCommand{
	Name:        "echo",
	Type:        CmdTypeChatInput,
	Description: "Repeats what you said back to you",
	Handler:     echoHandler,
	Options: []CommandOption{
		{
			Name:        "string",
			Description: "Echo... echo... echo...",
			Type:        CmdOptString,
			Required:    true,
		},
	},
}

var macroCommand = ApplicationCommand{
	Name:        "macro",
	Description: "Macros are a handy way to save a message and fetch it using a keyword",
	Type:        CmdTypeChatInput,
	Contexts:    []CommandContext{CmdContextGuild},
	Handler:     macroUseHandler,
	Options: []CommandOption{
		{
			Name:        "keyword",
			Description: "Keyword used to trigger the macro",
			Type:        CmdOptString,
			Required:    true,
		},
	},
}

var editMacroCommand = ApplicationCommand{
	Name:        "editmacro",
	Description: "Set or delete a macro",
	Type:        CmdTypeChatInput,
	Permissions: PermAdministrator,
	Contexts:    []CommandContext{CmdContextGuild},
	Options: []CommandOption{
		{
			Name:        "set",
			Description: "Set a macro",
			Type:        CmdOptSubcommand,
			Handler:     macroSetHandler,
			Options: []CommandOption{
				{
					Name:        "keyword",
					Description: "Keyword used to trigger the macro",
					Type:        CmdOptString,
					Required:    true,
				},
				{
					Name:        "response",
					Description: "The text Elaina will respond with",
					Type:        CmdOptString,
					MinLength:   1,
					MaxLength:   280,
					Required:    true,
				},
			},
		},
		{
			Name:        "delete",
			Description: "Delete a macro",
			Type:        CmdOptSubcommand,
			Handler:     macroDeleteHandler,
			Options: []CommandOption{
				{
					Name:        "keyword",
					Description: "Keyword used to trigger the macro",
					Type:        CmdOptString,
					Required:    true,
				},
			},
		},
	},
}

func echoHandler(params CommandParams) error {
	echo := params.GetOption("string").AsString()
	return SendInteractionMessageResponse(Message{Content: echo, Flags: MsgFlagEphemeral}, params.InteractionId, params.InteractionToken)
}

func macroSetHandler(params CommandParams) (err error) {
	macro := Macro{
		Guild:    params.GuildId,
		Key:      params.GetOption("keyword").AsString(),
		Response: params.GetOption("response").AsString(),
	}

	if err = CreateOrUpdateMacro(macro); err != nil {
		return err
	}
	slog.Info("Macro set:", slog.String("key", macro.Key), slog.String("response", macro.Response))

	return SendInteractionMessageResponse(Message{Content: "Macro set!", Flags: MsgFlagEphemeral}, params.InteractionId, params.InteractionToken)
}

func macroDeleteHandler(params CommandParams) error {
	key := params.GetOption("keyword").AsString()

	var response string
	if deleted, err := DeleteMacro(params.GuildId, key); err != nil {
		return err
	} else if deleted {
		response = "Macro deleted"
		slog.Info("[Elaina] Macro deleted: \"" + key + "\"")
	} else {
		response = "No macro found for \"" + key + "\""
	}

	return SendInteractionMessageResponse(Message{Content: response, Flags: MsgFlagEphemeral}, params.InteractionId, params.InteractionToken)
}

func macroUseHandler(params CommandParams) error {
	key := params.GetOption("keyword").AsString()
	macro, err := GetMacro(params.GuildId, key)
	if err != nil {
		return err
	}

	var response Message
	if macro != nil {
		response = Message{Content: macro.Response}
	} else {
		response = Message{Content: "No macro found for \"" + key + "\"", Flags: MsgFlagEphemeral}
	}

	return SendInteractionMessageResponse(response, params.InteractionId, params.InteractionToken)
}
