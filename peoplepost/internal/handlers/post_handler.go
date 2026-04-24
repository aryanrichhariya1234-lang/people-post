package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"peoplepost/internal/cache"
	"peoplepost/internal/config"
	"peoplepost/pkg/utils"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type createPostRequest struct {
	category    string      `json:"category"`
	description string      `json:"description"`
	location    interface{} `json:"location"`
	images      []string    `json:"images"`
	status      string      `json:"status,omitempty"`
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "fail",
			"message": "Invalid form data",
		})
		return
	}

	category := r.FormValue("category")
	description := r.FormValue("description")
	locationStr := r.FormValue("location")

	// images
	files := r.MultipartForm.File["images"]

	if category == "" || description == "" || locationStr == "" || len(files) == 0 {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "fail",
			"message": "category, description, location and images are required",
		})
		return
	}

	// parse location JSON
	var location interface{}
	if err := json.Unmarshal([]byte(locationStr), &location); err != nil {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "fail",
			"message": "Invalid location format",
		})
		return
	}

	// 🔥 IMPORTANT: upload images (Cloudinary)
	var imageURLs []string

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}
		defer file.Close()

		uploadRes, err := config.Cloudinary.Upload.Upload(
			r.Context(),
			file,
			uploader.UploadParams{Folder: "posts"},
		)

		if err == nil {
			imageURLs = append(imageURLs, uploadRes.SecureURL)
		}
	}

	userID := r.Context().Value("userID").(string)
	objID, _ := primitive.ObjectIDFromHex(userID)

	post := bson.M{
		"user":        objID,
		"category":    category,
		"description": description,
		"images":      imageURLs,
		"location":    location,

		// defaults
		"status": "OPEN",
		"likes":  []interface{}{},

		"createdAt": time.Now(),
	}

	collection := config.DB.Collection("posts")

	res, err := collection.InsertOne(context.Background(), post)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to create post",
		})
		return
	}

	cache.Delete("posts")
	cache.Delete("dashboard:insights")

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"status": "success",
		"data":   res.InsertedID,
	})
}
func GetAllPosts(w http.ResponseWriter, r *http.Request) {
	var cached []bson.M

	if err := cache.Get("posts", &cached); err == nil && len(cached) > 0 {
		utils.JSON(w, http.StatusOK, map[string]interface{}{
			"status": "success",
			"data":   cached,
		})
		return
	}

	collection := config.DB.Collection("posts")

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to fetch posts",
		})
		return
	}

	var posts []bson.M
	_ = cursor.All(context.Background(), &posts)

	_ = cache.Set("posts", posts, 30*time.Minute)

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   posts,
	})
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	id := extractID(r)

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "fail",
			"message": "Invalid ID",
		})
		return
	}

	collection := config.DB.Collection("posts")

	// 🔥 check ownership
	var existing bson.M
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&existing)
	if err != nil {
		utils.JSON(w, http.StatusNotFound, map[string]interface{}{
			"status":  "fail",
			"message": "Post not found",
		})
		return
	}

	

	// 🔥 decode body
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "fail",
			"message": "Invalid body",
		})
		return
	}

	update := bson.M{}

	// ✅ STATUS VALIDATION
	if status, ok := body["status"].(string); ok {

		status = strings.ToUpper(status)

		validStatus := map[string]bool{
			"OPEN":        true,
			"IN_PROCESS": true,
			"RESOLVED":    true,
		}

		if !validStatus[status] {
			utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
				"status":  "fail",
				"message": "Invalid status",
			})
			return
		}

		update["status"] = status
	}

	if len(update) == 0 {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "fail",
			"message": "No valid fields",
		})
		return
	}

	// ✅ return updated doc
	var updatedDoc bson.M

	err = collection.FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": update},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updatedDoc)

	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Update failed",
		})
		return
	}

	cache.Delete("posts")
	cache.Delete("dashboard:insights")

	// ✅ MATCH FRONTEND
	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   updatedDoc,
	})
}
// ================= DELETE POST =================
func DeletePost(w http.ResponseWriter, r *http.Request) {
	id := extractID(r)
	objID, _ := primitive.ObjectIDFromHex(id)

	collection := config.DB.Collection("posts")

	var existing bson.M
	err := collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&existing)
	if err != nil {
		utils.JSON(w, http.StatusNotFound, map[string]interface{}{
			"msg": "Post not found",
		})
		return
	}

	userID := r.Context().Value("userID").(string)

	if existing["user"].(primitive.ObjectID).Hex() != userID {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"msg": "Unauthorized",
		})
		return
	}

	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": objID})
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"msg": "Delete failed",
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"msg": "Post removed successfully",
	})
}

// ================= TOGGLE LIKE =================
func ToggleLike(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")
	id := strings.Split(path, "/")[0]

	objID, _ := primitive.ObjectIDFromHex(id)

	collection := config.DB.Collection("posts")

	var post bson.M
	err := collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&post)
	if err != nil {
		utils.JSON(w, http.StatusNotFound, map[string]interface{}{
			"message": "Post not found",
		})
		return
	}

	userID := r.Context().Value("userID").(string)

	likes, _ := post["likes"].([]interface{})
	alreadyLiked := false
	var updatedLikes []interface{}

	for _, like := range likes {
		likeUser := like.(bson.M)["user"].(primitive.ObjectID).Hex()
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

	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"likes": updatedLikes}},
	)

	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Update failed",
		})
		return
	}

	cache.Delete("posts")
	cache.Delete("dashboard:insights")

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":     "success",
		"liked":      !alreadyLiked,
		"likesCount": len(updatedLikes),
	})
}

// ================= HELPER =================
func extractID(r *http.Request) string {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")
	return strings.Split(path, "/")[0]
}