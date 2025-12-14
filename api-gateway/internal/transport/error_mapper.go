package transport

import (
	"encoding/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

func handleGRPCError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	st, ok := status.FromError(err)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Internal server error",
		})
		return
	}

	httpStatus, errorCode := mapGRPCCodeToHTTP(st.Code())
	errorMessage := st.Message()

	if errorMessage == "" {
		errorMessage = st.Code().String()
	}

	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   errorMessage,
		Code:    errorCode,
		Message: errorMessage,
	})
}

func mapGRPCCodeToHTTP(grpcCode codes.Code) (int, string) {
	switch grpcCode {
	case codes.OK:
		return http.StatusOK, "OK"
	case codes.Canceled:
		return http.StatusRequestTimeout, "CANCELLED"
	case codes.Unknown:
		return http.StatusInternalServerError, "UNKNOWN"
	case codes.InvalidArgument:
		return http.StatusBadRequest, "INVALID_ARGUMENT"
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout, "DEADLINE_EXCEEDED"
	case codes.NotFound:
		return http.StatusNotFound, "NOT_FOUND"
	case codes.AlreadyExists:
		return http.StatusConflict, "ALREADY_EXISTS"
	case codes.PermissionDenied:
		return http.StatusForbidden, "PERMISSION_DENIED"
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests, "RESOURCE_EXHAUSTED"
	case codes.FailedPrecondition:
		return http.StatusBadRequest, "FAILED_PRECONDITION"
	case codes.Aborted:
		return http.StatusConflict, "ABORTED"
	case codes.OutOfRange:
		return http.StatusBadRequest, "OUT_OF_RANGE"
	case codes.Unimplemented:
		return http.StatusNotImplemented, "UNIMPLEMENTED"
	case codes.Internal:
		return http.StatusInternalServerError, "INTERNAL"
	case codes.Unavailable:
		return http.StatusServiceUnavailable, "UNAVAILABLE"
	case codes.DataLoss:
		return http.StatusInternalServerError, "DATA_LOSS"
	case codes.Unauthenticated:
		return http.StatusUnauthorized, "UNAUTHENTICATED"
	default:
		return http.StatusInternalServerError, "INTERNAL"
	}
}
