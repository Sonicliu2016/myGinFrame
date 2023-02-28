package model

type NError struct {
	Code int
	Msg  string
}

func (e NError) Error() string {
	return e.Msg
}
