// internal/models/absence_period.go
package models

import "time"

type AbsencePeriod struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	StartDate time.Time `gorm:"type:date;not null" json:"start_date"`
	EndDate   time.Time `gorm:"type:date;not null" json:"end_date"`
	Type      string    `gorm:"type:varchar(20);not null" json:"type"` // vacation, sick_leave, day_off
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	User         User          `gorm:"foreignKey:UserID" json:"user"`
	WorkSessions []WorkSession `gorm:"foreignKey:AbsencePeriodID" json:"work_sessions"`
}

func (AbsencePeriod) TableName() string {
	return "absence_periods"
}

const (
	AbsenceTypeVacation  = "vacation"
	AbsenceTypeSickLeave = "sick_leave"
	AbsenceTypeDayOff    = "day_off"
)