package positionevent

import (
	"contact-monitoring-ingest-api/pkg/geo"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PositionEvent represents a spatial and temporal position of a device
type PositionEvent struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	DeviceID    string             `bson:"device" json:"device" binding:"required"`
	Time        int64              `bson:"time" json:"time" binding:"required"`
	LonLat      geo.Coord          `bson:"lonlat" json:"lonlat" binding:"required"`
	Accuracy    float32            `bson:"acc" json:"acc" binding:"required"`
	Floor       int16              `bson:"floor" json:"floor" binding:"required"`
	UserConsent bool               `bson:"userConsent" json:"userConsent" binding:"required"`
	Venue       string             `bson:"venue" json:"venue" binding:"required"`
	TimeBucket  uint32             `bson:"timeBucket"`
}

// PartialPositionEvent represents a small view of a position event used in
// a minute aggregation
type PartialPositionEvent struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	DeviceID string             `bson:"device"`
	LonLat   geo.Coord          `bson:"lonlat"`
	Accuracy float32            `bson:"accuracy"`
}

// MinuteAggregate represents a contact between two people at a time derived from two position events
type MinuteAggregate struct {
	ID         primitive.ObjectID      `bson:"_id,omitempty"`
	TimeBucket uint32                  `bson:"timeBucket,omitempty"`
	Events     [2]PartialPositionEvent `bson:"events"`
	Distance   float64                 `bson:"distance,omitempty"`
	Floor      int16                   `bson:"floor"`
}

// ContactEvent is the aggregation of the MinuteAggregate between two people over a length of time
type ContactEvent struct {
	ID               primitive.ObjectID   `bson:"_id,omitempty"`
	Devices          [2]string            `bson:"devices"`
	Start            uint32               `bson:"start"`
	End              uint32               `bson:"end"`
	MinuteAggregates []primitive.ObjectID `bson:"minuteaggregates"`
	Duration         int                  `bson:"duration"`
	FirstContact     MinuteAggregate      `bson:"firstcontact"`
	MinDistance      float64              `bson:"mindistance"`
	MaxDistance      float64              `bson:"maxdistance"`
}
