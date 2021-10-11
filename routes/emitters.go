package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/emitters"
	"github.com/cyops-se/dd-inserter/types"
	"github.com/gofiber/fiber/v2"
)

func RegisterEmitterRoutes(api fiber.Router) {
	api.Get("/emitter", GetAll)
	api.Get("/emitter/types", GetAllTypes)
	api.Post("/emitter/:type", NewEmitter)
	api.Put("/emitter/:id", UpdateEmitter)
}

func GetAll(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(emitters.Emitters)
}

func GetAllTypes(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(emitters.GetTypeNames())
}

type emitterIntermediateType struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Settings string `json:"settings"`
}

func NewEmitter(c *fiber.Ctx) error {
	typename := c.Params("type")
	item, err := emitters.CreateEmitter(typename)
	if err != nil {
		db.Log("error", fmt.Sprintf("Failed to create emitter type", typename), err.Error())
		return c.Status(503).SendString(err.Error())
	}

	if err := c.BodyParser(&item); err != nil {
		db.Log("error", fmt.Sprintf("Failed to map provided data to emitter type", typename), err.Error())
		return c.Status(503).SendString(err.Error())
	}

	if err := db.DB.Create(item).Error; err != nil {
		msg := db.Log("error", "Failed to create emitter", fmt.Sprintf("Type: %s, data: %#v, error: %s", typename, item, err.Error()))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"msg": msg})
	}

	db.Log("trace", "Emitter created", fmt.Sprintf("Type: %s, item: %#v", typename, item))
	emitters.LoadEmitter(item)

	return c.Status(http.StatusOK).JSON(item)
}

func UpdateEmitter(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	var emitter *types.Emitter
	for i, _ := range emitters.Emitters {
		if emitters.Emitters[i].ID == uint(id) {
			emitter = &emitters.Emitters[i]
			break
		}
	}

	if emitter == nil {
		err := fmt.Errorf("Emitter ID not found in emitter list: %d", id)
		db.Log("error", "Unknown emitter ID", err.Error())
		return c.Status(503).SendString(err.Error())
	}

	if err := c.BodyParser(emitter); err != nil {
		db.Log("error", "Failed to map provided data to type", err.Error())
		return c.Status(503).SendString(err.Error())
	}

	db.DB.Save(emitter)
	db.Log("trace", "Item updated", fmt.Sprintf("ID: %d, item: %#v, instance: %#v", id, emitter, emitter.Instance))
	emitters.LoadEmitter(emitter)

	c.Status(200)
	return c.JSON(emitter)
}
