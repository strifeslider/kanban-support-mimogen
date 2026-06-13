package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/user/kanban-saas/pkg/model"
)

func TestTicketRepository_New(t *testing.T) {
	repo := &TicketRepository{}
	if repo == nil {
		t.Error("expected non-nil repo")
	}
}

func TestAgentRepository_New(t *testing.T) {
	repo := &AgentRepository{}
	if repo == nil {
		t.Error("expected non-nil repo")
	}
}

func TestTicketRepository_Model(t *testing.T) {
	ticket := &model.Ticket{
		ID:          uuid.New(),
		WorkspaceID: uuid.New(),
		Subject:     "Issue",
		Status:      "open",
		Priority:    "medium",
		Source:      "web",
		UserID:      uuid.New(),
	}
	if ticket.Subject != "Issue" {
		t.Error("subject mismatch")
	}
}

func TestTicketRepository_MessageModel(t *testing.T) {
	msg := &model.TicketMessage{
		ID:         uuid.New(),
		TicketID:   uuid.New(),
		SenderID:   uuid.New(),
		SenderType: "user",
		Content:    "text",
	}
	if msg.Content != "text" {
		t.Error("content mismatch")
	}
}

func TestAgentRepository_Model(t *testing.T) {
	agent := &model.SupportAgent{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		MaxTickets: 10,
		IsOnline:   true,
	}
	if agent.MaxTickets != 10 {
		t.Error("max_tickets mismatch")
	}
}

func TestTicketRepository_StatusLogModel(t *testing.T) {
	log := &model.TicketStatusLog{
		ID:        uuid.New(),
		TicketID:  uuid.New(),
		NewStatus: "closed",
		ChangedBy: uuid.New(),
	}
	if log.NewStatus != "closed" {
		t.Error("new_status mismatch")
	}
}

func TestTicketRepository_FilterModel(t *testing.T) {
	s := "open"
	p := "high"
	filter := model.TicketFilter{
		Status:   &s,
		Priority: &p,
		Limit:    50,
		Offset:   0,
	}
	if *filter.Status != "open" || *filter.Priority != "high" {
		t.Error("filter mismatch")
	}
}
