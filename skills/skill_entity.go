package skills

import (
	"time"

	"github.com/jxncyjq/stardust.mini/databases"
)

// Skill represents a skill entity in the database
type Skill struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	Level       string    `gorm:"column:level;type:varchar(50)" json:"level"`
	Category    string    `gorm:"column:category;type:varchar(100)" json:"category"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName returns the table name for the Skill model
func (s *Skill) TableName() string {
	return "skills"
}

// PrimaryKey returns the primary key value
func (s *Skill) PrimaryKey() interface{} {
	return s.ID
}

// SkillEntity wraps the Skill with entity operations
type SkillEntity struct {
	databases.BaseEntity
	Skill *Skill
}

// NewSkillEntity creates a new skill entity
func NewSkillEntity(dao databases.BaseDao, skill *Skill) *SkillEntity {
	return &SkillEntity{
		BaseEntity: databases.NewEntity(dao, skill),
		Skill:      skill,
	}
}
