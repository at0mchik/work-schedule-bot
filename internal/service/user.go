package service

import (
	"fmt"
	"strings"
	"work-schedule-bot/internal/models"
	"work-schedule-bot/internal/repository"
)

type UserService struct {
    repo *repository.GormUserRepository
}

func NewUserService(repo *repository.GormUserRepository) *UserService {
    return &UserService{repo: repo}
}

// CreateUser создает нового пользователя с ролью client по умолчанию
func (s *UserService) CreateUser(chatID int64, username, firstName, lastName string) (*models.User, error) {
    // Проверяем валидность данных
    if firstName == "" {
        return nil, fmt.Errorf("имя не может быть пустым")
    }

    // Создаем пользователя с ролью client
    user := &models.User{
        ChatID:    chatID,
        Username:  username,
        FirstName: firstName,
        LastName:  lastName,
        Role:      models.RoleClient, // По умолчанию client
    }

    err := s.repo.Create(user)
    if err != nil {
        return nil, fmt.Errorf("ошибка создания пользователя: %v", err)
    }

    return user, nil
}

// GetUser возвращает пользователя по chatID
func (s *UserService) GetUser(chatID int64) (*models.User, error) {
    user, err := s.repo.GetByChatID(chatID)
    if err != nil {
        return nil, fmt.Errorf("ошибка получения пользователя: %v", err)
    }
    
    if user == nil {
        return nil, fmt.Errorf("пользователь не найден")
    }
    
    return user, nil
}

// UpdateUser обновляет данные пользователя
func (s *UserService) UpdateUser(chatID int64, username, firstName, lastName string) (*models.User, error) {
    user, err := s.repo.GetByChatID(chatID)
    if err != nil {
        return nil, fmt.Errorf("ошибка получения пользователя: %v", err)
    }
    
    if user == nil {
        return nil, fmt.Errorf("пользователь не найден")
    }

    // Обновляем поля (кроме роли)
    if username != "" {
        user.Username = username
    }
    if firstName != "" {
        user.FirstName = firstName
    }
    if lastName != "" {
        user.LastName = lastName
    }

    err = s.repo.Update(user)
    if err != nil {
        return nil, fmt.Errorf("ошибка обновления пользователя: %v", err)
    }

    return user, nil
}

// UpdateRole обновляет роль пользователя (только для админов)
func (s *UserService) UpdateRole(adminChatID, targetChatID int64, role models.Role) error {
    // Проверяем, что админ существует и является админом
    admin, err := s.repo.GetByChatID(adminChatID)
    if err != nil {
        return fmt.Errorf("ошибка проверки админа: %v", err)
    }
    
    if admin == nil || !admin.IsAdmin() {
        return fmt.Errorf("доступ запрещен: только администраторы могут менять роли")
    }

    // Проверяем, что целевой пользователь существует
    targetUser, err := s.repo.GetByChatID(targetChatID)
    if err != nil {
        return fmt.Errorf("ошибка поиска пользователя: %v", err)
    }
    
    if targetUser == nil {
        return fmt.Errorf("пользователь не найден")
    }

    // Обновляем роль
    return s.repo.UpdateRole(targetChatID, role)
}

// FormatUserInfo форматирует информацию о пользователе для вывода
func (s *UserService) FormatUserInfo(user *models.User) string {
    var lines []string
    
    lines = append(lines, "👤 Профиль пользователя:")
    lines = append(lines, "")
    lines = append(lines, fmt.Sprintf("🆔 ID чата: %d", user.ChatID))
    
    if user.Username != "" {
        lines = append(lines, fmt.Sprintf("📛 Никнейм: @%s", user.Username))
    }
    
    lines = append(lines, fmt.Sprintf("👨‍💼 Имя: %s", user.FirstName))
    
    if user.LastName != "" {
        lines = append(lines, fmt.Sprintf("👨‍💼 Фамилия: %s", user.LastName))
    }
    
    // Добавляем информацию о роли
    roleEmoji := "👤"
    if user.IsAdmin() {
        roleEmoji = "👑"
    }
    lines = append(lines, fmt.Sprintf("%s Роль: %s", roleEmoji, string(user.Role)))
    
    return strings.Join(lines, "\n")
}

// DeleteUser удаляет пользователя
func (s *UserService) DeleteUser(chatID int64) error {
    exists, err := s.repo.Exists(chatID)
    if err != nil {
        return fmt.Errorf("ошибка проверки пользователя: %v", err)
    }
    
    if !exists {
        return fmt.Errorf("пользователь не найден")
    }
    
    return s.repo.Delete(chatID)
}

// GetAllUsers возвращает всех пользователей
func (s *UserService) GetAllUsers() ([]*models.User, error) {
    return s.repo.GetAll()
}

// GetAdmins возвращает всех администраторов
func (s *UserService) GetAdmins() ([]*models.User, error) {
    return s.repo.GetAdmins()
}

// GetStats возвращает статистику
func (s *UserService) GetStats() (int, int, error) {
    return s.repo.GetStats()
}

// FormatAllUsers форматирует список всех пользователей
func (s *UserService) FormatAllUsers() (string, error) {
    users, err := s.GetAllUsers()
    if err != nil {
        return "", err
    }

    if len(users) == 0 {
        return "📭 Список пользователей пуст.", nil
    }

    var lines []string
    lines = append(lines, "📋 Все пользователи:")
    lines = append(lines, "")
    
    for i, user := range users {
        roleEmoji := "👤"
        if user.IsAdmin() {
            roleEmoji = "👑"
        }
        
        userInfo := fmt.Sprintf("%d. %s ", i+1, roleEmoji)
        if user.FirstName != "" {
            userInfo += user.FirstName + " "
        }
        if user.LastName != "" {
            userInfo += user.LastName + " "
        }
        if user.Username != "" {
            userInfo += fmt.Sprintf("(@%s) ", user.Username)
        }
        userInfo += fmt.Sprintf("- ID: %d", user.ChatID)
        lines = append(lines, userInfo)
    }

    total, admins, _ := s.GetStats()
    lines = append(lines, "")
    lines = append(lines, fmt.Sprintf("📊 Всего пользователей: %d", total))
    lines = append(lines, fmt.Sprintf("👑 Администраторов: %d", admins))

    return strings.Join(lines, "\n"), nil
}

// IsAdmin проверяет, является ли пользователь администратором
func (s *UserService) IsAdmin(chatID int64) (bool, error) {
    user, err := s.repo.GetByChatID(chatID)
    if err != nil {
        return false, err
    }
    
    return user != nil && user.IsAdmin(), nil
}

// InitializeAdmin инициализирует администратора из конфига
func (s *UserService) InitializeAdmin(adminChatID int64) error {
    if adminChatID == 0 {
        return nil // Админ не задан в конфиге
    }

    // Проверяем, существует ли уже пользователь с таким chatID
    existingUser, err := s.repo.GetByChatID(adminChatID)
    if err != nil {
        return err
    }

    if existingUser != nil {
        // Если пользователь существует, обновляем его роль на админа
        return s.repo.UpdateRole(adminChatID, "admin")
    }

    // Создаем нового администратора
    adminUser := &models.User{
        ChatID:    adminChatID,
        Username:  "admin",
        FirstName: "Администратор",
        LastName:  "",
        Role:      models.RoleAdmin,
    }

    return s.repo.Create(adminUser)
}