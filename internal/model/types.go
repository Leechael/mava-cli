package model

// Customer represents a ticket customer.
type Customer struct {
	ID         string            `json:"_id"`
	Name       string            `json:"name"`
	Email      string            `json:"email"`
	Attributes []TicketAttribute `json:"attributes"`
}

// Tag represents a ticket tag.
type Tag struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
}

// Category represents a ticket category.
type Category struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
}

// CSAT represents customer satisfaction rating.
type CSAT struct {
	Value int `json:"value"`
}

// TicketAttribute represents a custom attribute value on a ticket or customer.
type TicketAttribute struct {
	AttributeID string `json:"attributeId"`
	Name        string `json:"name"`
	Content     string `json:"content"`
}

// Attachment represents a file attachment on a message.
type Attachment struct {
	ID       string `json:"_id"`
	URL      string `json:"url"`
	FileName string `json:"fileName"`
}

// Message represents a ticket message.
type Message struct {
	ID                  string       `json:"_id"`
	Content             string       `json:"content"`
	FromCustomer        bool         `json:"fromCustomer"`
	MessageType         string       `json:"messageType"`
	SenderReferenceType string       `json:"senderReferenceType"`
	SenderType          string       `json:"senderType"`
	SenderName          string       `json:"senderName"`
	CreatedAt           string       `json:"createdAt"`
	Sender              interface{}  `json:"sender"`
	Attachments         []Attachment `json:"attachments"`
}

// Ticket represents a support ticket.
type Ticket struct {
	ID             string            `json:"_id"`
	Status         string            `json:"status"`
	Priority       int               `json:"priority"`
	SourceType     string            `json:"sourceType"`
	AIStatus       string            `json:"aiStatus"`
	AssignedTo     string            `json:"assignedTo"`
	Customer       Customer          `json:"customer"`
	Messages       []Message         `json:"messages"`
	Tags           []Tag             `json:"tags"`
	Category       *Category         `json:"category"`
	CSAT           *CSAT             `json:"csat"`
	ResolutionTime *int              `json:"resolutionTime"`
	Attributes     []TicketAttribute `json:"attributes"`
	CreatedAt      string            `json:"createdAt"`
	UpdatedAt      string            `json:"updatedAt"`
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

// SessionMember represents the current logged-in user from /session/refresh.
type SessionMember struct {
	ID              string `json:"_id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Type            string `json:"type"`
	IsEmailVerified bool   `json:"isEmailVerified"`
	IsArchived      bool   `json:"isArchived"`
	Client          struct {
		ID      string   `json:"_id"`
		Name    string   `json:"name"`
		Members []Member `json:"members"`
	} `json:"client"`
}

// SessionSubscription represents the plan/subscription info from /session/refresh.
type SessionSubscription struct {
	Name                     string `json:"name"`
	Period                   string `json:"period"`
	ExpirationDate           string `json:"expirationDate"`
	UsedSupportRequests      int    `json:"usedSupportRequests"`
	SupportRequestsAllowance int    `json:"supportRequestsAllowance"`
}

// SessionRefreshResponse is the response from /session/refresh.
type SessionRefreshResponse struct {
	Member       SessionMember       `json:"member"`
	Subscription SessionSubscription `json:"subscription"`
	TicketCount  int                 `json:"ticketCount"`
}

// Integration represents a connected channel (Discord, Telegram, WebChat, Email, etc.).
type Integration struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Name      string `json:"name"`
	AddedDate string `json:"addedDate"`
	GuildID   string `json:"guildId,omitempty"`
}

// ClientAttribute represents a custom attribute definition on the client (account).
type ClientAttribute struct {
	ID              string `json:"_id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	IconType        string `json:"iconType"`
	AffiliationType string `json:"affiliationType"`
	IsArchived      bool   `json:"isArchived"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
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
