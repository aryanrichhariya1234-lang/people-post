package services

import (
	"context"

	"peoplepost/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetUserByID(userID string) (bson.M, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var user bson.M
	err = config.DB.Collection("users").FindOne(
		context.Background(),
		bson.M{"_id": objID},
	).Decode(&user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func UpdateUser(userID string, updates map[string]interface{}) (bson.M, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	update := bson.M{"$set": updates}

	var updated bson.M
	err = config.DB.Collection("users").FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": objID},
		update,
	).Decode(&updated)

	if err != nil {
		return nil, err
	}

	return updated, nil
}

func DeleteUser(userID string) error {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	_, err = config.DB.Collection("users").DeleteOne(
		context.Background(),
		bson.M{"_id": objID},
	)

	return err
}