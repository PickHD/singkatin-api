package repository

import (
	"context"

	"singkatin-api/user/internal/config"
	"singkatin-api/user/internal/model"
	shortenerpb "singkatin-api/user/pkg/api/v1/proto/shortener"
	uploadpb "singkatin-api/user/pkg/api/v1/proto/upload"
	"singkatin-api/user/pkg/logger"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/protobuf/proto"
)

type (
	// UserRepository is an interface that has all the function to be implemented inside user repository
	UserRepository interface {
		FindByEmail(ctx context.Context, email string) (*model.User, error)
		PublishCreateUserShortener(ctx context.Context, req *model.GenerateShortUserMessage) error
		UpdateProfileByID(ctx context.Context, userID string, req *model.EditProfileRequest) error
		PublishUploadAvatarUser(ctx context.Context, req *model.UploadAvatarRequest) error
		UpdateAvatarUserByID(ctx context.Context, fileURL string, userID string) error
		PublishUpdateUserShortener(ctx context.Context, shortID string, req *model.ShortUserRequest) error
		PublishDeleteUserShortener(ctx context.Context, shortID string) error
	}

	// userRepositoryImpl is an app user struct that consists of all the dependencies needed for user repository
	userRepositoryImpl struct {
		Config   *config.Config
		Tracer   *trace.TracerProvider
		DB       *mongo.Database
		RabbitMQ *amqp.Channel
	}
)

// NewUserRepository return new instances user repository
func NewUserRepository(config *config.Config, tracer *trace.TracerProvider, db *mongo.Database, amqp *amqp.Channel) UserRepository {
	return &userRepositoryImpl{
		Config:   config,
		Tracer:   tracer,
		DB:       db,
		RabbitMQ: amqp,
	}
}

func (r *userRepositoryImpl) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	tr := r.Tracer.Tracer("User-FindByEmail Repository")
	_, span := tr.Start(ctx, "Start FindByEmail")
	defer span.End()

	user := model.User{}

	err := r.DB.Collection(r.Config.Database.UsersCollection).FindOne(ctx, bson.D{{Key: "email", Value: email}}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, model.NewError(model.NotFound, "users not found")
		}

		logger.Errorf("UserRepositoryImpl.FindByEmail FindOne ERROR, %v", err)
		return nil, err
	}

	return &user, nil
}

func (r *userRepositoryImpl) PublishCreateUserShortener(ctx context.Context, req *model.GenerateShortUserMessage) error {
	tr := r.Tracer.Tracer("User-PublishCreateUserShortener Repository")
	_, span := tr.Start(ctx, "Start PublishCreateUserShortener")
	defer span.End()

	logger.Info("data req before publish", req)

	// transform data to proto
	msg := r.prepareProtoPublishCreateUserShortenerMessage(req)

	b, err := proto.Marshal(msg)
	if err != nil {
		logger.Errorf("UserRepositoryImpl.PublishCreateUserShortener Marshal proto CreateShortenerMessage ERROR, %v", err)
		return err
	}

	message := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(b),
	}

	// Attempt to publish a message to the queue.
	if err := r.RabbitMQ.Publish(
		"",                                     // exchange
		r.Config.RabbitMQ.QueueCreateShortener, // queue name
		false,                                  // mandatory
		false,                                  // immediate
		message,                                // message to publish
	); err != nil {
		logger.Errorf("UserRepositoryImpl.PublishCreateUserShortener RabbitMQ.Publish ERROR, %v", err)
		return err
	}

	logger.Infof("Success Publish User Shortener to Queue: %s", r.Config.RabbitMQ.QueueCreateShortener)

	return nil
}

func (r *userRepositoryImpl) UpdateProfileByID(ctx context.Context, userID string, req *model.EditProfileRequest) error {
	tr := r.Tracer.Tracer("User-UpdateProfileByID Repository")
	_, span := tr.Start(ctx, "Start UpdateProfileByID")
	defer span.End()

	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logger.Errorf("UserRepositoryImpl.UpdateProfileByID primitive.ObjectIDFromHex ERROR, %v", err)
		return err
	}

	_, err = r.DB.Collection(r.Config.Database.UsersCollection).UpdateOne(ctx,
		bson.D{{Key: "_id", Value: objUserID}}, bson.M{
			"$set": bson.D{{Key: "fullname", Value: req.FullName}},
		})
	if err != nil {
		logger.Errorf("UserRepositoryImpl.UpdateProfileByID UpdateOne ERROR, %v", err)
		return err
	}

	return nil
}

func (r *userRepositoryImpl) PublishUploadAvatarUser(ctx context.Context, req *model.UploadAvatarRequest) error {
	tr := r.Tracer.Tracer("User-PublishUploadAvatarUser Repository")
	_, span := tr.Start(ctx, "Start PublishUploadAvatarUser")
	defer span.End()

	logger.Info("data req before publish", req)

	// transform data to proto
	msg := r.prepareProtoPublishUploadAvatarUserMessage(req)

	b, err := proto.Marshal(msg)
	if err != nil {
		logger.Errorf("UserRepositoryImpl.PublishUploadAvatarUser Marshal proto UploadAvatarMessage ERROR, %v", err)
		return err
	}

	message := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(b),
	}

	// Attempt to publish a message to the queue.
	if err := r.RabbitMQ.Publish(
		"",                                  // exchange
		r.Config.RabbitMQ.QueueUploadAvatar, // queue name
		false,                               // mandatory
		false,                               // immediate
		message,                             // message to publish
	); err != nil {
		logger.Errorf("UserRepositoryImpl.PublishUploadAvatarUser RabbitMQ.Publish ERROR, %v", err)
		return err
	}

	logger.Infof("Success Publish Upload Avatar Users to Queue: %s", r.Config.RabbitMQ.QueueUploadAvatar)

	return nil
}

func (r *userRepositoryImpl) UpdateAvatarUserByID(ctx context.Context, fileURL string, userID string) error {
	tr := r.Tracer.Tracer("User-UpdateAvatarUserByID Repository")
	_, span := tr.Start(ctx, "Start UpdateAvatarUserByID")
	defer span.End()

	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logger.Errorf("UserRepositoryImpl.UpdateAvatarUserByID primitive.ObjectIDFromHex ERROR, %v", err)
		return err
	}

	_, err = r.DB.Collection(r.Config.Database.UsersCollection).UpdateOne(ctx,
		bson.D{{Key: "_id", Value: objUserID}}, bson.M{
			"$set": bson.D{{Key: "avatar_url", Value: fileURL}},
		})
	if err != nil {
		logger.Errorf("UserRepositoryImpl.UpdateAvatarUserByID UpdateOne ERROR, %v", err)
		return err
	}

	return nil
}

func (r *userRepositoryImpl) PublishUpdateUserShortener(ctx context.Context, shortID string, req *model.ShortUserRequest) error {
	tr := r.Tracer.Tracer("User-PublishUpdateUserShortener Repository")
	_, span := tr.Start(ctx, "Start PublishUpdateUserShortener")
	defer span.End()

	logger.Info("data req before publish", req)

	// transform data to proto
	msg := r.prepareProtoPublishUpdateUserShortenerMessage(shortID, req)

	b, err := proto.Marshal(msg)
	if err != nil {
		logger.Errorf("UserRepositoryImpl.PublishUpdateUserShortener Marshal proto UpdateShortenerMessage ERROR, %v", err)
		return err
	}

	message := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(b),
	}

	// Attempt to publish a message to the queue.
	if err := r.RabbitMQ.Publish(
		"",                                     // exchange
		r.Config.RabbitMQ.QueueUpdateShortener, // queue name
		false,                                  // mandatory
		false,                                  // immediate
		message,                                // message to publish
	); err != nil {
		logger.Errorf("UserRepositoryImpl.PublishUpdateUserShortener RabbitMQ.Publish ERROR, %v", err)
		return err
	}

	logger.Infof("Success Publish User Shortener to Queue: %s", r.Config.RabbitMQ.QueueUpdateShortener)

	return nil
}

func (r *userRepositoryImpl) PublishDeleteUserShortener(ctx context.Context, shortID string) error {
	tr := r.Tracer.Tracer("User-PublishDeleteUserShortener Repository")
	_, span := tr.Start(ctx, "Start PublishDeleteUserShortener")
	defer span.End()

	logger.Info("data req before publish", shortID)

	// transform data to proto
	msg := r.prepareProtoPublishDeleteUserShortenerMessage(shortID)

	b, err := proto.Marshal(msg)
	if err != nil {
		logger.Errorf("UserRepositoryImpl.PublishDeleteUserShortener Marshal proto DeleteShortenerMessage ERROR, %v", err)
		return err
	}

	message := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(b),
	}

	// Attempt to publish a message to the queue.
	if err := r.RabbitMQ.Publish(
		"",                                     // exchange
		r.Config.RabbitMQ.QueueDeleteShortener, // queue name
		false,                                  // mandatory
		false,                                  // immediate
		message,                                // message to publish
	); err != nil {
		logger.Errorf("UserRepositoryImpl.PublishDeleteUserShortener RabbitMQ.Publish ERROR, %v", err)
		return err
	}

	logger.Infof("Success Publish User Shortener to Queue: %s", r.Config.RabbitMQ.QueueDeleteShortener)

	return nil
}

func (r *userRepositoryImpl) prepareProtoPublishCreateUserShortenerMessage(req *model.GenerateShortUserMessage) *shortenerpb.CreateShortenerMessage {
	return &shortenerpb.CreateShortenerMessage{
		FullUrl:  req.FullURL,
		UserId:   req.UserID,
		ShortUrl: req.ShortURL,
	}
}

func (r *userRepositoryImpl) prepareProtoPublishUploadAvatarUserMessage(req *model.UploadAvatarRequest) *uploadpb.UploadAvatarMessage {
	return &uploadpb.UploadAvatarMessage{
		FileName:    req.FileName,
		ContentType: req.ContentType,
		Avatars:     req.Avatars,
	}
}

func (r *userRepositoryImpl) prepareProtoPublishUpdateUserShortenerMessage(shortID string, req *model.ShortUserRequest) *shortenerpb.UpdateShortenerMessage {
	return &shortenerpb.UpdateShortenerMessage{
		Id:      shortID,
		FullUrl: req.FullURL,
	}
}

func (r *userRepositoryImpl) prepareProtoPublishDeleteUserShortenerMessage(shortID string) *shortenerpb.DeleteShortenerMessage {
	return &shortenerpb.DeleteShortenerMessage{
		Id: shortID,
	}
}
