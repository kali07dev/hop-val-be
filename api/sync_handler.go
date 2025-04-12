package api

import (
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/hopekali04/valuations/services"
	"github.com/hopekali04/valuations/utils"
)

type SyncHandler struct {
	Service *services.SyncService
}

func NewSyncHandler(service *services.SyncService) *SyncHandler {
	return &SyncHandler{Service: service}
}

// TriggerSync handles POST /sync
func (h *SyncHandler) TriggerSync(c *fiber.Ctx) error {
	log.Println("Received request to trigger property sync...")

	result, err := h.Service.FetchAndSyncProperties()
	if err != nil {
		log.Printf("Sync process failed: %v\n", err)
		// Return the sync result structure even on failure, it contains error details
		if result != nil {
			return c.Status(http.StatusInternalServerError).JSON(result)
		}
		// If result is nil (e.g., config error), return a generic error
		return utils.HandleError(c, err)
	}

	log.Printf("Sync process completed successfully.")
	return c.Status(http.StatusOK).JSON(result)
}
