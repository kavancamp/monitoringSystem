package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/kavancamp/monitoringSystem/internal/database/db"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	q *db.Queries
}

func NewServer(q *db.Queries) *Server {
	return &Server{q: q}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/devices", s.handleDevices)

	return mux
}

type createDeviceRequest struct {
	Name       string `json:"name"`
	Site       string `json:"site"`
	DeviceType string `json:"device_type"`
}

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.createDevice(w, r)
	case http.MethodGet:
		s.listDevices(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) createDevice(w http.ResponseWriter, r *http.Request) {
	var req createDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Site = strings.TrimSpace(req.Site)
	req.DeviceType = strings.TrimSpace(req.DeviceType)

	if err := validateCreateDevice(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	created, err := s.q.CreateDevice(ctx, db.CreateDeviceParams{
		ID:         uuid.New(),
		Name:       req.Name,
		Site:       req.Site,
		DeviceType: req.DeviceType,
	})
	if err != nil {
		// temporary: expose error so we can debug
		http.Error(w, "failed to create device: "+err.Error(), http.StatusInternalServerError)
		log.Printf("CreateDevice error: %v", err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func validateCreateDevice(req createDeviceRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.Site == "" {
		return errors.New("site is required")
	}
	if req.DeviceType == "" {
		return errors.New("device_type is required")
	}
	return nil
}
func (s *Server) listDevices(w http.ResponseWriter, r *http.Request) {
	qp := r.URL.Query()

	siteParam := strings.TrimSpace(qp.Get("site"))
	statusParam := strings.TrimSpace(qp.Get("status"))

	limit := int32(50) // default limit
	offset := int32(0) // default offset

	if l := qp.Get("limit"); l != "" {
		n, err := strconv.Atoi(l)
		if err != nil || n <= 0 {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
		if n > 200 {
			n = 200
		}
		limit = int32(n)
	}

	if l := qp.Get("offset"); l != "" {
		n, err := strconv.Atoi(l)
		if err != nil || n < 0 {
			http.Error(w, "invalid offset parameter", http.StatusBadRequest)
			return
		}
		offset = int32(n)
	}

	site := textNull()
	if siteParam != "" {
		site = textVal(siteParam)
	}

	status := textNull()
	if statusParam != "" {
		status = textVal(statusParam)
	}

	devices, err := s.q.ListDevices(r.Context(), db.ListDevicesParams{
		Site:   site,
		Status: status,
		Lim:    limit,
		Off:    offset,
	})
	if err != nil {
		http.Error(w, "failed to list devies: "+err.Error(), http.StatusInternalServerError)
	}

	writeJSON(w, http.StatusOK, devices)
}
func textNull() pgtype.Text {
	return pgtype.Text{Valid: false}
}

func textVal(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// context timeouts l
//
//	func withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
//		return context.WithTimeout(ctx, 3_000_000_000) // 3 seconds
//	}
