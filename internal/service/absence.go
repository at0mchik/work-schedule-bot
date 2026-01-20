// internal/service/absence_service.go
package service

import (
	"fmt"
	"time"
	"work-schedule-bot/internal/models"
	"work-schedule-bot/internal/repository"

	"github.com/sirupsen/logrus"
)

type AbsenceService struct {
	absenceRepo          repository.AbsencePeriodRepository
	workSessionRepo      repository.WorkSessionRepository
	userRepo             repository.UserRepository
	workScheduleRepo     repository.WorkScheduleRepository
	nonWorkingDayService *NonWorkingDayService
	logger               *logrus.Logger
}

func NewAbsenceService(
	absenceRepo repository.AbsencePeriodRepository,
	workSessionRepo repository.WorkSessionRepository,
	userRepo repository.UserRepository,
	workScheduleRepo repository.WorkScheduleRepository,
	nonWorkingDayService *NonWorkingDayService,
) *AbsenceService {
	return &AbsenceService{
		absenceRepo:          absenceRepo,
		workSessionRepo:      workSessionRepo,
		userRepo:             userRepo,
		workScheduleRepo:     workScheduleRepo,
		nonWorkingDayService: nonWorkingDayService,
		logger:               logrus.New(),
	}
}

// AddVacation добавляет отпуск (только будущие даты)
func (s *AbsenceService) AddVacation(userID uint, startDate, endDate time.Time) (*models.AbsencePeriod, error) {
	// Нормализуем даты (оставляем только дату)
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.Local)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.Local)

	// Проверяем, что даты в будущем
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	if startDate.Before(today) {
		return nil, fmt.Errorf("отпуск можно добавить только на будущие даты")
	}

	return s.addAbsencePeriod(userID, startDate, endDate, models.AbsenceTypeVacation)
}

// AddSickLeave добавляет больничный (можно на прошедшие дни)
func (s *AbsenceService) AddSickLeave(userID uint, startDate, endDate time.Time) (*models.AbsencePeriod, error) {
	// Нормализуем даты
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.Local)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.Local)

	return s.addAbsencePeriod(userID, startDate, endDate, models.AbsenceTypeSickLeave)
}

// AddDayOff добавляет отгул (один день)
func (s *AbsenceService) AddDayOff(userID uint, date time.Time) (*models.AbsencePeriod, error) {
	// Нормализуем дату
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)

	return s.addAbsencePeriod(userID, date, date, models.AbsenceTypeDayOff)
}

// addAbsencePeriod общий метод добавления периода отсутствия
func (s *AbsenceService) addAbsencePeriod(
	userID uint,
	startDate, endDate time.Time,
	absenceType string,
) (*models.AbsencePeriod, error) {

	// Проверяем, что даты корректны
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("дата окончания не может быть раньше даты начала")
	}

	// Проверяем пересечения с существующими периодами
	conflicts, err := s.absenceRepo.CheckPeriodConflict(userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки конфликтов: %v", err)
	}
	if conflicts {
		return nil, fmt.Errorf("период пересекается с существующим отпуском/больничным/отгулом")
	}

	// Проверяем, что все дни в периоде доступны (нет других сессий)
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		available, err := s.workSessionRepo.CheckDateAvailability(userID, date)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки доступности даты %s: %v", date.Format("02.01.2006"), err)
		}
		if !available {
			return nil, fmt.Errorf("на дату %s уже есть запись", date.Format("02.01.2006"))
		}
	}

	// Проверяем, что дни рабочие (для отпуска и отгула)
	// if absenceType == models.AbsenceTypeVacation || absenceType == models.AbsenceTypeDayOff {
	// 	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
	// 		isNonWorking, err := s.nonWorkingDayService.IsNonWorkingDay(date)
	// 		if err != nil {
	// 			s.logger.Warnf("Failed to check if day %s is non-working: %v", date.Format("02.01.2006"), err)
	// 		} else if isNonWorking {
	// 			return nil, fmt.Errorf("дата %s является выходным днем", date.Format("02.01.2006"))
	// 		}
	// 	}
	// }

	// Создаем период отсутствия
	period := &models.AbsencePeriod{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
		Type:      absenceType,
	}

	err = s.absenceRepo.Create(period)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания периода: %v", err)
	}

	// Создаем work sessions для каждого дня периода
	createdCount, err := s.createWorkSessionsForPeriod(period)
	if err != nil {
		// Откатываем создание периода если не удалось создать сессии
		s.absenceRepo.Delete(period.ID)
		return nil, fmt.Errorf("ошибка создания рабочих сессий: %v", err)
	}

	s.logger.Infof("Created absence period ID %d with %d work sessions", period.ID, createdCount)
	return period, nil
}

// createWorkSessionsForPeriod создает work sessions для каждого дня периода
func (s *AbsenceService) createWorkSessionsForPeriod(period *models.AbsencePeriod) (int, error) {
	var sessionType string
	var requiredMinutes int = 520 // 8 часов 40 минут = 520 минут

	// Определяем тип сессии и требуемое время
	switch period.Type {
	case models.AbsenceTypeVacation:
		sessionType = models.SessionTypeVacation
	case models.AbsenceTypeSickLeave:
		sessionType = models.SessionTypeSickLeave
	case models.AbsenceTypeDayOff:
		sessionType = models.SessionTypeDayOff
	default:
		return 0, fmt.Errorf("неизвестный тип отсутствия: %s", period.Type)
	}

	createdCount := 0

	// Создаем сессию для каждого дня периода
	for date := period.StartDate; !date.After(period.EndDate); date = date.AddDate(0, 0, 1) {
		// Создаем сессию отсутствия
		session := &models.WorkSession{
			UserID:          period.UserID,
			Date:            date,
			SessionType:     sessionType,
			ClockInTime:     date.Add(9 * time.Hour),                                  // Условное время начала 09:00
			ClockOutTime:    &[]time.Time{date.Add(17*time.Hour + 40*time.Minute)}[0], // 17:40
			RequiredMinutes: requiredMinutes,
			WorkedMinutes:   requiredMinutes,
			DiffMinutes:     0,
			Status:          models.StatusCompleted,
			AbsencePeriodID: &period.ID,
		}

		err := s.workSessionRepo.CreateAbsenceSession(session)
		if err != nil {
			return createdCount, fmt.Errorf("ошибка создания сессии на %s: %v", date.Format("02.01.2006"), err)
		}

		createdCount++
	}

	return createdCount, nil
}

// GetUserAbsences возвращает все периоды отсутствия пользователя
func (s *AbsenceService) GetUserAbsences(userID uint) ([]models.AbsencePeriod, error) {
	return s.absenceRepo.GetByUserID(userID)
}

// GetCurrentAbsence возвращает текущий период отсутствия пользователя
func (s *AbsenceService) GetCurrentAbsence(userID uint, date time.Time) (*models.AbsencePeriod, error) {
	return s.absenceRepo.GetCurrentAbsence(userID, date)
}

// DeleteAbsence удаляет период отсутствия
func (s *AbsenceService) DeleteAbsence(periodID uint) error {
	// Сначала удаляем связанные work sessions
	_, err := s.absenceRepo.GetByID(periodID)
	if err != nil {
		return err
	}

	// Удаляем период
	err = s.absenceRepo.Delete(periodID)
	if err != nil {
		return err
	}

	s.logger.Infof("Deleted absence period ID %d", periodID)
	return nil
}
