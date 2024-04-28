package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type UserProfile struct {
	FullName string `json:"full_name"`
	Age      int    `json:"age"`
	Gender   string `json:"gender"`
}

func (h *Handler) CreateProfile(c *gin.Context) {
	// Extract user ID from token
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var profile UserProfile
	if err := c.BindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}
	if profile.FullName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "full name is required"})
		return
	}

	// Start a transaction
	tx, err := h.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer func() {
		if err != nil {

			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Error rolling back transaction: %v", rbErr)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create profile"})
			return
		}
	}()

	// Insert profile data into the user_profile table
	_, err = tx.Exec("INSERT INTO user_profile (user_id, full_name, age, gender) VALUES ($1, $2, $3, $4)",
		userID, profile.FullName, profile.Age, profile.Gender)
	if err != nil {
		// Check if the error is a foreign key constraint violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user does not exist"})
			return
		}

		log.Printf("Error inserting profile data into user_profile table: %v", err)
		return
	}

	// Update the is_verified flag in the users table
	_, err = tx.Exec("UPDATE users SET is_verified = true WHERE id = $1", userID)
	if err != nil {
		log.Printf("Error updating is_verified flag in users table: %v", err)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Profile created successfully"})
}

//Update Profile

func (h *Handler) UpdateProfile(c *gin.Context) {
	// Extract user ID from token
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var profile UserProfile
	if err := c.BindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	query := "UPDATE user_profile SET"
	var params []interface{}
	var index int = 1 // Index for parameters

	if profile.FullName != "" {
		query += " full_name = $" + strconv.Itoa(index) + ","
		params = append(params, profile.FullName)
		index++
	}
	if profile.Age != 0 {
		query += " age = $" + strconv.Itoa(index) + ","
		params = append(params, profile.Age)
		index++
	}
	if profile.Gender != "" {
		query += " gender = $" + strconv.Itoa(index) + ","
		params = append(params, profile.Gender)
		index++
	}

	// Remove the trailing comma
	query = strings.TrimSuffix(query, ",")

	// Add the WHERE clause
	query += " WHERE user_id = $" + strconv.Itoa(index)
	params = append(params, userID)

	// Execute the SQL query
	_, err := h.DB.Exec(query, params...)
	if err != nil {
		log.Printf("Error updating profile data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	// Extract user ID from token
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Start a transaction
	tx, err := h.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}

	// Update isDelete in users table
	_, err = tx.Exec("UPDATE users SET is_delete = true WHERE id = $1", userID)
	if err != nil {
		tx.Rollback()
		log.Printf("Error updating users table: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	// Update isDelete in user_profile table
	_, err = tx.Exec("UPDATE user_profile SET is_delete = true WHERE user_id = $1", userID)
	if err != nil {
		tx.Rollback()
		log.Printf("Error updating user_profile table: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user profile"})
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *Handler) CreateWithParams(c *gin.Context) {
	// Extract user ID from URL parameters
	userID := c.Param("id")

	var profile UserProfile
	if err := c.BindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}
	if profile.FullName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "full name is required"})
		return
	}

	// Start a transaction
	tx, err := h.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Error rolling back transaction: %v", rbErr)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create profile"})
			return
		}
	}()

	// Insert profile data into the user_profile table
	_, err = tx.Exec("INSERT INTO user_profile (user_id, full_name, age, gender) VALUES ($1, $2, $3, $4)",
		userID, profile.FullName, profile.Age, profile.Gender)
	if err != nil {
		// Check if the error is a foreign key constraint violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user does not exist"})
			return
		}

		log.Printf("Error inserting profile data into user_profile table: %v", err)
		return
	}

	// Update the is_verified flag in the users table
	_, err = tx.Exec("UPDATE users SET is_verified = true WHERE id = $1", userID)
	if err != nil {
		log.Printf("Error updating is_verified flag in users table: %v", err)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Profile created successfully"})
}

// update with params
func (h *Handler) UpdateWithParam(c *gin.Context) {
	userID := c.Param("id")

	var profile UserProfile
	if err := c.BindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	query := "UPDATE user_profile SET"
	var params []interface{}
	var index int = 1 // Index for parameters

	if profile.FullName != "" {
		query += " full_name = $" + strconv.Itoa(index) + ","
		params = append(params, profile.FullName)
		index++
	}
	if profile.Age != 0 {
		query += " age = $" + strconv.Itoa(index) + ","
		params = append(params, profile.Age)
		index++
	}
	if profile.Gender != "" {
		query += " gender = $" + strconv.Itoa(index) + ","
		params = append(params, profile.Gender)
		index++
	}

	// Remove the trailing comma
	query = strings.TrimSuffix(query, ",")

	// Add the WHERE clause
	query += " WHERE user_id = $" + strconv.Itoa(index)
	params = append(params, userID)

	// Execute the SQL query
	_, err := h.DB.Exec(query, params...)
	if err != nil {
		log.Printf("Error updating profile data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func (h *Handler) DeleteWithParam(c *gin.Context) {
	userID := c.Param("id")
	// Start a transaction
	tx, err := h.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}

	// Update isDelete in users table
	_, err = tx.Exec("UPDATE users SET is_delete = true WHERE id = $1", userID)
	if err != nil {
		tx.Rollback()
		log.Printf("Error updating users table: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	// Update isDelete in user_profile table
	_, err = tx.Exec("UPDATE user_profile SET is_delete = true WHERE user_id = $1", userID)
	if err != nil {
		tx.Rollback()
		log.Printf("Error updating user_profile table: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user profile"})
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
