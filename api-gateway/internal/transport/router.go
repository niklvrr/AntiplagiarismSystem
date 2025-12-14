package transport

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	middleware1 "api-gateway/internal/transport/middleware"

	middleware2 "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(handler *Handler, logger *zap.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware2.RequestID)
	router.Use(middleware1.LoggingMiddleware(logger))
	router.Use(middleware1.RecoveryMiddleware)

	SetupSwaggerRoutes(router)

	router.Route("/api/v1", func(r chi.Router) {
		r.Post("/task", handler.UploadTask)
		r.Get("/task/{task_id}", handler.GetTask)
		r.Post("/analyse", handler.AnalyseTask)
		r.Get("/report/{task_id}", handler.GetReport)
		r.Get("/wordcloud/{task_id}", handler.GetWordCloud)
	})
	return router
}
