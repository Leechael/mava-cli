package main

import (
	"testing"

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
