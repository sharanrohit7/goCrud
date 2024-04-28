package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/lib/pq"
)

// User represents a user entity
type User struct {
	ID       int
	Username string
	Email    string
	Password string
}

// Handler represents the dependencies needed by handler functions
type Handler struct {
	DB *sql.DB
}

// NewHandler initializes a new instance of the Handler struct
func NewHandler(db *sql.DB) *Handler {
	return &Handler{DB: db}
}

// GetUsers retrieves users from the database
func (h *Handler) GetUsers(c *gin.Context) {
	// Query the database for users
	rows, err := h.DB.Query("SELECT * FROM users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve users"})
		return
	}
	defer rows.Close()

	// Iterate over the rows and populate the users slice
	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan user row"})
			return
		}
		users = append(users, user)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to iterate over user rows"})
		return
	}

	// Return the users
	c.JSON(http.StatusOK, users)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var user User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}
	if user.Username == "" || user.Email == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username, email, and password are required"})
		return
	}
	var id int64
	err := h.DB.QueryRow("INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id",
		user.Username, user.Email, user.Password).Scan(&id)

	if err != nil {
		// Check if the error is a duplicate key violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			// Handle duplicate key violation
			c.JSON(http.StatusConflict, gin.H{"error": "username or email already exists"})
			return
		}

		// Handle other database errors
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Account created successfully", "id": id, "is_verified": false})
}

//SignIn

type Claims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

func (h *Handler) SignInHandler(c *gin.Context) {
	var reqBody struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&reqBody); err != nil {
		log.Printf("Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Fetch user from the database
	var userID int
	var storedPassword string
	var isVerified bool
	err := h.DB.QueryRow("SELECT id, password, is_verified FROM users WHERE username = $1 AND is_delete = false", reqBody.Username).Scan(&userID, &storedPassword, &isVerified)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
			return
		}
		log.Printf("Error querying database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate user"})
		return
	}

	// Compare passwords
	if reqBody.Password != storedPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
	})
	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		log.Printf("Error generating JWT token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Return token and user ID
	c.JSON(http.StatusOK, gin.H{"token": tokenString, "user_id": userID, "is_verified": isVerified})
}

func (h *Handler) GetUserById(c *gin.Context) {
	// Get the user ID from the URL parameters
	userID := c.Param("id")

	// Query the database to retrieve user details using JOIN
	query := `
		SELECT u.id, u.username, u.email, p.full_name, p.age, p.gender
		FROM users u
		LEFT JOIN user_profile p ON u.id = p.user_id
		WHERE u.id = $1
	`

	// Execute the query
	row := h.DB.QueryRow(query, userID)

	// Scan the row into a User struct
	var user User
	var profile UserProfile
	err := row.Scan(&user.ID, &user.Username, &user.Email, &profile.FullName, &profile.Age, &profile.Gender)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user"})
		return
	}

	// Return the user details
	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"profile": profile,
	})
}
