package spotify

type Error struct {
	Function  string
	Operation string
	Err       error
}

func (e Error) Error() string {
	return "spotify." + e.Function + ": " + e.Operation + ": " + e.Err.Error()
}
