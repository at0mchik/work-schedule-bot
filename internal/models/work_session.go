package models

import (
	"fmt"
	"time"
)

type WorkSession struct {
    ID              uint       `gorm:"primarykey" json:"id"`
    UserID          uint       `gorm:"not null;index" json:"user_id"`
    Date            time.Time  `gorm:"type:date;not null;index" json:"date"`
    
    // Время прихода/ухода
    ClockInTime     time.Time  `gorm:"not null" json:"clock_in_time"`
    ClockOutTime    *time.Time `json:"clock_out_time"`
    
    // Плановые показатели
    RequiredMinutes int        `gorm:"not null;default:480" json:"required_minutes"`
    
    // Фактические показатели (рассчитываются)
    WorkedMinutes   int        `gorm:"not null;default:0" json:"worked_minutes"`
    DiffMinutes     int        `gorm:"not null;default:0" json:"diff_minutes"`
    
    // Статус
    Status          string     `gorm:"type:varchar(20);not null;default:'active';index" json:"status"`
    
    Notes           string     `json:"notes"`
    CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
    
    User            User       `gorm:"foreignKey:UserID"`
}

func (WorkSession) TableName() string {
    return "work_sessions"
}

// Статусы рабочих сессий
const (
    StatusActive    = "active"    // На работе
    StatusCompleted = "completed" // Рабочий день завершен
    StatusAbsent    = "absent"    // Отсутствовал
)

// CalculateWorkedMinutes вычисляет отработанные минуты
func (ws *WorkSession) CalculateWorkedMinutes() int {
    if ws.ClockOutTime == nil || ws.ClockOutTime.IsZero() {
        return 0
    }
    
    duration := ws.ClockOutTime.Sub(ws.ClockInTime)
    minutes := int(duration.Minutes())
    
    // Округляем до целых минут
    return minutes
}

// CalculateDiffMinutes вычисляет разницу между отработанным и требуемым временем
func (ws *WorkSession) CalculateDiffMinutes() int {
    return ws.WorkedMinutes - ws.RequiredMinutes
}

// UpdateCalculatedFields обновляет вычисляемые поля
func (ws *WorkSession) UpdateCalculatedFields() {
    ws.WorkedMinutes = ws.CalculateWorkedMinutes()
    ws.DiffMinutes = ws.CalculateDiffMinutes()
    
    // Обновляем статус если пользователь вышел
    if ws.ClockOutTime != nil && !ws.ClockOutTime.IsZero() {
        ws.Status = StatusCompleted
    }
}

// IsActive проверяет, активна ли сессия (пользователь на работе)
func (ws *WorkSession) IsActive() bool {
    return ws.Status == StatusActive && ws.ClockOutTime == nil
}

// IsCompleted проверяет, завершена ли сессия
func (ws *WorkSession) IsCompleted() bool {
    return ws.Status == StatusCompleted
}

// IsToday проверяет, является ли дата сессии сегодняшней
func (ws *WorkSession) IsToday() bool {
    now := time.Now()
    return ws.Date.Year() == now.Year() && 
           ws.Date.Month() == now.Month() && 
           ws.Date.Day() == now.Day()
}

// Duration возвращает продолжительность работы как строку
func (ws *WorkSession) Duration() string {
    if ws.ClockOutTime == nil || ws.ClockOutTime.IsZero() {
        return "еще на работе"
    }
    
    duration := ws.ClockOutTime.Sub(ws.ClockInTime)
    hours := int(duration.Hours())
    minutes := int(duration.Minutes()) % 60
    
    if minutes == 0 {
        return fmt.Sprintf("%dч", hours)
    }
    return fmt.Sprintf("%dч %dм", hours, minutes)
}

// FormatTime форматирует время для отображения
func (ws *WorkSession) FormatTime() string {
    if ws.ClockOutTime == nil || ws.ClockOutTime.IsZero() {
        return fmt.Sprintf("⏰ Пришел: %s", ws.ClockInTime.Format("15:04"))
    }
    
    inTime := ws.ClockInTime.Format("15:04")
    outTime := ws.ClockOutTime.Format("15:04")
    return fmt.Sprintf("⏰ Пришел: %s | Ушел: %s", inTime, outTime)
}

// IsValid проверяет валидность данных
func (ws *WorkSession) IsValid() bool {
    if ws.UserID == 0 {
        return false
    }
    if ws.Date.IsZero() {
        return false
    }
    if ws.ClockInTime.IsZero() {
        return false
    }
    if ws.RequiredMinutes <= 0 {
        return false
    }
    if ws.Status != StatusActive && ws.Status != StatusCompleted && ws.Status != StatusAbsent {
        return false
    }
    return true
}