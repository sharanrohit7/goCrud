package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/goCrud/config"
	"github.com/goCrud/handlers"
	"github.com/goCrud/middlewares"
)

//Every main working part in this entire project has a comment to help identify what its doing

func main() {
	// Initialize the database connection
	database, err := db.Init()
	if err != nil {
		panic(err)
	}
	handler := handlers.NewHandler(database)

	r := gin.Default()

	// Middleware to check API key
	r.Use(handlers.CheckAPIKey())

	// Ping route
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.POST("/users", handler.CreateUser)
	r.GET("/users", handler.GetUsers)
	r.POST("/login", handler.SignInHandler)

	//These routes has middleware in-between thus dont require dynamic routing where user id is exposed, it uses JWT tokens to perform CRUD
	r.POST("/profile", middlewares.AuthMiddleware(), handler.CreateProfile)
	r.POST("/updateProfile", middlewares.AuthMiddleware(), handler.UpdateProfile)
	r.PATCH("/deleteUser", middlewares.AuthMiddleware(), handler.DeleteUser)

	//In the mail , I was asked to make use of params to perform this so here is this
	r.POST("/profile/:id", handler.CreateWithParams)
	r.PUT("/profile/:id", handler.UpdateWithParam)
	r.DELETE("/profile/:id", handler.DeleteWithParam)
	//Get User Details By id
	r.GET("/profile/:id", handler.GetUserById)

	r.Run()
}
