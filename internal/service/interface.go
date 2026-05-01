package service

import "schedule/internal/model"

type Svc interface {
	GetUnscheduledQuotes() ([]model.Quote, error)
	CreateJob(input CreateJobInput) (*model.Job, error)
	CompleteJob(id uint) error
	GetNotifications(userType string, userID uint) ([]model.Notification, error)
}
