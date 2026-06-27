package utils

import "testing"

func TestGetRESTPort(t *testing.T) {
	t.Run("Check PORT", func(t *testing.T) {
		portVal := 1000
		SetKey("REST_PORT", portVal)
		if GetRESTPort() != portVal {
			t.Error("Port ENV fetch wrong")
		}
	})
}

func TestGetXInternalToken(t *testing.T) {
	t.Run("GET Internal Token Value", func(t *testing.T) {
		internalToken := "Test"
		SetKey("X_INTERNAL_TOKEN", internalToken)
		if GetXInternalToken() != internalToken {
			t.Error("Port ENV fetch wrong")
		}
	})
}
