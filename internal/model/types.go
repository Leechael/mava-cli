package model

// Customer represents a ticket customer.
type Customer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Message represents a ticket message.
type Message struct {
	ID                  string   `json:"_id"`
	Content             string   `json:"content"`
	FromCustomer        bool     `json:"fromCustomer"`
	MessageType         string   `json:"messageType"`
	SenderReferenceType string   `json:"senderReferenceType"`
	CreatedAt           string   `json:"createdAt"`
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

// AgentMapping is the known agent name -> ID mapping.
var AgentMapping = map[string]string{
	"doyle":      "68b7a10fc8899d2bd7e3e98c",
	"hangyin":    "68a3b7d48641a83d9130a9a7",
	"hugo":       "68811aa37686516d67e2c1df",
	"kingsley":   "68a3b7668641a83d9130a925",
	"leechael":   "68904ba393a35df39d8fe7d8",
	"marvintong": "6892f7fe9e5e9753fa50e1e0",
}

// AgentNameByID returns agent name for an ID, or the raw ID if unknown.
func AgentNameByID(id string) string {
	for name, aid := range AgentMapping {
		if aid == id {
			return name
		}
	}
	return id
}
