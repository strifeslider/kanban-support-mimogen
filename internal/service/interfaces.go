package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/user/kanban-saas/pkg/model"
)

type TicketRepository interface {
	Create(ctx context.Context, ticket *model.Ticket) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Ticket, error)
	List(ctx context.Context, filter model.TicketFilter) ([]model.Ticket, error)
	Update(ctx context.Context, ticket *model.Ticket) error
	Close(ctx context.Context, id uuid.UUID) error
	AddMessage(ctx context.Context, msg *model.TicketMessage) error
	ListMessages(ctx context.Context, ticketID uuid.UUID) ([]model.TicketMessage, error)
	AddStatusLog(ctx context.Context, log *model.TicketStatusLog) error
	GetStats(ctx context.Context) (map[string]interface{}, error)
}

type AgentRepository interface {
	Create(ctx context.Context, agent *model.SupportAgent) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.SupportAgent, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*model.SupportAgent, error)
	List(ctx context.Context) ([]model.SupportAgent, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, isOnline bool) error
	GetAvailableAgent(ctx context.Context) (*model.SupportAgent, error)
}
