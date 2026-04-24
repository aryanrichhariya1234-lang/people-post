package routes

import (
	"net/http"
	"strings"

	"peoplepost/internal/handlers"
	"peoplepost/internal/middleware"
)

func SetupRouter() http.Handler {
	mux := http.NewServeMux()

	// ================= ROOT =================
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"success","message":"API running"}`))
	})

	// ================= USERS =================
	mux.HandleFunc("/api/v1/users/me", middleware.Protect(handlers.GetMe))
	mux.HandleFunc("/api/v1/users/image", middleware.Protect(handlers.StoreMetaDataUser))
	mux.HandleFunc("/api/v1/users/forgotPassword", handlers.ForgotPassword)
	mux.HandleFunc("/api/v1/users/updateMe", middleware.Protect(handlers.UpdateMe))
	mux.HandleFunc("/api/v1/users/updatePassword", middleware.Protect(handlers.UpdatePassword))
	mux.HandleFunc("/api/v1/users/signup", handlers.SignUp)
	mux.HandleFunc("/api/v1/users/logout", handlers.Logout)
	mux.HandleFunc("/api/v1/users/login", handlers.Login)
	mux.HandleFunc("/api/v1/users/deleteMe", middleware.Protect(handlers.DeleteMe))
	mux.HandleFunc("/api/v1/users/resetPassword", handlers.ResetPassword)

	// ================= AI =================
	mux.HandleFunc("/api/v1/ai/insights", middleware.Protect(handlers.GetDashboardInsights))

	// ================= POSTS (RESTFUL) =================

	// GET all posts + CREATE post
	mux.HandleFunc("/api/v1/posts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetAllPosts(w, r)

		case http.MethodPost:
			middleware.Protect(handlers.CreatePost)(w, r)

		default:
			http.NotFound(w, r)
		}
	})

	// UPDATE / DELETE / LIKE (with ID)
	mux.HandleFunc("/api/v1/posts/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")

		// /:id/like
		if strings.HasSuffix(path, "/like") && r.Method == http.MethodPost {
			middleware.Protect(handlers.ToggleLike)(w, r)
			return
		}

		// /:id
		switch r.Method {
		case http.MethodPatch:
			middleware.Protect(handlers.UpdatePost)(w, r)

		case http.MethodDelete:
			middleware.Protect(handlers.DeletePost)(w, r)

		default:
			http.NotFound(w, r)
		}
	})

	// ================= CORS =================
	return middleware.CORS(mux)
}