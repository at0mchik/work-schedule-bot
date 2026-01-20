package repository

import (
	"errors"
	"work-schedule-bot/internal/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) (UserRepository, error) {
	// Автомиграция - создает таблицы если их нет
	err := db.AutoMigrate(&models.User{})
	if err != nil {
		return UserRepository{}, err
	}

	return UserRepository{db: db}, nil
}

func (r *UserRepository) Create(user *models.User) error {
	// Проверяем, существует ли уже пользователь
	var existingUser models.User
	result := r.db.Where("chat_id = ?", user.ChatID).First(&existingUser)
	if result.Error == nil {
		return errors.New("пользователь уже существует")
	}

	result = r.db.Create(user)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *UserRepository) GetByChatID(chatID int64) (*models.User, error) {
	var user models.User
	result := r.db.Where("chat_id = ?", chatID).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	// Проверяем существование пользователя
	var existingUser models.User
	result := r.db.Where("chat_id = ?", user.ChatID).First(&existingUser)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("пользователь не найден")
	}

	result = r.db.Save(user)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *UserRepository) Delete(chatID int64) error {
	result := r.db.Where("chat_id = ?", chatID).Delete(&models.User{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("пользователь не найден")
	}

	return nil
}

func (r *UserRepository) Exists(chatID int64) (bool, error) {
	var count int64
	result := r.db.Model(&models.User{}).Where("chat_id = ?", chatID).Count(&count)

	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}

func (r *UserRepository) GetAll() ([]*models.User, error) {
	var users []*models.User
	result := r.db.Find(&users)

	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}

func (r *UserRepository) UpdateRole(chatID int64, role models.Role) error {
	result := r.db.Model(&models.User{}).
		Where("chat_id = ?", chatID).
		Update("role", role)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("пользователь не найден")
	}

	return nil
}

func (r *UserRepository) GetAdmins() ([]*models.User, error) {
	var admins []*models.User
	result := r.db.Where("role = ?", models.RoleAdmin).Find(&admins)

	if result.Error != nil {
		return nil, result.Error
	}

	return admins, nil
}

func (r *UserRepository) GetStats() (int, int, error) {
	var total int64
	var admins int64

	// Получаем общее количество пользователей
	result := r.db.Model(&models.User{}).Count(&total)
	if result.Error != nil {
		return 0, 0, result.Error
	}

	// Получаем количество администраторов
	result = r.db.Model(&models.User{}).
		Where("role = ?", models.RoleAdmin).
		Count(&admins)
	if result.Error != nil {
		return 0, 0, result.Error
	}

	return int(total), int(admins), nil
}

// Дополнительные методы для GORM
func (r *UserRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
