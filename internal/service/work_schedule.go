package service

import (
    "fmt"
    "strconv"
    "strings"
    "time"
    "work-schedule-bot/internal/models"
    "work-schedule-bot/internal/repository"

    "github.com/sirupsen/logrus"
)

type WorkScheduleService struct {
    repo                   repository.WorkScheduleRepository
    userMonthlyStatService *UserMonthlyStatService // ДОБАВЛЕНО
    logger                 *logrus.Logger
}

func NewWorkScheduleService(
    repo repository.WorkScheduleRepository,
    userMonthlyStatService *UserMonthlyStatService, // ДОБАВЛЕНО
) *WorkScheduleService {
    return &WorkScheduleService{
        repo:                   repo,
        userMonthlyStatService: userMonthlyStatService, // ДОБАВЛЕНО
        logger:                 logrus.New(),
    }
}

// CreateSchedule создает новый рабочий график
// service/work_schedule_service.go
func (s *WorkScheduleService) CreateSchedule(year, month, workDays, workMinutesPerDay int) (*models.WorkSchedule, error) {
    s.logger.WithFields(logrus.Fields{
        "year":               year,
        "month":              month,
        "work_days":          workDays,
        "work_minutes_per_day": workMinutesPerDay,
    }).Info("Creating new work schedule")

    schedule := &models.WorkSchedule{
        Year:              year,
        Month:             month,
        WorkDays:          workDays,
        WorkMinutesPerDay: workMinutesPerDay,
    }

    if !schedule.IsValid() {
        s.logger.Warn("Invalid schedule data provided")
        return nil, fmt.Errorf("некорректные данные: год 2000-2100, месяц 1-12, дни 0-31, минуты в день 1-1440")
    }

    // Создаем график
    err := s.repo.Create(schedule) // Только ошибка, schedule обновляется по ссылке
    if err != nil {
        s.logger.WithError(err).Error("Failed to create schedule")
        return nil, err
    }

    s.logger.WithFields(logrus.Fields{
        "id":    schedule.ID,
        "total_minutes": schedule.TotalMinutes,
    }).Info("Schedule created successfully")

    // Создаем статистику для всех пользователей для нового графика
    go func() {
        if err := s.userMonthlyStatService.CreateStatsForNewSchedule(schedule); err != nil {
            s.logger.WithError(err).Error("Failed to create monthly stats for new schedule")
        }
    }()
    
    return schedule, nil
}

func (s *WorkScheduleService) UpdateSchedule(id uint, workDays, workMinutesPerDay int) (*models.WorkSchedule, error) {
    s.logger.WithFields(logrus.Fields{
        "id":                  id,
        "work_days":          workDays,
        "work_minutes_per_day": workMinutesPerDay,
    }).Info("Updating work schedule")

    // Получаем существующий график
    schedule, err := s.repo.GetByID(id)
    if err != nil {
        s.logger.WithError(err).Error("Failed to get schedule for update")
        return nil, err
    }
    
    if schedule == nil {
        s.logger.WithField("id", id).Warn("Schedule not found")
        return nil, fmt.Errorf("график с ID %d не найден", id)
    }

    // Обновляем поля
    schedule.WorkDays = workDays
    schedule.WorkMinutesPerDay = workMinutesPerDay
    schedule.TotalMinutes = schedule.CalculateTotalMinutes()

    if !schedule.IsValid() {
        s.logger.Warn("Invalid schedule data after update")
        return nil, fmt.Errorf("некорректные данные после обновления")
    }

    // Обновляем график
    err = s.repo.Update(schedule) // Только ошибка
    if err != nil {
        s.logger.WithError(err).Error("Failed to update schedule")
        return nil, err
    }

    s.logger.WithFields(logrus.Fields{
        "id":           schedule.ID,
        "year":         schedule.Year,
        "month":        schedule.Month,
        "total_minutes": schedule.TotalMinutes,
    }).Info("Schedule updated successfully")

    // Обновляем статистику для всех пользователей при изменении графика
    go func() {
        if err := s.userMonthlyStatService.UpdateStatsForSchedule(schedule); err != nil {
            s.logger.WithError(err).Error("Failed to update monthly stats after schedule update")
        }
    }()
    
    return schedule, nil
}

// DeleteSchedule удаляет график
func (s *WorkScheduleService) DeleteSchedule(id uint) error {
    s.logger.WithField("id", id).Info("Deleting work schedule")

    // Проверяем существование
    schedule, err := s.repo.GetByID(id)
    if err != nil {
        s.logger.WithError(err).Error("Failed to get schedule for deletion")
        return err
    }
    
    if schedule == nil {
        s.logger.WithField("id", id).Warn("Schedule not found for deletion")
        return fmt.Errorf("график с ID %d не найден", id)
    }

    err = s.repo.Delete(id)
    if err != nil {
        s.logger.WithError(err).Error("Failed to delete schedule")
        return err
    }

    s.logger.WithField("id", id).Info("Schedule deleted successfully")
    return nil
}

// GetScheduleByID возвращает график по ID
func (s *WorkScheduleService) GetScheduleByID(id uint) (*models.WorkSchedule, error) {
    s.logger.WithField("id", id).Debug("Getting schedule by ID")
    return s.repo.GetByID(id)
}

// GetScheduleByYearMonth возвращает график по году и месяцу
func (s *WorkScheduleService) GetScheduleByYearMonth(year, month int) (*models.WorkSchedule, error) {
    s.logger.WithFields(logrus.Fields{
        "year":  year,
        "month": month,
    }).Debug("Getting schedule by year/month")
    
    return s.repo.GetByYearMonth(year, month)
}

// GetCurrentSchedule возвращает график на текущий месяц
func (s *WorkScheduleService) GetCurrentSchedule() (*models.WorkSchedule, error) {
    s.logger.Debug("Getting current month schedule")
    return s.repo.GetCurrentMonth()
}

// GetAllSchedules возвращает все графики
func (s *WorkScheduleService) GetAllSchedules() ([]*models.WorkSchedule, error) {
    s.logger.Debug("Getting all schedules")
    return s.repo.GetAll()
}

// GetSchedulesByYear возвращает графики за год
func (s *WorkScheduleService) GetSchedulesByYear(year int) ([]*models.WorkSchedule, error) {
    s.logger.WithField("year", year).Debug("Getting schedules by year")
    return s.repo.GetByYear(year)
}

// FormatSchedule форматирует график для отображения
func (s *WorkScheduleService) FormatSchedule(schedule *models.WorkSchedule) string {
    if schedule == nil {
        return "❌ График не найден"
    }

    // Конвертируем минуты в часы:минуты
    hoursPerDay := schedule.WorkMinutesPerDay / 60
    minutesPerDay := schedule.WorkMinutesPerDay % 60
    
    totalHours := schedule.TotalMinutes / 60
    totalMinutes := schedule.TotalMinutes % 60

    var timePerDay string
    if minutesPerDay == 0 {
        timePerDay = fmt.Sprintf("%dч", hoursPerDay)
    } else {
        timePerDay = fmt.Sprintf("%dч %dм", hoursPerDay, minutesPerDay)
    }

    var totalTime string
    if totalMinutes == 0 {
        totalTime = fmt.Sprintf("%dч", totalHours)
    } else {
        totalTime = fmt.Sprintf("%dч %dм", totalHours, totalMinutes)
    }

    monthName := time.Month(schedule.Month).String()

    return fmt.Sprintf(
        `📅 **График работы: %s %d**

🆔 ID: %d
📊 Рабочих дней: %d
⏰ Время в день: %s
📈 Всего времени: %s
📅 Создан: %s
🔄 Обновлен: %s`,
        monthName, schedule.Year,
        schedule.ID,
        schedule.WorkDays,
        timePerDay,
        totalTime,
        schedule.CreatedAt.Format("02.01.2006 15:04"),
        schedule.UpdatedAt.Format("02.01.2006 15:04"),
    )
}

// FormatScheduleList форматирует список графиков
func (s *WorkScheduleService) FormatScheduleList(schedules []*models.WorkSchedule) string {
    if len(schedules) == 0 {
        return "📭 Графиков работы пока нет"
    }

    var result strings.Builder
    result.WriteString("📋 **Все графики работы:**\n\n")

    for i, schedule := range schedules {
        monthName := time.Month(schedule.Month).String()
        
        hoursPerDay := schedule.WorkMinutesPerDay / 60
        minutesPerDay := schedule.WorkMinutesPerDay % 60
        
        var timePerDay string
        if minutesPerDay == 0 {
            timePerDay = fmt.Sprintf("%dч", hoursPerDay)
        } else {
            timePerDay = fmt.Sprintf("%dч %dм", hoursPerDay, minutesPerDay)
        }

        result.WriteString(fmt.Sprintf(
            "%d. %s %d - %d дней × %s (ID: %d)\n",
            i+1,
            monthName,
            schedule.Year,
            schedule.WorkDays,
            timePerDay,
            schedule.ID,
        ))
    }

    return result.String()
}

// ParseScheduleData парсит данные графика из строки
func (s *WorkScheduleService) ParseScheduleData(input string) (year, month, workDays, workMinutesPerDay int, err error) {
    // Формат: "2024 12 22 480" (год месяц дни минуты_в_день)
    parts := strings.Fields(input)
    if len(parts) != 4 {
        return 0, 0, 0, 0, fmt.Errorf("неверный формат. Используйте: год месяц дни минуты_в_день")
    }

    // Парсим год
    year, err = strconv.Atoi(parts[0])
    if err != nil || year < 2000 || year > 2100 {
        return 0, 0, 0, 0, fmt.Errorf("неверный год. Должен быть между 2000 и 2100")
    }

    // Парсим месяц
    month, err = strconv.Atoi(parts[1])
    if err != nil || month < 1 || month > 12 {
        return 0, 0, 0, 0, fmt.Errorf("неверный месяц. Должен быть между 1 и 12")
    }

    // Парсим рабочие дни
    workDays, err = strconv.Atoi(parts[2])
    if err != nil || workDays < 0 || workDays > 31 {
        return 0, 0, 0, 0, fmt.Errorf("неверное количество дней. Должно быть между 0 и 31")
    }

    // Парсим минуты в день
    workMinutesPerDay, err = strconv.Atoi(parts[3])
    if err != nil || workMinutesPerDay <= 0 || workMinutesPerDay > 1440 {
        return 0, 0, 0, 0, fmt.Errorf("неверное количество минут в день. Должно быть между 1 и 1440")
    }

    return year, month, workDays, workMinutesPerDay, nil
}

// ParseTime парсит время из строки "8:30" в минуты
func (s *WorkScheduleService) ParseTime(timeStr string) (int, error) {
    parts := strings.Split(timeStr, ":")
    if len(parts) != 2 {
        return 0, fmt.Errorf("неверный формат времени. Используйте ЧЧ:ММ")
    }

    hours, err := strconv.Atoi(parts[0])
    if err != nil || hours < 0 || hours > 23 {
        return 0, fmt.Errorf("неверное количество часов. Должно быть между 0 и 23")
    }

    minutes, err := strconv.Atoi(parts[1])
    if err != nil || minutes < 0 || minutes > 59 {
        return 0, fmt.Errorf("неверное количество минут. Должно быть между 0 и 59")
    }

    return hours*60 + minutes, nil
}