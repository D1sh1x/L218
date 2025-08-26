package calendar

import (
	"testing"
	"time"
)

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("parse date: %v", err)
	}
	return d
}

func TestCreateAndGetByDay(t *testing.T) {
	s := NewService()
	day := mustDate(t, "2023-12-31")
	if _, err := s.CreateEvent(1, day, "New Year Eve"); err != nil {
		t.Fatalf("CreateEvent: %v", err)
	}
	events, err := s.EventsForDay(1, day)
	if err != nil {
		t.Fatalf("EventsForDay: %v", err)
	}
	if len(events) != 1 || events[0].Text != "New Year Eve" {
		t.Fatalf("unexpected events: %+v", events)
	}
}

func TestWeekAndMonth(t *testing.T) {
	s := NewService()
	_ = mustDate(t, "2023-12-25")
	if _, err := s.CreateEvent(1, mustDate(t, "2023-12-25"), "Mon"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateEvent(1, mustDate(t, "2023-12-31"), "Sun"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateEvent(1, mustDate(t, "2024-01-01"), "NextMonth"); err != nil {
		t.Fatal(err)
	}

	weekEvents, err := s.EventsForWeek(1, mustDate(t, "2023-12-27"))
	if err != nil {
		t.Fatal(err)
	}
	if len(weekEvents) != 2 {
		t.Fatalf("want 2 events in week, got %d", len(weekEvents))
	}
	monthEvents, err := s.EventsForMonth(1, mustDate(t, "2023-12-10"))
	if err != nil {
		t.Fatal(err)
	}
	if len(monthEvents) != 2 {
		t.Fatalf("want 2 events in Dec, got %d", len(monthEvents))
	}
}

func TestUpdateAndDelete(t *testing.T) {
	s := NewService()
	ev, err := s.CreateEvent(2, mustDate(t, "2023-03-05"), "Text")
	if err != nil {
		t.Fatal(err)
	}
	ev2, err := s.UpdateEvent(ev.ID, 2, mustDate(t, "2023-03-06"), "Text2")
	if err != nil {
		t.Fatal(err)
	}
	if ev2.Text != "Text2" || !ev2.Date.Equal(mustDate(t, "2023-03-06")) {
		t.Fatalf("unexpected updated: %+v", ev2)
	}
	if err := s.DeleteEvent(ev.ID, 2); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteEvent(ev.ID, 2); !errorsIs(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func errorsIs(err, target error) bool {
	if err == nil {
		return target == nil
	}
	return err.Error() == target.Error()
}
