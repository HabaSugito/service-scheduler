package model

import "time"

type Manager struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Technician struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Quote struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	Status      string    `gorm:"type:enum('unscheduled','scheduled');default:'unscheduled'" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Job struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	QuoteID      uint      `gorm:"not null" json:"quote_id"`
	TechnicianID uint      `gorm:"not null" json:"technician_id"`
	ManagerID    uint      `gorm:"not null" json:"manager_id"`
	StartAt      time.Time `gorm:"not null" json:"start_at"`
	EndAt        time.Time `gorm:"not null" json:"end_at"`
	Status       string    `gorm:"type:enum('scheduled','completed');default:'scheduled'" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Notification struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserType  string     `gorm:"type:enum('manager','technician');not null" json:"user_type"`
	UserID    uint       `gorm:"not null" json:"user_id"`
	Message   string     `gorm:"type:text;not null" json:"message"`
	JobID     uint       `gorm:"not null" json:"job_id"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
}
