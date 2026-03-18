package ticket

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/phalahq/mava-api/internal/model"
	"github.com/phalahq/mava-api/internal/output"
)

// CheckNeedsReply examines message history to determine if a human reply is needed.
// Returns (needsReply, lastCustomerMsg, aiRepliesAfterCustomer).
func CheckNeedsReply(messages []model.Message) (bool, *model.Message, []model.Message) {
	var real []model.Message
	for _, msg := range messages {
		switch msg.MessageType {
		case "StatusAction", "ChatbotButton", "InternalNote":
			continue
		}
		if msg.Content == "" {
			continue
		}
		real = append(real, msg)
	}

	sort.Slice(real, func(i, j int) bool {
		return real[i].CreatedAt < real[j].CreatedAt
	})

	if len(real) == 0 {
		return false, nil, nil
	}

	lastCustIdx := -1
	for i, msg := range real {
		if msg.FromCustomer {
			lastCustIdx = i
		}
	}
	if lastCustIdx < 0 {
		return false, nil, nil
	}

	lastCustMsg := real[lastCustIdx]

	var aiReplies []model.Message
	for i := lastCustIdx + 1; i < len(real); i++ {
		msg := real[i]
		if !msg.FromCustomer && msg.SenderReferenceType == "ClientMember" {
			return false, nil, nil
		}
		if !msg.FromCustomer {
			aiReplies = append(aiReplies, msg)
		}
	}

	return true, &lastCustMsg, aiReplies
}

// CalcWaitTime returns a human-readable duration since createdAt.
func CalcWaitTime(createdAt string) string {
	s := strings.Replace(createdAt, "Z", "+00:00", 1)
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, s)
		if err != nil {
			return ""
		}
	}
	delta := time.Since(t)
	days := int(delta.Hours() / 24)
	if days > 0 {
		return fmt.Sprintf("%dd", days)
	}
	hours := int(delta.Hours())
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dm", int(delta.Minutes()))
}

// BuildNeedsReplyItem constructs a NeedsReplyItem from check results.
func BuildNeedsReplyItem(t model.Ticket, lastCustMsg *model.Message, aiReplies []model.Message, dashboardURL string) model.NeedsReplyItem {
	return model.NeedsReplyItem{
		Ticket:         t,
		LastMessage:    lastCustMsg,
		AIRepliesAfter: aiReplies,
		LastMsgTime:    output.FormatDatetime(lastCustMsg.CreatedAt),
		WaitTime:       CalcWaitTime(lastCustMsg.CreatedAt),
		DashboardURL:   dashboardURL,
	}
}
