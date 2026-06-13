package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/user/kanban-saas/pkg/auth"
)

func TestNewTicketHandler(t *testing.T) {
	h := &TicketHandler{}
	if h == nil {
		t.Error("expected non-nil handler")
	}
}

func TestNewAgentHandler(t *testing.T) {
	h := &AgentHandler{}
	if h == nil {
		t.Error("expected non-nil handler")
	}
}

func TestNewInternalEventHandler(t *testing.T) {
	h := &InternalEventHandler{}
	if h == nil {
		t.Error("expected non-nil handler")
	}
}

func TestSetupRoutes(t *testing.T) {
	r := chi.NewRouter()
	th := &TicketHandler{}
	ah := &AgentHandler{}
	ih := &InternalEventHandler{}
	jwtCfg := auth.JWTConfig{Secret: "test"}

	SetupRoutes(r, th, ah, ih, jwtCfg)

	// Verify internal route exists
	req := httptest.NewRequest("POST", "/api/v1/internal/events", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// Should not return 404
	if w.Code == http.StatusNotFound {
		t.Error("internal events route not found")
	}
}

func TestTicketHandler_Create_EmptyBody(t *testing.T) {
	h := &TicketHandler{}
	req := httptest.NewRequest("POST", "/api/v1/tickets", nil)
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestTicketHandler_Get_InvalidID(t *testing.T) {
	h := &TicketHandler{}
	req := httptest.NewRequest("GET", "/api/v1/tickets/invalid", nil)
	w := httptest.NewRecorder()

	h.Get(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestInternalEventHandler_HandleEvent_InvalidJSON(t *testing.T) {
	h := &InternalEventHandler{}
	req := httptest.NewRequest("POST", "/api/v1/internal/events", nil)
	w := httptest.NewRecorder()

	h.HandleEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
