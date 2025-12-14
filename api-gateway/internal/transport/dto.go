package transport

// ==== UPLOAD TASK ====
type UploadTaskRequest struct {
	Filename   string `json:"filename"`
	UploadedBy string `json:"uploaded_by"`
}

type UploadTaskResponse struct {
	FileId    string `json:"file_id"`
	UploadUrl string `json:"upload_url"`
}

// ==== GET TASK ====
type GetTaskResponse struct {
	FileId     string `json:"file_id"`
	Filename   string `json:"filename"`
	Url        string `json:"url"`
	UploadedBy string `json:"uploaded_by"`
	UploadedAt string `json:"uploaded_at"`
}

// ==== ANALYSE TASK ====
type AnalyzeTaskRequest struct {
	TaskId   string `json:"task_id"`
	Filename string `json:"filename"`
}

type AnalyzeTaskResponse struct {
	Status bool `json:"status"`
}

// ==== GET REPORT ====
type GetReportResponse struct {
	TaskId               string  `json:"task_id"`
	IsPlagiarism         bool    `json:"is_plagiarism"`
	PlagiarismPercentage float64 `json:"plagiarism_percentage"`
}
