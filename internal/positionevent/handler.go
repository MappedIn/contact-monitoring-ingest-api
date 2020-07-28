package positionevent

import (
	"contact-monitoring-ingest-api/internal/auth"
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func positionEventProcessor(
	event PositionEvent,
	col *mongo.Collection,
	eventChan chan PositionEvent,
) httpResponse {
	if event.UserConsent != true {
		return httpResponse{
			Message: "userConsent was not set to true so the event was ignored",
			Status:  http.StatusBadRequest,
		}
	}

	res, err := col.InsertOne(context.Background(), event)

	if err != nil {
		if merr, ok := err.(mongo.WriteException); ok {
			// single duplicate key error which we are ok with
			if len(merr.WriteErrors) == 1 && merr.WriteErrors[0].Code == 11000 {
				return httpResponse{
					Message: "There is already a position for this device at this time",
					Status:  http.StatusConflict,
				}
			}
		}

		log.Println("error inserting position event", err)
		return httpResponse{
			Message: "An unexpected server error has occured",
			Status:  http.StatusInternalServerError,
		}
	}

	event.ID = res.InsertedID.(primitive.ObjectID)
	eventChan <- event

	return httpResponse{
		Status: http.StatusOK,
	}
}

type httpResponse struct {
	Message string `json:"message" binding:"omitempty"`
	Status  int    `json:"status"`
}

const timeBucketSize int64 = 60 * 1000 // 60 seconds in milliseconds

// PostHandler accepts a body of an array of position.Events
// it determines the best fit of those events to process by selecting
// the events nearest to each time bucket
func PostHandler(
	col *mongo.Collection,
	eventChan chan PositionEvent,
	accuracyThreshold float64,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := c.Get("claims")
		if !ok {
			c.Status(http.StatusUnauthorized)
			return
		}

		venueClaims, ok := claims.(*auth.Claims)
		if !ok {
			c.Status(http.StatusUnauthorized)
			return
		}

		var events []PositionEvent
		err := c.BindJSON(&events)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// filtering to only process subset of events
		// given that a batch of events may be sent that
		// include multiple positions per minute

		// // make sure events are sorted by timestamp
		sort.Slice(events, func(i, j int) bool {
			return events[i].Time < events[j].Time
		})

		// only process events that have good enough accuracy
		// and are closest to the time bucket
		response := make([]httpResponse, len(events))
		var currentBucket uint32 = 0
		for i, event := range events {
			event.TimeBucket = uint32(math.Round(float64(event.Time / timeBucketSize)))
			if event.Venue != venueClaims.Venue {
				response[i] = httpResponse{
					Message: fmt.Sprintf("Unauthorized venue %v specified; token only has access to %v", event.Venue, venueClaims.Venue),
					Status:  http.StatusUnauthorized,
				}
			} else if float64(event.Accuracy) > accuracyThreshold {
				// filter out events that don't have good enough accuracy
				response[i] = httpResponse{
					Message: fmt.Sprintf("Accuracy of %f exceeds threshold of %f", event.Accuracy, accuracyThreshold),
					Status:  http.StatusBadRequest,
				}
			} else if event.TimeBucket != currentBucket {
				// if we are looking at a newer time bucket then use this event
				currentBucket = event.TimeBucket
				response[i] = positionEventProcessor(event, col, eventChan)
			} else {
				// we already have an event for this time bucket so return conflict
				response[i] = httpResponse{
					Message: "There is already a position for this device at this time",
					Status:  http.StatusConflict,
				}
			}
		}

		c.JSON(http.StatusMultiStatus, response)
		c.Done()
	}
}
