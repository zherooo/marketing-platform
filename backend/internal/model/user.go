package model

import "time"

// User 用户表
type User struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	Username   string    `gorm:"uniqueIndex;size:50;not null;comment:用户名" json:"username"`
	Password   string    `gorm:"size:255;not null;comment:密码(加密)" json:"-"` // 不返回密码
	Email      string    `gorm:"size:100;comment:邮箱" json:"email"`
	Nickname   string    `gorm:"size:100;comment:昵称" json:"nickname"`
	Status     int       `gorm:"default:1;comment:状态(1正常,0禁用)" json:"status"`
	LastLogin  *time.Time `gorm:"comment:最后登录时间" json:"last_login"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
