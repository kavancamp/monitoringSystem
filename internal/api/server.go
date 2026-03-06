package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/kavancamp/monitoringSystem/internal/database/db"
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
	mux.HandleFunc("/devices/", s.handleDeviceByID)
	mux.HandleFunc("/readings", s.handleReadings)

	return mux
}

type createDeviceRequest struct {
	Name       string `json:"name"`
	Site       string `json:"site"`
	DeviceType string `json:"device_type"`
}
type createReadingRequest struct {
	DeviceID     string   `json:"device_id"`
	TemperatureC *float64 `json:"temperature_c"`
	PressureKpa  *float64 `json:"pressure_kpa"`
	Rpm          *float64 `json:"rpm"`
	Vibration    *float64 `json:"vibration"`
	//Payload      []byte    `json:"payload"`
}

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.createDevice(w, r)
	case http.MethodGet:
		s.listDevices(w, r)
	default:
		w.Header().Set("Allow", "GET, POST") // Inform allowed methods
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
		http.Error(w, "failed to list devices: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, devices)
}
func (s *Server) handleDeviceByID(w http.ResponseWriter, r *http.Request) {
	// Validate method
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet) // Inform allowed method
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Extract device ID from URL path
	id := strings.TrimPrefix(r.URL.Path, "/devices/")
	if id == "" {
		http.Error(w, "device_id is required", http.StatusBadRequest)
		return
	}
	if strings.Contains(id, "/") {
		http.Error(w, "Invalid route", http.StatusBadRequest)
		return
	}
	parsedID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid device ID format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	device, err := s.q.GetDevice(ctx, parsedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Device not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, device)
}

// Handle POST /readings
func (s *Server) handleReadings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req createReadingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.DeviceID = strings.TrimSpace(req.DeviceID)
	if req.DeviceID == "" {
		http.Error(w, "device_id is required", http.StatusBadRequest)
		return
	}

	deviceID, err := uuid.Parse(req.DeviceID)
	if err != nil {
		http.Error(w, "invalid device_id format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	// Check if device exists (404 if not)

	_, err = s.q.GetDevice(ctx, deviceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Reading not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	//Current UTC time
	ts := time.Now().UTC()

	params := db.InsertReadingParams{
		DeviceID:     deviceID,
		Ts:           timestamptzVal(ts),
		TemperatureC: floatFromPtr(req.TemperatureC),
		PressureKpa:  floatFromPtr(req.PressureKpa),
		Rpm:          floatFromPtr(req.Rpm),
		Vibration:    floatFromPtr(req.Vibration),
		Payload:      []byte("{}"),
	}

	reading, err := s.q.InsertReading(ctx, params)
	if err != nil {
		log.Printf("InsertReading error: %v", err)
		http.Error(w, "Failed to insert reading", http.StatusInternalServerError)
		return
	}

	// Update heartbeat
	if err := s.q.UpdateDeviceLastSeen(ctx, db.UpdateDeviceLastSeenParams{
		ID:         deviceID,
		LastSeenAt: timestamptzVal(ts),
	}); err != nil {
		log.Printf("UpdateDeviceLastSeen error: %v", err)
		http.Error(w, "failed to update device heartbeat", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, reading)
}

func textNull() pgtype.Text {
	return pgtype.Text{Valid: false}
}

func textVal(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func floatNull() pgtype.Float8 {
	return pgtype.Float8{Valid: false}
}
func floatVal(f float64) pgtype.Float8 {
	return pgtype.Float8{Float64: f, Valid: true}
}
func floatFromPtr(p *float64) pgtype.Float8 {
	if p == nil {
		return floatNull()
	}
	return floatVal(*p)
}
func timestamptzVal(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
