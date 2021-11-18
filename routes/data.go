package routes

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/types"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm/clause"
)

func RegisterDataRoutes(api fiber.Router) {
	api.Get("/data/:type", GetAllOfType)
	api.Get("/data/:type/id/:id", GetDataByID)
	api.Get("/data/:type/field/:field/:value", GetDataByField)
	api.Post("/data/:type", NewData)
	api.Put("/data/:type", UpdateData)
	api.Delete("/data/:type/", DeleteDataAll)
	api.Delete("/data/:type/:id", DeleteDataByID)
	api.Delete("/data/:type/field/:field/:value", DeleteDataByField)
}

func GetAllOfType(c *fiber.Ctx) error {
	table := c.Params("type")
	items := types.CreateSlice(table)
	if items == nil {
		e := db.Error("Database error", "Failed to find type %s in registry", table)
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": e.Error()})
	}

	db.DB.Table(table).Preload(clause.Associations).Find(items)

	return c.Status(http.StatusOK).JSON(items)
}

func GetDataByID(c *fiber.Ctx) (err error) {
	id := c.Params("id")
	table := c.Params("type")
	item := types.CreateType(table)
	if item == nil {
		e := db.Error("Database error", "Failed to find type %s in registry", table)
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": e.Error()})
	}

	if err = db.DB.Take(item, id).Error; err != nil {
		e := db.Error("Database error", "Failed to find item of type '%s', id: %s, error: %s", table, id, err.Error())
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": e.Error()})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"item": item})
}

func GetDataByField(c *fiber.Ctx) error {
	field := c.Params("field")
	value := c.Params("value")
	table := c.Params("type")
	items := types.CreateSlice(table)
	if items == nil {
		e := db.Error("Database error", "Failed to find type %s in registry", table)
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": e.Error()})
	}

	conditions := map[string]interface{}{field: value}
	if result := db.DB.Find(items, conditions); result.Error != nil {
		e := db.Error("Database error", "Failed to find item of type '%s', field: %s, value: %s, database error: %s", table, field, value, result.Error)
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": e.Error()})
	}

	return c.Status(http.StatusOK).JSON(items)
}

func NewData(c *fiber.Ctx) error {
	table := c.Params("type")
	item := types.CreateType(table)
	if item == nil {
		e := db.Error("Database error", "Failed to find type %s in registry", table)
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": e.Error()})
	}

	if err := c.BodyParser(&item); err != nil {
		e := db.Error("Database error", "Failed to map provided data to type %s while updating item", table)
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": e.Error()})
	}

	if err := db.DB.Create(item).Error; err != nil {
		e := db.Error("Database error", "Failed to create item of type '%s', data: %#v, error: %s", table, item, err.Error())
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": e.Error()})
	}

	db.Log("trace", "Item created", fmt.Sprintf("Type: %s, item: %#v", table, item))

	return c.Status(http.StatusOK).JSON(item)
}

func UpdateData(c *fiber.Ctx) error {
	table := c.Params("type")
	item := types.CreateType(table)
	if item == nil {
		e := db.Error("Database error", "Failed to find type %s in registry", table)
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": e.Error()})
	}

	if err := c.BodyParser(&item); err != nil {
		e := db.Error("Database error", "Failed to map provided data to type %s while updating item, error: %s", table, err.Error())
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": e.Error()})
	}

	db.DB.Save(item)

	c.Status(200)
	return c.JSON(item)
}

func DeleteDataAll(c *fiber.Ctx) error {
	table := c.Params("type")
	item := types.CreateType(table)
	if item == nil {
		e := db.Error("Database error", "Failed to find type %s in registry", table)
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": e.Error()})
	}

	if err := db.DB.Unscoped().Delete(item, "1=1").Error; err != nil {
		e := db.Error("Database error", "Failed to delete all items, error", table)
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": e.Error()})
	}

	return c.Status(http.StatusOK).JSON(item)
}

func DeleteDataByID(c *fiber.Ctx) error {
	id := c.Params("id")
	table := c.Params("type")
	item := types.CreateType(table)
	if item == nil {
		e := db.Error("Database error", "Failed to find type %s in registry", table)
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": e.Error()})
	}

	if err := db.DB.Unscoped().Delete(item, id).Error; err != nil {
		e := db.Error("Database error", "Failed to delete item id %d, type %s, error: %s", id, table, err.Error())
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": e.Error()})
	}

	return c.Status(http.StatusOK).JSON(item)
}

func DeleteDataByField(c *fiber.Ctx) error {
	field := c.Params("field")
	value, _ := url.QueryUnescape(c.Params("value"))
	table := c.Params("type")
	item := types.CreateSlice(table)
	if item == nil {
		e := db.Error("Database error", "Failed to find type %s in registry", table)
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": e.Error()})
	}

	conditions := map[string]interface{}{field: value}
	if result := db.DB.Delete(item, conditions); result.Error != nil {
		e := db.Error("Database error", "Failed to delete items of type %s where %s = %s, error: %s", table, field, value, result.Error.Error())
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": e.Error()})
	}

	return c.Status(http.StatusOK).JSON(item)
}
