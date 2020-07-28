package device

import (
	"contact-monitoring-ingest-api/internal/invitecode"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type activateBody struct {
	Code       string `json:"code"`
	DeviceType string `json:"deviceType"`
}

func ActivateHandler(codeRepo invitecode.Repo, deviceRepo Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var body activateBody
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		invite, err := codeRepo.UseOne(body.Code)
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "code is invalid"})
			return
		}

		device, err := deviceRepo.Activate(id, body.DeviceType, invite.Venue)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"venue": device.Venue})
	}
}

func GetHandler(deviceRepo Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		device, err := deviceRepo.Get(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, device)
	}
}

func CreateHandler(deviceRepo Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Status(http.StatusNotImplemented)
	}
}
