package main

import (
    "database/sql"
    "log"
    "net/http"
    "papibiyi/directories/Models"
    "strconv"
    "time"

    _ "github.com/mattn/go-sqlite3"
    "github.com/gin-gonic/gin"
)


func main() {
	app := &models.App{}
    app.InitializeDB()

	router := gin.Default()

	router.GET("/directories", func(c *gin.Context) {
		getDirectories(c, app.Db)
	})

	router.GET("/directories/:id", func (c *gin.Context) {
		getDirectoryByID(c, app.Db)
	})
	
    router.POST("/directories", func(c *gin.Context) {
        postDirectory(c, app.Db)
    })

	router.PUT("/directories/:id", func(c *gin.Context) {
		updateDirectory(c, app.Db)
	})

	router.DELETE("/directories/:id", func(ctx *gin.Context) {
		deleteDirectory(ctx, app.Db)
	})

	router.Run("localhost:8080")
}

func getDirectories(c *gin.Context, db *sql.DB) {
    rows, err := db.Query(`
        SELECT d.id, d.name, d.phone_number, d.created_at, d.updated_at,
               a.address_line_1, a.address_line_2, a.city, a.state, a.country
        FROM directory d
        LEFT JOIN address a ON a.directory_id = d.id
        ORDER BY d.id`)
    if err != nil {
        log.Printf("error querying directories: %v", err)
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to query directories"})
        return
    }
	
    var result []models.Directory
    for rows.Next() {
        var (
            id int64
            name, phone, createdAt, updatedAt sql.NullString
            addr1, addr2, city, state, country sql.NullString
        )
        if err := rows.Scan(&id, &name, &phone, &createdAt, &updatedAt, &addr1, &addr2, &city, &state, &country); err != nil {
            log.Printf("error scanning row: %v", err)
            c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to read directory"})
            return
        }
        d := models.Directory{
            ID:          strconv.FormatInt(id, 10),
            Name:        name.String,
            PhoneNumber: phone.String,
            Address: models.Address{
                AddressLine1: addr1.String,
                AddressLine2: addr2.String,
                City:         city.String,
                State:        state.String,
                Country:      country.String,
            },
            CreatedAt:   createdAt.String,
            UpdatedAt:   updatedAt.String,
        }
        result = append(result, d)
    }

    if err := rows.Err(); err != nil {
        log.Printf("row iteration error: %v", err)
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to read directories"})
        return
    }

    c.IndentedJSON(http.StatusOK, result)
}


func getDirectoryByID(c *gin.Context, db *sql.DB) {
	id := c.Param("id")

    rows, err := db.Query(`
        SELECT d.id, d.name, d.phone_number, d.created_at, d.updated_at,
               a.address_line_1, a.address_line_2, a.city, a.state, a.country
        FROM directory d
        LEFT JOIN address a ON a.directory_id = d.id
        WHERE d.id = ?
	`, id)
    if err != nil {
        log.Printf("error querying directories: %v", err)
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to query directory"})
        return
    }

	if !rows.Next() {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "directory not found"})
		rows.Close()
		return
	}

    var directory models.Directory
    for rows.Next() {
        var (
            id int64
            name, phone, createdAt, updatedAt sql.NullString
            addr1, addr2, city, state, country sql.NullString
        )
        if err := rows.Scan(&id, &name, &phone, &createdAt, &updatedAt, &addr1, &addr2, &city, &state, &country); err != nil {
            log.Printf("error scanning row: %v", err)
            c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to read directory"})
            return
        }
        d := models.Directory{
            ID:          strconv.FormatInt(id, 10),
            Name:        name.String,
            PhoneNumber: phone.String,
            Address: models.Address{
                AddressLine1: addr1.String,
                AddressLine2: addr2.String,
                City:         city.String,
                State:        state.String,
                Country:      country.String,
            },
            CreatedAt:   createdAt.String,
            UpdatedAt:   updatedAt.String,
        }
        directory = d
        break
    }

    if err := rows.Err(); err != nil {
        log.Printf("row iteration error: %v", err)
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to read directory"})
        return
    }

    c.IndentedJSON(http.StatusOK, directory)
}

func postDirectory(c *gin.Context, db *sql.DB) {
    var newDirectory models.Directory

    if err := c.BindJSON(&newDirectory); err != nil {
        c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
        return
    }

    tx, err := db.Begin()
    if err != nil {
        log.Printf("error starting tx: %v", err)
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
        return
    }

    now := time.Now().UTC().Format(time.RFC3339)

    res, err := tx.Exec(
        `INSERT INTO directory (name, phone_number, created_at, updated_at) VALUES (?, ?, ?, ?)`,
        newDirectory.Name, newDirectory.PhoneNumber, now, now,
    )
    if err != nil {
        _ = tx.Rollback()
        log.Printf("error inserting directory: %v", err)
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to create directory"})
        return
    }
    dirID, err := res.LastInsertId()
    if err != nil {
        _ = tx.Rollback()
        log.Printf("error fetching inserted id: %v", err)
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to create directory"})
        return
    }

    // Insert address only if any field is provided
    addr := newDirectory.Address
    if addr.AddressLine1 != "" || addr.AddressLine2 != "" || addr.City != "" || addr.State != "" || addr.Country != "" {
        if _, err := tx.Exec(
            `INSERT INTO address (directory_id, address_line_1, address_line_2, city, state, country) VALUES (?, ?, ?, ?, ?, ?)`,
            dirID, addr.AddressLine1, addr.AddressLine2, addr.City, addr.State, addr.Country,
        ); err != nil {
            _ = tx.Rollback()
            log.Printf("error inserting address: %v", err)
            c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to create address"})
            return
        }
    }

    if err := tx.Commit(); err != nil {
        log.Printf("error committing tx: %v", err)
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to save directory"})
        return
    }

    created := models.Directory{
        ID:          strconv.FormatInt(dirID, 10),
        Name:        newDirectory.Name,
        PhoneNumber: newDirectory.PhoneNumber,
        Address:     newDirectory.Address,
        CreatedAt:   now,
        UpdatedAt:   now,
    }

    c.IndentedJSON(http.StatusCreated, created)
}

func updateDirectory(c *gin.Context, db *sql.DB) {
	id := c.Param("id")

	var updatedDirectory models.Directory

	if err := c.BindJSON(&updatedDirectory); err != nil {
		return
	}

	tx, err := db.Begin()
	if err != nil {
        log.Printf("error starting tx: %v", err)
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
        return
    }

    now := time.Now().UTC().Format(time.RFC3339)

    _, err = tx.Exec(
        `UPDATE directory SET name = ?, phone_number = ?, updated_at = ? WHERE id = ?`,
        updatedDirectory.Name, updatedDirectory.PhoneNumber, now, id,
    )
    if err != nil {
        _ = tx.Rollback()
        log.Printf("error updating directory: %v", err)
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to update directory"})
        return
    }

    if _, err := tx.Exec(
        `UPDATE address SET address_line_1 = ?, address_line_2 = ?, city = ?, state = ?, country = ? WHERE directory_id = ?`,
        updatedDirectory.Address.AddressLine1, updatedDirectory.Address.AddressLine2, updatedDirectory.Address.City, updatedDirectory.Address.State, updatedDirectory.Address.Country, id,
    ); err != nil {
        _ = tx.Rollback()
        log.Printf("error updating address: %v", err)
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to update address"})
        return
    }

    if err := tx.Commit(); err != nil {
        log.Printf("error committing tx: %v", err)
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to save directory"})
        return
    }

    updatedDirectory.ID = id
    updatedDirectory.UpdatedAt = now

    c.IndentedJSON(http.StatusOK, updatedDirectory)
}

func deleteDirectory(c *gin.Context, db *sql.DB) {
	id := c.Param("id")

    tx, err := db.Begin()

    if err != nil {
        log.Printf("error starting tx: %v", err)
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
        return
    }

    _, err = tx.Exec(`DELETE FROM directory WHERE id = ?`, id)

    if err != nil {
        _ = tx.Rollback()
        log.Printf("error deleting directory: %v", err)
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to delete directory"})
        return
    }

    if err := tx.Commit(); err != nil {
        log.Printf("error committing tx: %v", err)
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to delete directory"})
        return
    }

    c.IndentedJSON(http.StatusOK, gin.H{"message": "directory deleted"})
    
}
