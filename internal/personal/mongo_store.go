package personal

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	notes        *mongo.Collection
	reminders    *mongo.Collection
	transactions *mongo.Collection
	habits       *mongo.Collection
	timeout      time.Duration
}

func NewMongoStore(database *mongo.Database) *MongoStore {
	return &MongoStore{
		notes:        database.Collection("personal_notes"),
		reminders:    database.Collection("personal_reminders"),
		transactions: database.Collection("personal_transactions"),
		habits:       database.Collection("personal_habits"),
		timeout:      5 * time.Second,
	}
}

func (s *MongoStore) Summary() (State, error) {
	notes, err := s.ListNotes()
	if err != nil {
		return State{}, err
	}
	reminders, err := s.ListReminders()
	if err != nil {
		return State{}, err
	}
	transactions, err := s.ListTransactions()
	if err != nil {
		return State{}, err
	}
	habits, err := s.ListHabits()
	if err != nil {
		return State{}, err
	}
	return State{Notes: notes, Reminders: reminders, Transactions: transactions, Habits: habits}, nil
}

func (s *MongoStore) ListNotes() ([]Note, error) {
	ctx, cancel := s.context()
	defer cancel()
	cursor, err := s.notes.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "pinned", Value: -1}, {Key: "updatedAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []mongoNote
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	notes := make([]Note, 0, len(docs))
	for _, doc := range docs {
		notes = append(notes, doc.toDomain())
	}
	return notes, nil
}

func (s *MongoStore) CreateNote(note Note) (Note, error) {
	ctx, cancel := s.context()
	defer cancel()
	_, err := s.notes.InsertOne(ctx, mongoNoteFromDomain(note))
	return note, err
}

func (s *MongoStore) DeleteNote(id string) (bool, error) {
	return deleteOne(s.notes, id, s.timeout)
}

func (s *MongoStore) ListReminders() ([]Reminder, error) {
	ctx, cancel := s.context()
	defer cancel()
	cursor, err := s.reminders.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "done", Value: 1}, {Key: "dueAt", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []mongoReminder
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	reminders := make([]Reminder, 0, len(docs))
	for _, doc := range docs {
		reminders = append(reminders, doc.toDomain())
	}
	return reminders, nil
}

func (s *MongoStore) CreateReminder(reminder Reminder) (Reminder, error) {
	ctx, cancel := s.context()
	defer cancel()
	_, err := s.reminders.InsertOne(ctx, mongoReminderFromDomain(reminder))
	return reminder, err
}

func (s *MongoStore) ToggleReminder(id string) (Reminder, bool, error) {
	ctx, cancel := s.context()
	defer cancel()
	var current mongoReminder
	if err := s.reminders.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&current); err != nil {
		if err == mongo.ErrNoDocuments {
			return Reminder{}, false, nil
		}
		return Reminder{}, false, err
	}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "done", Value: !current.Done}, {Key: "updatedAt", Value: time.Now().UTC()}}}}
	result := s.reminders.FindOneAndUpdate(ctx, bson.D{{Key: "_id", Value: id}}, update, options.FindOneAndUpdate().SetReturnDocument(options.After))
	if result.Err() != nil {
		return Reminder{}, false, result.Err()
	}
	var updated mongoReminder
	if err := result.Decode(&updated); err != nil {
		return Reminder{}, false, err
	}
	return updated.toDomain(), true, nil
}

func (s *MongoStore) DeleteReminder(id string) (bool, error) {
	return deleteOne(s.reminders, id, s.timeout)
}

func (s *MongoStore) ListTransactions() ([]Transaction, error) {
	ctx, cancel := s.context()
	defer cancel()
	cursor, err := s.transactions.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "occurredAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []mongoTransaction
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	transactions := make([]Transaction, 0, len(docs))
	for _, doc := range docs {
		transactions = append(transactions, doc.toDomain())
	}
	return transactions, nil
}

func (s *MongoStore) CreateTransaction(transaction Transaction) (Transaction, error) {
	ctx, cancel := s.context()
	defer cancel()
	_, err := s.transactions.InsertOne(ctx, mongoTransactionFromDomain(transaction))
	return transaction, err
}

func (s *MongoStore) DeleteTransaction(id string) (bool, error) {
	return deleteOne(s.transactions, id, s.timeout)
}

func (s *MongoStore) ListHabits() ([]Habit, error) {
	ctx, cancel := s.context()
	defer cancel()
	cursor, err := s.habits.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "streak", Value: -1}, {Key: "updatedAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []mongoHabit
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	habits := make([]Habit, 0, len(docs))
	for _, doc := range docs {
		habits = append(habits, doc.toDomain())
	}
	return habits, nil
}

func (s *MongoStore) CreateHabit(habit Habit) (Habit, error) {
	ctx, cancel := s.context()
	defer cancel()
	_, err := s.habits.InsertOne(ctx, mongoHabitFromDomain(habit))
	return habit, err
}

func (s *MongoStore) CheckInHabit(id string) (Habit, bool, error) {
	ctx, cancel := s.context()
	defer cancel()
	var current mongoHabit
	if err := s.habits.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&current); err != nil {
		if err == mongo.ErrNoDocuments {
			return Habit{}, false, nil
		}
		return Habit{}, false, err
	}
	habit := checkedInHabit(current.toDomain(), time.Now().UTC())
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "streak", Value: habit.Streak}, {Key: "lastCheckedAt", Value: habit.LastCheckedAt}, {Key: "updatedAt", Value: habit.UpdatedAt}}}}
	if _, err := s.habits.UpdateByID(ctx, id, update); err != nil {
		return Habit{}, false, err
	}
	return habit, true, nil
}

func (s *MongoStore) DeleteHabit(id string) (bool, error) {
	return deleteOne(s.habits, id, s.timeout)
}

func (s *MongoStore) context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.timeout)
}

func deleteOne(collection *mongo.Collection, id string, timeout time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	result, err := collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return false, err
	}
	return result.DeletedCount > 0, nil
}

type mongoNote struct {
	ID        string    `bson:"_id"`
	Title     string    `bson:"title"`
	Body      string    `bson:"body"`
	Tags      []string  `bson:"tags"`
	Pinned    bool      `bson:"pinned"`
	CreatedAt time.Time `bson:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

type mongoReminder struct {
	ID        string    `bson:"_id"`
	Title     string    `bson:"title"`
	Notes     string    `bson:"notes"`
	DueAt     time.Time `bson:"dueAt"`
	Done      bool      `bson:"done"`
	CreatedAt time.Time `bson:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

type mongoTransaction struct {
	ID         string    `bson:"_id"`
	Type       string    `bson:"type"`
	Amount     float64   `bson:"amount"`
	Category   string    `bson:"category"`
	Note       string    `bson:"note"`
	OccurredAt time.Time `bson:"occurredAt"`
	CreatedAt  time.Time `bson:"createdAt"`
	UpdatedAt  time.Time `bson:"updatedAt"`
}

type mongoHabit struct {
	ID            string    `bson:"_id"`
	Name          string    `bson:"name"`
	Target        string    `bson:"target"`
	Frequency     string    `bson:"frequency"`
	Streak        int       `bson:"streak"`
	LastCheckedAt time.Time `bson:"lastCheckedAt"`
	CreatedAt     time.Time `bson:"createdAt"`
	UpdatedAt     time.Time `bson:"updatedAt"`
}

func mongoNoteFromDomain(note Note) mongoNote {
	return mongoNote{
		ID: note.ID, Title: note.Title, Body: note.Body, Tags: note.Tags, Pinned: note.Pinned,
		CreatedAt: note.CreatedAt, UpdatedAt: note.UpdatedAt,
	}
}

func (n mongoNote) toDomain() Note {
	return Note{
		ID: n.ID, Title: n.Title, Body: n.Body, Tags: n.Tags, Pinned: n.Pinned,
		CreatedAt: n.CreatedAt, UpdatedAt: n.UpdatedAt,
	}
}

func mongoReminderFromDomain(reminder Reminder) mongoReminder {
	return mongoReminder{
		ID: reminder.ID, Title: reminder.Title, Notes: reminder.Notes, DueAt: reminder.DueAt, Done: reminder.Done,
		CreatedAt: reminder.CreatedAt, UpdatedAt: reminder.UpdatedAt,
	}
}

func (r mongoReminder) toDomain() Reminder {
	return Reminder{
		ID: r.ID, Title: r.Title, Notes: r.Notes, DueAt: r.DueAt, Done: r.Done,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func mongoTransactionFromDomain(transaction Transaction) mongoTransaction {
	return mongoTransaction{
		ID: transaction.ID, Type: transaction.Type, Amount: transaction.Amount, Category: transaction.Category, Note: transaction.Note,
		OccurredAt: transaction.OccurredAt, CreatedAt: transaction.CreatedAt, UpdatedAt: transaction.UpdatedAt,
	}
}

func (t mongoTransaction) toDomain() Transaction {
	return Transaction{
		ID: t.ID, Type: t.Type, Amount: t.Amount, Category: t.Category, Note: t.Note,
		OccurredAt: t.OccurredAt, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}

func mongoHabitFromDomain(habit Habit) mongoHabit {
	return mongoHabit{
		ID: habit.ID, Name: habit.Name, Target: habit.Target, Frequency: habit.Frequency, Streak: habit.Streak,
		LastCheckedAt: habit.LastCheckedAt, CreatedAt: habit.CreatedAt, UpdatedAt: habit.UpdatedAt,
	}
}

func (h mongoHabit) toDomain() Habit {
	return Habit{
		ID: h.ID, Name: h.Name, Target: h.Target, Frequency: h.Frequency, Streak: h.Streak,
		LastCheckedAt: h.LastCheckedAt, CreatedAt: h.CreatedAt, UpdatedAt: h.UpdatedAt,
	}
}
