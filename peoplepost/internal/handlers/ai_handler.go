package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"peoplepost/internal/cache"
	"peoplepost/internal/config"
	"peoplepost/pkg/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetDashboardInsights(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	cacheKey := "dashboard:insights"

	var cached map[string]interface{}
	if err := cache.Get(cacheKey, &cached); err == nil {
		utils.JSON(w, http.StatusOK, cached)
		return
	}

	postCollection := config.DB.Collection("posts")

	cityStats, _ := aggregate(postCollection, "$city")
	categoryStats, _ := aggregate(postCollection, "$category")
	userStats, _ := aggregate(postCollection, "$user")

	now := time.Now()
	last7Days := now.AddDate(0, 0, -7)
	prev7Days := now.AddDate(0, 0, -14)

	currentWeek, _ := postCollection.CountDocuments(ctx, bson.M{
		"createdAt": bson.M{"$gte": last7Days},
	})

	previousWeek, _ := postCollection.CountDocuments(ctx, bson.M{
		"createdAt": bson.M{
			"$gte": prev7Days,
			"$lt":  last7Days,
		},
	})

	trend := 100.0
	if previousWeek != 0 {
		trend = float64(currentWeek-previousWeek) / float64(previousWeek) * 100
	}

	totalReports, _ := postCollection.CountDocuments(ctx, bson.M{})
	estimatedImpact := totalReports * 50

	priority := "LOW"
	if trend > 30 || totalReports > 20 {
		priority = "HIGH"
	} else if trend > 10 {
		priority = "MEDIUM"
	}

	cityJSON, _ := json.Marshal(cityStats)
	categoryJSON, _ := json.Marshal(categoryStats)
	userJSON, _ := json.Marshal(userStats)

	prompt := fmt.Sprintf(`
You are a smart city decision intelligence system.

Give:
1. Most affected city
2. Most common issue
3. Most active reporter
4. Trend (with %%)
5. Priority level
6. Estimated people affected
7. Action recommendation

City Stats: %s
Category Stats: %s
User Stats: %s

Trend: %.2f%%
Total Reports: %d
Impact: %d
Priority: %s
`, cityJSON, categoryJSON, userJSON, trend, totalReports, estimatedImpact, priority)

	insights := "AI unavailable"

	responseText, err := callGemini(prompt)
	if err != nil {
		fmt.Println("Gemini error:", err)
	} else if responseText != "" {
		insights = responseText
	}

	result := map[string]interface{}{
		"status":   "success",
		"insights": insights,
		"meta": map[string]interface{}{
			"totalReports":    totalReports,
			"trend":           trend,
			"priority":        priority,
			"estimatedImpact": estimatedImpact,
		},
	}

	_ = cache.Set(cacheKey, result, 15*time.Minute)

	utils.JSON(w, http.StatusOK, result)
}

func aggregate(collection *mongo.Collection, field string) ([]bson.M, error) {
	ctx := context.Background()

	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   field,
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"count": -1},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func callGemini(prompt string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")

	if apiKey == "" {
		return "", fmt.Errorf("missing GEMINI_API_KEY")
	}

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(
		"POST",
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-3-flash-preview:generateContent?key="+apiKey,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	candidates, ok := result["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return "", fmt.Errorf("no candidates returned")
	}

	content := candidates[0].(map[string]interface{})["content"].(map[string]interface{})
	parts := content["parts"].([]interface{})

	return parts[0].(map[string]interface{})["text"].(string), nil
}