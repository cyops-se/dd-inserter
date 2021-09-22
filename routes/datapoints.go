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

	api.Get("/proxy/group", GetGroups)
	api.Get("/proxy/point", GetDataPoints)
	api.Put("/proxy/point", UpdateDataPoint)
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

// GROUPS

func GetGroups(c *fiber.Ctx) (err error) {
	items, _ := engine.GetGroups()
	return c.Status(http.StatusOK).JSON(items)
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
		log.Println("Updating datapoint meta", item)
		if err = db.DB.Updates(&item).Error; err != nil {
			return handleError(c, err)
		}
	} else {
		return handleError(c, err)
	}

	engine.InitDataPointMap()
	return c.Status(200).JSON(&fiber.Map{"item": item})
}
