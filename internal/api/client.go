package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/phalahq/mava-api/internal/config"
	"github.com/phalahq/mava-api/internal/model"
)

// Client wraps the HTTP client for Mava API calls.
type Client struct {
	http  *http.Client
	token string
}

// NewClient creates a new API client.
func NewClient() (*Client, error) {
	token, err := config.GetToken()
	if err != nil {
		return nil, err
	}
	return &Client{
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		token: token,
	}, nil
}

func (c *Client) doGet(path string, params url.Values) ([]byte, error) {
	u := config.BaseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(&http.Cookie{Name: "x-auth-token", Value: c.token})
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func (c *Client) doPost(path string, payload interface{}) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", config.BaseURL+path, strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "x-auth-token", Value: c.token})
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

// ListTicketsParams holds query parameters for listing tickets.
type ListTicketsParams struct {
	Limit        int
	Skip         int
	Sort         string
	Order        string
	Status       []string
	Priority     *int
	Category     string
	AssignedTo   string
	Tag          string
	AIStatus     string
	SourceType   string
	IncludeEmpty bool
}

// ListTickets fetches tickets with the given filters.
func (c *Client) ListTickets(p ListTicketsParams) (*model.TicketListResponse, []byte, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", p.Limit))
	params.Set("skip", fmt.Sprintf("%d", p.Skip))
	params.Set("sort", p.Sort)
	params.Set("order", p.Order)
	params.Set("filterVersion", "2")
	params.Set("filterLastUpdated", time.Now().UTC().Format(time.RFC3339))
	if p.IncludeEmpty {
		params.Set("skipEmptyMessages", "false")
	} else {
		params.Set("skipEmptyMessages", "true")
	}

	if len(p.Status) > 0 {
		params.Set("status", strings.Join(p.Status, ","))
		params.Set("hasStatusFilter", "true")
	} else {
		params.Set("status", "Open,Pending,Waiting,Resolved,Spam")
		params.Set("hasStatusFilter", "true")
	}

	if p.Priority != nil {
		params.Set("priority", fmt.Sprintf("%d", *p.Priority))
		params.Set("hasPriorityFilter", "true")
	} else {
		params.Set("priority", "")
		params.Set("hasPriorityFilter", "false")
	}

	if p.Category != "" {
		params.Set("category", p.Category)
		params.Set("hasCategoryFilter", "true")
	} else {
		params.Set("category", "")
		params.Set("hasCategoryFilter", "false")
	}

	if p.AssignedTo != "" {
		params.Set("assignedTo", p.AssignedTo)
		params.Set("hasAgentFilter", "true")
	} else {
		params.Set("assignedTo", "")
		params.Set("hasAgentFilter", "false")
	}

	if p.Tag != "" {
		params.Set("tag", p.Tag)
		params.Set("hasTagFilter", "true")
	} else {
		params.Set("tag", "")
		params.Set("hasTagFilter", "false")
	}

	if p.AIStatus != "" {
		params.Set("aiStatus", p.AIStatus)
		params.Set("hasAiStatusFilter", "true")
	} else {
		params.Set("aiStatus", "")
		params.Set("hasAiStatusFilter", "false")
	}

	if p.SourceType != "" {
		params.Set("sourceType", p.SourceType)
		params.Set("hasSourceTypeFilter", "true")
	} else {
		params.Set("sourceType", "")
		params.Set("hasSourceTypeFilter", "false")
	}

	body, err := c.doGet("/ticket/list", params)
	if err != nil {
		return nil, nil, err
	}

	var result model.TicketListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, body, err
	}
	return &result, body, nil
}

// GetTicket fetches a single ticket by ID.
func (c *Client) GetTicket(ticketID string) (*model.Ticket, []byte, error) {
	body, err := c.doGet("/ticket/"+ticketID, nil)
	if err != nil {
		return nil, nil, err
	}
	var ticket model.Ticket
	if err := json.Unmarshal(body, &ticket); err != nil {
		return nil, body, err
	}
	return &ticket, body, nil
}

// SearchMessages searches message content.
func (c *Client) SearchMessages(query string) ([]model.SearchResult, []byte, error) {
	body, err := c.doPost("/search/message-content", map[string]string{"query": query})
	if err != nil {
		return nil, nil, err
	}
	var results []model.SearchResult
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, body, err
	}
	return results, body, nil
}

// FetchMembers fetches team members from /session/refresh.
func (c *Client) FetchMembers() ([]model.Member, error) {
	body, err := c.doGet("/session/refresh", nil)
	if err != nil {
		return nil, err
	}
	var resp model.SessionRefreshResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp.Member.Client.Members, nil
}
