package repository

import (
    "errors"
    "time"
    "work-schedule-bot/internal/models"

    "github.com/sirupsen/logrus"
    "gorm.io/gorm"
)

type WorkScheduleRepository interface {
    Create(schedule *models.WorkSchedule) error
    Update(schedule *models.WorkSchedule) error
    Delete(id uint) error
    GetByID(id uint) (*models.WorkSchedule, error)
    GetByYearMonth(year, month int) (*models.WorkSchedule, error)
    GetByYear(year int) ([]*models.WorkSchedule, error)
    GetAll() ([]*models.WorkSchedule, error)
    Exists(year, month int) (bool, error)
    GetCurrentMonth() (*models.WorkSchedule, error)
}

type GormWorkScheduleRepository struct {
    db     *gorm.DB
    logger *logrus.Logger
}

func NewGormWorkScheduleRepository(db *gorm.DB) (*GormWorkScheduleRepository, error) {
    logger := logrus.New()
    logger.SetFormatter(&logrus.TextFormatter{
        FullTimestamp:   true,
        TimestampFormat: "2006-01-02 15:04:05",
    })
    
    // Автомиграция
    if err := db.AutoMigrate(&models.WorkSchedule{}); err != nil {
        logger.WithError(err).Error("Failed to auto-migrate work_schedules table")
        return nil, err
    }

    logger.Info("Work schedule repository initialized")
    
    return &GormWorkScheduleRepository{
        db:     db,
        logger: logger,
    }, nil
}

func (r *GormWorkScheduleRepository) Create(schedule *models.WorkSchedule) error {
    r.logger.WithFields(logrus.Fields{
        "year":  schedule.Year,
        "month": schedule.Month,
    }).Info("Creating work schedule")

    // Проверяем валидность
    if !schedule.IsValid() {
        r.logger.WithFields(logrus.Fields{
            "year":  schedule.Year,
            "month": schedule.Month,
        }).Warn("Invalid work schedule data")
        return errors.New("некорректные данные графика")
    }

    // Проверяем, существует ли уже график на этот месяц
    exists, err := r.Exists(schedule.Year, schedule.Month)
    if err != nil {
        r.logger.WithError(err).Error("Failed to check schedule existence")
        return err
    }
    
    if exists {
        r.logger.WithFields(logrus.Fields{
            "year":  schedule.Year,
            "month": schedule.Month,
        }).Warn("Work schedule already exists")
        return errors.New("график на этот месяц уже существует")
    }

    // Вычисляем total_minutes
    schedule.TotalMinutes = schedule.CalculateTotalMinutes()

    result := r.db.Create(schedule)
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to create work schedule")
        return result.Error
    }

    r.logger.WithFields(logrus.Fields{
        "id":    schedule.ID,
        "year":  schedule.Year,
        "month": schedule.Month,
    }).Info("Work schedule created successfully")
    
    return nil
}

func (r *GormWorkScheduleRepository) Update(schedule *models.WorkSchedule) error {
    r.logger.WithFields(logrus.Fields{
        "id":    schedule.ID,
        "year":  schedule.Year,
        "month": schedule.Month,
    }).Info("Updating work schedule")

    // Проверяем валидность
    if !schedule.IsValid() {
        r.logger.WithFields(logrus.Fields{
            "id":    schedule.ID,
            "year":  schedule.Year,
            "month": schedule.Month,
        }).Warn("Invalid work schedule data for update")
        return errors.New("некорректные данные графика")
    }

    // Проверяем существование
    existingSchedule, err := r.GetByID(schedule.ID)
    if err != nil {
        r.logger.WithError(err).Error("Failed to get work schedule for update")
        return err
    }
    
    if existingSchedule == nil {
        r.logger.WithField("id", schedule.ID).Warn("Work schedule not found for update")
        return errors.New("график не найден")
    }

    // Вычисляем total_minutes
    schedule.TotalMinutes = schedule.CalculateTotalMinutes()
    schedule.UpdatedAt = time.Now()

    result := r.db.Save(schedule)
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to update work schedule")
        return result.Error
    }

    r.logger.WithFields(logrus.Fields{
        "id":    schedule.ID,
        "year":  schedule.Year,
        "month": schedule.Month,
    }).Info("Work schedule updated successfully")
    
    return nil
}

func (r *GormWorkScheduleRepository) Delete(id uint) error {
    r.logger.WithField("id", id).Info("Deleting work schedule")

    result := r.db.Delete(&models.WorkSchedule{}, id)
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to delete work schedule")
        return result.Error
    }
    
    if result.RowsAffected == 0 {
        r.logger.WithField("id", id).Warn("Work schedule not found for deletion")
        return errors.New("график не найден")
    }

    r.logger.WithField("id", id).Info("Work schedule deleted successfully")
    return nil
}

func (r *GormWorkScheduleRepository) GetByID(id uint) (*models.WorkSchedule, error) {
    var schedule models.WorkSchedule
    result := r.db.First(&schedule, id)
    
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        r.logger.WithField("id", id).Debug("Work schedule not found")
        return nil, nil
    }
    
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to get work schedule by ID")
        return nil, result.Error
    }

    return &schedule, nil
}

func (r *GormWorkScheduleRepository) GetByYearMonth(year, month int) (*models.WorkSchedule, error) {
    var schedule models.WorkSchedule
    result := r.db.Where("year = ? AND month = ?", year, month).First(&schedule)
    
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        r.logger.WithFields(logrus.Fields{
            "year":  year,
            "month": month,
        }).Debug("Work schedule not found for year/month")
        return nil, nil
    }
    
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to get work schedule by year/month")
        return nil, result.Error
    }

    return &schedule, nil
}

func (r *GormWorkScheduleRepository) GetByYear(year int) ([]*models.WorkSchedule, error) {
    var schedules []*models.WorkSchedule
    result := r.db.Where("year = ?", year).Order("month ASC").Find(&schedules)
    
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to get work schedules by year")
        return nil, result.Error
    }

    r.logger.WithFields(logrus.Fields{
        "year":   year,
        "count":  len(schedules),
    }).Debug("Retrieved work schedules by year")
    
    return schedules, nil
}

func (r *GormWorkScheduleRepository) GetAll() ([]*models.WorkSchedule, error) {
    var schedules []*models.WorkSchedule
    result := r.db.Order("year ASC, month ASC").Find(&schedules)
    
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to get all work schedules")
        return nil, result.Error
    }

    r.logger.WithField("count", len(schedules)).Debug("Retrieved all work schedules")
    return schedules, nil
}

func (r *GormWorkScheduleRepository) Exists(year, month int) (bool, error) {
    var count int64
    result := r.db.Model(&models.WorkSchedule{}).
        Where("year = ? AND month = ?", year, month).
        Count(&count)
    
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to check work schedule existence")
        return false, result.Error
    }

    exists := count > 0
    r.logger.WithFields(logrus.Fields{
        "year":   year,
        "month":  month,
        "exists": exists,
    }).Debug("Checked work schedule existence")
    
    return exists, nil
}

func (r *GormWorkScheduleRepository) GetCurrentMonth() (*models.WorkSchedule, error) {
    now := time.Now()
    year := now.Year()
    month := int(now.Month())
    
    return r.GetByYearMonth(year, month)
}