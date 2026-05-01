package service

import (
	"errors"
	"schedule/internal/model"
	"schedule/internal/repository"
	"time"
)

type Service struct {
	repo repository.Repo
}

func New(repo repository.Repo) *Service {
	return &Service{repo: repo}
}

type CreateJobInput struct {
	QuoteID      uint      `json:"quote_id" binding:"required"`
	TechnicianID uint      `json:"technician_id" binding:"required"`
	ManagerID    uint      `json:"manager_id" binding:"required"`
	StartAt      time.Time `json:"start_at" binding:"required"`
}

var (
	ErrStartInPast          = errors.New("start_at must be in the future")
	ErrNotFound             = errors.New("referenced resource not found")
	ErrConflict             = errors.New("technician has a conflicting job in this time window")
	ErrQuoteAlreadyAssigned = errors.New("quote is already assigned to a job")
	ErrAlreadyDone          = errors.New("job is already completed")
)

func (s *Service) GetUnscheduledQuotes() ([]model.Quote, error) {
	return s.repo.GetUnscheduledQuotes()
}

func (s *Service) CreateJob(input CreateJobInput) (*model.Job, error) {
	if !input.StartAt.After(time.Now()) {
		return nil, ErrStartInPast
	}

	if ok, err := s.repo.ExistsTechnician(input.TechnicianID); err != nil || !ok {
		return nil, ErrNotFound
	}
	if ok, err := s.repo.ExistsManager(input.ManagerID); err != nil || !ok {
		return nil, ErrNotFound
	}
	if ok, err := s.repo.ExistsQuote(input.QuoteID); err != nil || !ok {
		return nil, ErrNotFound
	}

	job := &model.Job{
		QuoteID:      input.QuoteID,
		TechnicianID: input.TechnicianID,
		ManagerID:    input.ManagerID,
		StartAt:      input.StartAt,
		EndAt:        input.StartAt.Add(2 * time.Hour),
		Status:       "scheduled",
	}

	if err := s.repo.CreateJobAndUpdateQuote(job); err != nil {
		switch {
		case errors.Is(err, repository.ErrConflict):
			return nil, ErrConflict
		case errors.Is(err, repository.ErrQuoteAlreadyAssigned):
			return nil, ErrQuoteAlreadyAssigned
		}
		return nil, err
	}

	return job, nil
}

func (s *Service) CompleteJob(id uint) error {
	job, err := s.repo.GetJob(id)
	if err != nil {
		return ErrNotFound
	}
	if job.Status == "completed" {
		return ErrAlreadyDone
	}
	return s.repo.CompleteJob(id)
}

func (s *Service) GetNotifications(userType string, userID uint) ([]model.Notification, error) {
	return s.repo.GetNotifications(userType, userID)
}
