package discord

const apiUrl = "https://discord.com/api/v" + apiVersion
const apiVersion = "10"
const apiEncoding = "json"

var application *Application

type Application struct {
	name   string // Name of the discord bot
	id     string // Client ID
	secret string // Client Secret
	token  string // Bot user token
}

func Initialize(name string, clientId string, clientSecret string, token string) {
	application = &Application{
		name:   name,
		id:     clientId,
		secret: clientSecret,
		token:  token,
	}
}
