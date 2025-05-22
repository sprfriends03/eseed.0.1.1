package db

import (
	"context"
	"strings"

	"github.com/nhnghia272/gopkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type repo struct {
	col *mongo.Collection
}

func newrepo(col *mongo.Collection) *repo {
	return &repo{col}
}

func (s repo) Save(ctx context.Context, id primitive.ObjectID, domain any, opts ...*options.UpdateOptions) (primitive.ObjectID, error) {
	if id.IsZero() {
		obj, err := s.col.InsertOne(ctx, domain)
		if err != nil {
			return primitive.NilObjectID, err
		}
		return obj.InsertedID.(primitive.ObjectID), nil
	} else {
		obj, err := s.col.UpdateByID(ctx, id, M{"$set": domain}, opts...)
		if err != nil {
			return primitive.NilObjectID, err
		}
		if obj.UpsertedID != nil {
			return obj.UpsertedID.(primitive.ObjectID), nil
		}
		return id, nil
	}
}

func (s repo) InsertMany(ctx context.Context, domains []any, opts ...*options.InsertManyOptions) ([]primitive.ObjectID, error) {
	result, err := s.col.InsertMany(ctx, domains, opts...)
	if err != nil {
		return nil, err
	}
	return gopkg.MapFunc(result.InsertedIDs, func(e any) primitive.ObjectID { return e.(primitive.ObjectID) }), nil
}

func (s repo) FindOne(ctx context.Context, filter M, domain any, opts ...*options.FindOneOptions) error {
	return s.col.FindOne(ctx, filter, opts...).Decode(domain)
}

func (s repo) FindOneAndDelete(ctx context.Context, filter M, domain any, opts ...*options.FindOneAndDeleteOptions) error {
	return s.col.FindOneAndDelete(ctx, filter, opts...).Decode(domain)
}

func (s repo) FindOneAndUpdate(ctx context.Context, filter M, update M, domain any, opts ...*options.FindOneAndUpdateOptions) error {
	return s.col.FindOneAndUpdate(ctx, filter, update, opts...).Decode(domain)
}

func (s repo) CountDocuments(ctx context.Context, q Query, opts ...*options.CountOptions) int64 {
	cnt, _ := s.col.CountDocuments(ctx, q.Filter, opts...)
	return cnt
}

func (s repo) FindAll(ctx context.Context, q Query, domains any, opts ...*options.FindOptions) error {
	sorts := bson.D{}
	for sort := range strings.SplitSeq(q.Sorts, ",") {
		if p := strings.Split(strings.TrimSpace(sort), "."); len(p) == 2 {
			if strings.ToLower(p[1]) == "asc" {
				sorts = append(sorts, bson.E{Key: p[0], Value: 1})
			} else if strings.ToLower(p[1]) == "desc" {
				sorts = append(sorts, bson.E{Key: p[0], Value: -1})
			}
		}
	}
	if len(sorts) == 0 {
		sorts = bson.D{{Key: "_id", Value: -1}}
	}

	opt := options.Find().SetSkip((q.Page - 1) * q.Limit).SetLimit(q.Limit).SetSort(sorts)
	if len(opts) > 0 {
		opt = opts[0].SetSkip((q.Page - 1) * q.Limit).SetLimit(q.Limit).SetSort(sorts)
	}

	cursor, err := s.col.Find(ctx, q.Filter, opt)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	return cursor.All(ctx, domains)
}

func (s repo) UpdateOne(ctx context.Context, filter M, update M, opts ...*options.UpdateOptions) error {
	_, err := s.col.UpdateOne(ctx, filter, update, opts...)
	return err
}

func (s repo) UpdateMany(ctx context.Context, filter M, update M, opts ...*options.UpdateOptions) error {
	_, err := s.col.UpdateMany(ctx, filter, update, opts...)
	return err
}

func (s repo) DeleteOne(ctx context.Context, filter M, opts ...*options.DeleteOptions) error {
	_, err := s.col.DeleteOne(ctx, filter, opts...)
	return err
}

func (s repo) DeleteMany(ctx context.Context, filter M, opts ...*options.DeleteOptions) error {
	_, err := s.col.DeleteMany(ctx, filter, opts...)
	return err
}

func (s repo) Aggregate(ctx context.Context, pipeline []M, results any, opts ...*options.AggregateOptions) error {
	cursor, err := s.col.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	return cursor.All(ctx, results)
}

func (s repo) Distinct(ctx context.Context, field string, filter M, results any, opts ...*options.DistinctOptions) error {
	fields, err := s.col.Distinct(ctx, field, filter, opts...)
	if err != nil {
		return err
	}
	return gopkg.Convert(fields, results)
}
