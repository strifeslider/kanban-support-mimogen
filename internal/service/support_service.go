package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/user/kanban-saas/pkg/model"
)

type SupportService struct {
	ticketRepo TicketRepository
	agentRepo  AgentRepository
}

func NewSupportService(
	ticketRepo TicketRepository,
	agentRepo AgentRepository,
) *SupportService {
	return &SupportService{
		ticketRepo: ticketRepo,
		agentRepo:  agentRepo,
	}
}

func (s *SupportService) CreateTicket(ctx context.Context, userID uuid.UUID, req model.CreateTicketRequest) (*model.Ticket, error) {
	priority := "medium"
	if req.Priority != "" {
		priority = req.Priority
	}

	ticket := &model.Ticket{
		ID:          uuid.New(),
		WorkspaceID: req.WorkspaceID,
		Subject:     req.Subject,
		Status:      "open",
		Priority:    priority,
		Source:      "web",
		UserID:      userID,
	}

	if err := s.ticketRepo.Create(ctx, ticket); err != nil {
		return nil, fmt.Errorf("create ticket: %w", err)
	}

	agent, err := s.agentRepo.GetAvailableAgent(ctx)
	if err == nil && agent != nil {
		ticket.AgentID = &agent.ID
		ticket.Status = "in_progress"
		s.ticketRepo.Update(ctx, ticket)

		s.ticketRepo.AddStatusLog(ctx, &model.TicketStatusLog{
			ID:        uuid.New(),
			TicketID:  ticket.ID,
			OldStatus: strPtr("open"),
			NewStatus: "in_progress",
			ChangedBy: agent.UserID,
		})
	}

	return ticket, nil
}

func (s *SupportService) CreateTicketFromBot(ctx context.Context, req model.CreateTicketRequest, source, channelID string, userID *uuid.UUID) (*model.Ticket, error) {
	priority := "medium"
	if req.Priority != "" {
		priority = req.Priority
	}

	ticket := &model.Ticket{
		ID:          uuid.New(),
		WorkspaceID: req.WorkspaceID,
		Subject:     req.Subject,
		Status:      "open",
		Priority:    priority,
		Source:      source,
		ChannelID:   &channelID,
		UserID:      uuid.Nil,
	}

	if userID != nil {
		ticket.UserID = *userID
	}

	if err := s.ticketRepo.Create(ctx, ticket); err != nil {
		return nil, fmt.Errorf("create ticket: %w", err)
	}

	agent, err := s.agentRepo.GetAvailableAgent(ctx)
	if err == nil && agent != nil {
		ticket.AgentID = &agent.ID
		ticket.Status = "in_progress"
		s.ticketRepo.Update(ctx, ticket)
	}

	return ticket, nil
}

func (s *SupportService) GetTicket(ctx context.Context, id uuid.UUID) (*model.Ticket, error) {
	return s.ticketRepo.GetByID(ctx, id)
}

func (s *SupportService) ListTickets(ctx context.Context, filter model.TicketFilter) ([]model.Ticket, error) {
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	return s.ticketRepo.List(ctx, filter)
}

func (s *SupportService) UpdateTicket(ctx context.Context, id uuid.UUID, req model.UpdateTicketRequest, changedBy uuid.UUID) (*model.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldStatus := ticket.Status

	if req.Status != nil {
		ticket.Status = *req.Status
	}
	if req.Priority != nil {
		ticket.Priority = *req.Priority
	}
	if req.AgentID != nil {
		ticket.AgentID = req.AgentID
	}

	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return nil, fmt.Errorf("update ticket: %w", err)
	}

	if req.Status != nil && *req.Status != oldStatus {
		s.ticketRepo.AddStatusLog(ctx, &model.TicketStatusLog{
			ID:        uuid.New(),
			TicketID:  ticket.ID,
			OldStatus: &oldStatus,
			NewStatus: *req.Status,
			ChangedBy: changedBy,
		})
	}

	return ticket, nil
}

func (s *SupportService) CloseTicket(ctx context.Context, id uuid.UUID) error {
	return s.ticketRepo.Close(ctx, id)
}

func (s *SupportService) AssignAgent(ctx context.Context, ticketID, agentID uuid.UUID, changedBy uuid.UUID) error {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return err
	}

	oldStatus := ticket.Status
	ticket.AgentID = &agentID
	ticket.Status = "in_progress"

	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return fmt.Errorf("assign agent: %w", err)
	}

	s.ticketRepo.AddStatusLog(ctx, &model.TicketStatusLog{
		ID:        uuid.New(),
		TicketID:  ticket.ID,
		OldStatus: &oldStatus,
		NewStatus: "in_progress",
		ChangedBy: changedBy,
	})

	return nil
}

func (s *SupportService) AddMessage(ctx context.Context, ticketID, senderID uuid.UUID, senderType, content string, platform *string) (*model.TicketMessage, error) {
	msg := &model.TicketMessage{
		ID:         uuid.New(),
		TicketID:   ticketID,
		SenderID:   senderID,
		SenderType: senderType,
		Content:    content,
		Platform:   platform,
	}

	if err := s.ticketRepo.AddMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("add message: %w", err)
	}

	return msg, nil
}

func (s *SupportService) AddMessageFromBot(ctx context.Context, ticketID uuid.UUID, senderID uuid.UUID, senderType, content, externalID string, platform string) (*model.TicketMessage, error) {
	msg := &model.TicketMessage{
		ID:         uuid.New(),
		TicketID:   ticketID,
		SenderID:   senderID,
		SenderType: senderType,
		Content:    content,
		Platform:   &platform,
		ExternalID: &externalID,
	}

	if err := s.ticketRepo.AddMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("add message from bot: %w", err)
	}

	return msg, nil
}

func (s *SupportService) ListMessages(ctx context.Context, ticketID uuid.UUID) ([]model.TicketMessage, error) {
	return s.ticketRepo.ListMessages(ctx, ticketID)
}

func (s *SupportService) ListAgents(ctx context.Context) ([]model.SupportAgent, error) {
	return s.agentRepo.List(ctx)
}

func (s *SupportService) UpdateAgentStatus(ctx context.Context, agentID uuid.UUID, isOnline bool) error {
	return s.agentRepo.UpdateStatus(ctx, agentID, isOnline)
}

func (s *SupportService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	return s.ticketRepo.GetStats(ctx)
}

func strPtr(s string) *string {
	return &s
}
