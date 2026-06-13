package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/user/kanban-saas/pkg/mock"
	"github.com/user/kanban-saas/pkg/model"
)

func newTestSupportService() (*SupportService, *mock.MockTicketRepo, *mock.MockAgentRepo) {
	ticketRepo := mock.NewMockTicketRepo()
	agentRepo := mock.NewMockAgentRepo()
	svc := NewSupportService(ticketRepo, agentRepo)
	return svc, ticketRepo, agentRepo
}

func TestSupportService_CreateTicket(t *testing.T) {
	svc, _, _ := newTestSupportService()
	ctx := context.Background()

	ticket, err := svc.CreateTicket(ctx, uuid.New(), model.CreateTicketRequest{
		WorkspaceID: uuid.New(),
		Subject:     "Help needed",
		Priority:    "high",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ticket.Subject != "Help needed" {
		t.Errorf("expected subject 'Help needed', got '%s'", ticket.Subject)
	}
	if ticket.Priority != "high" {
		t.Errorf("expected priority 'high', got '%s'", ticket.Priority)
	}
}

func TestSupportService_CreateTicket_DefaultPriority(t *testing.T) {
	svc, _, _ := newTestSupportService()
	ctx := context.Background()

	ticket, err := svc.CreateTicket(ctx, uuid.New(), model.CreateTicketRequest{
		WorkspaceID: uuid.New(),
		Subject:     "Issue",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ticket.Priority != "medium" {
		t.Errorf("expected default priority 'medium', got '%s'", ticket.Priority)
	}
}

func TestSupportService_CreateTicket_AutoAssign(t *testing.T) {
	svc, _, agentRepo := newTestSupportService()
	ctx := context.Background()

	agentID := uuid.New()
	agentRepo.Agents[agentID] = &model.SupportAgent{
		ID:       agentID,
		UserID:   uuid.New(),
		IsOnline: true,
	}

	ticket, err := svc.CreateTicket(ctx, uuid.New(), model.CreateTicketRequest{
		WorkspaceID: uuid.New(),
		Subject:     "Issue",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ticket.AgentID == nil {
		t.Error("expected agent to be assigned")
	}
	if ticket.Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got '%s'", ticket.Status)
	}
}

func TestSupportService_CreateTicket_NoAgents(t *testing.T) {
	svc, _, _ := newTestSupportService()
	ctx := context.Background()

	ticket, err := svc.CreateTicket(ctx, uuid.New(), model.CreateTicketRequest{
		WorkspaceID: uuid.New(),
		Subject:     "Issue",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ticket.AgentID != nil {
		t.Error("expected no agent assigned")
	}
	if ticket.Status != "open" {
		t.Errorf("expected status 'open', got '%s'", ticket.Status)
	}
}

func TestSupportService_GetTicket(t *testing.T) {
	svc, ticketRepo, _ := newTestSupportService()
	ctx := context.Background()

	ticketID := uuid.New()
	ticketRepo.Tickets[ticketID] = &model.Ticket{
		ID:      ticketID,
		Subject: "Test",
	}

	ticket, err := svc.GetTicket(ctx, ticketID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ticket.Subject != "Test" {
		t.Errorf("expected subject 'Test', got '%s'", ticket.Subject)
	}
}

func TestSupportService_ListTickets(t *testing.T) {
	svc, ticketRepo, _ := newTestSupportService()
	ctx := context.Background()

	ticketRepo.Tickets[uuid.New()] = &model.Ticket{Status: "open"}
	ticketRepo.Tickets[uuid.New()] = &model.Ticket{Status: "closed"}

	tickets, err := svc.ListTickets(ctx, model.TicketFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tickets) != 2 {
		t.Errorf("expected 2 tickets, got %d", len(tickets))
	}
}

func TestSupportService_ListTickets_Filter(t *testing.T) {
	svc, ticketRepo, _ := newTestSupportService()
	ctx := context.Background()

	ticketRepo.Tickets[uuid.New()] = &model.Ticket{Status: "open"}
	ticketRepo.Tickets[uuid.New()] = &model.Ticket{Status: "closed"}

	status := "open"
	tickets, err := svc.ListTickets(ctx, model.TicketFilter{Status: &status})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tickets) != 1 {
		t.Errorf("expected 1 ticket, got %d", len(tickets))
	}
}

func TestSupportService_UpdateTicket(t *testing.T) {
	svc, ticketRepo, _ := newTestSupportService()
	ctx := context.Background()

	ticketID := uuid.New()
	ticketRepo.Tickets[ticketID] = &model.Ticket{
		ID:     ticketID,
		Status: "open",
	}

	newStatus := "in_progress"
	ticket, err := svc.UpdateTicket(ctx, ticketID, model.UpdateTicketRequest{
		Status: &newStatus,
	}, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ticket.Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got '%s'", ticket.Status)
	}
}

func TestSupportService_CloseTicket(t *testing.T) {
	svc, ticketRepo, _ := newTestSupportService()
	ctx := context.Background()

	ticketID := uuid.New()
	ticketRepo.Tickets[ticketID] = &model.Ticket{
		ID:     ticketID,
		Status: "open",
	}

	err := svc.CloseTicket(ctx, ticketID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ticketRepo.Tickets[ticketID].Status != "closed" {
		t.Errorf("expected status 'closed', got '%s'", ticketRepo.Tickets[ticketID].Status)
	}
}

func TestSupportService_AssignAgent(t *testing.T) {
	svc, ticketRepo, _ := newTestSupportService()
	ctx := context.Background()

	ticketID := uuid.New()
	ticketRepo.Tickets[ticketID] = &model.Ticket{
		ID:     ticketID,
		Status: "open",
	}

	agentID := uuid.New()
	err := svc.AssignAgent(ctx, ticketID, agentID, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ticketRepo.Tickets[ticketID].Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got '%s'", ticketRepo.Tickets[ticketID].Status)
	}
}

func TestSupportService_AddMessage(t *testing.T) {
	svc, _, _ := newTestSupportService()
	ctx := context.Background()

	ticketID := uuid.New()
	msg, err := svc.AddMessage(ctx, ticketID, uuid.New(), "user", "Hello", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Content != "Hello" {
		t.Errorf("expected content 'Hello', got '%s'", msg.Content)
	}
}

func TestSupportService_AddMessageFromBot(t *testing.T) {
	svc, _, _ := newTestSupportService()
	ctx := context.Background()

	ticketID := uuid.New()
	msg, err := svc.AddMessageFromBot(ctx, ticketID, uuid.New(), "bot", "Bot reply", "ext123", "telegram")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Content != "Bot reply" {
		t.Errorf("expected content 'Bot reply', got '%s'", msg.Content)
	}
	if *msg.Platform != "telegram" {
		t.Errorf("expected platform 'telegram', got '%s'", *msg.Platform)
	}
}

func TestSupportService_ListMessages(t *testing.T) {
	svc, ticketRepo, _ := newTestSupportService()
	ctx := context.Background()

	ticketID := uuid.New()
	ticketRepo.Messages[ticketID] = []model.TicketMessage{
		{Content: "msg1"},
		{Content: "msg2"},
	}

	messages, err := svc.ListMessages(ctx, ticketID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}
}

func TestSupportService_ListAgents(t *testing.T) {
	svc, _, agentRepo := newTestSupportService()
	ctx := context.Background()

	agentRepo.Agents[uuid.New()] = &model.SupportAgent{IsOnline: true}
	agentRepo.Agents[uuid.New()] = &model.SupportAgent{IsOnline: false}

	agents, err := svc.ListAgents(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(agents))
	}
}

func TestSupportService_UpdateAgentStatus(t *testing.T) {
	svc, _, agentRepo := newTestSupportService()
	ctx := context.Background()

	agentID := uuid.New()
	agentRepo.Agents[agentID] = &model.SupportAgent{IsOnline: false}

	err := svc.UpdateAgentStatus(ctx, agentID, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !agentRepo.Agents[agentID].IsOnline {
		t.Error("expected agent to be online")
	}
}

func TestSupportService_GetStats(t *testing.T) {
	svc, ticketRepo, _ := newTestSupportService()
	ctx := context.Background()

	ticketRepo.Tickets[uuid.New()] = &model.Ticket{Status: "open"}
	ticketRepo.Tickets[uuid.New()] = &model.Ticket{Status: "open"}
	ticketRepo.Tickets[uuid.New()] = &model.Ticket{Status: "closed"}

	stats, err := svc.GetStats(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats["open_tickets"] != 2 {
		t.Errorf("expected 2 open tickets, got %v", stats["open_tickets"])
	}
	if stats["total_tickets"] != 3 {
		t.Errorf("expected 3 total tickets, got %v", stats["total_tickets"])
	}
}
