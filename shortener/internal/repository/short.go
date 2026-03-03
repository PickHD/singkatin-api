package repository

import (
	"context"
	"fmt"
	"time"

	"singkatin-api/shortener/internal/config"
	"singkatin-api/shortener/internal/model"
	shortenerpb "singkatin-api/shortener/pkg/api/v1/proto/shortener"
	"singkatin-api/shortener/pkg/logger"

	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/protobuf/proto"
)

type (
	// ShortRepository is an interface that has all the function to be implemented inside short repository
	ShortRepository interface {
		GetListShortenerByUserID(ctx context.Context, userID string) ([]model.Short, error)
		Create(ctx context.Context, req *model.Short) error
		GetByShortURL(ctx context.Context, shortURL string) (*model.Short, error)
		GetFullURLByKey(ctx context.Context, shortURL string) (string, error)
		GetByID(ctx context.Context, ID string) (*model.Short, error)
		SetFullURLByKey(ctx context.Context, shortURL string, fullURL string, duration time.Duration) error
		PublishUpdateVisitorCount(ctx context.Context, req *model.UpdateVisitorRequest) error
		UpdateVisitorByShortURL(ctx context.Context, req *model.UpdateVisitorRequest, lastVisitedCount int64) error
		UpdateFullURLByID(ctx context.Context, req *model.UpdateShortRequest) error
		DeleteByID(ctx context.Context, req *model.DeleteShortRequest) error
		DeleteFullURLByKey(ctx context.Context, shortURL string) error
	}

	// shortRepositoryImpl is an app short struct that consists of all the dependencies needed for short repository
	shortRepositoryImpl struct {
		Config   *config.Config
		Tracer   *trace.TracerProvider
		DB       *mongo.Database
		Redis    *redis.Client
		RabbitMQ *amqp.Channel
	}
)

// NewShortRepository return new instances short repository
func NewShortRepository(config *config.Config, tracer *trace.TracerProvider, db *mongo.Database, rds *redis.Client, amqp *amqp.Channel) ShortRepository {
	return &shortRepositoryImpl{
		Config:   config,
		Tracer:   tracer,
		DB:       db,
		Redis:    rds,
		RabbitMQ: amqp,
	}
}

func (r *shortRepositoryImpl) GetListShortenerByUserID(ctx context.Context, userID string) ([]model.Short, error) {
	tr := r.Tracer.Tracer("Shortener-GetListShortenerByUserID Repository")
	ctx, span := tr.Start(ctx, "Start GetListShortenerByUserID")
	defer span.End()

	shorts := []model.Short{}

	cur, err := r.DB.Collection(r.Config.Database.ShortenersCollection).Find(ctx,
		bson.D{{Key: "user_id", Value: userID}},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}))
	if err != nil {
		logger.Error("ShortRepositoryImpl.GetListShortenerByUserID Find ERROR, ", err)
		return nil, err
	}

	for cur.Next(ctx) {
		var short model.Short

		err := cur.Decode(&short)
		if err != nil {
			logger.Error("ShortRepositoryImpl.GetListShortenerByUserID Decode ERROR, ", err)
		}

		shorts = append(shorts, short)
	}

	if err := cur.Err(); err != nil {
		logger.Error("ShortRepositoryImpl.GetListShortenerByUserID Cursors ERROR, ", err)
		return nil, err
	}

	return shorts, nil
}

func (r *shortRepositoryImpl) Create(ctx context.Context, req *model.Short) error {
	tr := r.Tracer.Tracer("Shortener-Create Repository")
	ctx, span := tr.Start(ctx, "Start Create")
	defer span.End()

	_, err := r.DB.Collection(r.Config.Database.ShortenersCollection).InsertOne(ctx,
		bson.D{{Key: "full_url", Value: req.FullURL},
			{Key: "user_id", Value: req.UserID},
			{Key: "short_url", Value: req.ShortURL},
			{Key: "visited", Value: 0}, {Key: "created_at", Value: time.Now()}})
	if err != nil {
		logger.Error("ShortRepositoryImpl.Create InsertOne ERROR, ", err)
		return err
	}

	return nil
}

func (r *shortRepositoryImpl) GetByShortURL(ctx context.Context, shortURL string) (*model.Short, error) {
	tr := r.Tracer.Tracer("Shortener-GetByShortURL Repository")
	ctx, span := tr.Start(ctx, "Start GetByShortURL")
	defer span.End()

	short := &model.Short{}

	err := r.DB.Collection(r.Config.Database.ShortenersCollection).FindOne(ctx, bson.D{{Key: "short_url", Value: shortURL}}).Decode(&short)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, model.NewError(model.NotFound, "short_url not found")
		}

		logger.Error("ShortRepositoryImpl.GetByShortURL FindOne ERROR,", err)
		return nil, err
	}

	return short, nil
}

func (r *shortRepositoryImpl) GetFullURLByKey(ctx context.Context, shortURL string) (string, error) {
	tr := r.Tracer.Tracer("Shortener-GetFullURLByKey Repository")
	ctx, span := tr.Start(ctx, "Start GetFullURLByKey")
	defer span.End()

	result := r.Redis.Get(ctx, fmt.Sprintf(model.KeyShortURL, shortURL))
	if result.Err() != nil {
		logger.Error("ShortRepositoryImpl.GetFullURLByKey Get ERROR, ", result.Err())

		return "", result.Err()
	}

	return result.Val(), nil
}

func (r *shortRepositoryImpl) GetByID(ctx context.Context, id string) (*model.Short, error) {
	tr := r.Tracer.Tracer("Shortener-GetByID Repository")
	ctx, span := tr.Start(ctx, "Start GetByID")
	defer span.End()

	objShortID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.Error("ShortRepositoryImpl.GetByID primitive.ObjectIDFromHex ERROR, ", err)
		return nil, err
	}

	short := &model.Short{}

	err = r.DB.Collection(r.Config.Database.ShortenersCollection).FindOne(ctx, bson.D{{Key: "_id", Value: objShortID}}).Decode(&short)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, model.NewError(model.NotFound, "short_url not found")
		}

		logger.Error("ShortRepositoryImpl.GetByID FindOne ERROR,", err)
		return nil, err
	}

	return short, nil
}

func (r *shortRepositoryImpl) SetFullURLByKey(ctx context.Context, shortURL string, fullURL string, duration time.Duration) error {
	tr := r.Tracer.Tracer("Shortener-SetFullURLByKey Repository")
	ctx, span := tr.Start(ctx, "Start SetFullURLByKey")
	defer span.End()

	err := r.Redis.SetEx(ctx, fmt.Sprintf(model.KeyShortURL, shortURL), fullURL, duration).Err()
	if err != nil {
		logger.Error("ShortRepositoryImpl.SetFullURLByKey SetEx ERROR, ", err)

		return err
	}

	return nil
}

func (r *shortRepositoryImpl) PublishUpdateVisitorCount(ctx context.Context, req *model.UpdateVisitorRequest) error {
	tr := r.Tracer.Tracer("Shortener-PublishUpdateVisitorCount Repository")
	_, span := tr.Start(ctx, "Start PublishUpdateVisitorCount")
	defer span.End()

	logger.Info("data req before publish", req)

	// transform data to proto
	msg := r.prepareProtoPublishUpdateVisitorCountMessage(req)

	b, err := proto.Marshal(msg)
	if err != nil {
		logger.Error("ShortRepositoryImpl.PublishUpdateVisitorCount Marshal proto UpdateVisitorCountMessage ERROR, ", err)
		return err
	}

	message := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(b),
	}

	// Attempt to publish a message to the queue.
	if err := r.RabbitMQ.Publish(
		"",                                   // exchange
		r.Config.RabbitMQ.QueueUpdateVisitor, // queue name
		false,                                // mandatory
		false,                                // immediate
		message,                              // message to publish
	); err != nil {
		logger.Error("ShortRepositoryImpl.PublishUpdateVisitorCount RabbitMQ.Publish ERROR, ", err)
		return err
	}

	logger.Info("Success Publish Update Visitor Count to Queue: ", r.Config.RabbitMQ.QueueUpdateVisitor)

	return nil
}

func (r *shortRepositoryImpl) UpdateVisitorByShortURL(ctx context.Context, req *model.UpdateVisitorRequest, lastVisitedCount int64) error {
	tr := r.Tracer.Tracer("Shortener-UpdateVisitorByShortURL Repository")
	ctx, span := tr.Start(ctx, "Start UpdateVisitorByShortURL")
	defer span.End()

	_, err := r.DB.Collection(r.Config.Database.ShortenersCollection).UpdateOne(ctx,
		bson.D{{Key: "short_url", Value: req.ShortURL}}, bson.M{
			"$set": bson.D{{Key: "visited", Value: lastVisitedCount + 1}, {Key: "updated_at", Value: time.Now()}},
		})
	if err != nil {
		logger.Error("ShortRepositoryImpl.UpdateVisitorByShortURL UpdateOne ERROR, ", err)
		return err
	}

	return nil
}

func (r *shortRepositoryImpl) UpdateFullURLByID(ctx context.Context, req *model.UpdateShortRequest) error {
	tr := r.Tracer.Tracer("Shortener-UpdateFullURLByID Repository")
	ctx, span := tr.Start(ctx, "Start UpdateFullURLByID")
	defer span.End()

	objShortID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		logger.Error("ShortRepositoryImpl.UpdateFullURLByID primitive.ObjectIDFromHex ERROR, ", err)
		return err
	}

	_, err = r.DB.Collection(r.Config.Database.ShortenersCollection).UpdateOne(ctx,
		bson.D{{Key: "_id", Value: objShortID}}, bson.M{
			"$set": bson.D{{Key: "full_url", Value: req.FullURL}, {Key: "updated_at", Value: time.Now()}},
		})
	if err != nil {
		logger.Error("ShortRepositoryImpl.UpdateFullURLByID UpdateOne ERROR, ", err)
		return err
	}

	return nil
}

func (r *shortRepositoryImpl) DeleteByID(ctx context.Context, req *model.DeleteShortRequest) error {
	tr := r.Tracer.Tracer("Shortener-DeleteByID Repository")
	ctx, span := tr.Start(ctx, "Start DeleteByID")
	defer span.End()

	objShortID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		logger.Error("ShortRepositoryImpl.DeleteByID primitive.ObjectIDFromHex ERROR, ", err)
		return err
	}

	_, err = r.DB.Collection(r.Config.Database.ShortenersCollection).DeleteOne(ctx,
		bson.D{{Key: "_id", Value: objShortID}})
	if err != nil {
		logger.Error("ShortRepositoryImpl.DeleteByID DeleteOne ERROR, ", err)
		return err
	}

	return nil
}

func (r *shortRepositoryImpl) DeleteFullURLByKey(ctx context.Context, shortURL string) error {
	tr := r.Tracer.Tracer("Shortener-DeleteFullURLByKey Repository")
	ctx, span := tr.Start(ctx, "Start DeleteFullURLByKey")
	defer span.End()

	err := r.Redis.Del(ctx, fmt.Sprintf(model.KeyShortURL, shortURL)).Err()
	if err != nil {
		logger.Error("ShortRepositoryImpl.DeleteFullURLByKey Del ERROR, ", err)

		return err
	}

	return nil
}

func (r *shortRepositoryImpl) prepareProtoPublishUpdateVisitorCountMessage(req *model.UpdateVisitorRequest) *shortenerpb.UpdateVisitorCountMessage {
	return &shortenerpb.UpdateVisitorCountMessage{
		ShortUrl: req.ShortURL,
	}
}
