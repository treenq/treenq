package handlers

type Error struct {
	Code string            `json:"code"`
	Meta map[string]string `json:"meta"`
}

func (e *Error) Error() string {
	return e.Code
}
