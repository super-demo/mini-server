package main

import (
	"fmt"
	"ku-research/sdk"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ResearchPaper represents a research paper
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
}

// Database to store papers
var (
	papers []ResearchPaper
	mu     sync.Mutex
)

func main() {
	// Initialize with sample data
	papers = getSamplePapers()

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

	// Get all papers endpoint
	app.Post("/get-research", func(c *fiber.Ctx) error {
		mu.Lock()
		allPapers := papers
		mu.Unlock()

		log.Printf("📤 Sending response: %v\n", allPapers)
		return c.JSON(fiber.Map{
			"papers": allPapers,
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
		})
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
			CoverImage:    "https://scontent-bkk1-2.xx.fbcdn.net/v/t39.30808-6/476509205_555409170843779_2189911961991409505_n.jpg?stp=dst-jpg_s1080x2048_tt6&_nc_cat=107&ccb=1-7&_nc_sid=cc71e4&_nc_ohc=dvVyY7fKDE4Q7kNvgEUsxY5&_nc_oc=Adm7pMZkVfAy3GQbposPMK9VpKk81CyRhNdrr2JqPgAp0TnQlYyuOJT7nelQ2oDVmBM&_nc_zt=23&_nc_ht=scontent-bkk1-2.xx&_nc_gid=eBOCDkseztFEgrAKMda5Xg&oh=00_AYGH3QWile1EBMI5YsbJZHdZDO8Ut6TAyNEjyCzqITjDEw&oe=67DF7CDD",
			PublishedYear: 2023,
			Field:         "Computer Science",
			Classifications: []string{
				"Quantum Computing",
				"Theoretical Physics",
				"Cryptography",
			},
			DOI:     "10.1038/s41586-019-1666-5",
			Journal: "Nature Quantum Information",
		},
		{
			ID:            "2",
			Title:         "Climate Change Impact on Marine Ecosystems: A Comprehensive Analysis",
			Authors:       "Dr. Sarah Johnson, Dr. Michael Rodriguez",
			Abstract:      "This research presents findings from a decade-long study on the effects of rising ocean temperatures and acidification on coral reefs and marine biodiversity across multiple climate zones.",
			CoverImage:    "https://scontent-bkk1-2.xx.fbcdn.net/v/t39.30808-6/476509205_555409170843779_2189911961991409505_n.jpg?stp=dst-jpg_s1080x2048_tt6&_nc_cat=107&ccb=1-7&_nc_sid=cc71e4&_nc_ohc=dvVyY7fKDE4Q7kNvgEUsxY5&_nc_oc=Adm7pMZkVfAy3GQbposPMK9VpKk81CyRhNdrr2JqPgAp0TnQlYyuOJT7nelQ2oDVmBM&_nc_zt=23&_nc_ht=scontent-bkk1-2.xx&_nc_gid=eBOCDkseztFEgrAKMda5Xg&oh=00_AYGH3QWile1EBMI5YsbJZHdZDO8Ut6TAyNEjyCzqITjDEw&oe=67DF7CDD",
			PublishedYear: 2023,
			Field:         "Environmental Science",
			Classifications: []string{
				"Climate Change", "Marine Biology", "Ecology",
			},
			DOI:     "10.1126/science.abc1234",
			Journal: "Science",
		},
		{
			ID:            "3",
			Title:         "Artificial Intelligence Ethics: Frameworks for Responsible Development",
			Authors:       "Dr. Alan Turing Institute, Dr. Grace Hopper",
			Abstract:      "This paper examines ethical considerations in AI development, proposing frameworks for addressing bias, transparency, privacy, and accountability in machine learning systems.",
			CoverImage:    "https://scontent-bkk1-2.xx.fbcdn.net/v/t39.30808-6/476509205_555409170843779_2189911961991409505_n.jpg?stp=dst-jpg_s1080x2048_tt6&_nc_cat=107&ccb=1-7&_nc_sid=cc71e4&_nc_ohc=dvVyY7fKDE4Q7kNvgEUsxY5&_nc_oc=Adm7pMZkVfAy3GQbposPMK9VpKk81CyRhNdrr2JqPgAp0TnQlYyuOJT7nelQ2oDVmBM&_nc_zt=23&_nc_ht=scontent-bkk1-2.xx&_nc_gid=eBOCDkseztFEgrAKMda5Xg&oh=00_AYGH3QWile1EBMI5YsbJZHdZDO8Ut6TAyNEjyCzqITjDEw&oe=67DF7CDD",
			PublishedYear: 2023,
			Field:         "Computer Science",
			Classifications: []string{
				"Artificial Intelligence", "Ethics", "Policy",
			},
			DOI:     "10.1145/3442188.3445901",
			Journal: "ACM Conference on Fairness, Accountability, and Transparency",
		},
		{
			ID:            "4",
			Title:         "CRISPR-Cas9 Applications in Treating Genetic Disorders",
			Authors:       "Dr. Jennifer Doudna, Dr. Feng Zhang",
			Abstract:      "This research explores recent advances in CRISPR-Cas9 gene editing technology and its potential applications in treating genetic disorders such as cystic fibrosis, sickle cell anemia, and Huntington's disease.",
			CoverImage:    "https://scontent-bkk1-2.xx.fbcdn.net/v/t39.30808-6/476509205_555409170843779_2189911961991409505_n.jpg?stp=dst-jpg_s1080x2048_tt6&_nc_cat=107&ccb=1-7&_nc_sid=cc71e4&_nc_ohc=dvVyY7fKDE4Q7kNvgEUsxY5&_nc_oc=Adm7pMZkVfAy3GQbposPMK9VpKk81CyRhNdrr2JqPgAp0TnQlYyuOJT7nelQ2oDVmBM&_nc_zt=23&_nc_ht=scontent-bkk1-2.xx&_nc_gid=eBOCDkseztFEgrAKMda5Xg&oh=00_AYGH3QWile1EBMI5YsbJZHdZDO8Ut6TAyNEjyCzqITjDEw&oe=67DF7CDD",
			PublishedYear: 2022,
			Field:         "Biotechnology",
			Classifications: []string{
				"CRISPR-Cas9", "Gene Editing", "Genetic Disorders",
			},
			DOI:     "10.1016/j.cell.2022.01.035",
			Journal: "Cell",
		},
		{
			ID:            "5",
			Title:         "Neuroplasticity and Cognitive Rehabilitation After Traumatic Brain Injury",
			Authors:       "Dr. Maya Rodriguez, Dr. James Wilson",
			Abstract:      "This paper presents findings on brain plasticity mechanisms and their implications for developing effective cognitive rehabilitation strategies for patients recovering from traumatic brain injuries.",
			CoverImage:    "https://scontent-bkk1-2.xx.fbcdn.net/v/t39.30808-6/476509205_555409170843779_2189911961991409505_n.jpg?stp=dst-jpg_s1080x2048_tt6&_nc_cat=107&ccb=1-7&_nc_sid=cc71e4&_nc_ohc=dvVyY7fKDE4Q7kNvgEUsxY5&_nc_oc=Adm7pMZkVfAy3GQbposPMK9VpKk81CyRhNdrr2JqPgAp0TnQlYyuOJT7nelQ2oDVmBM&_nc_zt=23&_nc_ht=scontent-bkk1-2.xx&_nc_gid=eBOCDkseztFEgrAKMda5Xg&oh=00_AYGH3QWile1EBMI5YsbJZHdZDO8Ut6TAyNEjyCzqITjDEw&oe=67DF7CDD",
			PublishedYear: 2022,
			Field:         "Neuroscience",
			Classifications: []string{
				"Neuroplasticity", "Cognitive Rehabilitation", "Brain Injury",
			},
			DOI:     "10.1093/brain/awab123",
			Journal: "Brain",
		},
		{
			ID:            "6",
			Title:         "Renewable Energy Integration: Challenges and Solutions for Power Grids",
			Authors:       "Dr. Elena Patel, Dr. Thomas Schmidt",
			Abstract:      "This research addresses the technical challenges of integrating large-scale renewable energy sources into existing power grids, proposing solutions for energy storage, demand response, and grid stability.",
			CoverImage:    "https://scontent-bkk1-2.xx.fbcdn.net/v/t39.30808-6/476509205_555409170843779_2189911961991409505_n.jpg?stp=dst-jpg_s1080x2048_tt6&_nc_cat=107&ccb=1-7&_nc_sid=cc71e4&_nc_ohc=dvVyY7fKDE4Q7kNvgEUsxY5&_nc_oc=Adm7pMZkVfAy3GQbposPMK9VpKk81CyRhNdrr2JqPgAp0TnQlYyuOJT7nelQ2oDVmBM&_nc_zt=23&_nc_ht=scontent-bkk1-2.xx&_nc_gid=eBOCDkseztFEgrAKMda5Xg&oh=00_AYGH3QWile1EBMI5YsbJZHdZDO8Ut6TAyNEjyCzqITjDEw&oe=67DF7CDD",
			PublishedYear: 2023,
			Field:         "Energy Engineering",
			Classifications: []string{
				"Renewable Energy", "Power Systems", "Energy Storage",
			},
			DOI:     "10.1109/tpwrs.2022.3156789",
			Journal: "IEEE Transactions on Power Systems",
		},
		{
			ID:            "7",
			Title:         "Machine Learning Approaches to Drug Discovery and Development",
			Authors:       "Dr. David Kim, Dr. Rachel Martinez",
			Abstract:      "This paper reviews machine learning techniques applied to drug discovery, including virtual screening, de novo drug design, and prediction of pharmacokinetic properties and toxicity.",
			CoverImage:    "https://scontent-bkk1-2.xx.fbcdn.net/v/t39.30808-6/476509205_555409170843779_2189911961991409505_n.jpg?stp=dst-jpg_s1080x2048_tt6&_nc_cat=107&ccb=1-7&_nc_sid=cc71e4&_nc_ohc=dvVyY7fKDE4Q7kNvgEUsxY5&_nc_oc=Adm7pMZkVfAy3GQbposPMK9VpKk81CyRhNdrr2JqPgAp0TnQlYyuOJT7nelQ2oDVmBM&_nc_zt=23&_nc_ht=scontent-bkk1-2.xx&_nc_gid=eBOCDkseztFEgrAKMda5Xg&oh=00_AYGH3QWile1EBMI5YsbJZHdZDO8Ut6TAyNEjyCzqITjDEw&oe=67DF7CDD",
			PublishedYear: 2023,
			Field:         "Pharmaceutical Science",
			Classifications: []string{
				"Machine Learning",
				"Drug Discovery",
				"Computational Chemistry",
			},
			DOI:     "10.1021/acs.jmedchem.2c01699",
			Journal: "Journal of Medicinal Chemistry",
		},
		{
			ID:            "8",
			Title:         "Sustainable Urban Planning: Integrating Green Infrastructure for Climate Resilience",
			Authors:       "Dr. Carlos Mendez, Dr. Sophia Lee",
			Abstract:      "This research examines strategies for incorporating green infrastructure into urban planning to enhance climate resilience, reduce urban heat islands, and improve stormwater management.",
			CoverImage:    "https://scontent-bkk1-2.xx.fbcdn.net/v/t39.30808-6/476509205_555409170843779_2189911961991409505_n.jpg?stp=dst-jpg_s1080x2048_tt6&_nc_cat=107&ccb=1-7&_nc_sid=cc71e4&_nc_ohc=dvVyY7fKDE4Q7kNvgEUsxY5&_nc_oc=Adm7pMZkVfAy3GQbposPMK9VpKk81CyRhNdrr2JqPgAp0TnQlYyuOJT7nelQ2oDVmBM&_nc_zt=23&_nc_ht=scontent-bkk1-2.xx&_nc_gid=eBOCDkseztFEgrAKMda5Xg&oh=00_AYGH3QWile1EBMI5YsbJZHdZDO8Ut6TAyNEjyCzqITjDEw&oe=67DF7CDD",
			PublishedYear: 2022,
			Field:         "Urban Planning",
			Classifications: []string{
				"Sustainability",
				"Climate Resilience",
				"Green Infrastructure",
			},
			DOI:     "10.1016/j.landurbplan.2022.104567",
			Journal: "Landscape and Urban Planning",
		},
		{
			ID:            "9",
			Title:         "Blockchain Technology for Supply Chain Transparency and Traceability",
			Authors:       "Dr. Satoshi Nakamoto, Dr. Vitalik Buterin",
			Abstract:      "This paper explores applications of blockchain technology in enhancing supply chain transparency, traceability, and security, with case studies from food, pharmaceutical, and luxury goods industries.",
			CoverImage:    "https://scontent-bkk1-2.xx.fbcdn.net/v/t39.30808-6/476509205_555409170843779_2189911961991409505_n.jpg?stp=dst-jpg_s1080x2048_tt6&_nc_cat=107&ccb=1-7&_nc_sid=cc71e4&_nc_ohc=dvVyY7fKDE4Q7kNvgEUsxY5&_nc_oc=Adm7pMZkVfAy3GQbposPMK9VpKk81CyRhNdrr2JqPgAp0TnQlYyuOJT7nelQ2oDVmBM&_nc_zt=23&_nc_ht=scontent-bkk1-2.xx&_nc_gid=eBOCDkseztFEgrAKMda5Xg&oh=00_AYGH3QWile1EBMI5YsbJZHdZDO8Ut6TAyNEjyCzqITjDEw&oe=67DF7CDD",
			PublishedYear: 2023,
			Field:         "Computer Science",
			Classifications: []string{
				"Blockchain",
				"Supply Chain Management",
				"Information Security",
			},
			DOI:     "10.1109/access.2023.1234567",
			Journal: "IEEE Access",
		},
		{
			ID:            "10",
			Title:         "Immunotherapy Advances in Cancer Treatment: Personalized Approaches",
			Authors:       "Dr. James Allison, Dr. Tasuku Honjo",
			Abstract:      "This research reviews recent advances in cancer immunotherapy, focusing on personalized approaches such as CAR-T cell therapy, checkpoint inhibitors, and neoantigen vaccines.",
			CoverImage:    "https://scontent-bkk1-2.xx.fbcdn.net/v/t39.30808-6/476509205_555409170843779_2189911961991409505_n.jpg?stp=dst-jpg_s1080x2048_tt6&_nc_cat=107&ccb=1-7&_nc_sid=cc71e4&_nc_ohc=dvVyY7fKDE4Q7kNvgEUsxY5&_nc_oc=Adm7pMZkVfAy3GQbposPMK9VpKk81CyRhNdrr2JqPgAp0TnQlYyuOJT7nelQ2oDVmBM&_nc_zt=23&_nc_ht=scontent-bkk1-2.xx&_nc_gid=eBOCDkseztFEgrAKMda5Xg&oh=00_AYGH3QWile1EBMI5YsbJZHdZDO8Ut6TAyNEjyCzqITjDEw&oe=67DF7CDD",
			PublishedYear: 2023,
			Field:         "Oncology",
			Classifications: []string{
				"Immunotherapy",
				"Cancer Research",
				"Personalized Medicine",
			},
			DOI:     "10.1038/s41591-023-02345-0",
			Journal: "Nature Medicine",
		},
	}
}
