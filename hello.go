package main

import (
	"math/rand"
	"net/http"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Team represents a baseball team
type Team struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Name    string `gorm:"type:varchar(100)" json:"name"`
	City    string `gorm:"type:varchar(100)" json:"city"`
	Founded int    `json:"founded"`
}

// Player represents a baseball player
type Player struct {
	ID         uint    `gorm:"primaryKey" json:"id"`
	TeamID     uint    `json:"team_id"`
	FirstName  string  `gorm:"type:varchar(50)" json:"first_name"`
	LastName   string  `gorm:"type:varchar(50)" json:"last_name"`
	Position   string  `gorm:"type:varchar(50)" json:"position"`
	BattingAvg float64 `json:"batting_avg"`
}

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"unique" json:"username"`
	Password string `json:"password"`
}

var db *gorm.DB

func main() {
	// Initialize MySQL connection
	dsn := "user:password@tcp(127.0.0.1:3306)/baseball_db?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto-migrate the schema
	db.AutoMigrate(&Team{}, &Player{}, &User{})

	// Seed the database with fake data
	seedDatabase()

	// Set up Gin router
	r := gin.Default()
	r.Use(cors.Default())

	// API endpoints
	r.GET("/teams", getTeams)
	r.GET("/players", getPlayers)
	r.GET("/players/:team_id", getPlayersByTeam)

	// Signup endpoint
	r.POST("/signup", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := db.Create(&user).Error; err != nil {
			c.JSON(400, gin.H{"error": "Username already exists"})
			return
		}
		c.JSON(200, gin.H{"message": "Signup successful"})
	})

	// Login endpoint
	r.POST("/login", func(c *gin.Context) {
		var req User
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		var user User
		if err := db.Where("username = ? AND password = ?", req.Username, req.Password).First(&user).Error; err != nil {
			c.JSON(401, gin.H{"error": "Invalid username or password"})
			return
		}
		c.JSON(200, gin.H{"message": "Login successful"})
	})

	// Start server
	r.Run(":8080")
}

func seedDatabase() {
	// Seed teams
	for i := 0; i < 10; i++ {
		team := Team{
			Name:    gofakeit.Company(),
			City:    gofakeit.City(),
			Founded: gofakeit.Year(),
		}
		db.Create(&team)

		// Seed players for each team
		for j := 0; j < 15; j++ {
			player := Player{
				TeamID:     team.ID,
				FirstName:  gofakeit.FirstName(),
				LastName:   gofakeit.LastName(),
				Position:   randomPosition(),
				BattingAvg: gofakeit.Float64Range(0.200, 0.400),
			}
			db.Create(&player)
		}
	}
}

func randomPosition() string {
	positions := []string{"Pitcher", "Catcher", "First Base", "Second Base", "Third Base", "Shortstop", "Left Field", "Center Field", "Right Field"}
	return positions[rand.Intn(len(positions))]
}

func getTeams(c *gin.Context) {
	var teams []Team
	db.Find(&teams)
	c.JSON(http.StatusOK, teams)
}

func getPlayers(c *gin.Context) {
	var players []Player
	db.Find(&players)
	c.JSON(http.StatusOK, players)
}

func getPlayersByTeam(c *gin.Context) {
	teamID := c.Param("team_id")
	var players []Player
	db.Where("team_id = ?", teamID).Find(&players)
	c.JSON(http.StatusOK, players)
}
