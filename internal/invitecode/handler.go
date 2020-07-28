package invitecode

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type createBody struct {
	MaxUses int `form:"maxUses"`
}

// CreateHandler returns a gin HandlerFunc which creates
// a new invite code for venue provided by venue param in route
func CreateHandler(codeRepo Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		venue := c.Param("venue")
		var body createBody
		err := c.BindJSON(&body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		code, err := codeRepo.Create(venue, body.MaxUses)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create invite code"})
			return
		}

		c.JSON(http.StatusCreated, code)
	}
}

// DeleteHandler returns a gin HandlerFunc which deletes
// invite codes based on provided code param in route
func DeleteHandler(codeRepo Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")
		_, err := codeRepo.Delete(code)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusOK)
	}
}
