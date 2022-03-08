package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/listeners"
	"github.com/cyops-se/dd-inserter/types"
	"github.com/gofiber/fiber/v2"
)

func RegisterListenerRoutes(api fiber.Router) {
	api.Get("/listener", GetAllListeners)
	api.Get("/listener/types", GetAllListenerTypes)
	api.Post("/listener/:type", NewListener)
	api.Put("/listener/:id", UpdateListener)
}

func GetAllListeners(c *fiber.Ctx) error {
	var items []types.Listener
	if err := db.DB.Find(&items).Error; err != nil {
		db.Log("error", "Failed to find listeners in database", err.Error())
		return c.Status(503).SendString(err.Error())
	}

	return c.Status(http.StatusOK).JSON(listeners.Listeners)
}

func GetAllListenerTypes(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(listeners.GetTypeNames())
}

type listenerIntermediateType struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Settings string `json:"settings"`
}

func NewListener(c *fiber.Ctx) error {
	typename := c.Params("type")
	item, err := listeners.CreateListener(typename)
	if err != nil {
		db.Log("error", fmt.Sprintf("Failed to create listener type: %s", typename), err.Error())
		return c.Status(503).SendString(err.Error())
	}

	if err := c.BodyParser(&item); err != nil {
		db.Log("error", fmt.Sprintf("Failed to map provided data to listener type: %s", typename), err.Error())
		return c.Status(503).SendString(err.Error())
	}

	if err := db.DB.Create(item).Error; err != nil {
		msg := db.Log("error", "Failed to create listener", fmt.Sprintf("Type: %s, data: %#v, error: %s", typename, item, err.Error()))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"msg": msg})
	}

	db.Log("trace", "Listener created", fmt.Sprintf("Type: %s, item: %#v", typename, item))
	listeners.LoadListener(item, nil)

	return c.Status(http.StatusOK).JSON(item)
}

func UpdateListener(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	var listener *types.Listener
	for i, _ := range listeners.Listeners {
		if listeners.Listeners[i].ID == uint(id) {
			listener = &listeners.Listeners[i]
			break
		}
	}

	if listener == nil {
		err := fmt.Errorf("Listener ID not found in listener list: %d", id)
		db.Log("error", "Unknown listener ID", err.Error())
		return c.Status(503).SendString(err.Error())
	}

	if err := c.BodyParser(listener); err != nil {
		db.Log("error", "Failed to map provided data to type", err.Error())
		return c.Status(503).SendString(err.Error())
	}

	db.DB.Save(listener)
	listeners.LoadListener(listener, nil)

	return c.Status(http.StatusOK).JSON(listener)
}
