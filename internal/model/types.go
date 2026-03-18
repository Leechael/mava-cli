package model

// Customer represents a ticket customer.
type Customer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Message represents a ticket message.
type Message struct {
	ID                  string      `json:"_id"`
	Content             string      `json:"content"`
	FromCustomer        bool        `json:"fromCustomer"`
	MessageType         string      `json:"messageType"`
	SenderReferenceType string      `json:"senderReferenceType"`
	CreatedAt           string      `json:"createdAt"`
	Sender              interface{} `json:"sender"`
}

// Ticket represents a support ticket.
type Ticket struct {
	ID         string    `json:"_id"`
	Status     string    `json:"status"`
	Priority   int       `json:"priority"`
	SourceType string    `json:"sourceType"`
	AIStatus   string    `json:"aiStatus"`
	AssignedTo string    `json:"assignedTo"`
	Customer   Customer  `json:"customer"`
	Messages   []Message `json:"messages"`
	CreatedAt  string    `json:"createdAt"`
	UpdatedAt  string    `json:"updatedAt"`
}

// TicketListResponse is the API response for listing tickets.
type TicketListResponse struct {
	Tickets []Ticket `json:"tickets"`
}

// SearchResult represents a search result item.
type SearchResult struct {
	ID              string   `json:"_id"`
	Status          string   `json:"status"`
	Customer        Customer `json:"customer"`
	RelevantMessage *struct {
		Content string `json:"content"`
	} `json:"relevantMessage"`
}

// NeedsReplyItem holds a ticket that needs a human reply.
type NeedsReplyItem struct {
	Ticket         Ticket    `json:"ticket"`
	LastMessage    *Message  `json:"last_message"`
	AIRepliesAfter []Message `json:"ai_replies_after"`
	LastMsgTime    string    `json:"last_message_time"`
	WaitTime       string    `json:"wait_time"`
	DashboardURL   string    `json:"dashboard_url"`
}

// PriorityString returns the human-readable priority name.
func PriorityString(p int) string {
	switch p {
	case 1:
		return "Low"
	case 2:
		return "Medium"
	case 3:
		return "High"
	case 4:
		return "Urgent"
	default:
		return "Medium"
	}
}

// Member represents a team member from the session/refresh API.
type Member struct {
	ID         string `json:"_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Type       string `json:"type"`
	IsArchived bool   `json:"isArchived"`
}

// SessionRefreshResponse is the response from /session/refresh.
type SessionRefreshResponse struct {
	Member struct {
		Client struct {
			Members []Member `json:"members"`
		} `json:"client"`
	} `json:"member"`
}

// AgentNameByID returns agent name for an ID, or the raw ID if unknown.
// It uses the cached members list if available (call SetMembers first).
func AgentNameByID(id string) string {
	for _, m := range cachedMembers {
		if m.ID == id {
			return m.Name
		}
	}
	return id
}

var cachedMembers []Member

// SetMembers caches the members list for AgentNameByID lookups.
func SetMembers(members []Member) {
	cachedMembers = members
}
