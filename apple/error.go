package apple

type Error struct {
	Function  string
	Operation string
	Err       error
}

func (e Error) Error() string {
	return "apple." + e.Function + ": " + e.Operation + ": " + e.Err.Error()
}
