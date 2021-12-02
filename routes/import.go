package routes

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/emitters"
	"github.com/cyops-se/dd-inserter/types"
	"github.com/gofiber/fiber/v2"
)

func RegisterImportRoutes(api fiber.Router) {
	api.Get("/meta/all", GetMetaAll)
	api.Post("/meta/import", PostMetaImport)
	api.Post("/meta/changes", PostMetaChanges)
}

func GetMetaAll(c *fiber.Ctx) error {
	var items []*types.Meta
	// if rows, err := emitters.TimescaleDB.Query("select tag_id,name,description,location,type,unit,min,max from measurements.tags"); err == nil {
	if rows, err := emitters.TimescaleDBConn.Query(context.Background(), "select tag_id,name,description,location,type,unit,min,max from measurements.tags"); err == nil {
		for rows.Next() {
			item := &types.Meta{}
			rows.Scan(&item.TagId, &item.Name, &item.Description, &item.Location, &item.Type, &item.Unit, &item.Min, &item.Max)
			items = append(items, item)
			log.Printf("META init of %s", item.Name)
		}

		rows.Close()
	}
	return c.Status(http.StatusOK).JSON(items)
}

func PostMetaImport(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		db.Log("error", "No file provided", err.Error())
		return c.Status(503).SendString(err.Error())
	}

	// Make sure ./upload exists
	if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
		os.Mkdir("./uploads", 0755)
	}

	filename := fmt.Sprintf("./uploads/%s", file.Filename)
	if err := c.SaveFile(file, filename); err != nil {
		msg := fmt.Sprintf("failed to save file, name: '%s', size: %d, error: %s", file.Filename, file.Size, err.Error())
		db.Log("error", "Import meta request filed", msg)
		return c.Status(503).SendString(msg)
	} else {
		db.Log("trace", "Import meta request", fmt.Sprintf("name: '%s', size: %d", file.Filename, file.Size))
	}

	records, err := processCSVFile(filename)
	if err != nil {
		msg := fmt.Sprintf("CSV processing error, name: '%s', size: %d, error: %s", file.Filename, file.Size, err.Error())
		db.Log("trace", "Import meta request", msg)
		return c.Status(503).SendString(msg)
	}

	return c.Status(http.StatusOK).JSON(records)
}

func PostMetaChanges(c *fiber.Ctx) error {
	var items []types.Meta
	if err := c.BodyParser(&items); err != nil {
		db.Log("error", "Failed to map provided data to types.Meta array", err.Error())
		return c.Status(503).SendString(err.Error())
	}

	for _, item := range items {
		// if _, err := emitters.TimescaleDB.Exec("update measurements.tags set description=$2,location=$3,type=$4,unit=$5,min=$6,max=$7 where tag_id = $1",
		if _, err := emitters.TimescaleDBConn.Exec(context.Background(), "update measurements.tags set description=$2,location=$3,type=$4,unit=$5,min=$6,max=$7 where tag_id = $1",
			item.TagId, item.Description, item.Location, item.Type, item.Unit, item.Min, item.Max); err != nil {
			db.Log("error", "Failed to update meta", err.Error())
		}
	}

	return c.Status(http.StatusOK).JSON(items)
}

func processCSVFile(filename string) ([][]string, error) {
	// Read all content
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(bytes.NewReader(content))
	r.Comma = ';'
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}
