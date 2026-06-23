package main

import "fmt"

func isValidObjectID(id string) bool {
	if len(id) != 24 {
		return false
	}
	for _, c := range id {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

func validateTicketID(id string) error {
	if !isValidObjectID(id) {
		return fmt.Errorf("invalid ticket id %q: expected 24-character hex id", id)
	}
	return nil
}

func validateCustomerID(id string) error {
	if !isValidObjectID(id) {
		return fmt.Errorf("invalid customer id %q: expected 24-character hex id", id)
	}
	return nil
}

func validateMessageID(id string) error {
	if !isValidObjectID(id) {
		return fmt.Errorf("invalid message id %q: expected 24-character hex id", id)
	}
	return nil
}
