package repository

import (
	"schedule/internal/model"
)

type Repo interface {
	GetUnscheduledQuotes() ([]model.Quote, error)
	ExistsTechnician(id uint) (bool, error)
	ExistsManager(id uint) (bool, error)
	ExistsQuote(id uint) (bool, error)
	CreateJobAndUpdateQuote(job *model.Job) error
	GetJob(id uint) (*model.Job, error)
	CompleteJob(id uint) error
	GetNotifications(userType string, userID uint) ([]model.Notification, error)
}
