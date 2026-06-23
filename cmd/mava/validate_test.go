package main

import "testing"

func TestIsValidObjectID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"valid lowercase hex", "6a393bfd24a7951070951db5", true},
		{"valid uppercase hex", "6A393BFD24A7951070951DB5", true},
		{"valid mixed case hex", "6a393BFD24a7951070951Db5", true},
		{"empty string", "", false},
		{"too short", "6a393bfd24a7951070951db", false},
		{"too long", "6a393bfd24a7951070951db5a", false},
		{"invalid character g", "6a393bfd24a7951070951dbg", false},
		{"invalid character z", "zzzzzzzzzzzzzzzzzzzzzzzz", false},
		{"invalid character hyphen", "6a393bfd-4a7951070951db5", false},
		{"all zeros", "000000000000000000000000", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidObjectID(tt.id); got != tt.want {
				t.Errorf("isValidObjectID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestValidateTicketID(t *testing.T) {
	if err := validateTicketID("6a393bfd24a7951070951db5"); err != nil {
		t.Errorf("validateTicketID(valid) = %v, want nil", err)
	}
	if err := validateTicketID("xxx"); err == nil {
		t.Error("validateTicketID(invalid) = nil, want error")
	}
}

func TestValidateCustomerID(t *testing.T) {
	if err := validateCustomerID("6a393bfd24a7951070951db5"); err != nil {
		t.Errorf("validateCustomerID(valid) = %v, want nil", err)
	}
	if err := validateCustomerID("xxx"); err == nil {
		t.Error("validateCustomerID(invalid) = nil, want error")
	}
}

func TestValidateMessageID(t *testing.T) {
	if err := validateMessageID("6a393bfd24a7951070951db5"); err != nil {
		t.Errorf("validateMessageID(valid) = %v, want nil", err)
	}
	if err := validateMessageID("xxx"); err == nil {
		t.Error("validateMessageID(invalid) = nil, want error")
	}
}
