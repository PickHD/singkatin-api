package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/PickHD/singkatin-revamp/auth/internal/config"
	"github.com/PickHD/singkatin-revamp/auth/internal/model"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// AuthRepository is an interface that has all the function to be implemented inside auth repository
	AuthRepository interface {
		CreateUser(ctx context.Context, req *model.User) (*model.User, error)
		FindByEmail(ctx context.Context, email string) (*model.User, error)
		SetVerificationByEmail(ctx context.Context, email string, code string, duration time.Duration, verificationType model.VerificationType) error
		GetVerificationByCode(ctx context.Context, code string, verificationType model.VerificationType) (string, error)
		UpdateVerifyStatusByEmail(ctx context.Context, email string) error
		UpdatePasswordByEmail(ctx context.Context, email string, newPassword string) error
	}

	// AuthRepositoryImpl is an app auth struct that consists of all the dependencies needed for auth repository
	AuthRepositoryImpl struct {
		Context context.Context
		Config  *config.Configuration
		Logger  *logrus.Logger
		Tracer  *trace.TracerProvider
		DB      *mongo.Database
		Redis   *redis.Client
	}
)

// NewAuthRepository return new instances auth repository
func NewAuthRepository(ctx context.Context, config *config.Configuration, logger *logrus.Logger, tracer *trace.TracerProvider, db *mongo.Database, rds *redis.Client) *AuthRepositoryImpl {
	return &AuthRepositoryImpl{
		Context: ctx,
		Config:  config,
		Logger:  logger,
		Tracer:  tracer,
		DB:      db,
		Redis:   rds,
	}
}

func (ar *AuthRepositoryImpl) CreateUser(ctx context.Context, req *model.User) (*model.User, error) {
	tr := ar.Tracer.Tracer("Auth-CreateUser repository")
	ctx, span := tr.Start(ctx, "Start CreateUser")
	defer span.End()

	// check data users by email is already exists or not
	err := ar.DB.Collection(ar.Config.Database.UsersCollection).FindOne(ctx, bson.D{{Key: "email", Value: req.Email}}).Err()
	if err != nil {
		// if doc not exists, create new one
		if err == mongo.ErrNoDocuments {
			res, err := ar.DB.Collection(ar.Config.Database.UsersCollection).InsertOne(ctx, model.User{
				FullName:   req.FullName,
				Email:      req.Email,
				Password:   req.Password,
				CreatedAt:  time.Now(),
				IsVerified: false,
			})
			if err != nil {
				ar.Logger.Error("AuthRepositoryImpl.CreateUser InsertOne ERROR, ", err)
				return nil, err
			}

			id, ok := res.InsertedID.(primitive.ObjectID)
			if !ok {
				ar.Logger.Error("AuthRepositoryImpl.CreateUser Type Assertion ERROR, ", err)
				return nil, model.NewError(model.Type, "type assertion error")
			}
			req.ID = id

			return req, nil
		}

		ar.Logger.Error("AuthRepositoryImpl.CreateUser FindOne ERROR, ", err)
		return nil, err
	}

	return nil, model.NewError(model.Validation, "email already exists")
}

func (ar *AuthRepositoryImpl) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	tr := ar.Tracer.Tracer("Auth-FindByEmail repository")
	ctx, span := tr.Start(ctx, "Start FindByEmail")
	defer span.End()

	user := model.User{}

	err := ar.DB.Collection(ar.Config.Database.UsersCollection).FindOne(ctx, bson.D{{Key: "email", Value: email}, {Key: "is_verified", Value: true}}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, model.NewError(model.NotFound, "users not found")
		}

		ar.Logger.Error("AuthRepositoryImpl.FindByEmail FindOne ERROR, ", err)
		return nil, err
	}

	return &user, nil
}

func (ar *AuthRepositoryImpl) SetVerificationByEmail(ctx context.Context, email string, code string, duration time.Duration, verificationType model.VerificationType) error {
	tr := ar.Tracer.Tracer("Auth-SetVerificationByEmail repository")
	ctx, span := tr.Start(ctx, "Start SetVerificationByEmail")
	defer span.End()

	err := ar.Redis.SetEx(ctx, fmt.Sprintf(model.VerificationKey, verificationType, code), email, duration).Err()
	if err != nil {
		ar.Logger.Error("AuthRepositoryImpl.SetVerificationByEmail SetEx ERROR, ", err)

		return err
	}

	return nil
}

func (ar *AuthRepositoryImpl) GetVerificationByCode(ctx context.Context, code string, verificationType model.VerificationType) (string, error) {
	tr := ar.Tracer.Tracer("Auth-GetVerificationByCode repository")
	ctx, span := tr.Start(ctx, "Start GetVerificationByCode")
	defer span.End()

	result := ar.Redis.Get(ctx, fmt.Sprintf(model.VerificationKey, verificationType, code))
	if result.Err() != nil {
		ar.Logger.Error("AuthRepositoryImpl.GetVerificationByCode Get ERROR, ", result.Err())

		return "", result.Err()
	}

	return result.Val(), nil
}

func (ar *AuthRepositoryImpl) UpdateVerifyStatusByEmail(ctx context.Context, email string) error {
	tr := ar.Tracer.Tracer("Auth-UpdateVerifyStatusByEmail repository")
	ctx, span := tr.Start(ctx, "Start UpdateVerifyStatusByEmail")
	defer span.End()

	_, err := ar.DB.Collection(ar.Config.Database.UsersCollection).UpdateOne(ctx,
		bson.D{{Key: "email", Value: email}}, bson.M{
			"$set": bson.D{{Key: "is_verified", Value: true}},
		})
	if err != nil {
		ar.Logger.Error("AuthRepositoryImpl.UpdateVerifyStatusByEmail UpdateOne ERROR, ", err)
		return err
	}

	return nil
}

func (ar *AuthRepositoryImpl) UpdatePasswordByEmail(ctx context.Context, email string, newPassword string) error {
	tr := ar.Tracer.Tracer("Auth-UpdatePasswordByEmail repository")
	ctx, span := tr.Start(ctx, "Start UpdatePasswordByEmail")
	defer span.End()

	_, err := ar.DB.Collection(ar.Config.Database.UsersCollection).UpdateOne(ctx,
		bson.D{{Key: "email", Value: email}}, bson.M{
			"$set": bson.D{{Key: "password", Value: newPassword}},
		})
	if err != nil {
		ar.Logger.Error("AuthRepositoryImpl.UpdatePasswordByEmail UpdateOne ERROR, ", err)
		return err
	}

	return nil
}
