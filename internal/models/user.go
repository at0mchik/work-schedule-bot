package models

type Role string

const (
    RoleClient string = "client"
    RoleAdmin  string = "admin"
)

type User struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt int64          `json:"created_at"`
    UpdatedAt int64          `json:"updated_at"`
    ChatID    int64          `gorm:"uniqueIndex;not null" json:"chat_id"`
    Username  string         `json:"username"`
    FirstName string         `gorm:"not null" json:"first_name"`
    LastName  string         `json:"last_name"`
    Role      string         `gorm:"default:'client'" json:"role"` // string вместо Role
}

// IsAdmin проверяет, является ли пользователь администратором
func (u *User) IsAdmin() bool {
    return u.Role == "admin"
}

// SetRole устанавливает роль
func (u *User) SetRole(role Role) {
    u.Role = string(role)
}

// TableName задает имя таблицы в БД
func (User) TableName() string {
    return "users"
}