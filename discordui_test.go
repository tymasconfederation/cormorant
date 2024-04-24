package cormorant

import (
	"errors"
	"testing"
)

func Test_extractErrorMessage(t *testing.T) {
	tests := []struct {
		name    string
		arg     error
		wantRet string
	}{
		// test cases
		{"normal error", errors.New("just a standard error"), "just a standard error"},
		{"unknown role error", errors.New(`HTTP 404 Not Found, {"message": "Unknown Role", "code": 10011}`), "Unknown Role"},
		{"bad request error", errors.New(`HTTP 400 Bad Request, {"message": "Cannot send an empty message", "code": 50006}`), "Cannot send an empty message"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRet := extractErrorMessage(tt.arg); gotRet != tt.wantRet {
				t.Errorf("extractErrorMessage() = (%v), want (%v)", gotRet, tt.wantRet)
			}
		})
	}
}
