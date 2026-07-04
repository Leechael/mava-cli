package main

import (
	"testing"

	"github.com/phalahq/mava-api/internal/model"
)

func TestTicketAssignedTo(t *testing.T) {
	agentID := "68b7a10fc8899d2bd7e3e98c"
	if !ticketAssignedTo(&model.Ticket{AssignedTo: agentID}, agentID) {
		t.Fatal("ticketAssignedTo() = false, want true for matching id")
	}
	if ticketAssignedTo(&model.Ticket{AssignedTo: ""}, agentID) {
		t.Fatal("ticketAssignedTo() = true, want false for empty assignee")
	}
	if ticketAssignedTo(nil, agentID) {
		t.Fatal("ticketAssignedTo(nil) = true, want false")
	}
}
