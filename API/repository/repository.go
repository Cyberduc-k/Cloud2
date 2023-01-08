package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository[T any] struct {
	ctx  context.Context
	coll *mongo.Collection
}

func NewClient(ctx context.Context, connection string) (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(connection))

	if err != nil {
		return nil, err
	}

	if err = client.Connect(ctx); err != nil {
		return nil, err
	}

	return client, nil
}

func New[T any](ctx context.Context, client *mongo.Client, db, coll string) Repository[T] {
	collection := client.Database(db).Collection(coll)
	return Repository[T]{ctx, collection}
}

func (self *Repository[T]) GetAll() ([]T, error) {
	cursor, err := self.coll.Find(self.ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(self.ctx)
	var results []T

	if err = cursor.All(self.ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (self *Repository[T]) GetAllWhere(filter bson.M) ([]T, error) {
	cursor, err := self.coll.Find(self.ctx, filter)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(self.ctx)
	var results []T

	if err = cursor.All(self.ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (self *Repository[T]) GetById(id primitive.ObjectID) (T, error) {
	result := self.coll.FindOne(self.ctx, bson.M{"_id": id})
	var t T

	if err := result.Decode(&t); err != nil {
		return t, err
	}

	return t, nil
}

func (self *Repository[T]) Insert(value T) (primitive.ObjectID, error) {
	inserted, err := self.coll.InsertOne(self.ctx, value)

	if id, ok := inserted.InsertedID.(primitive.ObjectID); ok {
		return id, err
	}

	return primitive.NilObjectID, err
}

func (self *Repository[T]) Update(id primitive.ObjectID, value interface{}) error {
	_, err := self.coll.UpdateByID(self.ctx, id, value)
	return err
}

func (self *Repository[T]) Delete(id primitive.ObjectID) error {
	_, err := self.coll.DeleteOne(self.ctx, bson.M{"_id": id})
	return err
}
