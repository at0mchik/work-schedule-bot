package models

import (
    "time"
)

type UserMonthlyStat struct {
    ID               uint      `gorm:"primarykey" json:"id"`
    UserID           uint      `gorm:"not null;index" json:"user_id"`
    Year             int       `gorm:"not null;index" json:"year"`
    Month            int       `gorm:"not null;check:month >= 1 AND month <= 12;index" json:"month"`
    
    // Плановые показатели
    PlannedDays      int       `gorm:"not null;default:0" json:"planned_days"`
    PlannedMinutes   int       `gorm:"not null;default:0" json:"planned_minutes"`
    
    // Фактические показатели
    WorkedDays       int       `gorm:"not null;default:0" json:"worked_days"`
    WorkedMinutes    int       `gorm:"not null;default:0" json:"worked_minutes"`
    OvertimeMinutes  int       `gorm:"not null;default:0" json:"overtime_minutes"`
    DeficitMinutes   int       `gorm:"not null;default:0" json:"deficit_minutes"`
    
    CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
    
    User             User      `gorm:"foreignKey:UserID"`
}

func (UserMonthlyStat) TableName() string {
    return "user_monthly_stats"
}

// CalculateStats вычисляет переработку и недобор
func (ums *UserMonthlyStat) CalculateStats() {
    diff := ums.WorkedMinutes - ums.PlannedMinutes
    if diff > 0 {
        ums.OvertimeMinutes = diff
        ums.DeficitMinutes = 0
    } else {
        ums.OvertimeMinutes = 0
        ums.DeficitMinutes = -diff
    }
}

// UpdateFromWorkSchedule обновляет плановые показатели из графика
func (ums *UserMonthlyStat) UpdateFromWorkSchedule(schedule *WorkSchedule) {
    ums.PlannedDays = schedule.WorkDays
    ums.PlannedMinutes = schedule.TotalMinutes
    ums.CalculateStats()
}

// UpdateWorkedTime обновляет отработанное время и пересчитывает статистику
func (ums *UserMonthlyStat) UpdateWorkedTime(workedDays int, workedMinutes int) {
    ums.WorkedDays = workedDays
    ums.WorkedMinutes = workedMinutes
    ums.CalculateStats()
}

// IsValid проверяет валидность данных
func (ums *UserMonthlyStat) IsValid() bool {
    if ums.Month < 1 || ums.Month > 12 {
        return false
    }
    if ums.PlannedDays < 0 {
        return false
    }
    if ums.PlannedMinutes < 0 {
        return false
    }
    if ums.WorkedDays < 0 {
        return false
    }
    if ums.WorkedMinutes < 0 {
        return false
    }
    return true
}