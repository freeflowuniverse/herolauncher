package main

import (
	"encoding/json"
	"log"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// Register routes from OpenAPI spec

	// List all pets
	app.Get("/pets", func(c *fiber.Ctx) error {
		// Mock implementation for listPets

		// Return example response

		return c.Status(200).Type("application/json").Send([]byte(`[
  {"id": 1, "name": "Fluffy", "type": "cat"},
  {"id": 2, "name": "Buddy", "type": "dog"}
]`))


	})

	// Create a pet
	app.Post("/pets", func(c *fiber.Ctx) error {
		// Mock implementation for createPet

		// Return example response

		return c.Status(201).Type("application/json").Send([]byte(`{"id": 3, "name": "Rex", "type": "dog"}`))


	})

	// Get a pet by ID
	app.Get("/pets/:id", func(c *fiber.Ctx) error {
		// Mock implementation for getPet

		// Return example response

		return c.Status(200).Type("application/json").Send([]byte(`{"id": 1, "name": "Fluffy", "type": "cat"}`))


	})


	log.Println("Server started on :8080")
	log.Fatal(app.Listen(":8080"))
}
