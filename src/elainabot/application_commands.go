package main

import (
	. "elaina-common"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"
)

// Commands contains all the bot commands Elaina is currently using
var Commands CommandCollection

type CommandCollection []*ApplicationCommand

func (c CommandCollection) GetCommand(name string) *ApplicationCommand {
	for _, cmd := range c {
		if cmd.Name == name {
			return cmd
		}
	}
	return nil
}

// dispatchCommand attempts to execute the command given an input ApplicationCommandData from discord. The data should
// be verified to be of the correct type of command prior to calling dispatchCommand
func dispatchCommand(c *ApplicationCommand, guild Snowflake, interactionId Snowflake, interactionToken string, data ApplicationCommandData) error {
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

func DeployCommands(commands CommandCollection) {
	slog.Info("Deploying application commands...")
	for _, com := range commands {
		func() {
			resp, err := CreateOrUpdateCommand(com)
			if err != nil {
				slog.Error("Error registering command: ", slog.String("command", com.Name), slog.String("error", err.Error()))
				return
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case 200:
				slog.Info("Command updated successfully: " + com.Name)
			case 201:
				slog.Info("Command added successfully: " + com.Name)
			default:
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					panic(err) // Should never be hit
				}
				slog.Error("Command could not be created: ", slog.String("command", com.Name), slog.String("status_code", resp.Status), slog.String("body", string(body)))
			}
		}()
		time.Sleep(1 * time.Second) // This looks really stupid, but it's to avoid rate limiting
	}
}
