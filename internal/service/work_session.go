package service

import (
	"fmt"
	"strings"
	"time"
	"work-schedule-bot/internal/models"
	"work-schedule-bot/internal/repository"

	"github.com/sirupsen/logrus"
)

type WorkSessionService struct {
	sessionRepo         repository.WorkSessionRepository
	userMonthlyStatRepo repository.UserMonthlyStatRepository
	workScheduleRepo    repository.WorkScheduleRepository
	absenceRepo         repository.AbsencePeriodRepository
	logger              *logrus.Logger
}

func NewWorkSessionService(
	sessionRepo 		repository.WorkSessionRepository,
	userMonthlyStatRepo repository.UserMonthlyStatRepository,
	workScheduleRepo 	repository.WorkScheduleRepository,
	absenceRepo         repository.AbsencePeriodRepository,
) *WorkSessionService {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	return &WorkSessionService{
		sessionRepo:         sessionRepo,
		userMonthlyStatRepo: userMonthlyStatRepo,
		workScheduleRepo:    workScheduleRepo,
		logger:              logger,
	}
}

// ClockIn отмечает начало рабочего дня
func (s *WorkSessionService) ClockIn(userID uint, clockInTime time.Time, requiredMinutes int) (*models.WorkSession, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":          userID,
		"clock_in_time":    clockInTime.Format("15:04"),
		"required_minutes": requiredMinutes,
	}).Info("User clocking in")

	// Проверяем, есть ли активная сессия
	hasActive, err := s.sessionRepo.UserHasActiveSession(userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check active session")
		return nil, err
	}

	if hasActive {
		s.logger.WithField("user_id", userID).Warn("User already has active session")
		return nil, fmt.Errorf("у вас уже есть активная рабочая сессия")
	}

	// Проверяем, есть ли сессия на сегодня
	// hasToday, err := s.sessionRepo.UserHasSessionToday(userID)
	// if err != nil {
	//     s.logger.WithError(err).Error("Failed to check today's session")
	//     return nil, err
	// }

	// if hasToday {
	//     s.logger.WithField("user_id", userID).Warn("User already has session today")
	//     return nil, fmt.Errorf("сегодня вы уже отмечались")
	// }

	// Создаем новую сессию
	session := &models.WorkSession{
		UserID:          userID,
		Date:            clockInTime,
		ClockInTime:     clockInTime,
		ClockOutTime:    nil,
		RequiredMinutes: requiredMinutes,
		Status:          models.StatusActive,
	}

	// Вычисляем поля
	session.UpdateCalculatedFields()

	if !session.IsValid() {
		s.logger.Warn("Invalid work session data")
		return nil, fmt.Errorf("некорректные данные сессии")
	}

	err = s.sessionRepo.Create(session)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create work session")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"id":      session.ID,
		"user_id": userID,
		"date":    session.Date.Format("2006-01-02"),
	}).Info("User clocked in successfully")

	return session, nil
}

// ClockOut отмечает конец рабочего дня
func (s *WorkSessionService) ClockOut(userID uint, clockOutTime time.Time) (*models.WorkSession, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":        userID,
		"clock_out_time": clockOutTime.Format("15:04"),
	}).Info("User clocking out")

	// Завершаем сессию
	sessionID, err := s.sessionRepo.CompleteSession(userID, clockOutTime)
	if err != nil {
		s.logger.WithError(err).Error("Failed to complete work session")
		return nil, err
	}

	// Получаем обновленную сессию
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get completed session")
		return nil, err
	}

	// Обновляем статистику за месяц
	go func() {
		if err := s.updateMonthlyStats(userID, session); err != nil {
			s.logger.WithError(err).Error("Failed to update monthly stats after clock out")
		}
	}()

	s.logger.WithFields(logrus.Fields{
		"id":             session.ID,
		"user_id":        userID,
		"worked_minutes": session.WorkedMinutes,
		"diff_minutes":   session.DiffMinutes,
	}).Info("User clocked out successfully")

	return session, nil
}

// updateMonthlyStats обновляет месячную статистику после завершения рабочего дня
func (s *WorkSessionService) updateMonthlyStats(userID uint, session *models.WorkSession) error {
	year := session.Date.Year()
	month := int(session.Date.Month())

	// Получаем статистику за месяц
	days, minutes, err := s.sessionRepo.GetStatsByUserAndMonth(userID, year, month)
	if err != nil {
		return err
	}

	// Обновляем месячную статистику
	err = s.userMonthlyStatRepo.UpdateWorkedStats(userID, year, month, days, minutes)
	if err != nil {
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"year":    year,
		"month":   month,
		"days":    days,
		"minutes": minutes,
	}).Info("Monthly stats updated after clock out")

	return nil
}

// GetAllTodaySession возвращает все сессии на сегодня
func (s *WorkSessionService) GetAllTodaySessions(userID uint) (*[]models.WorkSession, error) {
	s.logger.WithField("user_id", userID).Debug("Getting today's work session")
	return s.sessionRepo.GetAllTodayByUserID(userID)
}

// GetTodaySession возвращает сессию на сегодня
func (s *WorkSessionService) GetTodaySession(userID uint) (*models.WorkSession, error) {
	s.logger.WithField("user_id", userID).Debug("Getting today's work session")
	return s.sessionRepo.GetTodayByUserID(userID)
}

// GetActiveSession возвращает активную сессию
func (s *WorkSessionService) GetActiveSession(userID uint) (*models.WorkSession, error) {
	s.logger.WithField("user_id", userID).Debug("Getting active work session")
	return s.sessionRepo.GetActiveByUserID(userID)
}

// GetSessionHistory возвращает историю сессий
func (s *WorkSessionService) GetSessionHistory(userID uint, limit int) ([]*models.WorkSession, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"limit":   limit,
	}).Debug("Getting work session history")

	return s.sessionRepo.GetByUserID(userID, limit)
}

// GetMonthSessions возвращает сессии за месяц
func (s *WorkSessionService) GetMonthSessions(userID uint, year, month int) ([]*models.WorkSession, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"year":    year,
		"month":   month,
	}).Debug("Getting work sessions for month")

	return s.sessionRepo.GetByUserIDAndMonth(userID, year, month)
}

// FormatSession форматирует сессию для отображения
func (s *WorkSessionService) FormatSession(session *models.WorkSession) string {
	if session == nil {
		return "❌ Сессия не найдена"
	}

	// Если это отсутствие
	if session.IsAbsence() {
		emoji := session.GetAbsenceEmoji()
		sessionType := session.FormatSessionType()

		return fmt.Sprintf(
			`%s %s: %s

📅 Дата: %s
⏰ Время: 09:00 - 17:40
⏳ Отметка: %d:%02d часов
📊 Статус: %s

%s`,
			emoji, sessionType,
			session.Date.Format("02.01.2006"),
			session.Date.Format("02.01.2006"),
			session.RequiredMinutes/60, session.RequiredMinutes%60,
			"✅ Засчитано",
			"💡 Этот день засчитан как рабочий согласно графику",
		)
	}

	dateStr := session.Date.Format("02.01.2006")
	timeStr := session.FormatTime()
	durationStr := session.Duration()

	// Конвертируем минуты в часы:минуты
	requiredHours := session.RequiredMinutes / 60
	requiredMinutes := session.RequiredMinutes % 60
	workedHours := session.WorkedMinutes / 60
	workedMinutes := session.WorkedMinutes % 60
	diffHours := session.DiffMinutes / 60
	diffMinutes := session.DiffMinutes % 60
	tempDiffMinutes := 0
	if session.WorkedMinutes == 0{
		currentTime := time.Now()
		diff := currentTime.Sub(session.ClockInTime)
		tempMinutes := int(diff.Minutes())
		workedHours = tempMinutes / 60
		workedMinutes = tempMinutes % 60
		
		tempDiffMinutes = session.RequiredMinutes - tempMinutes
		diffHours = tempDiffMinutes / 60
		diffMinutes = tempDiffMinutes % 60
	} 

	var requiredTime, workedTime, diffTime string

	if requiredMinutes == 0 {
		requiredTime = fmt.Sprintf("%dч", requiredHours)
	} else {
		requiredTime = fmt.Sprintf("%dч %dм", requiredHours, requiredMinutes)
	}

	if workedMinutes == 0 {
		workedTime = fmt.Sprintf("%dч", workedHours)
	} else {
		workedTime = fmt.Sprintf("%dч %dм", workedHours, workedMinutes)
	}

	if session.DiffMinutes != 0 {
		absDiffHours := diffHours
		if diffHours < 0 {
			absDiffHours = -diffHours
		}
		absDiffMinutes := diffMinutes
		if diffMinutes < 0 {
			absDiffMinutes = -diffMinutes
		}

		if absDiffMinutes == 0 {
			diffTime = fmt.Sprintf("%dч", absDiffHours)
		} else {
			diffTime = fmt.Sprintf("%dч %dм", absDiffHours, absDiffMinutes)
		}
	}

	statusEmoji := "🟢"
	if session.Status == models.StatusCompleted {
		statusEmoji = "✅"
	}

	diffStatus := ""
	if tempDiffMinutes < 0 {
		diffStatus = fmt.Sprintf("➕ Переработка: %s", diffTime)
	} else if tempDiffMinutes > 0 {
		diffStatus = fmt.Sprintf("➖ Недобор: %s", diffTime)
	}

	result := fmt.Sprintf(
		`📅 Рабочий день: %s
%s %s

%s
⏳ Продолжительность: %s

📊 Нормы:
   📋 Плановое время: %s
   ⏰ Отработано: %s`,
		dateStr,
		statusEmoji, session.Status,
		timeStr,
		durationStr,
		requiredTime,
		workedTime,
	)

	if diffStatus != "" {
		result += fmt.Sprintf("\n\n%s", diffStatus)
	}

	if session.Notes != "" {
		result += fmt.Sprintf("\n\n📝 Примечание: %s", session.Notes)
	}

	result += fmt.Sprintf("\n\n🕒 Обновлено: %s",
		session.UpdatedAt.Format("02.01.2006 15:04"))

	return result
}

// FormatSessionList форматирует список сессий
func (s *WorkSessionService) FormatSessionList(sessions []*models.WorkSession) string {
	if len(sessions) == 0 {
		return "📭 Рабочих сессий пока нет"
	}

	var result strings.Builder
	result.WriteString("📋 История рабочих дней:\n\n")

	for i, session := range sessions {
		dateStr := session.Date.Format("02.01")
		timeStr := session.FormatTime()

		statusEmoji := "🟢"
		if session.Status == models.StatusCompleted {
			statusEmoji = "✅"
		}

		workedHours := session.WorkedMinutes / 60
		workedMinutes := session.WorkedMinutes % 60

		var workedTime string
		if workedMinutes == 0 {
			workedTime = fmt.Sprintf("%dч", workedHours)
		} else {
			workedTime = fmt.Sprintf("%dч %dм", workedHours, workedMinutes)
		}

		fmt.Fprintf(&result, "%d. %s %s - %s (%s)\n",
			i+1,
			statusEmoji,
			dateStr,
			workedTime,
			timeStr)
	}

	return result.String()
}

// GetRequiredMinutesForToday возвращает необходимое время работы на сегодня
func (s *WorkSessionService) GetRequiredMinutesForToday(userID uint) (int, error) {
	now := time.Now()

	// Получаем график на текущий месяц
	schedule, err := s.workScheduleRepo.GetByYearMonth(now.Year(), int(now.Month()))
	if err != nil {
		return 480, err // По умолчанию 8 часов
	}

	if schedule == nil {
		return 480, nil // По умолчанию 8 часов
	}

	return schedule.WorkMinutesPerDay, nil
}

// CanClockIn проверяет, может ли пользователь начать работу
func (s *WorkSessionService) CanClockIn(userID uint, targetTime time.Time) (bool, string, error) {
	// Проверяем активную сессию
	activeSession, err := s.sessionRepo.GetActiveByUserID(userID)
	if err != nil {
		return false, "", err
	}

	if activeSession != nil {
		return false, "у вас уже есть активная рабочая сессия", nil
	}

	todaySession, err := s.sessionRepo.GetByUserAndDate(userID, targetTime)
	if err != nil {
		return false, "", err
	}

	if todaySession != nil {
		// Проверяем, не является ли это отсутствием
		if todaySession.IsAbsence() {
			return false, "сегодня у вас " + strings.ToLower(todaySession.FormatSessionType()), nil
		}

		// if todaySession.IsCompleted() {
		// 	return false, "вы уже завершили работу сегодня", nil
		// }
	}

	// Проверяем, не находится ли пользователь в отпуске/больничном
	if s.absenceRepo != nil {
		currentAbsence, err := s.absenceRepo.GetCurrentAbsence(userID, targetTime)
		if err != nil {
			s.logger.Warnf("Failed to check current absence: %v", err)
		} else if currentAbsence != nil {
			return false, "сегодня у вас " + getAbsenceTypeText(currentAbsence.Type), nil
		}
	}

	return true, "", nil
}

// Вспомогательная функция
func getAbsenceTypeText(absenceType string) string {
	switch absenceType {
	case models.AbsenceTypeVacation:
		return "отпуск"
	case models.AbsenceTypeSickLeave:
		return "больничный"
	case models.AbsenceTypeDayOff:
		return "отгул"
	default:
		return "отсутствие"
	}
}

// CanClockOut проверяет, может ли пользователь закончить работу
func (s *WorkSessionService) CanClockOut(userID uint) (bool, string, error) {
	// Проверяем активную сессию
	hasActive, err := s.sessionRepo.UserHasActiveSession(userID)
	if err != nil {
		return false, "", err
	}

	if !hasActive {
		return false, "у вас нет активной рабочей сессии", nil
	}

	return true, "", nil
}
