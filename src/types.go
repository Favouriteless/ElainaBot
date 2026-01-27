package main

import "ElainaBot/discord"

// Macro represents a text macro where a given trigger string sends a response message in the chat.
type Macro struct {
	Guild    discord.Snowflake `json:"guild"`
	Key      string            `json:"key"`
	Response string            `json:"response"`
}

// Ban represents a banned user-- this is not a discord.GuildBan
type Ban struct {
	Guild   discord.Snowflake `json:"guild"`
	User    discord.Snowflake `json:"user"`
	Expires int64             `json:"expires"` // Unix timestamp
	Reason  string            `json:"reason"`
}
