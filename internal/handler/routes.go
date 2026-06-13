package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/user/kanban-saas/pkg/auth"
)

func SetupRoutes(
	r chi.Router,
	ticketHandler *TicketHandler,
	agentHandler *AgentHandler,
	internalHandler *InternalEventHandler,
	jwtCfg auth.JWTConfig,
) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/internal/events", internalHandler.HandleEvent)

		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth(jwtCfg))

			r.Route("/tickets", func(r chi.Router) {
				r.Post("/", ticketHandler.Create)
				r.Get("/", ticketHandler.List)
				r.Get("/{id}", ticketHandler.Get)
				r.Put("/{id}", ticketHandler.Update)
				r.Post("/{id}/assign", ticketHandler.Assign)
				r.Post("/{id}/close", ticketHandler.Close)
				r.Post("/{id}/messages", ticketHandler.AddMessage)
				r.Get("/{id}/messages", ticketHandler.ListMessages)
			})

			r.Route("/agents", func(r chi.Router) {
				r.Get("/", agentHandler.List)
				r.Put("/{id}/status", agentHandler.UpdateStatus)
			})

			r.Get("/stats", agentHandler.Stats)
		})
	})
}
