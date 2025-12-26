package repository

import (
	"errors"
	"time"
	"work-schedule-bot/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserMonthlyStatRepository interface {
	Create(stat *models.UserMonthlyStat) error
	Update(stat *models.UserMonthlyStat) error
	GetByID(id uint) (*models.UserMonthlyStat, error)
	GetByUserID(userID uint) ([]*models.UserMonthlyStat, error)
	GetByUserAndMonth(userID uint, year, month int) (*models.UserMonthlyStat, error)
	GetByMonth(year, month int) ([]*models.UserMonthlyStat, error)
	UpdatePlannedStats(userID uint, year, month, plannedDays, plannedMinutes int) error
	UpdateWorkedStats(userID uint, year, month, workedDays, workedMinutes int) error
	DeleteByUserID(userID uint) error
	DeleteByID(id uint) error
	Exists(userID uint, year, month int) (bool, error)
	CreateForAllUsers(year, month, plannedDays, plannedMinutes int) error
	UpdateForAllUsers(year, month, plannedDays, plannedMinutes int) error
}

type GormUserMonthlyStatRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewGormUserMonthlyStatRepository(db *gorm.DB) (*GormUserMonthlyStatRepository, error) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Автомиграция
	if err := db.AutoMigrate(&models.UserMonthlyStat{}); err != nil {
		logger.WithError(err).Error("Failed to auto-migrate user_monthly_stats table")
		return nil, err
	}

	logger.Info("User monthly stat repository initialized")

	return &GormUserMonthlyStatRepository{
		db:     db,
		logger: logger,
	}, nil
}

func (r *GormUserMonthlyStatRepository) Create(stat *models.UserMonthlyStat) error {
	r.logger.WithFields(logrus.Fields{
		"user_id": stat.UserID,
		"year":    stat.Year,
		"month":   stat.Month,
	}).Debug("Creating user monthly stat")

	if !stat.IsValid() {
		r.logger.WithFields(logrus.Fields{
			"user_id": stat.UserID,
			"year":    stat.Year,
			"month":   stat.Month,
		}).Warn("Invalid monthly stat data")
		return errors.New("некорректные данные статистики")
	}

	// Проверяем существование
	exists, err := r.Exists(stat.UserID, stat.Year, stat.Month)
	if err != nil {
		return err
	}

	if exists {
		r.logger.WithFields(logrus.Fields{
			"user_id": stat.UserID,
			"year":    stat.Year,
			"month":   stat.Month,
		}).Debug("Monthly stat already exists")
		return errors.New("статистика за этот месяц уже существует")
	}

	// Вычисляем статистику
	stat.CalculateStats()

	result := r.db.Create(stat)
	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to create user monthly stat")
		return result.Error
	}

	r.logger.WithFields(logrus.Fields{
		"id":      stat.ID,
		"user_id": stat.UserID,
		"year":    stat.Year,
		"month":   stat.Month,
	}).Debug("User monthly stat created successfully")

	return nil
}

func (r *GormUserMonthlyStatRepository) Update(stat *models.UserMonthlyStat) error {
	r.logger.WithFields(logrus.Fields{
		"id":      stat.ID,
		"user_id": stat.UserID,
		"year":    stat.Year,
		"month":   stat.Month,
	}).Debug("Updating user monthly stat")

	if !stat.IsValid() {
		r.logger.WithFields(logrus.Fields{
			"id":      stat.ID,
			"user_id": stat.UserID,
		}).Warn("Invalid monthly stat data for update")
		return errors.New("некорректные данные статистики")
	}

	// Проверяем существование
	existingStat, err := r.GetByID(stat.ID)
	if err != nil {
		return err
	}

	if existingStat == nil {
		r.logger.WithField("id", stat.ID).Warn("Monthly stat not found for update")
		return errors.New("статистика не найдена")
	}

	// Вычисляем статистику
	stat.CalculateStats()
	stat.UpdatedAt = time.Now()

	result := r.db.Save(stat)
	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to update user monthly stat")
		return result.Error
	}

	r.logger.WithFields(logrus.Fields{
		"id":      stat.ID,
		"user_id": stat.UserID,
	}).Debug("User monthly stat updated successfully")

	return nil
}

func (r *GormUserMonthlyStatRepository) GetByID(id uint) (*models.UserMonthlyStat, error) {
	var stat models.UserMonthlyStat
	result := r.db.First(&stat, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		r.logger.WithField("id", id).Debug("Monthly stat not found")
		return nil, nil
	}

	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to get monthly stat by ID")
		return nil, result.Error
	}

	return &stat, nil
}

func (r *GormUserMonthlyStatRepository) GetByUserID(userID uint) ([]*models.UserMonthlyStat, error) {
	var stats []*models.UserMonthlyStat
	result := r.db.Where("user_id = ?", userID).Order("year ASC, month ASC").Find(&stats)

	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to get monthly stats by user ID")
		return nil, result.Error
	}

	r.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"count":   len(stats),
	}).Debug("Retrieved monthly stats by user ID")

	return stats, nil
}

func (r *GormUserMonthlyStatRepository) GetByUserAndMonth(userID uint, year, month int) (*models.UserMonthlyStat, error) {
	var stat models.UserMonthlyStat
	result := r.db.Where("user_id = ? AND year = ? AND month = ?", userID, year, month).First(&stat)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		r.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"year":    year,
			"month":   month,
		}).Debug("Monthly stat not found for user/month")
		return nil, nil
	}

	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to get monthly stat by user and month")
		return nil, result.Error
	}

	return &stat, nil
}

func (r *GormUserMonthlyStatRepository) GetByMonth(year, month int) ([]*models.UserMonthlyStat, error) {
	var stats []*models.UserMonthlyStat
	result := r.db.Where("year = ? AND month = ?", year, month).Find(&stats)

	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to get monthly stats by month")
		return nil, result.Error
	}

	r.logger.WithFields(logrus.Fields{
		"year":  year,
		"month": month,
		"count": len(stats),
	}).Debug("Retrieved monthly stats by month")

	return stats, nil
}

func (r *GormUserMonthlyStatRepository) UpdatePlannedStats(userID uint, year, month, plannedDays, plannedMinutes int) error {
	r.logger.WithFields(logrus.Fields{
		"user_id":         userID,
		"year":            year,
		"month":           month,
		"planned_days":    plannedDays,
		"planned_minutes": plannedMinutes,
	}).Debug("Updating planned stats")

	// Получаем существующую статистику
	stat, err := r.GetByUserAndMonth(userID, year, month)
	if err != nil {
		return err
	}

	if stat == nil {
		// Создаем новую запись если не существует
		stat = &models.UserMonthlyStat{
			UserID:         userID,
			Year:           year,
			Month:          month,
			PlannedDays:    plannedDays,
			PlannedMinutes: plannedMinutes,
		}
		return r.Create(stat)
	}

	// Обновляем существующую запись
	stat.PlannedDays = plannedDays
	stat.PlannedMinutes = plannedMinutes
	stat.CalculateStats()
	stat.UpdatedAt = time.Now()

	result := r.db.Save(stat)
	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to update planned stats")
		return result.Error
	}

	r.logger.WithFields(logrus.Fields{
		"id":      stat.ID,
		"user_id": userID,
	}).Debug("Planned stats updated successfully")

	return nil
}

func (r *GormUserMonthlyStatRepository) UpdateWorkedStats(userID uint, year, month, workedDays, workedMinutes int) error {
	r.logger.WithFields(logrus.Fields{
		"user_id":        userID,
		"year":           year,
		"month":          month,
		"worked_days":    workedDays,
		"worked_minutes": workedMinutes,
	}).Debug("Updating worked stats")

	// Получаем существующую статистику
	stat, err := r.GetByUserAndMonth(userID, year, month)
	if err != nil {
		return err
	}

	if stat == nil {
		// Создаем новую запись если не существует
		stat = &models.UserMonthlyStat{
			UserID:        userID,
			Year:          year,
			Month:         month,
			WorkedDays:    workedDays,
			WorkedMinutes: workedMinutes,
		}
		stat.CalculateStats()
		return r.Create(stat)
	}

	// Обновляем существующую запись
	stat.WorkedDays = workedDays
	stat.WorkedMinutes = workedMinutes
	stat.CalculateStats()
	stat.UpdatedAt = time.Now()

	result := r.db.Save(stat)
	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to update worked stats")
		return result.Error
	}

	r.logger.WithFields(logrus.Fields{
		"id":      stat.ID,
		"user_id": userID,
	}).Debug("Worked stats updated successfully")

	return nil
}

func (r *GormUserMonthlyStatRepository) DeleteByUserID(userID uint) error {
	r.logger.WithField("user_id", userID).Info("Deleting all monthly stats for user")

	result := r.db.Where("user_id = ?", userID).Delete(&models.UserMonthlyStat{})
	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to delete user monthly stats")
		return result.Error
	}

	r.logger.WithFields(logrus.Fields{
		"user_id":       userID,
		"rows_affected": result.RowsAffected,
	}).Info("User monthly stats deleted successfully")

	return nil
}

func (r *GormUserMonthlyStatRepository) DeleteByID(id uint) error {
	r.logger.WithField("id", id).Info("Deleting monthly stat by ID")

	result := r.db.Delete(&models.UserMonthlyStat{}, id)
	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to delete monthly stat")
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithField("id", id).Warn("Monthly stat not found for deletion")
		return errors.New("статистика не найдена")
	}

	r.logger.WithField("id", id).Info("Monthly stat deleted successfully")
	return nil
}

func (r *GormUserMonthlyStatRepository) Exists(userID uint, year, month int) (bool, error) {
	var count int64
	result := r.db.Model(&models.UserMonthlyStat{}).
		Where("user_id = ? AND year = ? AND month = ?", userID, year, month).
		Count(&count)

	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to check monthly stat existence")
		return false, result.Error
	}

	exists := count > 0
	r.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"year":    year,
		"month":   month,
		"exists":  exists,
	}).Debug("Checked monthly stat existence")

	return exists, nil
}

func (r *GormUserMonthlyStatRepository) CreateForAllUsers(year, month, plannedDays, plannedMinutes int) error {
	r.logger.WithFields(logrus.Fields{
		"year":            year,
		"month":           month,
		"planned_days":    plannedDays,
		"planned_minutes": plannedMinutes,
	}).Info("Creating monthly stats for all users")

	// Получаем всех пользователей
	var users []models.User
	result := r.db.Find(&users)
	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to get users for monthly stats creation")
		return result.Error
	}

	// Создаем статистику для каждого пользователя
	for _, user := range users {
		stat := &models.UserMonthlyStat{
			UserID:         user.ID,
			Year:           year,
			Month:          month,
			PlannedDays:    plannedDays,
			PlannedMinutes: plannedMinutes,
		}
		stat.CalculateStats()

		// Проверяем, существует ли уже статистика
		exists, err := r.Exists(user.ID, year, month)
		if err != nil {
			return err
		}

		if !exists {
			if err := r.Create(stat); err != nil {
				r.logger.WithError(err).Error("Failed to create monthly stat for user", "user_id", user.ID)
				return err
			}
		}
	}

	r.logger.WithFields(logrus.Fields{
		"year":  year,
		"month": month,
		"users": len(users),
	}).Info("Monthly stats created for all users")

	return nil
}

func (r *GormUserMonthlyStatRepository) UpdateForAllUsers(year, month, plannedDays, plannedMinutes int) error {
	r.logger.WithFields(logrus.Fields{
		"year":            year,
		"month":           month,
		"planned_days":    plannedDays,
		"planned_minutes": plannedMinutes,
	}).Info("Updating monthly stats for all users")

	// Обновляем все записи за указанный месяц
	result := r.db.Model(&models.UserMonthlyStat{}).
		Where("year = ? AND month = ?", year, month).
		Updates(map[string]interface{}{
			"planned_days":    plannedDays,
			"planned_minutes": plannedMinutes,
			"updated_at":      time.Now(),
		})

	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to update monthly stats for all users")
		return result.Error
	}

	// После обновления плановых показателей нужно пересчитать статистику
	// Получаем все обновленные записи
	var stats []*models.UserMonthlyStat
	result = r.db.Where("year = ? AND month = ?", year, month).Find(&stats)
	if result.Error != nil {
		return result.Error
	}

	// Пересчитываем статистику для каждой записи
	for _, stat := range stats {
		stat.CalculateStats()
		if err := r.db.Save(stat).Error; err != nil {
			r.logger.WithError(err).Error("Failed to recalculate stats after update")
			return err
		}
	}

	r.logger.WithFields(logrus.Fields{
		"year":         year,
		"month":        month,
		"rows_updated": result.RowsAffected,
	}).Info("Monthly stats updated for all users")

	return nil
}
