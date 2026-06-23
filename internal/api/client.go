package api

import (
	"bytes"
	"crypto/tls"
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

// APIError represents a non-2xx API response.
type APIError struct {
	StatusCode int
	Body       []byte
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API returned status %d: %s", e.StatusCode, string(e.Body))
}

// HTTPProtocol controls which HTTP protocol the client transport should use.
type HTTPProtocol string

const (
	HTTPProtocolAuto HTTPProtocol = "auto"
	HTTPProtocolH1   HTTPProtocol = "h1"
	HTTPProtocolH2   HTTPProtocol = "h2"
)

// ClientOptions controls API client transport behavior.
type ClientOptions struct {
	Protocol            HTTPProtocol
	MaxIdleConns        int
	MaxIdleConnsPerHost int
}

// NewClient creates a new API client.
func NewClient() (*Client, error) {
	return NewClientWithOptions(ClientOptions{})
}

// NewClientWithOptions creates a new API client with explicit transport options.
func NewClientWithOptions(opts ClientOptions) (*Client, error) {
	token, err := config.GetToken()
	if err != nil {
		return nil, err
	}
	transport, err := newTransport(opts)
	if err != nil {
		return nil, err
	}
	return &Client{
		http: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		token: token,
	}, nil
}

func newTransport(opts ClientOptions) (*http.Transport, error) {
	protocol := opts.Protocol
	if protocol == "" || protocol == HTTPProtocolAuto {
		protocol = HTTPProtocolH2
	}
	if protocol != HTTPProtocolH1 && protocol != HTTPProtocolH2 {
		return nil, fmt.Errorf("invalid HTTP protocol %q", opts.Protocol)
	}

	maxIdleConns := opts.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = 100
	}
	maxIdleConnsPerHost := opts.MaxIdleConnsPerHost
	if maxIdleConnsPerHost <= 0 {
		maxIdleConnsPerHost = 20
	}

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.MaxIdleConns = maxIdleConns
	tr.MaxIdleConnsPerHost = maxIdleConnsPerHost
	tr.IdleConnTimeout = 90 * time.Second
	tr.ResponseHeaderTimeout = 30 * time.Second

	if protocol == HTTPProtocolH1 {
		tr.ForceAttemptHTTP2 = false
		tr.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
		tlsConfig := &tls.Config{}
		if tr.TLSClientConfig != nil {
			tlsConfig = tr.TLSClientConfig.Clone()
		}
		tlsConfig.NextProtos = []string{"http/1.1"}
		tr.TLSClientConfig = tlsConfig
	} else {
		tr.ForceAttemptHTTP2 = true
	}

	return tr, nil
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
		return nil, &APIError{StatusCode: resp.StatusCode, Body: body}
	}
	return body, nil
}

// doGetMayEmpty is like doGet but also accepts 204 No Content (returns nil body).
func (c *Client) doGetMayEmpty(path string, params url.Values) ([]byte, error) {
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
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{StatusCode: resp.StatusCode, Body: body}
	}
	return body, nil
}

func (c *Client) doPost(path string, payload interface{}) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", config.BaseURL+path, bytes.NewReader(data))
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
		return nil, &APIError{StatusCode: resp.StatusCode, Body: body}
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

// SearchByCustomerName searches tickets by customer name.
func (c *Client) SearchByCustomerName(query string, skip int) ([]model.Ticket, []byte, error) {
	body, err := c.doPost("/search/customer-name", map[string]interface{}{"query": query, "skip": skip})
	if err != nil {
		return nil, nil, err
	}
	var tickets []model.Ticket
	if err := json.Unmarshal(body, &tickets); err != nil {
		return nil, body, err
	}
	return tickets, body, nil
}

// SearchByAttributes searches tickets by custom attributes.
func (c *Client) SearchByAttributes(query string, skip int) ([]model.Ticket, []byte, error) {
	body, err := c.doPost("/search/attributes", map[string]interface{}{"query": query, "skip": skip})
	if err != nil {
		return nil, nil, err
	}
	var tickets []model.Ticket
	if err := json.Unmarshal(body, &tickets); err != nil {
		return nil, body, err
	}
	return tickets, body, nil
}

// MarkMessagesRead marks the given message IDs as read in a ticket.
func (c *Client) MarkMessagesRead(ticketID string, messageIDs []string) error {
	_, err := c.doPost("/ticket/"+ticketID+"/readmessages", map[string]interface{}{"messageIds": messageIDs})
	return err
}

// FetchCustomerTickets fetches all tickets for a given customer ID.
// skip is an optional ticket ID cursor for pagination (empty string for first page).
func (c *Client) FetchCustomerTickets(customerID, skip string) ([]model.Ticket, []byte, error) {
	params := url.Values{}
	if skip != "" {
		params.Set("skip", skip)
	}
	u := "/ticket/customer/" + customerID
	body, err := c.doGetMayEmpty(u, params)
	if err != nil {
		return nil, nil, err
	}
	// 204 No Content means no more tickets
	if len(body) == 0 {
		return nil, body, nil
	}
	var tickets []model.Ticket
	if err := json.Unmarshal(body, &tickets); err != nil {
		return nil, body, err
	}
	return tickets, body, nil
}

// FetchSession fetches the full session refresh response.
func (c *Client) FetchSession() (*model.SessionRefreshResponse, error) {
	body, err := c.doGet("/session/refresh", nil)
	if err != nil {
		return nil, err
	}
	var resp model.SessionRefreshResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FetchIntegrations fetches the list of connected integrations.
func (c *Client) FetchIntegrations() ([]model.Integration, error) {
	body, err := c.doGet("/integrations/list", nil)
	if err != nil {
		return nil, err
	}
	var integrations []model.Integration
	if err := json.Unmarshal(body, &integrations); err != nil {
		return nil, err
	}
	return integrations, nil
}

// FetchClientAttributes fetches the custom attribute definitions for the client.
func (c *Client) FetchClientAttributes() ([]model.ClientAttribute, error) {
	body, err := c.doGet("/client/attribute", nil)
	if err != nil {
		return nil, err
	}
	var attrs []model.ClientAttribute
	if err := json.Unmarshal(body, &attrs); err != nil {
		return nil, err
	}
	return attrs, nil
}

// FetchMembers fetches team members from /session/refresh.
func (c *Client) FetchMembers() ([]model.Member, error) {
	resp, err := c.FetchSession()
	if err != nil {
		return nil, err
	}
	return resp.Member.Client.Members, nil
}
