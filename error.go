package agologo

import "fmt"

type Error struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func (de *Error) Error() string {
	return fmt.Sprintf("%s (%d)", de.Message, de.StatusCode)
}
