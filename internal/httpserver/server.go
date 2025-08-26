package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"calendar/internal/calendar"
)

type Server struct {
	Svc *calendar.Service
}

func New(svc *calendar.Service) *Server {
	return &Server{Svc: svc}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/create_event", s.handleCreateEvent)
	mux.HandleFunc("/update_event", s.handleUpdateEvent)
	mux.HandleFunc("/delete_event", s.handleDeleteEvent)
	mux.HandleFunc("/events_for_day", s.handleEventsForDay)
	mux.HandleFunc("/events_for_week", s.handleEventsForWeek)
	mux.HandleFunc("/events_for_month", s.handleEventsForMonth)
	return mux
}

type errorResponse struct {
	Error string `json:"error"`
}

type resultResponse[T any] struct {
	Result T `json:"result"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func badRequest(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, errorResponse{Error: msg})
}

func bizError(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: msg})
}

func internalError(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusInternalServerError, errorResponse{Error: msg})
}

func parseUserID(values map[string]string) (int64, error) {
	userStr := values["user_id"]
	if userStr == "" {
		return 0, errors.New("missing user_id")
	}
	uid, err := strconv.ParseInt(userStr, 10, 64)
	if err != nil || uid <= 0 {
		return 0, errors.New("invalid user_id")
	}
	return uid, nil
}

func parseDate(values map[string]string) (time.Time, error) {
	ds := values["date"]
	if ds == "" {
		return time.Time{}, errors.New("missing date")
	}
	d, err := time.Parse("2006-01-02", ds)
	if err != nil {
		return time.Time{}, errors.New("invalid date")
	}
	return d, nil
}

func parseBodyOrForm(r *http.Request) (map[string]string, error) {
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		defer r.Body.Close()
		var m map[string]string
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			return nil, err
		}
		return m, nil
	}
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	res := make(map[string]string)
	for k := range r.Form {
		res[k] = r.Form.Get(k)
	}
	return res, nil
}

func (s *Server) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	vals, err := parseBodyOrForm(r)
	if err != nil {
		internalError(w, "failed to parse body")
		return
	}
	uid, err := parseUserID(vals)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	date, err := parseDate(vals)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	text := vals["event"]
	if text == "" {
		badRequest(w, "missing event")
		return
	}
	ev, err := s.Svc.CreateEvent(uid, date, text)
	if err != nil {
		bizError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resultResponse[calendar.Event]{Result: ev})
}

func (s *Server) handleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	vals, err := parseBodyOrForm(r)
	if err != nil {
		internalError(w, "failed to parse body")
		return
	}
	id := vals["id"]
	uid, err := parseUserID(vals)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	date, err := parseDate(vals)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	text := vals["event"]
	if text == "" {
		badRequest(w, "missing event")
		return
	}
	ev, err := s.Svc.UpdateEvent(id, uid, date, text)
	if err != nil {
		if errors.Is(err, calendar.ErrNotFound) || errors.Is(err, calendar.ErrInvalidUserID) || errors.Is(err, calendar.ErrInvalidText) {
			bizError(w, err.Error())
			return
		}
		internalError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resultResponse[calendar.Event]{Result: ev})
}

func (s *Server) handleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	vals, err := parseBodyOrForm(r)
	if err != nil {
		internalError(w, "failed to parse body")
		return
	}
	id := vals["id"]
	var uid int64
	if u := vals["user_id"]; u != "" {
		uid, _ = strconv.ParseInt(u, 10, 64)
	}
	if err := s.Svc.DeleteEvent(id, uid); err != nil {
		if errors.Is(err, calendar.ErrNotFound) {
			bizError(w, err.Error())
			return
		}
		internalError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resultResponse[string]{Result: "deleted"})
}

func (s *Server) handleEventsForDay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	uidStr := r.URL.Query().Get("user_id")
	dateStr := r.URL.Query().Get("date")
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil || uid <= 0 {
		badRequest(w, "invalid user_id")
		return
	}
	day, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		badRequest(w, "invalid date")
		return
	}
	events, err := s.Svc.EventsForDay(uid, day)
	if err != nil {
		bizError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resultResponse[[]calendar.Event]{Result: events})
}

func (s *Server) handleEventsForWeek(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	uidStr := r.URL.Query().Get("user_id")
	dateStr := r.URL.Query().Get("date")
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil || uid <= 0 {
		badRequest(w, "invalid user_id")
		return
	}
	day, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		badRequest(w, "invalid date")
		return
	}
	events, err := s.Svc.EventsForWeek(uid, day)
	if err != nil {
		bizError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resultResponse[[]calendar.Event]{Result: events})
}

func (s *Server) handleEventsForMonth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	uidStr := r.URL.Query().Get("user_id")
	dateStr := r.URL.Query().Get("date")
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil || uid <= 0 {
		badRequest(w, "invalid user_id")
		return
	}
	day, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		badRequest(w, "invalid date")
		return
	}
	events, err := s.Svc.EventsForMonth(uid, day)
	if err != nil {
		bizError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resultResponse[[]calendar.Event]{Result: events})
}
