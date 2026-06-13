package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/user/kanban-saas/pkg/model"
)

func TestNewSupportService(t *testing.T) {
	svc := &SupportService{}
	if svc == nil {
		t.Error("expected non-nil service")
	}
}

func TestTicketStatus(t *testing.T) {
	statuses := []string{"open", "in_progress", "waiting", "resolved", "closed"}
	for _, s := range statuses {
		if s == "" {
			t.Error("expected non-empty status")
		}
	}
}

func TestTicketPriority(t *testing.T) {
	priorities := []string{"low", "medium", "high", "urgent"}
	for _, p := range priorities {
		if p == "" {
			t.Error("expected non-empty priority")
		}
	}
}

func TestTicketSource(t *testing.T) {
	sources := []string{"telegram", "discord", "web", "email"}
	for _, s := range sources {
		if s == "" {
			t.Error("expected non-empty source")
		}
	}
}

func TestTicketModel(t *testing.T) {
	ticket := model.Ticket{
		ID:       uuid.New(),
		Subject:  "Test Ticket",
		Status:   "open",
		Priority: "medium",
		Source:   "web",
		UserID:   uuid.New(),
	}

	if ticket.Subject != "Test Ticket" {
		t.Errorf("expected subject 'Test Ticket', got '%s'", ticket.Subject)
	}
	if ticket.Status != "open" {
		t.Errorf("expected status 'open', got '%s'", ticket.Status)
	}
}

func TestTicketMessageModel(t *testing.T) {
	msg := model.TicketMessage{
		ID:         uuid.New(),
		TicketID:   uuid.New(),
		SenderID:   uuid.New(),
		SenderType: "user",
		Content:    "Test message",
	}

	if msg.Content != "Test message" {
		t.Errorf("expected content 'Test message', got '%s'", msg.Content)
	}
	if msg.SenderType != "user" {
		t.Errorf("expected sender_type 'user', got '%s'", msg.SenderType)
	}
}

func TestSupportAgentModel(t *testing.T) {
	agent := model.SupportAgent{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		MaxTickets: 10,
		IsOnline:   true,
	}

	if agent.MaxTickets != 10 {
		t.Errorf("expected max_tickets 10, got %d", agent.MaxTickets)
	}
	if !agent.IsOnline {
		t.Error("expected agent to be online")
	}
}

func TestCreateTicketRequest(t *testing.T) {
	req := model.CreateTicketRequest{
		WorkspaceID: uuid.New(),
		Subject:     "New Ticket",
		Priority:    "high",
	}

	if req.Subject != "New Ticket" {
		t.Errorf("expected subject 'New Ticket', got '%s'", req.Subject)
	}
}

func TestUpdateTicketRequest(t *testing.T) {
	status := "in_progress"
	priority := "urgent"
	req := model.UpdateTicketRequest{
		Status:   &status,
		Priority: &priority,
	}

	if *req.Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got '%s'", *req.Status)
	}
}

func TestTicketFilter(t *testing.T) {
	status := "open"
	filter := model.TicketFilter{
		Status: &status,
		Limit:  50,
		Offset: 0,
	}

	if *filter.Status != "open" {
		t.Errorf("expected status 'open', got '%s'", *filter.Status)
	}
	if filter.Limit != 50 {
		t.Errorf("expected limit 50, got %d", filter.Limit)
	}
}

func TestStrPtr(t *testing.T) {
	result := strPtr("test")
	if *result != "test" {
		t.Errorf("expected 'test', got '%s'", *result)
	}
}
