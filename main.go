package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB

// User struct to map user data
type User struct {
	ID        int    `db:"id" json:"id"`
	Email     string `db:"email" json:"email" binding:"required,email"`
	Password  string `db:"password" json:"password" binding:"required"`
	CreatedAt string `db:"created_at" json:"created_at"`
	UpdatedAt string `db:"updated_at" json:"updated_at"`
}

// Initialize database connection
func initDB() {
	dbHost, exists := os.LookupEnv("DB_HOST")
	if !exists {
		log.Fatalf("DB_HOST environment variable not set")
	}

	dbPort, exists := os.LookupEnv("DB_PORT")
	if !exists {
		log.Fatalf("DB_PORT environment variable not set")
	}

	dbUser, exists := os.LookupEnv("DB_USER")
	if !exists {
		log.Fatalf("DB_USER environment variable not set")
	}

	dbPassword, exists := os.LookupEnv("DB_PASSWORD")
	if !exists {
		log.Fatalf("DB_PASSWORD environment variable not set")
	}

	dbName, exists := os.LookupEnv("DB_NAME")
	if !exists {
		log.Fatalf("DB_NAME environment variable not set")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	db, err = sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
}

// Function to run migrations
func runMigrations() {
	dbHost, exists := os.LookupEnv("DB_HOST")
	if !exists {
		log.Fatalf("DB_HOST environment variable not set")
	}

	dbPort, exists := os.LookupEnv("DB_PORT")
	if !exists {
		log.Fatalf("DB_PORT environment variable not set")
	}

	dbUser, exists := os.LookupEnv("DB_USER")
	if !exists {
		log.Fatalf("DB_USER environment variable not set")
	}

	dbPassword, exists := os.LookupEnv("DB_PASSWORD")
	if !exists {
		log.Fatalf("DB_PASSWORD environment variable not set")
	}

	dbName, exists := os.LookupEnv("DB_NAME")
	if !exists {
		log.Fatalf("DB_NAME environment variable not set")
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Could not connect to database for migration: %v", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres", driver)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("An error occurred while running the migrations: %v", err)
	}

	log.Println("Migrations ran successfully")
}

func main() {
	// Initialize the database
	initDB()

	// Run migrations
	runMigrations()

	r := gin.Default()

	// Configure CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.POST("/api/user", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			log.Printf("Error binding JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		query := `INSERT INTO users (email, password) VALUES ($1, $2)`
		_, err := db.Exec(query, user.Email, user.Password)
		if err != nil {
			log.Printf("Error inserting user into database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert user"})
			return
		}

		log.Printf("User registered: %s", user.Email)
		c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
	})

	r.GET("/api/users", func(c *gin.Context) {
		var users []User
		err := db.Select(&users, "SELECT id, email, password, created_at, updated_at FROM users")
		if err != nil {
			log.Printf("Error fetching users: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
			return
		}

		log.Printf("Fetched %d users", len(users))
		c.JSON(http.StatusOK, users)
	})

	err := r.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
	log.Println("Server started on port 8080")
}
