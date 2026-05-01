package service_test

import (
	"errors"
	"schedule/internal/model"
	"schedule/internal/repository"
	"schedule/internal/service"
	"testing"
	"time"
)

// mockRepo is a manual mock implementing repository.Repo.
type mockRepo struct {
	existsTechnician      func(id uint) (bool, error)
	existsManager         func(id uint) (bool, error)
	existsQuote           func(id uint) (bool, error)
	createJobAndUpdateQuote func(job *model.Job) error
	getJob                func(id uint) (*model.Job, error)
	completeJob           func(id uint) error
	getUnscheduledQuotes  func() ([]model.Quote, error)
	getNotifications      func(userType string, userID uint) ([]model.Notification, error)
}

func (m *mockRepo) ExistsTechnician(id uint) (bool, error) { return m.existsTechnician(id) }
func (m *mockRepo) ExistsManager(id uint) (bool, error)    { return m.existsManager(id) }
func (m *mockRepo) ExistsQuote(id uint) (bool, error)      { return m.existsQuote(id) }
func (m *mockRepo) CreateJobAndUpdateQuote(job *model.Job) error {
	return m.createJobAndUpdateQuote(job)
}
func (m *mockRepo) GetJob(id uint) (*model.Job, error) { return m.getJob(id) }
func (m *mockRepo) CompleteJob(id uint) error          { return m.completeJob(id) }
func (m *mockRepo) GetUnscheduledQuotes() ([]model.Quote, error) {
	return m.getUnscheduledQuotes()
}
func (m *mockRepo) GetNotifications(userType string, userID uint) ([]model.Notification, error) {
	return m.getNotifications(userType, userID)
}

func futureTime() time.Time { return time.Now().Add(24 * time.Hour) }
func pastTime() time.Time   { return time.Now().Add(-1 * time.Hour) }

func defaultMock() *mockRepo {
	return &mockRepo{
		existsTechnician:        func(id uint) (bool, error) { return true, nil },
		existsManager:           func(id uint) (bool, error) { return true, nil },
		existsQuote:             func(id uint) (bool, error) { return true, nil },
		createJobAndUpdateQuote: func(job *model.Job) error { return nil },
	}
}

// --- CreateJob tests ---

func TestCreateJob_StartInPast(t *testing.T) {
	svc := service.New(defaultMock())
	_, err := svc.CreateJob(service.CreateJobInput{
		QuoteID: 1, TechnicianID: 1, ManagerID: 1, StartAt: pastTime(),
	})
	if !errors.Is(err, service.ErrStartInPast) {
		t.Fatalf("expected ErrStartInPast, got %v", err)
	}
}

func TestCreateJob_TechnicianNotFound(t *testing.T) {
	m := defaultMock()
	m.existsTechnician = func(id uint) (bool, error) { return false, nil }
	svc := service.New(m)
	_, err := svc.CreateJob(service.CreateJobInput{
		QuoteID: 1, TechnicianID: 99, ManagerID: 1, StartAt: futureTime(),
	})
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateJob_ManagerNotFound(t *testing.T) {
	m := defaultMock()
	m.existsManager = func(id uint) (bool, error) { return false, nil }
	svc := service.New(m)
	_, err := svc.CreateJob(service.CreateJobInput{
		QuoteID: 1, TechnicianID: 1, ManagerID: 99, StartAt: futureTime(),
	})
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateJob_QuoteNotFound(t *testing.T) {
	m := defaultMock()
	m.existsQuote = func(id uint) (bool, error) { return false, nil }
	svc := service.New(m)
	_, err := svc.CreateJob(service.CreateJobInput{
		QuoteID: 99, TechnicianID: 1, ManagerID: 1, StartAt: futureTime(),
	})
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateJob_Conflict(t *testing.T) {
	m := defaultMock()
	m.createJobAndUpdateQuote = func(job *model.Job) error { return repository.ErrConflict }
	svc := service.New(m)
	_, err := svc.CreateJob(service.CreateJobInput{
		QuoteID: 1, TechnicianID: 1, ManagerID: 1, StartAt: futureTime(),
	})
	if !errors.Is(err, service.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestCreateJob_QuoteAlreadyAssigned(t *testing.T) {
	m := defaultMock()
	m.createJobAndUpdateQuote = func(job *model.Job) error { return repository.ErrQuoteAlreadyAssigned }
	svc := service.New(m)
	_, err := svc.CreateJob(service.CreateJobInput{
		QuoteID: 1, TechnicianID: 1, ManagerID: 1, StartAt: futureTime(),
	})
	if !errors.Is(err, service.ErrQuoteAlreadyAssigned) {
		t.Fatalf("expected ErrQuoteAlreadyAssigned, got %v", err)
	}
}

func TestCreateJob_EndAtIs2HoursAfterStartAt(t *testing.T) {
	var captured *model.Job
	m := defaultMock()
	m.createJobAndUpdateQuote = func(job *model.Job) error {
		captured = job
		return nil
	}
	svc := service.New(m)
	start := futureTime()
	_, err := svc.CreateJob(service.CreateJobInput{
		QuoteID: 1, TechnicianID: 1, ManagerID: 1, StartAt: start,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.EndAt != start.Add(2*time.Hour) {
		t.Fatalf("expected EndAt = StartAt+2h, got %v", captured.EndAt)
	}
}

// --- CompleteJob tests ---

func TestCompleteJob_NotFound(t *testing.T) {
	m := defaultMock()
	m.getJob = func(id uint) (*model.Job, error) { return nil, errors.New("not found") }
	svc := service.New(m)
	err := svc.CompleteJob(99)
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCompleteJob_AlreadyDone(t *testing.T) {
	m := defaultMock()
	m.getJob = func(id uint) (*model.Job, error) {
		return &model.Job{Status: "completed"}, nil
	}
	svc := service.New(m)
	err := svc.CompleteJob(1)
	if !errors.Is(err, service.ErrAlreadyDone) {
		t.Fatalf("expected ErrAlreadyDone, got %v", err)
	}
}

func TestCompleteJob_Success(t *testing.T) {
	m := defaultMock()
	m.getJob = func(id uint) (*model.Job, error) {
		return &model.Job{Status: "scheduled"}, nil
	}
	m.completeJob = func(id uint) error { return nil }
	svc := service.New(m)
	if err := svc.CompleteJob(1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
