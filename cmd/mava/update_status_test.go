package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/model"
)

func TestTicketStatusIs(t *testing.T) {
	if !ticketStatusIs(&model.Ticket{Status: "Waiting"}, "Waiting") {
		t.Fatal("ticketStatusIs() = false, want true for matching status")
	}
	if ticketStatusIs(&model.Ticket{Status: "Open"}, "Resolved") {
		t.Fatal("ticketStatusIs() = true, want false for different status")
	}
	if ticketStatusIs(nil, "Resolved") {
		t.Fatal("ticketStatusIs(nil) = true, want false")
	}
}

func TestReadbackFallbackAllowed(t *testing.T) {
	timeoutErr := fmt.Errorf("wrapped: %w", api.ErrWebSocketAckTimeout)
	if !readbackFallbackAllowed(timeoutErr, true, false) {
		t.Fatal("readbackFallbackAllowed() = false, want true for timeout after a known previous mismatch")
	}
	if readbackFallbackAllowed(errors.New("auth failed"), true, false) {
		t.Fatal("readbackFallbackAllowed() = true, want false for non-timeout error")
	}
	if readbackFallbackAllowed(timeoutErr, false, false) {
		t.Fatal("readbackFallbackAllowed() = true, want false when previous state was not checked")
	}
	if readbackFallbackAllowed(timeoutErr, true, true) {
		t.Fatal("readbackFallbackAllowed() = true, want false when previous state already matched")
	}
}
