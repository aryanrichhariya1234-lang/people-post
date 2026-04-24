package services

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ToggleLike(existingLikes []interface{}, userID string) ([]interface{}, bool) {
	var updatedLikes []interface{}
	alreadyLiked := false

	for _, like := range existingLikes {
		likeMap := like.(bson.M)
		likeUser := likeMap["user"].(primitive.ObjectID).Hex()

		if likeUser == userID {
			alreadyLiked = true
			continue
		}

		updatedLikes = append(updatedLikes, like)
	}

	if !alreadyLiked {
		uid, _ := primitive.ObjectIDFromHex(userID)
		updatedLikes = append(updatedLikes, bson.M{"user": uid})
	}

	return updatedLikes, !alreadyLiked
}