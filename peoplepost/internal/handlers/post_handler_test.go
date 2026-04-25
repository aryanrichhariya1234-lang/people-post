package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"mime/multipart"
	"peoplepost/internal/config"

	"go.mongodb.org/mongo-driver/bson"
)


func setupTestDB() {
	config.LoadEnv()
	config.ConnectMongo()
	config.InitCloudinary() 
}

func mockAuthContext(req *http.Request) *http.Request {
	ctx := context.WithValue(req.Context(), "userID", "507f1f77bcf86cd799439011")
	return req.WithContext(ctx)
}


func TestGetAllPosts(t *testing.T) {
	setupTestDB()

	req, _ := http.NewRequest("GET", "/api/v1/posts", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetAllPosts)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}


func TestCreatePost_MissingFields(t *testing.T) {
	setupTestDB()

	body := map[string]interface{}{
		"category": "Road",
		// missing description, location
	}

	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	req = mockAuthContext(req)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(CreatePost)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing fields, got %d", rr.Code)
	}
}


func TestCreatePost_Success(t *testing.T) {
	setupTestDB()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// form fields
	writer.WriteField("category", "Road")
	writer.WriteField("Address", "Somewhere")
	writer.WriteField("description", "Big pothole")
	writer.WriteField("location", `{"lat":19.07,"lng":72.87}`)

	// fake image file
	part, _ := writer.CreateFormFile("images", "test.jpg")
	part.Write([]byte("fake image content"))

	writer.Close()

	req, _ := http.NewRequest("POST", "/api/v1/posts", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	req = mockAuthContext(req)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(CreatePost)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated && rr.Code != http.StatusOK {
		t.Errorf("Expected success, got %d", rr.Code)
	}
}


func TestProtectedRoute_Unauthorized(t *testing.T) {
	req, _ := http.NewRequest("POST", "/api/v1/posts", nil)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(CreatePost)
	handler.ServeHTTP(rr, req)

	if rr.Code == http.StatusOK {
		t.Errorf("Expected unauthorized error")
	}
}


func TestStatusValidation(t *testing.T) {
	validStatuses := []string{"OPEN", "IN_PROGRESS", "RESOLVED"}
	invalidStatus := "DONE"

	for _, status := range validStatuses {
		if !isValidStatus(status) {
			t.Errorf("Expected %s to be valid", status)
		}
	}

	if isValidStatus(invalidStatus) {
		t.Errorf("Expected invalid status to fail")
	}
}


func isValidStatus(status string) bool {
	switch status {
	case "OPEN", "IN_PROGRESS", "RESOLVED":
		return true
	}
	return false
}


func TestDBConnection(t *testing.T) {
	setupTestDB()

	collection := config.DB.Collection("posts")

	count, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		t.Errorf("DB connection failed: %v", err)
	}

	t.Logf("DB connected, total posts: %d", count)
}