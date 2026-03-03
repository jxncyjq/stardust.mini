package skills

import (
	"github.com/jxncyjq/stardust.mini/databases"
)

// SkillDao provides database operations for skills
type SkillDao struct {
	databases.BaseDao
}

// NewSkillDao creates a new SkillDao instance
func NewSkillDao(dao databases.BaseDao) *SkillDao {
	return &SkillDao{BaseDao: dao}
}

// ListAll retrieves all skills from the database
func (d *SkillDao) ListAll() ([]*Skill, error) {
	var skills []*Skill
	err := d.FindMany(&skills, "id ASC")
	if err != nil {
		return nil, err
	}
	return skills, nil
}

// ListByCategory retrieves skills filtered by category
func (d *SkillDao) ListByCategory(category string) ([]*Skill, error) {
	var skills []*Skill
	condition := &Skill{Category: category}
	err := d.FindMany(&skills, "id ASC", condition)
	if err != nil {
		return nil, err
	}
	return skills, nil
}

// ListByLevel retrieves skills filtered by level
func (d *SkillDao) ListByLevel(level string) ([]*Skill, error) {
	var skills []*Skill
	condition := &Skill{Level: level}
	err := d.FindMany(&skills, "id ASC", condition)
	if err != nil {
		return nil, err
	}
	return skills, nil
}

// FindByID retrieves a skill by its ID
func (d *SkillDao) FindByID(id int64) (*Skill, error) {
	skill := &Skill{ID: id}
	found, err := d.FindById(id, skill)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return skill, nil
}
