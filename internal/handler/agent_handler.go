package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	apperr "github.com/user/kanban-saas/pkg/errors"
	"github.com/user/kanban-saas/pkg/model"
	"github.com/user/kanban-saas/services/support/internal/service"
)

type AgentHandler struct {
	supportService *service.SupportService
}

func NewAgentHandler(supportService *service.SupportService) *AgentHandler {
	return &AgentHandler{supportService: supportService}
}

func (h *AgentHandler) List(w http.ResponseWriter, r *http.Request) {
	agents, err := h.supportService.ListAgents(r.Context())
	if err != nil {
		apperr.RespondError(w, err)
		return
	}

	apperr.RespondJSON(w, http.StatusOK, agents)
}

func (h *AgentHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperr.RespondError(w, apperr.BadRequest("invalid agent id"))
		return
	}

	var req struct {
		IsOnline bool `json:"is_online"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperr.RespondError(w, apperr.BadRequest("invalid request body"))
		return
	}

	if err := h.supportService.UpdateAgentStatus(r.Context(), id, req.IsOnline); err != nil {
		apperr.RespondError(w, err)
		return
	}

	apperr.RespondJSON(w, http.StatusOK, map[string]string{"message": "status updated"})
}

func (h *AgentHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.supportService.GetStats(r.Context())
	if err != nil {
		apperr.RespondError(w, err)
		return
	}

	apperr.RespondJSON(w, http.StatusOK, stats)
}

type InternalEventHandler struct {
	supportService *service.SupportService
}

func NewInternalEventHandler(supportService *service.SupportService) *InternalEventHandler {
	return &InternalEventHandler{supportService: supportService}
}

func (h *InternalEventHandler) HandleEvent(w http.ResponseWriter, r *http.Request) {
	var event struct {
		Type      string      `json:"type"`
		Source    string      `json:"source"`
		ChannelID string      `json:"channel_id"`
		UserID    *uuid.UUID  `json:"user_id,omitempty"`
		Data      interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		apperr.RespondError(w, apperr.BadRequest("invalid event"))
		return
	}

	switch event.Type {
	case "ticket.created":
		var data struct {
			WorkspaceID uuid.UUID `json:"workspace_id"`
			Subject     string    `json:"subject"`
			Priority    string    `json:"priority"`
			Content     string    `json:"content"`
		}
		if err := json.Unmarshal(event.Data.([]byte), &data); err != nil {
			apperr.RespondError(w, apperr.BadRequest("invalid event data"))
			return
		}

		ticket, err := h.supportService.CreateTicketFromBot(r.Context(), model.CreateTicketRequest{
			WorkspaceID: data.WorkspaceID,
			Subject:     data.Subject,
			Priority:    data.Priority,
		}, event.Source, event.ChannelID, event.UserID)
		if err != nil {
			apperr.RespondError(w, err)
			return
		}

		if data.Content != "" && event.UserID != nil {
			h.supportService.AddMessage(r.Context(), ticket.ID, *event.UserID, "user", data.Content, &event.Source)
		}

		apperr.RespondJSON(w, http.StatusCreated, ticket)

	case "ticket.message":
		var data struct {
			TicketID   uuid.UUID `json:"ticket_id"`
			SenderID   uuid.UUID `json:"sender_id"`
			SenderType string    `json:"sender_type"`
			Content    string    `json:"content"`
			ExternalID string    `json:"external_id"`
		}
		if err := json.Unmarshal(event.Data.([]byte), &data); err != nil {
			apperr.RespondError(w, apperr.BadRequest("invalid event data"))
			return
		}

		msg, err := h.supportService.AddMessageFromBot(r.Context(), data.TicketID, data.SenderID, data.SenderType, data.Content, data.ExternalID, event.Source)
		if err != nil {
			apperr.RespondError(w, err)
			return
		}

		apperr.RespondJSON(w, http.StatusCreated, msg)

	default:
		apperr.RespondError(w, apperr.BadRequest("unknown event type"))
	}
}
