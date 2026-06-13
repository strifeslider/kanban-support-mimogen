package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/user/kanban-saas/pkg/model"
)

type AgentRepository struct {
	db *pgxpool.Pool
}

func NewAgentRepository(db *pgxpool.Pool) *AgentRepository {
	return &AgentRepository{db: db}
}

func (r *AgentRepository) Create(ctx context.Context, agent *model.SupportAgent) error {
	query := `
		INSERT INTO support_agents (id, user_id, max_tickets, is_online)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`

	return r.db.QueryRow(ctx, query,
		agent.ID, agent.UserID, agent.MaxTickets, agent.IsOnline,
	).Scan(&agent.CreatedAt)
}

func (r *AgentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.SupportAgent, error) {
	query := `
		SELECT id, user_id, max_tickets, is_online, created_at
		FROM support_agents WHERE id = $1`

	agent := &model.SupportAgent{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&agent.ID, &agent.UserID, &agent.MaxTickets, &agent.IsOnline, &agent.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get agent: %w", err)
	}
	return agent, nil
}

func (r *AgentRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.SupportAgent, error) {
	query := `
		SELECT id, user_id, max_tickets, is_online, created_at
		FROM support_agents WHERE user_id = $1`

	agent := &model.SupportAgent{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&agent.ID, &agent.UserID, &agent.MaxTickets, &agent.IsOnline, &agent.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get agent by user: %w", err)
	}
	return agent, nil
}

func (r *AgentRepository) List(ctx context.Context) ([]model.SupportAgent, error) {
	query := `
		SELECT id, user_id, max_tickets, is_online, created_at
		FROM support_agents ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list agents: %w", err)
	}
	defer rows.Close()

	var agents []model.SupportAgent
	for rows.Next() {
		var a model.SupportAgent
		if err := rows.Scan(&a.ID, &a.UserID, &a.MaxTickets, &a.IsOnline, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan agent: %w", err)
		}
		agents = append(agents, a)
	}
	return agents, nil
}

func (r *AgentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, isOnline bool) error {
	query := `UPDATE support_agents SET is_online = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, isOnline)
	return err
}

func (r *AgentRepository) GetAvailableAgent(ctx context.Context) (*model.SupportAgent, error) {
	query := `
		SELECT sa.id, sa.user_id, sa.max_tickets, sa.is_online, sa.created_at
		FROM support_agents sa
		LEFT JOIN (
			SELECT agent_id, COUNT(*) as ticket_count
			FROM tickets
			WHERE status IN ('open', 'in_progress') AND deleted_at IS NULL
			GROUP BY agent_id
		) tc ON sa.id = tc.agent_id
		WHERE sa.is_online = true
		ORDER BY COALESCE(tc.ticket_count, 0) ASC
		LIMIT 1`

	agent := &model.SupportAgent{}
	err := r.db.QueryRow(ctx, query).Scan(
		&agent.ID, &agent.UserID, &agent.MaxTickets, &agent.IsOnline, &agent.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get available agent: %w", err)
	}
	return agent, nil
}
