package repository

import (
	"context"

	"github.com/PickHD/singkatin-revamp/user/internal/config"
	"github.com/PickHD/singkatin-revamp/user/internal/model"
	shortenerpb "github.com/PickHD/singkatin-revamp/user/pkg/api/v1/proto/shortener"
	uploadpb "github.com/PickHD/singkatin-revamp/user/pkg/api/v1/proto/upload"
	"github.com/sirupsen/logrus"
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

	// UserRepositoryImpl is an app user struct that consists of all the dependencies needed for user repository
	UserRepositoryImpl struct {
		Context  context.Context
		Config   *config.Configuration
		Logger   *logrus.Logger
		Tracer   *trace.TracerProvider
		DB       *mongo.Database
		RabbitMQ *amqp.Channel
	}
)

// NewUserRepository return new instances user repository
func NewUserRepository(ctx context.Context, config *config.Configuration, logger *logrus.Logger, tracer *trace.TracerProvider, db *mongo.Database, amqp *amqp.Channel) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		Context:  ctx,
		Config:   config,
		Logger:   logger,
		Tracer:   tracer,
		DB:       db,
		RabbitMQ: amqp,
	}
}

func (ur *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	tr := ur.Tracer.Tracer("User-FindByEmail Repository")
	_, span := tr.Start(ctx, "Start FindByEmail")
	defer span.End()

	user := model.User{}

	err := ur.DB.Collection(ur.Config.Database.UsersCollection).FindOne(ctx, bson.D{{Key: "email", Value: email}}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, model.NewError(model.NotFound, "users not found")
		}

		ur.Logger.Error("UserRepositoryImpl.FindByEmail FindOne ERROR, ", err)
		return nil, err
	}

	return &user, nil
}

func (ur *UserRepositoryImpl) PublishCreateUserShortener(ctx context.Context, req *model.GenerateShortUserMessage) error {
	tr := ur.Tracer.Tracer("User-PublishCreateUserShortener Repository")
	_, span := tr.Start(ctx, "Start PublishCreateUserShortener")
	defer span.End()

	ur.Logger.Info("data req before publish", req)

	// transform data to proto
	msg := ur.prepareProtoPublishCreateUserShortenerMessage(req)

	b, err := proto.Marshal(msg)
	if err != nil {
		ur.Logger.Error("UserRepositoryImpl.PublishCreateUserShortener Marshal proto CreateShortenerMessage ERROR, ", err)
		return err
	}

	message := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(b),
	}

	// Attempt to publish a message to the queue.
	if err := ur.RabbitMQ.Publish(
		"",                                      // exchange
		ur.Config.RabbitMQ.QueueCreateShortener, // queue name
		false,                                   // mandatory
		false,                                   // immediate
		message,                                 // message to publish
	); err != nil {
		ur.Logger.Error("UserRepositoryImpl.PublishCreateUserShortener RabbitMQ.Publish ERROR, ", err)
		return err
	}

	ur.Logger.Info("Success Publish User Shortener to Queue: ", ur.Config.RabbitMQ.QueueCreateShortener)

	return nil
}

func (ur *UserRepositoryImpl) UpdateProfileByID(ctx context.Context, userID string, req *model.EditProfileRequest) error {
	tr := ur.Tracer.Tracer("User-UpdateProfileByID Repository")
	_, span := tr.Start(ctx, "Start UpdateProfileByID")
	defer span.End()

	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		ur.Logger.Error("UserRepositoryImpl.UpdateProfileByID primitive.ObjectIDFromHex ERROR, ", err)
		return err
	}

	_, err = ur.DB.Collection(ur.Config.Database.UsersCollection).UpdateOne(ctx,
		bson.D{{Key: "_id", Value: objUserID}}, bson.M{
			"$set": bson.D{{Key: "fullname", Value: req.FullName}},
		})
	if err != nil {
		ur.Logger.Error("UserRepositoryImpl.UpdateProfileByID UpdateOne ERROR, ", err)
		return err
	}

	return nil
}

func (ur *UserRepositoryImpl) PublishUploadAvatarUser(ctx context.Context, req *model.UploadAvatarRequest) error {
	tr := ur.Tracer.Tracer("User-PublishUploadAvatarUser Repository")
	_, span := tr.Start(ctx, "Start PublishUploadAvatarUser")
	defer span.End()

	ur.Logger.Info("data req before publish", req)

	// transform data to proto
	msg := ur.prepareProtoPublishUploadAvatarUserMessage(req)

	b, err := proto.Marshal(msg)
	if err != nil {
		ur.Logger.Error("UserRepositoryImpl.PublishUploadAvatarUser Marshal proto UploadAvatarMessage ERROR, ", err)
		return err
	}

	message := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(b),
	}

	// Attempt to publish a message to the queue.
	if err := ur.RabbitMQ.Publish(
		"",                                   // exchange
		ur.Config.RabbitMQ.QueueUploadAvatar, // queue name
		false,                                // mandatory
		false,                                // immediate
		message,                              // message to publish
	); err != nil {
		ur.Logger.Error("UserRepositoryImpl.PublishUploadAvatarUser RabbitMQ.Publish ERROR, ", err)
		return err
	}

	ur.Logger.Info("Success Publish Upload Avatar Users to Queue: ", ur.Config.RabbitMQ.QueueUploadAvatar)

	return nil
}

func (ur *UserRepositoryImpl) UpdateAvatarUserByID(ctx context.Context, fileURL string, userID string) error {
	tr := ur.Tracer.Tracer("User-UpdateAvatarUserByID Repository")
	_, span := tr.Start(ctx, "Start UpdateAvatarUserByID")
	defer span.End()

	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		ur.Logger.Error("UserRepositoryImpl.UpdateAvatarUserByID primitive.ObjectIDFromHex ERROR, ", err)
		return err
	}

	_, err = ur.DB.Collection(ur.Config.Database.UsersCollection).UpdateOne(ctx,
		bson.D{{Key: "_id", Value: objUserID}}, bson.M{
			"$set": bson.D{{Key: "avatar_url", Value: fileURL}},
		})
	if err != nil {
		ur.Logger.Error("UserRepositoryImpl.UpdateAvatarUserByID UpdateOne ERROR, ", err)
		return err
	}

	return nil
}

func (ur *UserRepositoryImpl) PublishUpdateUserShortener(ctx context.Context, shortID string, req *model.ShortUserRequest) error {
	tr := ur.Tracer.Tracer("User-PublishUpdateUserShortener Repository")
	_, span := tr.Start(ctx, "Start PublishUpdateUserShortener")
	defer span.End()

	ur.Logger.Info("data req before publish", req)

	// transform data to proto
	msg := ur.prepareProtoPublishUpdateUserShortenerMessage(shortID, req)

	b, err := proto.Marshal(msg)
	if err != nil {
		ur.Logger.Error("UserRepositoryImpl.PublishUpdateUserShortener Marshal proto UpdateShortenerMessage ERROR, ", err)
		return err
	}

	message := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(b),
	}

	// Attempt to publish a message to the queue.
	if err := ur.RabbitMQ.Publish(
		"",                                      // exchange
		ur.Config.RabbitMQ.QueueUpdateShortener, // queue name
		false,                                   // mandatory
		false,                                   // immediate
		message,                                 // message to publish
	); err != nil {
		ur.Logger.Error("UserRepositoryImpl.PublishUpdateUserShortener RabbitMQ.Publish ERROR, ", err)
		return err
	}

	ur.Logger.Info("Success Publish User Shortener to Queue: ", ur.Config.RabbitMQ.QueueUpdateShortener)

	return nil
}

func (ur *UserRepositoryImpl) PublishDeleteUserShortener(ctx context.Context, shortID string) error {
	tr := ur.Tracer.Tracer("User-PublishDeleteUserShortener Repository")
	_, span := tr.Start(ctx, "Start PublishDeleteUserShortener")
	defer span.End()

	ur.Logger.Info("data req before publish", shortID)

	// transform data to proto
	msg := ur.prepareProtoPublishDeleteUserShortenerMessage(shortID)

	b, err := proto.Marshal(msg)
	if err != nil {
		ur.Logger.Error("UserRepositoryImpl.PublishDeleteUserShortener Marshal proto DeleteShortenerMessage ERROR, ", err)
		return err
	}

	message := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(b),
	}

	// Attempt to publish a message to the queue.
	if err := ur.RabbitMQ.Publish(
		"",                                      // exchange
		ur.Config.RabbitMQ.QueueDeleteShortener, // queue name
		false,                                   // mandatory
		false,                                   // immediate
		message,                                 // message to publish
	); err != nil {
		ur.Logger.Error("UserRepositoryImpl.PublishDeleteUserShortener RabbitMQ.Publish ERROR, ", err)
		return err
	}

	ur.Logger.Info("Success Publish User Shortener to Queue: ", ur.Config.RabbitMQ.QueueDeleteShortener)

	return nil
}

func (ur *UserRepositoryImpl) prepareProtoPublishCreateUserShortenerMessage(req *model.GenerateShortUserMessage) *shortenerpb.CreateShortenerMessage {
	return &shortenerpb.CreateShortenerMessage{
		FullUrl:  req.FullURL,
		UserId:   req.UserID,
		ShortUrl: req.ShortURL,
	}
}

func (ur *UserRepositoryImpl) prepareProtoPublishUploadAvatarUserMessage(req *model.UploadAvatarRequest) *uploadpb.UploadAvatarMessage {
	return &uploadpb.UploadAvatarMessage{
		FileName:    req.FileName,
		ContentType: req.ContentType,
		Avatars:     req.Avatars,
	}
}

func (ur *UserRepositoryImpl) prepareProtoPublishUpdateUserShortenerMessage(shortID string, req *model.ShortUserRequest) *shortenerpb.UpdateShortenerMessage {
	return &shortenerpb.UpdateShortenerMessage{
		Id:      shortID,
		FullUrl: req.FullURL,
	}
}

func (ur *UserRepositoryImpl) prepareProtoPublishDeleteUserShortenerMessage(shortID string) *shortenerpb.DeleteShortenerMessage {
	return &shortenerpb.DeleteShortenerMessage{
		Id: shortID,
	}
}
