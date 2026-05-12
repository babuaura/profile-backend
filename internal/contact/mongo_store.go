package contact

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	collection *mongo.Collection
	timeout    time.Duration
}

func NewMongoStore(database *mongo.Database) *MongoStore {
	return &MongoStore{collection: database.Collection("contact_messages"), timeout: 5 * time.Second}
}

func (s *MongoStore) Save(message Message) error {
	ctx, cancel := s.context()
	defer cancel()
	_, err := s.collection.InsertOne(ctx, mongoMessageFromDomain(message))
	return err
}

func (s *MongoStore) List() ([]Message, error) {
	ctx, cancel := s.context()
	defer cancel()
	cursor, err := s.collection.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []mongoMessage
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	messages := make([]Message, 0, len(docs))
	for _, doc := range docs {
		messages = append(messages, doc.toDomain())
	}
	return messages, nil
}

func (s *MongoStore) UpdateStatus(id string, status string) (Message, bool, error) {
	ctx, cancel := s.context()
	defer cancel()
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: status}}}}
	result := s.collection.FindOneAndUpdate(ctx, bson.D{{Key: "_id", Value: id}}, update, options.FindOneAndUpdate().SetReturnDocument(options.After))
	if result.Err() == mongo.ErrNoDocuments {
		return Message{}, false, nil
	}
	if result.Err() != nil {
		return Message{}, false, result.Err()
	}
	var doc mongoMessage
	if err := result.Decode(&doc); err != nil {
		return Message{}, false, err
	}
	return doc.toDomain(), true, nil
}

func (s *MongoStore) Delete(id string) (bool, error) {
	ctx, cancel := s.context()
	defer cancel()
	result, err := s.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return false, err
	}
	return result.DeletedCount > 0, nil
}

func (s *MongoStore) context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.timeout)
}

type mongoMessage struct {
	ID        string    `bson:"_id"`
	Name      string    `bson:"name"`
	Email     string    `bson:"email"`
	Budget    string    `bson:"budget,omitempty"`
	Message   string    `bson:"message"`
	Source    string    `bson:"source"`
	Status    string    `bson:"status"`
	CreatedAt time.Time `bson:"createdAt"`
}

func mongoMessageFromDomain(message Message) mongoMessage {
	return mongoMessage{
		ID:        message.ID,
		Name:      message.Name,
		Email:     message.Email,
		Budget:    message.Budget,
		Message:   message.Message,
		Source:    message.Source,
		Status:    message.Status,
		CreatedAt: message.CreatedAt,
	}
}

func (m mongoMessage) toDomain() Message {
	return Message{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		Budget:    m.Budget,
		Message:   m.Message,
		Source:    m.Source,
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
	}
}
