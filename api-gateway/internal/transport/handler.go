package transport

import (
	"api-gateway/internal/infrastructure/analysis"
	"api-gateway/internal/infrastructure/storing"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	analysisClient *analysis.Client
	storingClient  *storing.Client
	logger         *zap.Logger
}

func NewHandler(analysisClient *analysis.Client, storingClient *storing.Client, logger *zap.Logger) *Handler {
	return &Handler{
		analysisClient: analysisClient,
		storingClient:  storingClient,
		logger:         logger,
	}
}

func (h *Handler) UploadTask(w http.ResponseWriter, r *http.Request) {
	req := &UploadTaskRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.logger.Warn("failed to decode upload task request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("upload task request",
		zap.String("filename", req.Filename),
		zap.String("uploaded_by", req.UploadedBy))

	res, err := h.storingClient.UploadTask(r.Context(), req.Filename, req.UploadedBy)
	if err != nil {
		h.logger.Error("failed to upload task", zap.Error(err))
		handleGRPCError(w, err)
		return
	}

	resp := &UploadTaskResponse{
		FileId:    res.FileId,
		UploadUrl: res.UploadUrl,
	}

	h.logger.Info("upload task success",
		zap.String("file_id", res.FileId))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode upload task response", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	taskId := chi.URLParam(r, "task_id")
	if taskId == "" {
		h.logger.Warn("get task request without task_id")
		http.Error(w, "task_id is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("get task request", zap.String("task_id", taskId))

	res, err := h.storingClient.GetTask(r.Context(), taskId)
	if err != nil {
		h.logger.Error("failed to get task", zap.String("task_id", taskId), zap.Error(err))
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

	h.logger.Info("get task success", zap.String("task_id", taskId))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode get task response", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) AnalyseTask(w http.ResponseWriter, r *http.Request) {
	req := &AnalyzeTaskRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.logger.Warn("failed to decode analyse task request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("analyse task request",
		zap.String("task_id", req.TaskId),
		zap.String("filename", req.Filename))

	res, err := h.analysisClient.AnalyseTask(r.Context(), req.TaskId, req.Filename)
	if err != nil {
		h.logger.Error("failed to analyse task",
			zap.String("task_id", req.TaskId),
			zap.Error(err))
		handleGRPCError(w, err)
		return
	}

	resp := &AnalyzeTaskResponse{
		Status: res.Status,
	}

	h.logger.Info("analyse task success",
		zap.String("task_id", req.TaskId),
		zap.Bool("status", res.Status))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode analyse task response", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	taskId := chi.URLParam(r, "task_id")
	if taskId == "" {
		h.logger.Warn("get report request without task_id")
		http.Error(w, "task_id is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("get report request", zap.String("task_id", taskId))

	res, err := h.analysisClient.GetReport(r.Context(), taskId)
	if err != nil {
		h.logger.Error("failed to get report",
			zap.String("task_id", taskId),
			zap.Error(err))
		handleGRPCError(w, err)
		return
	}

	resp := &GetReportResponse{
		TaskId:               res.TaskId,
		IsPlagiarism:         res.IsPlagiarism,
		PlagiarismPercentage: float64(res.PlagiarismPercentage),
	}

	h.logger.Info("get report success",
		zap.String("task_id", taskId),
		zap.Bool("is_plagiarism", res.IsPlagiarism),
		zap.Float64("plagiarism_percentage", float64(res.PlagiarismPercentage)))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode get report response", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) GetWordCloud(w http.ResponseWriter, r *http.Request) {
	taskId := chi.URLParam(r, "task_id")
	if taskId == "" {
		h.logger.Warn("get word cloud request without task_id")
		http.Error(w, "task_id is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("get word cloud request", zap.String("task_id", taskId))

	fileContent, err := h.storingClient.GetFileContent(r.Context(), taskId)
	if err != nil {
		h.logger.Error("failed to get file content",
			zap.String("task_id", taskId),
			zap.Error(err))
		handleGRPCError(w, err)
		return
	}

	wordCloudRes, err := h.analysisClient.GenerateWordCloud(r.Context(), fileContent.Content)
	if err != nil {
		h.logger.Error("failed to generate word cloud",
			zap.String("task_id", taskId),
			zap.Error(err))
		handleGRPCError(w, err)
		return
	}

	resp := map[string]string{
		"task_id":   taskId,
		"image_url": wordCloudRes.ImageUrl,
	}

	h.logger.Info("get word cloud success",
		zap.String("task_id", taskId),
		zap.String("image_url", wordCloudRes.ImageUrl))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode word cloud response", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
