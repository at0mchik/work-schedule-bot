package weekends

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// WeekendJSON - структура для парсинга исходного JSON
type WeekendJSON struct {
	Year        int             `json:"year"`
	Months      []MonthWeekends `json:"months"`
	Transitions []Transition    `json:"transitions"`
	Statistic   Statistic       `json:"statistic"`
}

type MonthWeekends struct {
	Month int    `json:"month"`
	Days  string `json:"days"`
}

type Transition struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Statistic struct {
	Workdays int     `json:"workdays"`
	Holidays int     `json:"holidays"`
	Hours40  float64 `json:"hours40"`
	Hours36  float64 `json:"hours36"`
	Hours24  float64 `json:"hours24"`
}

// NonWorkingDay - структура для хранения в базе данных
type NonWorkingDay struct {
	Date  time.Time `json:"date"`
	Year  int       `json:"year"`
	Month int       `json:"month"`
	Day   int       `json:"day"`
}

// ParseWeekendsJSON - парсит JSON и возвращает массив выходных дней
func ParseWeekendsJSON(filePath string) ([]NonWorkingDay, error) {
	// Читаем файл
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}

	// Парсим JSON
	var weekendJSON WeekendJSON
	if err := json.Unmarshal(data, &weekendJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	nonWorkingDays := []NonWorkingDay{}

	// Обрабатываем каждый месяц
	for _, monthData := range weekendJSON.Months {
		// Разбиваем строку с днями
		dayStrings := strings.Split(monthData.Days, ",")
		
		for _, dayStr := range dayStrings {
			// Убираем специальные символы (+, *)
			dayStr = strings.TrimSpace(dayStr)
			dayStr = strings.TrimSuffix(dayStr, "+")
			dayStr = strings.TrimSuffix(dayStr, "*")
			
			if dayStr == "" {
				continue
			}
			
			// Парсим день
			day, err := strconv.Atoi(dayStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse day '%s' in month %d: %w", 
					dayStr, monthData.Month, err)
			}
			
			// Создаем дату
			date := time.Date(weekendJSON.Year, time.Month(monthData.Month), day, 0, 0, 0, 0, time.Local)
			
			// Добавляем в результат
			nonWorkingDays = append(nonWorkingDays, NonWorkingDay{
				Date:  date,
				Year:  weekendJSON.Year,
				Month: monthData.Month,
				Day:   day,
			})
		}
	}
	
	// Обрабатываем переносы (transitions)
	// nonWorkingDays = applyTransitions(nonWorkingDays, weekendJSON.Transitions, weekendJSON.Year)

	return nonWorkingDays, nil
}

// GetNonWorkingDaysForMonth - возвращает выходные дни для конкретного месяца
func GetNonWorkingDaysForMonth(days []NonWorkingDay, year, month int) []NonWorkingDay {
	result := []NonWorkingDay{}
	for _, day := range days {
		if day.Year == year && day.Month == month {
			result = append(result, day)
		}
	}
	return result
}

// IsNonWorkingDay - проверяет, является ли дата выходным днем
func IsNonWorkingDay(days []NonWorkingDay, date time.Time) bool {
	for _, day := range days {
		if day.Date.Year() == date.Year() && 
		   day.Date.Month() == date.Month() && 
		   day.Date.Day() == date.Day() {
			return true
		}
	}
	return false
}

// PrintSummary - выводит статистику по выходным дням
func PrintSummary(days []NonWorkingDay, stats Statistic) {
	fmt.Printf("Год: %d\n", days[0].Year)
	fmt.Printf("Всего выходных дней: %d\n", len(days))
	fmt.Printf("Рабочих дней: %d\n", stats.Workdays)
	fmt.Printf("Часов при 40-часовой неделе: %.1f\n", stats.Hours40)
	fmt.Printf("Часов при 36-часовой неделе: %.1f\n", stats.Hours36)
	fmt.Printf("Часов при 24-часовой неделе: %.1f\n", stats.Hours24)
}