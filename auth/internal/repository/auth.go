package repository

import (
	"context"
	"fmt"
	"time"

	"singkatin-api/auth/internal/config"
	"singkatin-api/auth/internal/model"
	"singkatin-api/auth/pkg/logger"

	"github.com/redis/go-redis/v9"
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

	// authRepositoryImpl is an app auth struct that consists of all the dependencies needed for auth repository
	authRepositoryImpl struct {
		Config  *config.Config
		Tracer  *trace.TracerProvider
		DB      *mongo.Database
		Redis   *redis.Client
	}
)

// NewAuthRepository return new instances auth repository
func NewAuthRepository(config *config.Config, tracer *trace.TracerProvider, db *mongo.Database, rds *redis.Client) AuthRepository {
	return &authRepositoryImpl{
		Config:  config,
		Tracer:  tracer,
		DB:      db,
		Redis:   rds,
	}
}

func (r *authRepositoryImpl) CreateUser(ctx context.Context, req *model.User) (*model.User, error) {
	tr := r.Tracer.Tracer("Auth-CreateUser repository")
	ctx, span := tr.Start(ctx, "Start CreateUser")
	defer span.End()

	// check data users by email is already exists or not
	err := r.DB.Collection(r.Config.Database.UsersCollection).FindOne(ctx, bson.D{{Key: "email", Value: req.Email}}).Err()
	if err != nil {
		// if doc not exists, create new one
		if err == mongo.ErrNoDocuments {
			res, err := r.DB.Collection(r.Config.Database.UsersCollection).InsertOne(ctx, model.User{
				FullName:   req.FullName,
				Email:      req.Email,
				Password:   req.Password,
				CreatedAt:  time.Now(),
				IsVerified: false,
			})
			if err != nil {
				logger.Errorf("AuthRepositoryImpl.CreateUser InsertOne ERROR, %v", err)
				return nil, err
			}

			id, ok := res.InsertedID.(primitive.ObjectID)
			if !ok {
				logger.Errorf("AuthRepositoryImpl.CreateUser Type Assertion ERROR, %v", err)
				return nil, model.NewError(model.Type, "type assertion error")
			}
			req.ID = id

			return req, nil
		}

		logger.Errorf("AuthRepositoryImpl.CreateUser FindOne ERROR, %v", err)
		return nil, err
	}

	return nil, model.NewError(model.Validation, "email already exists")
}

func (r *authRepositoryImpl) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	tr := r.Tracer.Tracer("Auth-FindByEmail repository")
	ctx, span := tr.Start(ctx, "Start FindByEmail")
	defer span.End()

	user := model.User{}

	err := r.DB.Collection(r.Config.Database.UsersCollection).FindOne(ctx, bson.D{{Key: "email", Value: email}, {Key: "is_verified", Value: true}}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, model.NewError(model.NotFound, "users not found")
		}

		logger.Errorf("AuthRepositoryImpl.FindByEmail FindOne ERROR, %v", err)
		return nil, err
	}

	return &user, nil
}

func (r *authRepositoryImpl) SetVerificationByEmail(ctx context.Context, email string, code string, duration time.Duration, verificationType model.VerificationType) error {
	tr := r.Tracer.Tracer("Auth-SetVerificationByEmail repository")
	ctx, span := tr.Start(ctx, "Start SetVerificationByEmail")
	defer span.End()

	err := r.Redis.SetEx(ctx, fmt.Sprintf(model.VerificationKey, verificationType, code), email, duration).Err()
	if err != nil {
		logger.Errorf("AuthRepositoryImpl.SetVerificationByEmail SetEx ERROR, %v", err)

		return err
	}

	return nil
}

func (r *authRepositoryImpl) GetVerificationByCode(ctx context.Context, code string, verificationType model.VerificationType) (string, error) {
	tr := r.Tracer.Tracer("Auth-GetVerificationByCode repository")
	ctx, span := tr.Start(ctx, "Start GetVerificationByCode")
	defer span.End()

	result := r.Redis.Get(ctx, fmt.Sprintf(model.VerificationKey, verificationType, code))
	if result.Err() != nil {
		logger.Errorf("AuthRepositoryImpl.GetVerificationByCode Get ERROR, %v", result.Err())

		return "", result.Err()
	}

	return result.Val(), nil
}

func (r *authRepositoryImpl) UpdateVerifyStatusByEmail(ctx context.Context, email string) error {
	tr := r.Tracer.Tracer("Auth-UpdateVerifyStatusByEmail repository")
	ctx, span := tr.Start(ctx, "Start UpdateVerifyStatusByEmail")
	defer span.End()

	_, err := r.DB.Collection(r.Config.Database.UsersCollection).UpdateOne(ctx,
		bson.D{{Key: "email", Value: email}}, bson.M{
			"$set": bson.D{{Key: "is_verified", Value: true}},
		})
	if err != nil {
		logger.Errorf("AuthRepositoryImpl.UpdateVerifyStatusByEmail UpdateOne ERROR, %v", err)
		return err
	}

	return nil
}

func (r *authRepositoryImpl) UpdatePasswordByEmail(ctx context.Context, email string, newPassword string) error {
	tr := r.Tracer.Tracer("Auth-UpdatePasswordByEmail repository")
	ctx, span := tr.Start(ctx, "Start UpdatePasswordByEmail")
	defer span.End()

	_, err := r.DB.Collection(r.Config.Database.UsersCollection).UpdateOne(ctx,
		bson.D{{Key: "email", Value: email}}, bson.M{
			"$set": bson.D{{Key: "password", Value: newPassword}},
		})
	if err != nil {
		logger.Errorf("AuthRepositoryImpl.UpdatePasswordByEmail UpdateOne ERROR, %v", err)
		return err
	}

	return nil
}
