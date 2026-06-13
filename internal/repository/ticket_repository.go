package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/user/kanban-saas/pkg/model"
)

type TicketRepository struct {
	db *pgxpool.Pool
}

func NewTicketRepository(db *pgxpool.Pool) *TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) Create(ctx context.Context, ticket *model.Ticket) error {
	query := `
		INSERT INTO tickets (id, workspace_id, subject, status, priority, source, channel_id, user_id, agent_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`

	return r.db.QueryRow(ctx, query,
		ticket.ID, ticket.WorkspaceID, ticket.Subject, ticket.Status,
		ticket.Priority, ticket.Source, ticket.ChannelID, ticket.UserID, ticket.AgentID,
	).Scan(&ticket.CreatedAt, &ticket.UpdatedAt)
}

func (r *TicketRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Ticket, error) {
	query := `
		SELECT id, workspace_id, subject, status, priority, source, channel_id, user_id, agent_id, created_at, updated_at, resolved_at
		FROM tickets WHERE id = $1 AND deleted_at IS NULL`

	ticket := &model.Ticket{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ticket.ID, &ticket.WorkspaceID, &ticket.Subject, &ticket.Status,
		&ticket.Priority, &ticket.Source, &ticket.ChannelID, &ticket.UserID,
		&ticket.AgentID, &ticket.CreatedAt, &ticket.UpdatedAt, &ticket.ResolvedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get ticket: %w", err)
	}
	return ticket, nil
}

func (r *TicketRepository) List(ctx context.Context, filter model.TicketFilter) ([]model.Ticket, error) {
	query := `
		SELECT id, workspace_id, subject, status, priority, source, channel_id, user_id, agent_id, created_at, updated_at, resolved_at
		FROM tickets WHERE deleted_at IS NULL`

	args := []interface{}{}
	argIdx := 1

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.Priority != nil {
		query += fmt.Sprintf(" AND priority = $%d", argIdx)
		args = append(args, *filter.Priority)
		argIdx++
	}
	if filter.Source != nil {
		query += fmt.Sprintf(" AND source = $%d", argIdx)
		args = append(args, *filter.Source)
		argIdx++
	}
	if filter.AgentID != nil {
		query += fmt.Sprintf(" AND agent_id = $%d", argIdx)
		args = append(args, *filter.AgentID)
		argIdx++
	}
	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, *filter.UserID)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
		argIdx++
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	defer rows.Close()

	var tickets []model.Ticket
	for rows.Next() {
		var t model.Ticket
		if err := rows.Scan(
			&t.ID, &t.WorkspaceID, &t.Subject, &t.Status,
			&t.Priority, &t.Source, &t.ChannelID, &t.UserID,
			&t.AgentID, &t.CreatedAt, &t.UpdatedAt, &t.ResolvedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ticket: %w", err)
		}
		tickets = append(tickets, t)
	}
	return tickets, nil
}

func (r *TicketRepository) Update(ctx context.Context, ticket *model.Ticket) error {
	query := `
		UPDATE tickets SET status = $2, priority = $3, agent_id = $4, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at`

	return r.db.QueryRow(ctx, query,
		ticket.ID, ticket.Status, ticket.Priority, ticket.AgentID,
	).Scan(&ticket.UpdatedAt)
}

func (r *TicketRepository) Close(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE tickets SET status = 'closed', resolved_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *TicketRepository) AddMessage(ctx context.Context, msg *model.TicketMessage) error {
	query := `
		INSERT INTO ticket_messages (id, ticket_id, sender_id, sender_type, content, platform, external_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at`

	return r.db.QueryRow(ctx, query,
		msg.ID, msg.TicketID, msg.SenderID, msg.SenderType,
		msg.Content, msg.Platform, msg.ExternalID,
	).Scan(&msg.CreatedAt)
}

func (r *TicketRepository) ListMessages(ctx context.Context, ticketID uuid.UUID) ([]model.TicketMessage, error) {
	query := `
		SELECT id, ticket_id, sender_id, sender_type, content, platform, external_id, created_at
		FROM ticket_messages WHERE ticket_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var messages []model.TicketMessage
	for rows.Next() {
		var msg model.TicketMessage
		if err := rows.Scan(
			&msg.ID, &msg.TicketID, &msg.SenderID, &msg.SenderType,
			&msg.Content, &msg.Platform, &msg.ExternalID, &msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (r *TicketRepository) AddStatusLog(ctx context.Context, log *model.TicketStatusLog) error {
	query := `
		INSERT INTO ticket_status_log (id, ticket_id, old_status, new_status, changed_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`

	return r.db.QueryRow(ctx, query,
		log.ID, log.TicketID, log.OldStatus, log.NewStatus, log.ChangedBy,
	).Scan(&log.CreatedAt)
}

func (r *TicketRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var openCount int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM tickets WHERE status = 'open' AND deleted_at IS NULL`).Scan(&openCount)
	if err != nil {
		return nil, err
	}
	stats["open_tickets"] = openCount

	var inProgressCount int
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM tickets WHERE status = 'in_progress' AND deleted_at IS NULL`).Scan(&inProgressCount)
	if err != nil {
		return nil, err
	}
	stats["in_progress_tickets"] = inProgressCount

	var totalCount int
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM tickets WHERE deleted_at IS NULL`).Scan(&totalCount)
	if err != nil {
		return nil, err
	}
	stats["total_tickets"] = totalCount

	return stats, nil
}
