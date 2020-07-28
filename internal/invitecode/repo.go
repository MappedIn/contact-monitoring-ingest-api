package invitecode

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// InviteCode represents an invite code for devices
// to become active on a venue
type InviteCode struct {
	Code    string `json:"code" bson:"_id"`
	MaxUses int    `json:"maxUses" bson:"maxUses"`
	Venue   string `json:"venue" bson:"venue"`
}

// Repo is an interface for accessing invite code data
// from its persistence layer
type Repo interface {
	Create(venue string, maxUses int) (inviteCode *InviteCode, err error)
	Delete(code string) (deleted bool, err error)
	Exists(code string) (exists bool, err error)
	Get(code string) (inviteCode *InviteCode, err error)
	GetByVenue(venue string) (inviteCodes []InviteCode, err error)
	UseOne(code string) (inviteCode *InviteCode, err error)
}

type repo struct {
	col *mongo.Collection
}

// NewRepo returns a new Repo interface
func NewRepo(col *mongo.Collection) Repo {
	return &repo{
		col,
	}
}

// Create attempts to generate a new invite code and insert it into the
// DB. We attempt to get a unique invite code 10 times and if that
// fails then return an error.
func (i *repo) Create(venue string, maxUses int) (inviteCode *InviteCode, err error) {
	newCode := GenerateNumberString(8)
	exists := false

	for x := 0; x < 10; x++ {
		exists, err = i.Exists(newCode)
		if err != nil && err != mongo.ErrNoDocuments {
			return nil, err
		}

		if exists {
			newCode = GenerateNumberString(8)
		} else {
			inviteCode = &InviteCode{
				Code:    newCode,
				MaxUses: maxUses,
				Venue:   venue,
			}
			_, err = i.col.InsertOne(
				context.Background(),
				inviteCode,
			)

			return
		}
	}

	return nil, errors.New("after 10 attempts, could not generate unique code")
}

// Exists return true if a code already exists in DB, false if it does not exist
func (i *repo) Exists(code string) (exists bool, err error) {
	inviteCode, err := i.Get(code)
	if err != nil {
		return false, err
	}
	exists = inviteCode != nil && inviteCode.Code == code
	return
}

// Get returns invite code from DB by code
func (i *repo) Get(code string) (inviteCode *InviteCode, err error) {
	err = i.col.FindOne(
		context.Background(),
		bson.M{"_id": code},
	).Decode(&inviteCode)

	return inviteCode, err
}

// UseOne finds and updates an invite code with at least 1 use by code
// and decrements its maxUses by 1. If the invite code is 0 then the
// invite code is deleted.
func (i *repo) UseOne(code string) (inviteCode *InviteCode, err error) {
	err = i.col.FindOneAndUpdate(
		context.Background(),
		bson.M{
			"_id": code,
			"maxUses": bson.M{
				"$gt": 0,
			},
		},
		bson.M{
			"$inc": bson.M{
				"maxUses": -1,
			},
		},
	).Decode(&inviteCode)

	if err != nil {
		return
	}

	if inviteCode.MaxUses <= 0 {
		i.Delete(inviteCode.Code)
	}

	return
}

// GetByVenue returns all invite codes corresponding to provided
// venue slug
func (i *repo) GetByVenue(venue string) (inviteCodes []InviteCode, err error) {
	cursor, err := i.col.Find(
		context.Background(),
		bson.M{"venue": venue},
	)
	if err != nil {
		return
	}

	err = cursor.All(context.Background(), &inviteCodes)
	if err != nil {
		return
	}

	return inviteCodes, nil
}

// Delete removes an invite code from DB by code
func (i *repo) Delete(code string) (deleted bool, err error) {
	result, err := i.col.DeleteOne(context.Background(), bson.M{"code": code})
	deleted = result.DeletedCount > 0
	return
}
