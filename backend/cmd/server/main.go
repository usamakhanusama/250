package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sust-cse-portal/internal/attendance"
	"sust-cse-portal/internal/auth"
	"sust-cse-portal/internal/config"
	"sust-cse-portal/internal/course"
	"sust-cse-portal/internal/db"
	customMiddleware "sust-cse-portal/internal/middleware"
	"sust-cse-portal/internal/routine"
	"sust-cse-portal/internal/utils"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	log.Println("🎓 Starting SUST CSE 23/24 Academic Management System API...")

	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// 2. Connect to PostgreSQL
	database, err := db.Connect(cfg)
	if err != nil {
		log.Printf("⚠️ Warning: PostgreSQL connection failed (%v). Server will start in degraded mode.", err)
	} else {
		defer database.Close()

		// Run Database Migrations
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		migrationsDir := "migrations"
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			migrationsDir = "../../migrations"
		}

		if err := database.RunMigrations(ctx, migrationsDir); err != nil {
			log.Printf("⚠️ Migration warning: %v", err)
		}
	}

	// 3. Initialize Services and Handlers
	// Auth storage: PostgreSQL if connected, otherwise an in-memory store
	// (so signup/login still work in a local demo without a database).
	var authStore auth.Store
	if database != nil {
		authStore = auth.NewPostgresStore(database)
	} else {
		log.Println("⚠️ Using IN-MEMORY auth store (accounts are lost on restart). Demo users: 2023331020 / 2023331088, password Password@123")
		memStore := auth.NewMemoryStore()
		auth.SeedDemoStudents(context.Background(), memStore)
		authStore = memStore
	}
	authHandler := auth.NewHandler(auth.NewService(authStore, cfg))
	requireJWT := customMiddleware.AuthMiddleware(cfg)

	// ============================================================================
	// AFIF'S GO SERVICES (Authored by Afif Zaman Khan - 2023331020):
	// 1. Routine Engine (Timetable, Room Availability, Live Class Tracker)
	// 2. Course Catalog (Analytics, Prerequisites, Evaluation Schemes)
	// 3. Attendance Engine (SUST 75% Ordinance & Safe-Bunk Calculator)
	// ============================================================================
	routineService := routine.NewService()
	routineHandler := routine.NewHandler(routineService)

	courseService := course.NewService()
	courseHandler := course.NewHandler(courseService)

	attendanceService := attendance.NewService()
	attendanceHandler := attendance.NewHandler(attendanceService)

	// 4. Rate Limiters (token bucket written from scratch, see middleware/ratelimit.go)
	//    auth: 5 requests/min per IP (slows down password guessing & fake signups)
	//    api:  60 requests/min per IP (protects the server from request floods)
	authRateLimiter := customMiddleware.NewIPRateLimiter(cfg.RateLimitAuthPerMin, 5)
	apiRateLimiter := customMiddleware.NewIPRateLimiter(cfg.RateLimitAPIPerMin, 60)

	// 5. Setup Chi Router & Global Middlewares
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	// NOTE: chiMiddleware.RealIP is intentionally NOT used: it copies the client-controlled
	// X-Forwarded-For header into RemoteAddr, which would let attackers bypass the rate limiter.
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(customMiddleware.SecurityHeadersMiddleware)
	r.Use(customMiddleware.BodyLimitMiddleware(1024 * 1024)) // 1MB payload limit

	// CORS Setup
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001", cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "Retry-After"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// 6. Register API Routes
	r.Route("/api/v1", func(api chi.Router) {
		// Apply general API rate limiter (60 req/min)
		api.Use(customMiddleware.RateLimitMiddleware(apiRateLimiter))

		// Health Check
		api.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			dbStatus := "connected"
			if database == nil {
				dbStatus = "disconnected"
			}
			utils.Success(w, http.StatusOK, "SUST CSE 23/24 API is operational", map[string]interface{}{
				"status":    "healthy",
				"database":  dbStatus,
				"timestamp": time.Now().Format(time.RFC3339),
				"version":   "1.0.0",
			})
		})

		// ============================================================================
		// AFIF'S API DOMAIN (Authored by Afif Zaman Khan - 2023331020)
		// ============================================================================
		api.Route("/routines", func(routineRouter chi.Router) {
			routineRouter.Get("/", routineHandler.GetDayRoutine)
			routineRouter.Get("/weekly", routineHandler.GetWeeklyGrid)
			routineRouter.Get("/current", routineHandler.GetCurrentClass)
			routineRouter.Get("/rooms", routineHandler.GetRoomAvailability)
			routineRouter.Get("/check-room", routineHandler.CheckRoomConflict)
		})

		api.Route("/courses", func(courseRouter chi.Router) {
			courseRouter.Get("/", courseHandler.GetCourses)
			courseRouter.Get("/analytics", courseHandler.GetAnalytics)
			courseRouter.Post("/prerequisites/check", courseHandler.CheckPrerequisites)
			courseRouter.Get("/details/{code}", courseHandler.GetCourseDetails)
			courseRouter.Get("/assignments", courseHandler.GetAssignments)
			// Submitting work requires a logged-in student (JWT + role check)
			courseRouter.With(requireJWT, customMiddleware.RequireRoles("student")).
				Post("/assignments/submit", courseHandler.SubmitAssignment)
			courseRouter.Get("/{code}", courseHandler.GetCourseByCode)
		})

		// Attendance is personal data: JWT required, and the student is read from the token
		api.Route("/attendance", func(attRouter chi.Router) {
			attRouter.Use(requireJWT, customMiddleware.RequireRoles("student"))
			attRouter.Get("/", attendanceHandler.GetStudentAttendance)
			attRouter.Get("/safebunk", attendanceHandler.AnalyzeSafeBunk)
		})

		// Auth Routes
		api.Route("/auth", func(authRouter chi.Router) {
			// Public endpoints: protected by the strict auth rate limiter
			authRouter.Group(func(public chi.Router) {
				public.Use(customMiddleware.RateLimitMiddleware(authRateLimiter))
				public.Post("/signup", authHandler.Signup)
				public.Post("/login", authHandler.Login)
				public.Post("/refresh", authHandler.Refresh)
				public.Post("/logout", authHandler.Logout)
			})

			// Private endpoints: need "Authorization: Bearer <access_token>"
			authRouter.Group(func(private chi.Router) {
				private.Use(requireJWT)
				private.Get("/me", authHandler.Me)
				private.Post("/change-password", authHandler.ChangePassword)
			})
		})
	})

	// 7. Start HTTP Server with Graceful Shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🚀 Server running on port %s (http://localhost:%s)", cfg.Port, cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("❌ HTTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server gracefully stopped. Goodbye!")
}
