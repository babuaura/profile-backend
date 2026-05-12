package profile

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const ownerProfileID = "owner"

type MongoStore struct {
	collection *mongo.Collection
	timeout    time.Duration
}

func NewMongoStore(database *mongo.Database) *MongoStore {
	return &MongoStore{collection: database.Collection("profiles"), timeout: 5 * time.Second}
}

func (s *MongoStore) Get() (Profile, error) {
	ctx, cancel := s.context()
	defer cancel()
	var doc mongoProfile
	err := s.collection.FindOne(ctx, bson.D{{Key: "_id", Value: ownerProfileID}}).Decode(&doc)
	if err == nil {
		return doc.Profile, nil
	}
	if err != mongo.ErrNoDocuments {
		return Profile{}, err
	}
	profile := defaultProfile()
	_, err = s.collection.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: ownerProfileID}},
		bson.D{{Key: "$setOnInsert", Value: mongoProfile{ID: ownerProfileID, Profile: profile}}},
		options.Update().SetUpsert(true),
	)
	return profile, err
}

func (s *MongoStore) Save(profile Profile) (Profile, error) {
	ctx, cancel := s.context()
	defer cancel()
	_, err := s.collection.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: ownerProfileID}},
		bson.D{{Key: "$set", Value: mongoProfile{ID: ownerProfileID, Profile: profile}}},
		options.Update().SetUpsert(true),
	)
	return profile, err
}

func (s *MongoStore) context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.timeout)
}

type mongoProfile struct {
	ID      string `bson:"_id"`
	Profile `bson:",inline"`
}
