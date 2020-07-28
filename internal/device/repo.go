package device

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Device struct {
	ID    string `json:"id" bson:"_id"`
	Type  string `json:"type" bson:"type"`
	Venue string `json:"venue" bson:"venue"`
}

type Repo interface {
	Activate(id string, deviceType string, venue string) (device *Device, err error)
	Create(id string, deviceType string, venue string) (device *Device, err error)
	Get(id string) (device *Device, err error)
	Delete(id string) (deleted bool, err error)
}

type repo struct {
	col *mongo.Collection
}

func NewRepo(col *mongo.Collection) Repo {
	return &repo{
		col,
	}
}

func (d *repo) Create(id string, deviceType string, venue string) (device *Device, err error) {
	device = &Device{
		ID:    id,
		Venue: venue,
		Type:  deviceType,
	}

	_, err = d.col.InsertOne(
		context.Background(),
		device,
	)

	if err != nil {
		return nil, err
	}

	return
}

func (d *repo) Get(id string) (device *Device, err error) {
	err = d.col.FindOne(
		context.Background(),
		bson.M{"_id": id},
	).Decode(&device)
	return
}

func (d *repo) Activate(id string, deviceType string, venue string) (device *Device, err error) {
	err = d.col.FindOneAndUpdate(
		context.Background(),
		bson.M{
			"_id": id,
		},
		bson.M{
			"$set": bson.M{
				"venue": venue,
				"type":  deviceType,
			},
		},
		options.FindOneAndUpdate().
			SetUpsert(true).
			SetReturnDocument(options.After),
	).Decode(&device)

	return
}

func (d *repo) Delete(id string) (deleted bool, err error) {
	result, err := d.col.DeleteOne(
		context.Background(),
		bson.M{"_id": id},
	)
	deleted = result.DeletedCount > 0
	return
}
