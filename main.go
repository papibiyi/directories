package main

import (
	"net/http"
	"papibiyi/directories/Models"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var directories = []models.Directory {}

func main() {
	app := &models.App{}
    app.InitializeDB()

	router := gin.Default()
	router.GET("/directories", getDirectories)
	router.GET("/directories/:id", getDirectoryByID)
	router.POST("/directories", postDirectory)
	router.PUT("/directories/:id", updateDirectory)
	router.DELETE("/directories/:id", deleteDirectory)

	router.Run("localhost:8080")
}

func getDirectories(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, directories)
}


func getDirectoryByID(c *gin.Context) {
	id := c.Param("id")

	for _, a := range directories {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "directory not found"})
}

func postDirectory(c *gin.Context) {
	var newDirectory models.Directory

	if err := c.BindJSON(&newDirectory); err != nil {
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)

	newDirectory.ID = uuid.NewString()
	newDirectory.CreatedAt, newDirectory.UpdatedAt = now, now
	directories = append(directories, newDirectory)	
	c.IndentedJSON(http.StatusCreated, newDirectory)
}

func updateDirectory(c *gin.Context) {
	id := c.Param("id")

	var updatedDirectory models.Directory

	if err := c.BindJSON(&updatedDirectory); err != nil {
		return
	}

	for i, a := range directories {
		if a.ID == id {
			updatedDirectory.ID = a.ID
			updatedDirectory.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

			directories[i] = updatedDirectory
			c.IndentedJSON(http.StatusOK, updatedDirectory)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "directory not found"})
}

func deleteDirectory(c *gin.Context) {
	id := c.Param("id")

	var index int = -1
	for i, a := range directories {
		if a.ID == id {
			index = i
			break
		}
	}
	if index != -1 {
		directories = slices.Delete(directories, index, index+1)
		c.IndentedJSON(http.StatusOK, gin.H{"message": "directory deleted"})
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "directory not found"})
	}
}