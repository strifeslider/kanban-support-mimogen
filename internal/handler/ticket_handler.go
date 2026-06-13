package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/user/kanban-saas/pkg/auth"
	apperr "github.com/user/kanban-saas/pkg/errors"
	"github.com/user/kanban-saas/pkg/model"
	"github.com/user/kanban-saas/services/support/internal/service"
)

type TicketHandler struct {
	supportService *service.SupportService
}

func NewTicketHandler(supportService *service.SupportService) *TicketHandler {
	return &TicketHandler{supportService: supportService}
}

func (h *TicketHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)

	var req model.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperr.RespondError(w, apperr.BadRequest("invalid request body"))
		return
	}

	if req.Subject == "" {
		apperr.RespondError(w, apperr.BadRequest("subject is required"))
		return
	}

	ticket, err := h.supportService.CreateTicket(r.Context(), userID, req)
	if err != nil {
		apperr.RespondError(w, err)
		return
	}

	apperr.RespondJSON(w, http.StatusCreated, ticket)
}

func (h *TicketHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := model.TicketFilter{}

	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}
	if priority := r.URL.Query().Get("priority"); priority != "" {
		filter.Priority = &priority
	}
	if source := r.URL.Query().Get("source"); source != "" {
		filter.Source = &source
	}
	if agentIDStr := r.URL.Query().Get("agent_id"); agentIDStr != "" {
		if agentID, err := uuid.Parse(agentIDStr); err == nil {
			filter.AgentID = &agentID
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	tickets, err := h.supportService.ListTickets(r.Context(), filter)
	if err != nil {
		apperr.RespondError(w, err)
		return
	}

	apperr.RespondJSON(w, http.StatusOK, tickets)
}

func (h *TicketHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperr.RespondError(w, apperr.BadRequest("invalid ticket id"))
		return
	}

	ticket, err := h.supportService.GetTicket(r.Context(), id)
	if err != nil {
		apperr.RespondError(w, err)
		return
	}

	apperr.RespondJSON(w, http.StatusOK, ticket)
}

func (h *TicketHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperr.RespondError(w, apperr.BadRequest("invalid ticket id"))
		return
	}

	userID := auth.GetUserID(r)

	var req model.UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperr.RespondError(w, apperr.BadRequest("invalid request body"))
		return
	}

	ticket, err := h.supportService.UpdateTicket(r.Context(), id, req, userID)
	if err != nil {
		apperr.RespondError(w, err)
		return
	}

	apperr.RespondJSON(w, http.StatusOK, ticket)
}

func (h *TicketHandler) Close(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperr.RespondError(w, apperr.BadRequest("invalid ticket id"))
		return
	}

	if err := h.supportService.CloseTicket(r.Context(), id); err != nil {
		apperr.RespondError(w, err)
		return
	}

	apperr.RespondJSON(w, http.StatusOK, map[string]string{"message": "ticket closed"})
}

func (h *TicketHandler) Assign(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperr.RespondError(w, apperr.BadRequest("invalid ticket id"))
		return
	}

	userID := auth.GetUserID(r)

	var req struct {
		AgentID uuid.UUID `json:"agent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperr.RespondError(w, apperr.BadRequest("invalid request body"))
		return
	}

	if err := h.supportService.AssignAgent(r.Context(), ticketID, req.AgentID, userID); err != nil {
		apperr.RespondError(w, err)
		return
	}

	apperr.RespondJSON(w, http.StatusOK, map[string]string{"message": "agent assigned"})
}

func (h *TicketHandler) AddMessage(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperr.RespondError(w, apperr.BadRequest("invalid ticket id"))
		return
	}

	userID := auth.GetUserID(r)

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperr.RespondError(w, apperr.BadRequest("invalid request body"))
		return
	}

	if req.Content == "" {
		apperr.RespondError(w, apperr.BadRequest("content is required"))
		return
	}

	msg, err := h.supportService.AddMessage(r.Context(), ticketID, userID, "agent", req.Content, nil)
	if err != nil {
		apperr.RespondError(w, err)
		return
	}

	apperr.RespondJSON(w, http.StatusCreated, msg)
}

func (h *TicketHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperr.RespondError(w, apperr.BadRequest("invalid ticket id"))
		return
	}

	messages, err := h.supportService.ListMessages(r.Context(), ticketID)
	if err != nil {
		apperr.RespondError(w, err)
		return
	}

	apperr.RespondJSON(w, http.StatusOK, messages)
}
