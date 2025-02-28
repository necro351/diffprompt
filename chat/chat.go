package chat

type Completer interface {
	// Send a message to the chat backend
	Complete(message string) (string, error)
}
