package transport

import (
	"api-gateway/internal/infrastructure/analysis"
	"api-gateway/internal/infrastructure/storing"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Handler struct {
	analysisClient *analysis.Client
	storingClient  *storing.Client
}

func NewHandler(analysisClient *analysis.Client, storingClient *storing.Client) *Handler {
	return &Handler{
		analysisClient: analysisClient,
		storingClient:  storingClient,
	}
}

func (h *Handler) UploadTask(w http.ResponseWriter, r *http.Request) {
	req := &UploadTaskRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.storingClient.UploadTask(r.Context(), req.Filename, req.UploadedBy)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	resp := &UploadTaskResponse{
		FileId:    res.FileId,
		UploadUrl: res.UploadUrl,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	taskId := chi.URLParam(r, "task_id")
	if taskId == "" {
		http.Error(w, "task_id is required", http.StatusBadRequest)
		return
	}
	res, err := h.storingClient.GetTask(r.Context(), taskId)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	resp := &GetTaskResponse{
		FileId:     res.FileId,
		Filename:   res.Filename,
		Url:        res.Url,
		UploadedBy: res.UploadedBy,
		UploadedAt: res.UploadedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) AnalyseTask(w http.ResponseWriter, r *http.Request) {
	req := &AnalyzeTaskRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.analysisClient.AnalyseTask(r.Context(), req.TaskId, req.Filename)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	resp := &AnalyzeTaskResponse{
		Status: res.Status,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	taskId := chi.URLParam(r, "task_id")
	if taskId == "" {
		http.Error(w, "task_id is required", http.StatusBadRequest)
		return
	}

	res, err := h.analysisClient.GetReport(r.Context(), taskId)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	resp := &GetReportResponse{
		TaskId:               res.TaskId,
		IsPlagiarism:         res.IsPlagiarism,
		PlagiarismPercentage: float64(res.PlagiarismPercentage),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
