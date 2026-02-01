package common

const BaseApiUrl = "https://discord.com/api/v" + ApiVersion
const ApiVersion = "10"
const ApiEncoding = "json"

var CommonSecrets = struct {
	Id       string // Client ID
	Secret   string // Client Secret
	BotToken string // Bot user token
}{}
