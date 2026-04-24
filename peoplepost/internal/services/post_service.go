// package services

// import (
// 	"context"
// 	"time"

// 	"peoplepost/internal/config"

// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// )

// func CreatePost(userID string, data map[string]interface{}) (interface{}, error) {
// 	objID, err := primitive.ObjectIDFromHex(userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	post := bson.M{
// 		"user":        objID,
// 		"category":    data["category"],
// 		"Address":     data["Address"],
// 		"description": data["description"],
// 		"images":      data["images"],
// 		"location":    data["location"],
// 		"createdAt":   time.Now(),
// 	}

// 	res, err := config.DB.Collection("posts").InsertOne(context.Background(), post)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return res.InsertedID, nil
// }

// // func GetAllPosts() ([]bson.M, error) {
// // 	cursor, err := config.DB.Collection("posts").Find(context.Background(), bson.M{})
// // 	if err != nil {
// // 		return nil, err
// // 	}

// // 	var posts []bson.M
// // 	err = cursor.All(context.Background(), &posts)
// // 	if err != nil {
// // 		return nil, err
// // 	}

// // 	return posts, nil
// // }

// func UpdatePost(postID string, updates map[string]interface{}) error {
// 	objID, err := primitive.ObjectIDFromHex(postID)
// 	if err != nil {
// 		return err
// 	}

// 	update := bson.M{"$set": updates}

// 	_, err = config.DB.Collection("posts").UpdateOne(
// 		context.Background(),
// 		bson.M{"_id": objID},
// 		update,
// 	)

// 	return err
// }

// func DeletePost(postID string) error {
// 	objID, err := primitive.ObjectIDFromHex(postID)
// 	if err != nil {
// 		return err
// 	}

// 	_, err = config.DB.Collection("posts").DeleteOne(
// 		context.Background(),
// 		bson.M{"_id": objID},
// 	)

// 	return err
// }