package routes

import (
	"log"
	"net/http"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/engine"
	"github.com/cyops-se/dd-inserter/types"
	"github.com/gofiber/fiber/v2"
)

func RegisterProxyRoutes(api fiber.Router) {

	api.Get("/proxy/point", GetDataPoints)
	api.Put("/proxy/point", UpdateDataPoint)
	api.Post("/proxy/points", UpdateDataPoints)
	api.Delete("/proxy/point/:id", DeleteDataPoint)
}

func handlePanic(c *fiber.Ctx) {
	if r := recover(); r != nil {
		log.Println(r)
		c.Status(500)
		c.JSON(r)
		return
	}
}

func handleError(c *fiber.Ctx, err error) error {
	db.Log("error", "Database error", err.Error())
	return c.Status(400).JSON(&fiber.Map{"error": err.Error()})
}

// DATA POINTS

func GetDataPoints(c *fiber.Ctx) (err error) {
	var items []types.DataPointMeta
	if err = db.DB.Find(&items).Error; err != nil {
		return handleError(c, err)
	}

	return c.Status(http.StatusOK).JSON(items)
}

func UpdateDataPoint(c *fiber.Ctx) (err error) {
	var item types.DataPointMeta
	if err = c.BodyParser(&item); err == nil {
		if err = db.DB.Save(&item).Error; err != nil {
			return handleError(c, err)
		}
	} else {
		return handleError(c, err)
	}

	engine.InitDataPointMap()
	log.Printf("item saved: %#v", item)
	return c.Status(200).JSON(&fiber.Map{"item": item})
}

func UpdateDataPoints(c *fiber.Ctx) (err error) {
	var items []types.DataPointMeta
	failedcount := 0
	if err = c.BodyParser(&items); err == nil {
		for _, item := range items {
			if item.ID > 0 {
				if err = db.DB.Save(&item).Error; err != nil {
					log.Printf("Failed to save item: %#v, error: %s", item, err.Error())
					failedcount++
				}
			} else {
				if err = db.DB.Create(&item).Error; err != nil {
					log.Printf("Failed to create item: %#v, error: %s", item, err.Error())
					failedcount++
				}
			}
		}
	} else {
		log.Printf("Failed to parse body, error: %s", err.Error())
	}

	engine.InitDataPointMap()
	if failedcount > 0 || err != nil {
		return c.Status(500).JSON(&fiber.Map{"items": items, "err": err, "failedcount": failedcount})
	}

	return c.Status(200).JSON(&fiber.Map{"items": items})
}

func DeleteDataPoint(c *fiber.Ctx) (err error) {
	id := c.Params("id")
	var item types.DataPointMeta
	log.Println("Deleting datapoint meta, id:", id)
	if err = db.DB.Unscoped().Delete(&item, "id = ?", id).Error; err != nil {
		return handleError(c, err)
	}

	engine.InitDataPointMap()
	return c.Status(200).JSON(&fiber.Map{"item": item})
}
