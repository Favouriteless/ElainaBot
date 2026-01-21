package elaina

// Macro represents a text macro where a given trigger string sends a response message in the chat.
type Macro struct {
	Key      string `json:"key"`
	Response string `json:"response"`
}
