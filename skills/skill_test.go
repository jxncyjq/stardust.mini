package skills

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// MockSkillDao is a mock implementation of SkillDao for testing
type MockSkillDao struct {
	skills []*Skill
}

func (m *MockSkillDao) ListAll() ([]*Skill, error) {
	return m.skills, nil
}

func (m *MockSkillDao) ListByCategory(category string) ([]*Skill, error) {
	var filtered []*Skill
	for _, skill := range m.skills {
		if skill.Category == category {
			filtered = append(filtered, skill)
		}
	}
	return filtered, nil
}

func (m *MockSkillDao) ListByLevel(level string) ([]*Skill, error) {
	var filtered []*Skill
	for _, skill := range m.skills {
		if skill.Level == level {
			filtered = append(filtered, skill)
		}
	}
	return filtered, nil
}

func TestListSkillsAll(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock data
	mockSkills := []*Skill{
		{ID: 1, Name: "Go Programming", Description: "Backend development with Go", Level: "Advanced", Category: "Programming"},
		{ID: 2, Name: "JavaScript", Description: "Frontend development", Level: "Intermediate", Category: "Programming"},
		{ID: 3, Name: "Docker", Description: "Containerization", Level: "Intermediate", Category: "DevOps"},
	}

	// Setup
	router := gin.Default()
	mockDao := &MockSkillDao{skills: mockSkills}

	// Create handler function
	router.GET("/api/skills", func(c *gin.Context) {
		skills, _ := mockDao.ListAll()
		c.JSON(http.StatusOK, ListSkillsResponse{
			Skills: skills,
			Total:  len(skills),
		})
	})

	// Test
	req := httptest.NewRequest(http.MethodGet, "/api/skills", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response ListSkillsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 3, response.Total)
	assert.Equal(t, len(mockSkills), len(response.Skills))
}

func TestListSkillsByCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock data
	mockSkills := []*Skill{
		{ID: 1, Name: "Go Programming", Description: "Backend development with Go", Level: "Advanced", Category: "Programming"},
		{ID: 2, Name: "JavaScript", Description: "Frontend development", Level: "Intermediate", Category: "Programming"},
		{ID: 3, Name: "Docker", Description: "Containerization", Level: "Intermediate", Category: "DevOps"},
	}

	// Setup
	router := gin.Default()
	mockDao := &MockSkillDao{skills: mockSkills}

	router.GET("/api/skills", func(c *gin.Context) {
		category := c.Query("category")
		var skills []*Skill
		if category != "" {
			skills, _ = mockDao.ListByCategory(category)
		} else {
			skills, _ = mockDao.ListAll()
		}
		c.JSON(http.StatusOK, ListSkillsResponse{
			Skills: skills,
			Total:  len(skills),
		})
	})

	// Test filtering by Programming category
	req := httptest.NewRequest(http.MethodGet, "/api/skills?category=Programming", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response ListSkillsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, response.Total)
	assert.Equal(t, "Programming", response.Skills[0].Category)
	assert.Equal(t, "Programming", response.Skills[1].Category)
}

func TestListSkillsByLevel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock data
	mockSkills := []*Skill{
		{ID: 1, Name: "Go Programming", Description: "Backend development with Go", Level: "Advanced", Category: "Programming"},
		{ID: 2, Name: "JavaScript", Description: "Frontend development", Level: "Intermediate", Category: "Programming"},
		{ID: 3, Name: "Docker", Description: "Containerization", Level: "Intermediate", Category: "DevOps"},
	}

	// Setup
	router := gin.Default()
	mockDao := &MockSkillDao{skills: mockSkills}

	router.GET("/api/skills", func(c *gin.Context) {
		level := c.Query("level")
		var skills []*Skill
		if level != "" {
			skills, _ = mockDao.ListByLevel(level)
		} else {
			skills, _ = mockDao.ListAll()
		}
		c.JSON(http.StatusOK, ListSkillsResponse{
			Skills: skills,
			Total:  len(skills),
		})
	})

	// Test filtering by Intermediate level
	req := httptest.NewRequest(http.MethodGet, "/api/skills?level=Intermediate", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response ListSkillsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, response.Total)
	assert.Equal(t, "Intermediate", response.Skills[0].Level)
	assert.Equal(t, "Intermediate", response.Skills[1].Level)
}

func TestSkillTableName(t *testing.T) {
	skill := &Skill{}
	assert.Equal(t, "skills", skill.TableName())
}

func TestSkillPrimaryKey(t *testing.T) {
	skill := &Skill{ID: 123}
	assert.Equal(t, int64(123), skill.PrimaryKey())
}
