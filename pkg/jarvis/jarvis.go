package jarvis

type Jarvis interface {
	Ask(msg string) (reply string, err error)
	AskStream(msg string) error
}
