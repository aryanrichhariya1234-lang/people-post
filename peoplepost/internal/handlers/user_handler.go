package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"peoplepost/internal/config"
	"peoplepost/pkg/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	objID, _ := primitive.ObjectIDFromHex(userID)

	userCollection := config.DB.Collection("users")
	postCollection := config.DB.Collection("posts")

	var user bson.M
	err := userCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		utils.JSON(w, http.StatusNotFound, map[string]interface{}{
			"status":  "fail",
			"message": "User not found",
		})
		return
	}

	cursor, err := postCollection.Find(context.Background(), bson.M{"user": objID})
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to fetch posts",
		})
		return
	}

	var posts []bson.M
	_ = cursor.All(context.Background(), &posts)

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"user":   user,
		"data":   posts,
	})
}

func UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	objID, _ := primitive.ObjectIDFromHex(userID)

	var body map[string]interface{}
	json.NewDecoder(r.Body).Decode(&body)

	update := bson.M{"$set": bson.M{}}

	if name, ok := body["name"]; ok {
		update["$set"].(bson.M)["name"] = name
	}
	if email, ok := body["email"]; ok {
		update["$set"].(bson.M)["email"] = email
	}

	collection := config.DB.Collection("users")

	var updated bson.M
	err := collection.FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": objID},
		update,
	).Decode(&updated)

	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Update failed",
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"update": updated,
	})
}

func DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	objID, _ := primitive.ObjectIDFromHex(userID)

	collection := config.DB.Collection("users")

	_, err := collection.DeleteOne(context.Background(), bson.M{"_id": objID})
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Delete failed",
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
	})
}