package positionevent

import (
	"contact-monitoring-ingest-api/pkg/geo"
	"context"
	"log"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const maximumDistanceBetweenDevices = 5 // in meters

// EventWorkerConfig defines configuration values for an EventWorker
type EventWorkerConfig struct {
	DB                            *mongo.Database
	EventChan                     chan PositionEvent
	MinAggregateChan              chan MinuteAggregate
	WG                            *sync.WaitGroup
	MaximumDistanceBetweenDevices float64
	AccuracyThreshold             float64
	Name                          int
}

// EventWorker processes position.Events from an input channel
// and puts minute aggregate events on output channel
func EventWorker(c EventWorkerConfig) {
	defer c.WG.Done()

	eventCollection := c.DB.Collection("position-event")

	for event := range c.EventChan {
		query := bson.M{
			"_id": bson.M{
				"$ne": event.ID,
			},
			"floor": event.Floor,
			"lonlat": bson.M{
				"$geoWithin": bson.M{
					"$centerSphere": bson.A{
						event.LonLat,
						(float64(event.Accuracy) + c.MaximumDistanceBetweenDevices + c.AccuracyThreshold) / geo.EarthRadiusMeters,
					},
				},
			},
			"timeBucket": event.TimeBucket,
		}
		cursor, err := eventCollection.Find(context.Background(), query)

		if err != nil {
			log.Println(err)
			continue
		}

		var results []PositionEvent
		err = cursor.All(context.Background(), &results)
		if err != nil {
			log.Fatal(err)
		}

		for _, result := range results {
			distance := geo.Distance(event.LonLat, result.LonLat)
			if event.ID != result.ID && distance < float64(event.Accuracy+result.Accuracy+5) {
				events := [2]PartialPositionEvent{
					{
						ID:       event.ID,
						DeviceID: event.DeviceID,
						Accuracy: event.Accuracy,
						LonLat:   event.LonLat,
					},
					{
						ID:       result.ID,
						DeviceID: result.DeviceID,
						Accuracy: result.Accuracy,
						LonLat:   result.LonLat,
					},
				}
				if events[0].DeviceID > events[1].DeviceID {
					events[0], events[1] = events[1], events[0]
				}

				// normally this would be producing to kafka, but we're imitating with a channel
				minuteAggregate := MinuteAggregate{
					TimeBucket: event.TimeBucket,
					Events:     events,
					Distance:   distance,
					Floor:      event.Floor,
				}

				c.MinAggregateChan <- minuteAggregate
			}
		}

	}
}

func AggregateWorker(
	db *mongo.Database,
	minAggregatePartitionChannel chan MinuteAggregate,
	wg *sync.WaitGroup,
	workerNum int,
) {
	defer wg.Done()

	minuteAggregateCol := db.Collection("minute-aggregation")
	contactEventCol := db.Collection("contact-event")

	for minAggregate := range minAggregatePartitionChannel {
		devices := [2]string{minAggregate.Events[0].DeviceID, minAggregate.Events[1].DeviceID}
		res, err := minuteAggregateCol.InsertOne(context.Background(), minAggregate)

		if err != nil {
			if merr, ok := err.(mongo.WriteException); ok {
				// single duplicate key error which we are ok with
				if len(merr.WriteErrors) == 1 && merr.WriteErrors[0].Code == 11000 {
					continue
				}
			}

			log.Println("error inserting minute aggregate", err)
			continue
		}

		if res.InsertedID != nil && res.InsertedID != "" {
			aggID := res.InsertedID.(primitive.ObjectID)
			// if only the minute aggregate was "inserted" do we continue with the 5 min association
			// base contact event of 1 minute

			// high level algorithm:
			// 1. for a 1 min contact event at T
			// 2. check if contact event ending at T-1 exists, if so then merge
			// 3. check if contact event starting at T+1 exists, if so then merge
			// 4. either insert the contact event at T, or the newly merged event, plus also delete the obsolete events
			contact := ContactEvent{
				Devices:          devices,
				Start:            minAggregate.TimeBucket,
				End:              minAggregate.TimeBucket,
				MinuteAggregates: []primitive.ObjectID{aggID},
				Duration:         1,
				FirstContact: MinuteAggregate{
					Events: [2]PartialPositionEvent{
						{
							DeviceID: minAggregate.Events[0].DeviceID,
							LonLat:   minAggregate.Events[0].LonLat,
							Accuracy: minAggregate.Events[0].Accuracy,
						},
						{
							DeviceID: minAggregate.Events[1].DeviceID,
							LonLat:   minAggregate.Events[1].LonLat,
							Accuracy: minAggregate.Events[1].Accuracy,
						},
					},
					Floor: minAggregate.Floor,
				},
				MinDistance: minAggregate.Distance,
				MaxDistance: minAggregate.Distance,
			}
			// TODO this can at least be optimized to fetch both T-1 and T+1 in one go
			// Also can probably reuse one of the T-1 and T+1 for extension instead of having to delete
			var operations []mongo.WriteModel

			// find contact event of T-1 minute
			var contactBefore ContactEvent
			filter := bson.M{"devices": devices, "end": minAggregate.TimeBucket - 1}
			err := contactEventCol.FindOne(context.Background(), filter, options.FindOne()).Decode(&contactBefore)
			if err == nil {
				contact.Start = contactBefore.Start
				contact.MinuteAggregates = append(contactBefore.MinuteAggregates, contact.MinuteAggregates...)
				contact.Duration = contact.Duration + contactBefore.Duration
				contact.FirstContact = contactBefore.FirstContact
				if contactBefore.MinDistance < contact.MinDistance {
					contact.MinDistance = contactBefore.MinDistance
				}
				if contactBefore.MaxDistance > contact.MaxDistance {
					contact.MaxDistance = contactBefore.MaxDistance
				}

				// delete T-1
				operation := mongo.NewDeleteOneModel()
				operation.SetFilter(bson.M{"devices": contactBefore.Devices, "start": contactBefore.Start, "end": contactBefore.End})
				operations = append(operations, operation)
			}
			// find contact event of T+1 minute
			var contactAfter ContactEvent
			filter = bson.M{"devices": devices, "start": minAggregate.TimeBucket + 1}
			err = contactEventCol.FindOne(context.Background(), filter, options.FindOne()).Decode(&contactAfter)
			if err == nil {
				contact.End = contactAfter.End
				contact.MinuteAggregates = append(contact.MinuteAggregates, contactAfter.MinuteAggregates...)
				contact.Duration = contact.Duration + contactAfter.Duration
				if contactAfter.MinDistance < contact.MinDistance {
					contact.MinDistance = contactAfter.MinDistance
				}
				if contactAfter.MaxDistance > contact.MaxDistance {
					contact.MaxDistance = contactAfter.MaxDistance
				}

				// delete T+1
				operation := mongo.NewDeleteOneModel()
				operation.SetFilter(bson.M{"devices": contactAfter.Devices, "start": contactAfter.Start, "end": contactAfter.End})
				operations = append(operations, operation)

			}
			// insert the contact event at T
			operation := mongo.NewInsertOneModel()
			operation.SetDocument(contact)
			operations = append(operations, operation)

			_, err = contactEventCol.BulkWrite(context.Background(), operations, &options.BulkWriteOptions{})
			if err != nil {
				log.Println("error bulkwriting contact events", err)
			}
		}
	}
}
