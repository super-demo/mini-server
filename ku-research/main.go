package main

import (
	"fmt"
	"ku-research/sdk"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ResearchPaper represents a research paper with visibility settings
type ResearchPaper struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Authors         string   `json:"authors"`
	Abstract        string   `json:"abstract"`
	CoverImage      string   `json:"coverImage"`
	PublishedYear   int      `json:"publishedYear"`
	Field           string   `json:"field"`
	Classifications []string `json:"classifications"`
	DOI             string   `json:"doi,omitempty"`
	Journal         string   `json:"journal,omitempty"`

	// Visibility fields
	UserID          int    `json:"userId"`
	IsPublic        bool   `json:"isPublic"`
	PublicOption    string `json:"publicOption,omitempty"` // "workspace", "site", or "everyone"
	WorkspaceSiteID int    `json:"workspaceSiteID,omitempty"`
}

type WorkspaceUser struct {
	WorkspaceID int `json:"workspaceId"`
	UserID      int `json:"userId"`
}

var (
	papers         []ResearchPaper
	workspaceUsers []WorkspaceUser
	siteUsers      []int // User IDs that belong to site #1
	mu             sync.Mutex
)

func main() {
	// Initialize with sample data
	papers = getSamplePapers()
	workspaceUsers = getSampleWorkspaceUsers()
	siteUsers = getSampleSiteUsers()

	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024,
	})

	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusOK)
		}
		return c.Next()
	})

	app.Use(func(c *fiber.Ctx) error {
		log.Printf("📥 Incoming request to: %s %s\n", c.Method(), c.Path())
		log.Printf("📄 Request body: %s\n", string(c.Body()))
		return c.Next()
	})

	// Get research papers with access control
	app.Post("/get-research", func(c *fiber.Ctx) error {
		// Parse request body to get the requesting user's ID
		var request struct {
			UserID int `json:"userId"`
		}

		if err := c.BodyParser(&request); err != nil {
			log.Printf("❌ Error parsing request: %v\n", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request",
			})
		}

		userID := request.UserID
		log.Printf("🔍 Retrieving papers for user ID: %d\n", userID)

		mu.Lock()
		// Filter papers based on access permissions
		accessiblePapers := filterAccessiblePapers(papers, userID)
		mu.Unlock()

		log.Printf("📤 Sending %d accessible papers\n", len(accessiblePapers))
		return c.JSON(fiber.Map{
			"papers": accessiblePapers,
		})
	})

	// Add new paper endpoint
	app.Post("/add-paper", func(c *fiber.Ctx) error {
		var newPaper ResearchPaper
		if err := c.BodyParser(&newPaper); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid paper data",
			})
		}

		// Validate required fields
		if newPaper.Title == "" || newPaper.Authors == "" || newPaper.Abstract == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing required fields",
			})
		}

		// Generate ID if not provided
		if newPaper.ID == "" {
			newPaper.ID = generateID()
		}

		// Add paper to the database
		mu.Lock()
		papers = append(papers, newPaper)
		mu.Unlock()

		log.Printf("📤 Paper added: %v\n", newPaper)
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Paper added successfully",
			"paper":   newPaper,
		})
	})

	go func() {
		log.Println("📦 Ku Research running at http://localhost:8083")
		log.Println("✅ Ready to accept connections")
		if err := app.Listen(":8083"); err != nil {
			log.Fatalf("❌ Server error: %v\n", err)
		}
	}()

	time.Sleep(1 * time.Second)
	sdk := sdk.NewSuperAppSDK("super-secret-key")
	maxRetries := 5

	for i := range maxRetries {
		log.Printf("Attempting to register with Super App (attempt %d/%d)\n", i+1, maxRetries)
		err := sdk.Register("Ku Research", []string{
			"get-research",
			"add-paper",
		},
			"http://host.docker.internal:8083",
		)
		if err == nil {
			log.Println("✅ Ku Research registered successfully")
			break
		}
		log.Printf("❌ Registration attempt %d failed: %v\n", i+1, err)
		if i < maxRetries-1 {
			log.Println("Waiting before retry...")
			time.Sleep(2 * time.Second)
		} else {
			log.Println("⚠️ All registration attempts failed, but continuing...")
		}
	}

	select {}
}

func filterAccessiblePapers(allPapers []ResearchPaper, userID int) []ResearchPaper {
	var accessiblePapers []ResearchPaper

	for _, paper := range allPapers {
		if hasAccess(paper, userID) {
			accessiblePapers = append(accessiblePapers, paper)
		}
	}

	return accessiblePapers
}

func hasAccess(paper ResearchPaper, userID int) bool {
	if paper.UserID == userID {
		return true
	}

	if !paper.IsPublic {
		return false
	}

	if paper.PublicOption == "everyone" {
		return true
	}

	if paper.PublicOption == "site" {
		for _, siteUserID := range siteUsers {
			if siteUserID == userID {
				return true
			}
		}
		return false
	}

	if paper.PublicOption == "workspace" {
		for _, workspaceUser := range workspaceUsers {
			if workspaceUser.WorkspaceID == paper.WorkspaceSiteID && workspaceUser.UserID == userID {
				return true
			}
		}
		return false
	}

	return false
}

// generateID generates a simple ID for new papers
func generateID() string {
	mu.Lock()
	defer mu.Unlock()
	return fmt.Sprintf("%d", len(papers)+1)
}

func getSamplePapers() []ResearchPaper {
	return []ResearchPaper{
		{
			ID:            "1",
			Title:         "Quantum Computing: Recent Advances and Future Directions",
			Authors:       "Dr. Richard Feynman, Dr. Lisa Chen",
			Abstract:      "This paper reviews recent developments in quantum computing, focusing on quantum supremacy experiments and potential applications in cryptography, optimization, and simulation of quantum systems.",
			CoverImage:    "https://images.unsplash.com/photo-1635070041078-e363dbe005cb?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=2070&q=80",
			PublishedYear: 2023,
			Field:         "Computer Science",
			Classifications: []string{
				"Quantum Computing",
				"Theoretical Physics",
				"Cryptography",
			},
			DOI:          "10.1038/s41586-019-1666-5",
			Journal:      "Nature Quantum Information",
			UserID:       1,
			IsPublic:     true,
			PublicOption: "everyone",
		},
	}
}

func getSampleWorkspaceUsers() []WorkspaceUser {
	return []WorkspaceUser{
		{WorkspaceID: 1, UserID: 2},
		{WorkspaceID: 3, UserID: 2},
		{WorkspaceID: 1, UserID: 3},
		{WorkspaceID: 3, UserID: 3},
		{WorkspaceID: 9, UserID: 3},
		{WorkspaceID: 1, UserID: 4},
		{WorkspaceID: 3, UserID: 4},
		{WorkspaceID: 8, UserID: 4},
	}
}

func getSampleSiteUsers() []int {
	return []int{1, 2, 3, 4}
}
