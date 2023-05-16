package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	_"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Image    string `json:"image_filename"`
}

type Json struct {
	Items []Item `json:"item"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func getItem(c echo.Context) error {
	db, err := sql.Open("sqlite3", "../db/mercari.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM items")
	res := rowsToResponse(rows)
	return c.JSON(http.StatusOK, res)
}

func getItemWithId(c echo.Context) error {
	db, err := sql.Open("sqlite3", "../db/mercari.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	idString := c.Param("id")
	id, _ := strconv.Atoi(idString)
	rows, _ := db.Query("SELECT * FROM items WHERE id=$1", id)
	res := rowsToResponse(rows)
	return c.JSON(http.StatusOK, res)
}

func getItemWithName(c echo.Context) error {
	db, err := sql.Open("sqlite3", "../db/mercari.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	matchedName := c.QueryParam("keyword")
	rows, _ := db.Query("SELECT * FROM items WHERE name=$1", matchedName)
	res := rowsToResponse(rows)
	return c.JSON(http.StatusOK, res)
}

func rowToString(rows *sql.Rows) Item {
	var id_int int
	var item Item
	if err := rows.Scan(&id_int, &item.Name, &item.Category, &item.Image); err != nil {
		log.Fatal(err)
	}
	return item
}

func rowsToResponse(rows *sql.Rows) []Item {
	var res []Item
	for rows.Next() {
		item := rowToString(rows)
		res = append(res, item)
	}
	return res
}

func addItem(c echo.Context) error {
	// Get form data
	db, err := sql.Open("sqlite3", "../db/mercari.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	name := c.FormValue("name")
	category := c.FormValue("category")
	image := imageToHash(c.FormValue("image"))
	item := Item{Name: name, Category: category, Image: image}
	_, err = db.Exec("INSERT INTO items (name, category, image) VALUES ($1, $2, $3)", item.Name, item.Category, item.Image)
	if err != nil {
		log.Fatal(err)
	}

	c.Logger().Infof("Receive item: %s", name)

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func imageToHash(imagePass string) string {
	imageFile, _ := os.ReadFile(imagePass)
	imageHash32bytes := sha256.Sum256(imageFile)
	image := hex.EncodeToString(imageHash32bytes[:]) + ".jpg"
	return image
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	front_url := os.Getenv("FRONT_URL")
	if front_url == "" {
		front_url = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{front_url},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.GET("/items", getItem)
	e.POST("/items", addItem)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/items/:id", getItemWithId)
	e.GET("/search", getItemWithName)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
