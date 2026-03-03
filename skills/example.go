package skills

import (
	"fmt"

	"github.com/jxncyjq/stardust.mini/databases"
)

// Example demonstrates how to use the skills list feature
func Example() {
	// 1. Setup database (replace with actual database configuration)
	// dbConfig := []byte(`{"dsn": "your_database_dsn"}`)
	// db, err := databases.NewMySQLDB(dbConfig)
	// if err != nil {
	//     panic(err)
	// }
	// dao := databases.NewBaseDao(db.DB())

	// 2. Create HTTP server (example - commented out)
	// httpConfig := []byte(`{"address":"127.0.0.1","port":8080,"path":"/api"}`)
	// srv, err := httpServer.NewHttpServer(httpConfig)
	// if err != nil {
	//     panic(err)
	// }

	// 3. Initialize skill handler with your DAO
	// skillHandler := NewSkillHandler(dao)

	// 4. Register the list skills endpoint
	// srv.Get("skills", "", skillHandler.ListSkillsHandler())

	// 5. Start the server
	// if err := srv.Startup(); err != nil {
	//     panic(err)
	// }

	fmt.Println("Skills API endpoints:")
	fmt.Println("- GET /api/skills - List all skills")
	fmt.Println("- GET /api/skills?category=Programming - List skills by category")
	fmt.Println("- GET /api/skills?level=Advanced - List skills by level")
}

// ExampleUsage shows how to create and query skills
func ExampleUsage(dao databases.BaseDao) {
	skillDao := NewSkillDao(dao)

	// Create some example skills
	skill1 := &Skill{
		Name:        "Go Programming",
		Description: "Backend development with Go",
		Level:       "Advanced",
		Category:    "Programming",
	}
	skill2 := &Skill{
		Name:        "JavaScript",
		Description: "Frontend development",
		Level:       "Intermediate",
		Category:    "Programming",
	}

	// Use entity pattern to save skills
	entity1 := NewSkillEntity(dao, skill1)
	entity2 := NewSkillEntity(dao, skill2)

	// Create skills in database
	if _, err := entity1.Create(); err != nil {
		fmt.Printf("Failed to create skill1: %v\n", err)
	}
	if _, err := entity2.Create(); err != nil {
		fmt.Printf("Failed to create skill2: %v\n", err)
	}

	// List all skills
	skills, err := skillDao.ListAll()
	if err != nil {
		fmt.Printf("Failed to list skills: %v\n", err)
		return
	}
	fmt.Printf("Total skills: %d\n", len(skills))

	// List skills by category
	programmingSkills, err := skillDao.ListByCategory("Programming")
	if err != nil {
		fmt.Printf("Failed to list programming skills: %v\n", err)
		return
	}
	fmt.Printf("Programming skills: %d\n", len(programmingSkills))

	// List skills by level
	advancedSkills, err := skillDao.ListByLevel("Advanced")
	if err != nil {
		fmt.Printf("Failed to list advanced skills: %v\n", err)
		return
	}
	fmt.Printf("Advanced skills: %d\n", len(advancedSkills))
}
