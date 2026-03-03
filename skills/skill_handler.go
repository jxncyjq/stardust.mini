package skills

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust.mini/databases"
	httpServer "github.com/jxncyjq/stardust.mini/http_server"
)

// ListSkillsRequest represents the request for listing skills
type ListSkillsRequest struct {
	Category string `form:"category" json:"category"`
	Level    string `form:"level" json:"level"`
}

// ListSkillsResponse represents the response for listing skills
type ListSkillsResponse struct {
	Skills []*Skill `json:"skills"`
	Total  int      `json:"total"`
}

// SkillHandler handles skill-related HTTP requests
type SkillHandler struct {
	dao *SkillDao
}

// NewSkillHandler creates a new SkillHandler
func NewSkillHandler(dao databases.BaseDao) *SkillHandler {
	return &SkillHandler{
		dao: NewSkillDao(dao),
	}
}

// ListSkillsHandler creates a handler for listing skills
func (h *SkillHandler) ListSkillsHandler() httpServer.IHandler {
	return httpServer.NewHandler(
		"list_skills",
		[]string{"skills"},
		func(c *gin.Context, req ListSkillsRequest, resp ListSkillsResponse) error {
			var skills []*Skill
			var err error

			// Filter by category if provided
			if req.Category != "" {
				skills, err = h.dao.ListByCategory(req.Category)
			} else if req.Level != "" {
				// Filter by level if provided
				skills, err = h.dao.ListByLevel(req.Level)
			} else {
				// Return all skills
				skills, err = h.dao.ListAll()
			}

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return err
			}

			// Return the response
			c.JSON(http.StatusOK, ListSkillsResponse{
				Skills: skills,
				Total:  len(skills),
			})
			return nil
		},
	)
}
