package routes

import (
	"github.com/cyops-se/dd-inserter/engine"
	"github.com/gofiber/fiber/v2"
)

type SystemInformation struct {
	GitVersion string `json:"gitversion"`
	GitCommit  string `json:"gitcommit"`
}

var SysInfo SystemInformation

func RegisterSystemRoutes(api fiber.Router) {
	api.Get("/system/info", GetSysInfo)
	api.Post("/system/testmail", PostTestMail)
}

func GetSysInfo(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(SysInfo)
}

func PostTestMail(c *fiber.Ctx) error {
	var recipients string
	recipients = string(c.Body())

	// if err := c.BodyParser(&recipients); err != nil {
	// 	e := db.Error("Database error", "Failed to map string slice type while parsing body, error: %s", err.Error())
	// 	return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": e.Error()})
	// }

	engine.SendTestAlerts(recipients)

	return c.Status(fiber.StatusOK).JSON(recipients)
}
