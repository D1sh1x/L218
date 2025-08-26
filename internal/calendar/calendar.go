package calendar

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID     string    `json:"id"`
	UserID int64     `json:"user_id"`
	Date   time.Time `json:"date"`
	Text   string    `json:"event"`
}

var (
	ErrNotFound      = errors.New("event not found")
	ErrInvalidUserID = errors.New("invalid user id")
	ErrInvalidDate   = errors.New("invalid date")
	ErrInvalidText   = errors.New("invalid text")
)

type Service struct {
	mu              sync.RWMutex
	userToEventsMap map[int64]map[string]Event
}

func NewService() *Service {
	return &Service{userToEventsMap: make(map[int64]map[string]Event)}
}

func normalizeDate(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

func (s *Service) CreateEvent(userID int64, date time.Time, text string) (Event, error) {
	if userID <= 0 {
		return Event{}, ErrInvalidUserID
	}
	if text == "" {
		return Event{}, ErrInvalidText
	}
	date = normalizeDate(date)

	newEvent := Event{
		ID:     uuid.NewString(),
		UserID: userID,
		Date:   date,
		Text:   text,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.userToEventsMap[userID]; !ok {
		s.userToEventsMap[userID] = make(map[string]Event)
	}
	s.userToEventsMap[userID][newEvent.ID] = newEvent
	return newEvent, nil
}

func (s *Service) UpdateEvent(id string, userID int64, date time.Time, text string) (Event, error) {
	if id == "" {
		return Event{}, ErrNotFound
	}
	if userID <= 0 {
		return Event{}, ErrInvalidUserID
	}
	if text == "" {
		return Event{}, ErrInvalidText
	}
	date = normalizeDate(date)

	s.mu.Lock()
	defer s.mu.Unlock()

	userEvents, ok := s.userToEventsMap[userID]
	if !ok {
		return Event{}, ErrNotFound
	}
	ev, ok := userEvents[id]
	if !ok {
		return Event{}, ErrNotFound
	}
	ev.Date = date
	ev.Text = text
	userEvents[id] = ev
	return ev, nil
}

func (s *Service) DeleteEvent(id string, userID int64) error {
	if id == "" {
		return ErrNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if userID > 0 {
		userEvents, ok := s.userToEventsMap[userID]
		if !ok {
			return ErrNotFound
		}
		if _, ok := userEvents[id]; !ok {
			return ErrNotFound
		}
		delete(userEvents, id)
		return nil
	}

	for uid, userEvents := range s.userToEventsMap {
		if _, ok := userEvents[id]; ok {
			delete(userEvents, id)
			if len(userEvents) == 0 {
				delete(s.userToEventsMap, uid)
			}
			return nil
		}
	}
	return ErrNotFound
}

func (s *Service) EventsForDay(userID int64, day time.Time) ([]Event, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}
	day = normalizeDate(day)
	s.mu.RLock()
	defer s.mu.RUnlock()
	userEvents, ok := s.userToEventsMap[userID]
	if !ok {
		return []Event{}, nil
	}
	res := make([]Event, 0)
	for _, ev := range userEvents {
		if ev.Date.Equal(day) {
			res = append(res, ev)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Date.Before(res[j].Date) || (res[i].Date.Equal(res[j].Date) && res[i].ID < res[j].ID)
	})
	return res, nil
}

func (s *Service) EventsForWeek(userID int64, anyDay time.Time) ([]Event, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}
	anyDay = normalizeDate(anyDay)
	weekday := int(anyDay.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := anyDay.AddDate(0, 0, -(weekday - 1))
	end := start.AddDate(0, 0, 7)

	s.mu.RLock()
	defer s.mu.RUnlock()
	userEvents, ok := s.userToEventsMap[userID]
	if !ok {
		return []Event{}, nil
	}
	res := make([]Event, 0)
	for _, ev := range userEvents {
		if (ev.Date.Equal(start) || ev.Date.After(start)) && ev.Date.Before(end) {
			res = append(res, ev)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Date.Before(res[j].Date) || (res[i].Date.Equal(res[j].Date) && res[i].ID < res[j].ID)
	})
	return res, nil
}

func (s *Service) EventsForMonth(userID int64, anyDay time.Time) ([]Event, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}
	anyDay = normalizeDate(anyDay)
	start := time.Date(anyDay.Year(), anyDay.Month(), 1, 0, 0, 0, 0, anyDay.Location())
	end := start.AddDate(0, 1, 0)

	s.mu.RLock()
	defer s.mu.RUnlock()
	userEvents, ok := s.userToEventsMap[userID]
	if !ok {
		return []Event{}, nil
	}
	res := make([]Event, 0)
	for _, ev := range userEvents {
		if (ev.Date.Equal(start) || ev.Date.After(start)) && ev.Date.Before(end) {
			res = append(res, ev)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Date.Before(res[j].Date) || (res[i].Date.Equal(res[j].Date) && res[i].ID < res[j].ID)
	})
	return res, nil
}
