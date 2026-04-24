package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"peoplepost/internal/config"
	"peoplepost/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type signupRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"passwordConfirm"`
	GovernmentID    string `json:"governmentId"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ================= SIGNUP =================
func SignUp(w http.ResponseWriter, r *http.Request) {
	var body signupRequest
	json.NewDecoder(r.Body).Decode(&body)

	// ✅ validation
	if body.Name == "" || body.Email == "" || body.Password == "" {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "fail",
			"message": "Name, email and password are required",
		})
		return
	}

	if body.Password != body.PasswordConfirm {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "fail",
			"message": "Passwords do not match",
		})
		return
	}

	// 🔥 role logic
	role := "citizen"
	if body.GovernmentID != "" {
		role = "official"
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(body.Password), 12)

	user := bson.M{
		"name":         body.Name,
		"email":        body.Email,
		"password":     string(hashedPassword),
		"role":         role,
		"governmentId": body.GovernmentID,
		"createdAt":    time.Now(),
	}

	collection := config.DB.Collection("users")

	res, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "User creation failed",
		})
		return
	}

	id := res.InsertedID.(primitive.ObjectID)
	token := generateToken(id.Hex())

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Secure:   false,
	
		Path:     "/",
	})

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"token":  token,
		"user": map[string]interface{}{
			"id":    id.Hex(),
			"name":  body.Name,
			"email": body.Email,
			"role":  role,
		},
	})
}

// ================= LOGIN =================
func Login(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	json.NewDecoder(r.Body).Decode(&body)

	if body.Email == "" || body.Password == "" {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "fail",
			"message": "Please enter email and password",
		})
		return
	}

	collection := config.DB.Collection("users")

	var user bson.M
	err := collection.FindOne(context.Background(), bson.M{"email": body.Email}).Decode(&user)
	if err != nil {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"status":  "fail",
			"message": "User not found",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user["password"].(string)), []byte(body.Password))
	if err != nil {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"status":  "fail",
			"message": "Incorrect password",
		})
		return
	}

	id := user["_id"].(primitive.ObjectID)
	token := generateToken(id.Hex())

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Secure:   false,
	
		Path:     "/",
	})

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"token":  token,
	})
}

// ================= UPDATE PASSWORD =================
func UpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	objID, _ := primitive.ObjectIDFromHex(userID)

	var body struct {
		CurrentPassword string `json:"currentPassword"`
		Password        string `json:"password"`
		PasswordConfirm string `json:"passwordConfirm"`
	}

	json.NewDecoder(r.Body).Decode(&body)

	collection := config.DB.Collection("users")

	var user bson.M
	err := collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "fail",
			"message": "User not found",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user["password"].(string)), []byte(body.CurrentPassword))
	if err != nil {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"status":  "fail",
			"message": "Incorrect current password",
		})
		return
	}

	if body.Password != body.PasswordConfirm {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "fail",
			"message": "Passwords do not match",
		})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(body.Password), 12)

	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"password": string(hashedPassword)}},
	)

	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Update failed",
		})
		return
	}

	token := generateToken(userID)

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Secure:   false,
	
		Path:     "/",
	})

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Password updated",
	})
}

// ================= LOGOUT =================
func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",

		// 🔥 MUST MATCH LOGIN COOKIE
		HttpOnly: true,
		Secure:   false,
	

		// expire immediately
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Logged out successfully",
	})
}

// ================= FORGOT PASSWORD =================
func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	collection := config.DB.Collection("users")

	var user bson.M
	err := collection.FindOne(context.Background(), bson.M{"email": body.Email}).Decode(&user)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "fail",
			"message": "No user found with that email",
		})
		return
	}

	tokenData, _ := utils.GenerateResetToken()
	objID := user["_id"].(primitive.ObjectID)

	_, _ = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{
			"passwordResetToken": tokenData.HashedToken,
			"passwordResetTime":  tokenData.ExpiresAt,
		}},
	)

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Token sent",
	})
}

// ================= RESET PASSWORD =================
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	var body struct {
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])

	collection := config.DB.Collection("users")

	var user bson.M
	err := collection.FindOne(context.Background(), bson.M{
		"passwordResetToken": hashedToken,
	}).Decode(&user)

	if err != nil {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "fail",
			"message": "Invalid token",
		})
		return
	}

	expiry := user["passwordResetTime"].(primitive.DateTime).Time()
	if time.Now().After(expiry) {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "fail",
			"message": "Token expired",
		})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(body.Password), 12)

	objID := user["_id"].(primitive.ObjectID)

	_, _ = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"password": string(hashedPassword)}},
	)

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"token":  generateToken(objID.Hex()),
	})
}

// ================= JWT =================
func generateToken(id string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"exp": time.Now().Add(10 * 24 * time.Hour).Unix(),
	})

	tokenStr, _ := token.SignedString([]byte(os.Getenv("SECRET")))
	return tokenStr
}