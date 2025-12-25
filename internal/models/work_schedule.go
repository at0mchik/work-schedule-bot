package models

import (
    "time"
)

type WorkSchedule struct {
    ID                uint      `gorm:"primarykey" json:"id"`
    Year              int       `gorm:"not null;index" json:"year"`
    Month             int       `gorm:"not null;check:month >= 1 AND month <= 12;index" json:"month"`
    WorkDays          int       `gorm:"not null;default:0" json:"work_days"`
    WorkMinutesPerDay int       `gorm:"not null;default:480" json:"work_minutes_per_day"` // 8 часов = 480 минут
    TotalMinutes      int       `gorm:"not null;default:0" json:"total_minutes"`          // work_days * work_minutes_per_day
    CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (WorkSchedule) TableName() string {
    return "work_schedules"
}

// CalculateTotalMinutes вычисляет общее количество минут в месяце
func (ws *WorkSchedule) CalculateTotalMinutes() int {
    return ws.WorkDays * ws.WorkMinutesPerDay
}

// BeforeSave хук для пересчета total_minutes перед сохранением
func (ws *WorkSchedule) BeforeSave() error {
    ws.TotalMinutes = ws.CalculateTotalMinutes()
    return nil
}

// IsValid проверяет валидность данных
func (ws *WorkSchedule) IsValid() bool {
    if ws.Year < 2000 || ws.Year > 2100 {
        return false
    }
    if ws.Month < 1 || ws.Month > 12 {
        return false
    }
    if ws.WorkDays < 0 || ws.WorkDays > 31 {
        return false
    }
    if ws.WorkMinutesPerDay <= 0 || ws.WorkMinutesPerDay > 1440 { // 24 часа
        return false
    }
    return true
}