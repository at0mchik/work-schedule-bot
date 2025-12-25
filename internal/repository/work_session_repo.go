package repository

import (
    "errors"
    "time"
    "work-schedule-bot/internal/models"

    "github.com/sirupsen/logrus"
    "gorm.io/gorm"
)

type WorkSessionRepository interface {
    Create(session *models.WorkSession) error
    Update(session *models.WorkSession) error
    GetByID(id uint) (*models.WorkSession, error)
    GetActiveByUserID(userID uint) (*models.WorkSession, error)
    GetCompletedByUserID(userID uint) (*models.WorkSession, error)
    GetByUserAndDate(userID uint, date time.Time) (*models.WorkSession, error)
    GetTodayByUserID(userID uint) (*models.WorkSession, error)
    GetByUserID(userID uint, limit int) ([]*models.WorkSession, error)
    GetByUserIDAndMonth(userID uint, year, month int) ([]*models.WorkSession, error)
    CompleteSession(userID uint, clockOutTime time.Time) (uint, error)
    DeleteByID(id uint) error
    DeleteByUserID(userID uint) error
    GetStatsByUserAndMonth(userID uint, year, month int) (int, int, error) // дни, минуты
    UserHasActiveSession(userID uint) (bool, error)
    UserHasSessionToday(userID uint) (bool, error)
}

type GormWorkSessionRepository struct {
    db     *gorm.DB
    logger *logrus.Logger
}

func NewGormWorkSessionRepository(db *gorm.DB) (*GormWorkSessionRepository, error) {
    logger := logrus.New()
    logger.SetFormatter(&logrus.TextFormatter{
        FullTimestamp:   true,
        TimestampFormat: "2006-01-02 15:04:05",
    })
    
    // Автомиграция
    if err := db.AutoMigrate(&models.WorkSession{}); err != nil {
        logger.WithError(err).Error("Failed to auto-migrate work_sessions table")
        return nil, err
    }

    logger.Info("Work session repository initialized")
    
    return &GormWorkSessionRepository{
        db:     db,
        logger: logger,
    }, nil
}

func (r *GormWorkSessionRepository) Create(session *models.WorkSession) error {
    r.logger.WithFields(logrus.Fields{
        "user_id": session.UserID,
        "date":    session.Date.Format("2006-01-02"),
    }).Info("Creating work session")

    if !session.IsValid() {
        r.logger.WithFields(logrus.Fields{
            "user_id": session.UserID,
            "date":    session.Date.Format("2006-01-02"),
        }).Warn("Invalid work session data")
        return errors.New("некорректные данные рабочей сессии")
    }

    // Проверяем, есть ли уже сессия на эту дату
    existing, err := r.GetByUserAndDate(session.UserID, session.Date)
    if err != nil {
        r.logger.WithError(err).Error("Failed to check existing session")
        return err
    }
    
    if existing != nil {
        r.logger.WithFields(logrus.Fields{
            "user_id": session.UserID,
            "date":    session.Date.Format("2006-01-02"),
        }).Warn("Work session already exists for this date")
        return errors.New("рабочая сессия на эту дату уже существует")
    }

    // Вычисляем поля
    session.UpdateCalculatedFields()

    result := r.db.Create(session)
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to create work session")
        return result.Error
    }

    r.logger.WithFields(logrus.Fields{
        "id":      session.ID,
        "user_id": session.UserID,
        "status":  session.Status,
    }).Info("Work session created successfully")
    
    return nil
}

func (r *GormWorkSessionRepository) Update(session *models.WorkSession) error {
    r.logger.WithFields(logrus.Fields{
        "id":      session.ID,
        "user_id": session.UserID,
    }).Info("Updating work session")

    if !session.IsValid() {
        r.logger.WithFields(logrus.Fields{
            "id":      session.ID,
            "user_id": session.UserID,
        }).Warn("Invalid work session data for update")
        return errors.New("некорректные данные рабочей сессии")
    }

    // Проверяем существование
    existing, err := r.GetByID(session.ID)
    if err != nil {
        return err
    }
    
    if existing == nil {
        r.logger.WithField("id", session.ID).Warn("Work session not found for update")
        return errors.New("рабочая сессия не найдена")
    }

    // Вычисляем поля
    session.UpdateCalculatedFields()
    session.UpdatedAt = time.Now()

    result := r.db.Save(session)
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to update work session")
        return result.Error
    }

    r.logger.WithFields(logrus.Fields{
        "id":      session.ID,
        "user_id": session.UserID,
        "status":  session.Status,
    }).Info("Work session updated successfully")
    
    return nil
}

func (r *GormWorkSessionRepository) GetByID(id uint) (*models.WorkSession, error) {
    var session models.WorkSession
    result := r.db.Preload("User").First(&session, id)
    
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        r.logger.WithField("id", id).Debug("Work session not found")
        return nil, nil
    }
    
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to get work session by ID")
        return nil, result.Error
    }

    return &session, nil
}

func (r *GormWorkSessionRepository) GetActiveByUserID(userID uint) (*models.WorkSession, error) {
    var session models.WorkSession
    // result := r.db.Where("user_id = ? AND status = ?", userID, models.StatusActive).First(&session)
    result := r.db.Where("user_id = ? AND status = ?", userID, models.StatusActive).First(&session)

    
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        r.logger.WithField("user_id", userID).Debug("No active work session found")
        return nil, nil
    }
    
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to get active work session")
        return nil, result.Error
    }

    return &session, nil
}

func (r *GormWorkSessionRepository) GetCompletedByUserID(userID uint) (*models.WorkSession, error) {
    var session models.WorkSession
    result := r.db.Where("user_id = ? AND status = ?", userID, models.StatusCompleted).First(&session)

    
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        r.logger.WithField("user_id", userID).Debug("No completed work session found")
        return nil, nil
    }
    
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to get completed work session")
        return nil, result.Error
    }

    return &session, nil
}

func (r *GormWorkSessionRepository) GetByUserAndDate(userID uint, date time.Time) (*models.WorkSession, error) {
    var session models.WorkSession
    result := r.db.Where("user_id = ? AND date = ?", userID, date.Format("2006-01-02")).First(&session)
    
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        r.logger.WithFields(logrus.Fields{
            "user_id": userID,
            "date":    date.Format("2006-01-02"),
        }).Debug("Work session not found for user/date")
        return nil, nil
    }
    
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to get work session by user and date")
        return nil, result.Error
    }

    return &session, nil
}

func (r *GormWorkSessionRepository) GetTodayByUserID(userID uint) (*models.WorkSession, error) {
    return r.GetByUserAndDate(userID, time.Now())
}

func (r *GormWorkSessionRepository) GetByUserID(userID uint, limit int) ([]*models.WorkSession, error) {
    var sessions []*models.WorkSession
    
    query := r.db.Where("user_id = ?", userID).Order("date DESC")
    if limit > 0 {
        query = query.Limit(limit)
    }
    
    result := query.Find(&sessions)
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to get work sessions by user ID")
        return nil, result.Error
    }

    r.logger.WithFields(logrus.Fields{
        "user_id": userID,
        "count":   len(sessions),
        "limit":   limit,
    }).Debug("Retrieved work sessions by user ID")
    
    return sessions, nil
}

func (r *GormWorkSessionRepository) GetByUserIDAndMonth(userID uint, year, month int) ([]*models.WorkSession, error) {
    var sessions []*models.WorkSession
    
    startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
    endDate := startDate.AddDate(0, 1, -1)
    
    result := r.db.Where("user_id = ? AND date BETWEEN ? AND ?", 
        userID, 
        startDate.Format("2006-01-02"), 
        endDate.Format("2006-01-02")).
        Order("date DESC").
        Find(&sessions)
    
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to get work sessions by user and month")
        return nil, result.Error
    }

    r.logger.WithFields(logrus.Fields{
        "user_id": userID,
        "year":    year,
        "month":   month,
        "count":   len(sessions),
    }).Debug("Retrieved work sessions by user and month")
    
    return sessions, nil
}

func (r *GormWorkSessionRepository) CompleteSession(userID uint, clockOutTime time.Time) (uint, error) {
    r.logger.WithFields(logrus.Fields{
        "user_id":      userID,
        "clock_out_time": clockOutTime.Format("15:04"),
    }).Info("Completing work session")

    // Находим активную сессию
    session, err := r.GetActiveByUserID(userID)
    if err != nil {
        return 0, err
    }
    
    if session == nil {
        r.logger.WithField("user_id", userID).Warn("No active work session found to complete")
        return 0, errors.New("нет активной рабочей сессии")
    }

    // Обновляем время выхода
    session.ClockOutTime = &clockOutTime
    session.UpdateCalculatedFields()
    session.UpdatedAt = time.Now()

    result := r.db.Save(session)
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to complete work session")
        return 0, result.Error
    }

    r.logger.WithFields(logrus.Fields{
        "id":              session.ID,
        "user_id":         userID,
        "worked_minutes":  session.WorkedMinutes,
        "diff_minutes":    session.DiffMinutes,
    }).Info("Work session completed successfully")
    
    return session.ID, nil
}

func (r *GormWorkSessionRepository) DeleteByID(id uint) error {
    r.logger.WithField("id", id).Info("Deleting work session by ID")

    result := r.db.Delete(&models.WorkSession{}, id)
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to delete work session")
        return result.Error
    }
    
    if result.RowsAffected == 0 {
        r.logger.WithField("id", id).Warn("Work session not found for deletion")
        return errors.New("рабочая сессия не найдена")
    }

    r.logger.WithField("id", id).Info("Work session deleted successfully")
    return nil
}

func (r *GormWorkSessionRepository) DeleteByUserID(userID uint) error {
    r.logger.WithField("user_id", userID).Info("Deleting all work sessions for user")

    result := r.db.Where("user_id = ?", userID).Delete(&models.WorkSession{})
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to delete user work sessions")
        return result.Error
    }

    r.logger.WithFields(logrus.Fields{
        "user_id":       userID,
        "rows_affected": result.RowsAffected,
    }).Info("User work sessions deleted successfully")
    
    return nil
}

func (r *GormWorkSessionRepository) GetStatsByUserAndMonth(userID uint, year, month int) (int, int, error) {
    var data struct{
	    Days    int64
	    Minutes int64
    }
    
    startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
    endDate := startDate.AddDate(0, 1, -1)
    
    // Подсчитываем дни и минуты
    result := r.db.Model(&models.WorkSession{}).
        Select("COUNT(DISTINCT date) as days, COALESCE(SUM(worked_minutes), 0) as minutes").
        Where("user_id = ? AND date BETWEEN ? AND ? AND status = ?",
            userID,
            startDate.Format("2006-01-02"),
            endDate.Format("2006-01-02"),
            models.StatusCompleted).
        Scan(&data)
    
    if result.Error != nil {
        r.logger.WithError(result.Error).Error("Failed to get work session stats")
        return 0, 0, result.Error
    }

    r.logger.WithFields(logrus.Fields{
        "user_id": userID,
        "year":    year,
        "month":   month,
        "days":    data.Days,
        "minutes": data.Minutes,
    }).Debug("Retrieved work session stats")
    
    return int(data.Days), int(data.Minutes), nil
}

func (r *GormWorkSessionRepository) UserHasActiveSession(userID uint) (bool, error) {
    session, err := r.GetActiveByUserID(userID)
    if err != nil {
        return false, err
    }
    return session != nil, nil
}

func (r *GormWorkSessionRepository) UserHasSessionToday(userID uint) (bool, error) {
    session, err := r.GetTodayByUserID(userID)
    if err != nil {
        return false, err
    }
    return session != nil, nil
}