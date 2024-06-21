package gen

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Meta    map[string]string `json:"meta"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s, %s", e.Code, e.Message)
}

func CheckResp(resp *http.Response) error {
	if resp.StatusCode >= 500 {
		return &Error{
			Code:    "UNKNOWN",
			Message: fmt.Sprintf("failed to get info (code: %d)", resp.StatusCode),
		}
	} else if resp.StatusCode >= 400 {
		var errResp Error
		err := json.NewDecoder(resp.Body).Decode(&errResp)
		if err != nil {
			return &Error{
				Code:    "UNKNOWN",
				Message: "failed to decode error response: " + err.Error(),
			}
		}
		return &errResp
	}
	return nil
}
