// internal/repository/absence_period_repo.go
package repository

import (
	"time"
	"work-schedule-bot/internal/models"

	"gorm.io/gorm"
)

type AbsencePeriodRepository interface {
	Create(period *models.AbsencePeriod) error
	GetByID(id uint) (*models.AbsencePeriod, error)
	GetByUserID(userID uint) ([]models.AbsencePeriod, error)
	GetByUserIDAndType(userID uint, absenceType string) ([]models.AbsencePeriod, error)
	GetCurrentAbsence(userID uint, date time.Time) (*models.AbsencePeriod, error)
	CheckPeriodConflict(userID uint, startDate, endDate time.Time) (bool, error)
	Delete(id uint) error
	DeleteByUserID(userID uint) error
}

type GormAbsencePeriodRepository struct {
	db *gorm.DB
}

func NewGormAbsencePeriodRepository(db *gorm.DB) (AbsencePeriodRepository, error) {
	if err := db.AutoMigrate(&models.AbsencePeriod{}); err != nil {
		return nil, err
	}
	return &GormAbsencePeriodRepository{db: db}, nil
}

func (r *GormAbsencePeriodRepository) Create(period *models.AbsencePeriod) error {
	return r.db.Create(period).Error
}

func (r *GormAbsencePeriodRepository) GetByID(id uint) (*models.AbsencePeriod, error) {
	var period models.AbsencePeriod
	err := r.db.First(&period, id).Error
	if err != nil {
		return nil, err
	}
	return &period, nil
}

func (r *GormAbsencePeriodRepository) GetByUserID(userID uint) ([]models.AbsencePeriod, error) {
	var periods []models.AbsencePeriod
	err := r.db.Where("user_id = ?", userID).
		Order("start_date DESC").
		Find(&periods).Error
	return periods, err
}

func (r *GormAbsencePeriodRepository) GetByUserIDAndType(userID uint, absenceType string) ([]models.AbsencePeriod, error) {
	var periods []models.AbsencePeriod
	err := r.db.Where("user_id = ? AND type = ?", userID, absenceType).
		Order("start_date DESC").
		Find(&periods).Error
	return periods, err
}

func (r *GormAbsencePeriodRepository) GetCurrentAbsence(userID uint, date time.Time) (*models.AbsencePeriod, error) {
	var period models.AbsencePeriod
	err := r.db.Where("user_id = ? AND start_date <= ? AND end_date >= ?", 
		userID, date, date).
		First(&period).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &period, nil
}

func (r *GormAbsencePeriodRepository) CheckPeriodConflict(userID uint, startDate, endDate time.Time) (bool, error) {
	var count int64
	err := r.db.Model(&models.AbsencePeriod{}).
		Where("user_id = ? AND "+
			"(start_date BETWEEN ? AND ? OR "+
			"end_date BETWEEN ? AND ? OR "+
			"(start_date <= ? AND end_date >= ?))",
			userID,
			startDate, endDate,
			startDate, endDate,
			startDate, endDate).
		Count(&count).Error
	return count > 0, err
}

func (r *GormAbsencePeriodRepository) Delete(id uint) error {
	return r.db.Delete(&models.AbsencePeriod{}, id).Error
}

func (r *GormAbsencePeriodRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.AbsencePeriod{}).Error
}