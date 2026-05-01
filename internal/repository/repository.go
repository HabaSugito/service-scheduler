package repository

import (
	"schedule/internal/model"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetUnscheduledQuotes() ([]model.Quote, error) {
	var quotes []model.Quote
	err := r.db.Where("status = ?", "unscheduled").Find(&quotes).Error
	return quotes, err
}

func (r *Repository) ExistsTechnician(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Technician{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *Repository) ExistsManager(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Manager{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *Repository) ExistsQuote(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Quote{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// IsQuoteUnscheduled checks within a transaction that the quote is still unscheduled (SELECT FOR UPDATE).
func (r *Repository) IsQuoteUnscheduled(tx *gorm.DB, quoteID uint) (bool, error) {
	var quotes []model.Quote
	err := tx.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Where("id = ? AND status = 'unscheduled'", quoteID).
		Find(&quotes).Error
	return len(quotes) > 0, err
}

// HasOverlappingJob checks for time conflicts using SELECT FOR UPDATE (must be called within a transaction).
func (r *Repository) HasOverlappingJob(tx *gorm.DB, technicianID uint, start, end time.Time) (bool, error) {
	var jobs []model.Job
	err := tx.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Where("technician_id = ? AND status != 'completed' AND start_at < ? AND end_at > ?", technicianID, end, start).
		Find(&jobs).Error
	return len(jobs) > 0, err
}

func (r *Repository) CreateJobAndUpdateQuote(job *model.Job) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Lock the quote row to prevent duplicate assignment under concurrent requests.
		unscheduled, err := r.IsQuoteUnscheduled(tx, job.QuoteID)
		if err != nil {
			return err
		}
		if !unscheduled {
			return ErrQuoteAlreadyAssigned
		}

		// Lock existing jobs for the technician to prevent race conditions.
		overlaps, err := r.HasOverlappingJob(tx, job.TechnicianID, job.StartAt, job.EndAt)
		if err != nil {
			return err
		}
		if overlaps {
			return ErrConflict
		}

		if err := tx.Create(job).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Quote{}).Where("id = ?", job.QuoteID).Update("status", "scheduled").Error; err != nil {
			return err
		}

		notification := &model.Notification{
			UserType: "technician",
			UserID:   job.TechnicianID,
			Message:  "新しいJobが割り当てられました",
			JobID:    job.ID,
		}
		return tx.Create(notification).Error
	})
}

func (r *Repository) GetJob(id uint) (*model.Job, error) {
	var job model.Job
	err := r.db.First(&job, id).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *Repository) CompleteJob(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var job model.Job
		if err := tx.First(&job, id).Error; err != nil {
			return err
		}

		if err := tx.Model(&job).Update("status", "completed").Error; err != nil {
			return err
		}

		notification := &model.Notification{
			UserType: "manager",
			UserID:   job.ManagerID,
			Message:  "Jobが完了しました",
			JobID:    job.ID,
		}
		return tx.Create(notification).Error
	})
}

func (r *Repository) GetNotifications(userType string, userID uint) ([]model.Notification, error) {
	var notifications []model.Notification
	err := r.db.Where("user_type = ? AND user_id = ?", userType, userID).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}
