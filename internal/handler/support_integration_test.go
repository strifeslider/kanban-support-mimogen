package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/user/kanban-saas/pkg/auth"
)

func TestTicketHandler_List_NoAuth(t *testing.T) {
	h := &TicketHandler{}
	req := httptest.NewRequest("GET", "/api/v1/tickets", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestTicketHandler_Update_InvalidID(t *testing.T) {
	h := &TicketHandler{}
	body := map[string]string{"status": "closed"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/v1/tickets/invalid", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	h.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestTicketHandler_Close_InvalidID(t *testing.T) {
	h := &TicketHandler{}
	req := httptest.NewRequest("POST", "/api/v1/tickets/invalid/close", nil)
	w := httptest.NewRecorder()

	h.Close(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestTicketHandler_Assign_InvalidID(t *testing.T) {
	h := &TicketHandler{}
	body := map[string]string{"agent_id": "invalid"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/tickets/invalid/assign", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	h.Assign(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestTicketHandler_AddMessage_InvalidID(t *testing.T) {
	h := &TicketHandler{}
	body := map[string]string{"content": "test"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/tickets/invalid/messages", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	h.AddMessage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestTicketHandler_ListMessages_InvalidID(t *testing.T) {
	h := &TicketHandler{}
	req := httptest.NewRequest("GET", "/api/v1/tickets/invalid/messages", nil)
	w := httptest.NewRecorder()

	h.ListMessages(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestAgentHandler_List_NoAuth(t *testing.T) {
	h := &AgentHandler{}
	req := httptest.NewRequest("GET", "/api/v1/agents", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAgentHandler_UpdateStatus_InvalidID(t *testing.T) {
	h := &AgentHandler{}
	body := map[string]bool{"is_online": true}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/v1/agents/invalid/status", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	h.UpdateStatus(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestAgentHandler_Stats_NoAuth(t *testing.T) {
	h := &AgentHandler{}
	req := httptest.NewRequest("GET", "/api/v1/stats", nil)
	w := httptest.NewRecorder()

	h.Stats(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestInternalEventHandler_HandleEvent_UnknownType(t *testing.T) {
	h := &InternalEventHandler{}
	body := map[string]interface{}{
		"type":       "unknown.event",
		"source":     "test",
		"channel_id": "123",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/internal/events", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestSetupRoutes_Internal(t *testing.T) {
	r := chi.NewRouter()
	th := &TicketHandler{}
	ah := &AgentHandler{}
	ih := &InternalEventHandler{}
	jwtCfg := auth.JWTConfig{Secret: "test"}

	SetupRoutes(r, th, ah, ih, jwtCfg)

	routes := []string{
		"/api/v1/internal/events",
		"/api/v1/tickets",
		"/api/v1/agents",
		"/api/v1/stats",
	}

	for _, route := range routes {
		req := httptest.NewRequest("GET", route, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusNotFound {
			t.Errorf("route %s not found", route)
		}
	}
}
