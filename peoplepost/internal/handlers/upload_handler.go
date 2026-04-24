package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"peoplepost/internal/config"
	"peoplepost/pkg/utils"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func UploadImage(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "fail",
			"message": "File required",
		})
		return
	}
	defer file.Close()

	result, err := config.Cloudinary.Upload.Upload(
		r.Context(),
		file,
		uploader.UploadParams{
			Folder: "posts",
		},
	)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Upload failed",
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":    "success",
		"url":       result.SecureURL,
		"public_id": result.PublicID,
	})
}

func StoreMetaDataUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL      string `json:"url"`
		PublicID string `json:"public_id"`
	}

	json.NewDecoder(r.Body).Decode(&body)

	if body.URL == "" || body.PublicID == "" {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "fail",
			"message": "Missing fields",
		})
		return
	}

	userID := r.Context().Value("userID").(string)
	objID, _ := primitive.ObjectIDFromHex(userID)

	collection := config.DB.Collection("users")

	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"photo":     body.URL,
				"public_id": body.PublicID,
			},
		},
	)

	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Update failed",
		})
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"status": "success",
	})
}

func StoreMetaDataPost(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL      string `json:"url"`
		PublicID string `json:"public_id"`
	}

	json.NewDecoder(r.Body).Decode(&body)

	if body.URL == "" || body.PublicID == "" {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "fail",
			"message": "Missing fields",
		})
		return
	}

	id := r.URL.Query().Get("id")
	objID, _ := primitive.ObjectIDFromHex(id)

	collection := config.DB.Collection("posts")

	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"url":       body.URL,
				"public_id": body.PublicID,
			},
		},
	)

	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Update failed",
		})
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"status": "success",
	})
}