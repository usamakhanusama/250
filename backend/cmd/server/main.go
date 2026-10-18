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
	"sust-cse-portal/internal/result"
	"sust-cse-portal/internal/routine"
	"sust-cse-portal/internal/student"
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
	var authHandler *auth.Handler
	if database != nil {
		authService := auth.NewService(database, cfg)
		authHandler = auth.NewHandler(authService)
	}

	// ============================================================================
	// AFIF'S GO SERVICES (50% Backend - 3 Packages):
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

	// Usama's Go Services (50% Backend): Student Identity & Result Engine
	studentService := student.NewService()
	studentHandler := student.NewHandler(studentService)

	resultService := result.NewService()
	resultHandler := result.NewHandler(resultService)

	// 4. Rate Limiters (Written from scratch with token bucket & cleanup)
	authRateLimiter := customMiddleware.NewIPRateLimiter(cfg.RateLimitAuthPerMin, 5) // 5 req/min
	apiRateLimiter := customMiddleware.NewIPRateLimiter(cfg.RateLimitAPIPerMin, 60)   // 60 req/min

	// 5. Setup Chi Router & Global Middlewares
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(customMiddleware.SecurityHeadersMiddleware)
	r.Use(customMiddleware.BodyLimitMiddleware(1024 * 1024)) // 1MB payload limit

	// CORS Setup
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", cfg.FrontendURL},
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
		// AFIF'S API DOMAIN (50% Go Backend Ownership: routine, course, attendance)
		// ============================================================================
		api.Route("/routines", func(routineRouter chi.Router) {
			routineRouter.Get("/", routineHandler.GetDayRoutine)
			routineRouter.Get("/weekly", routineHandler.GetWeeklyGrid)
			routineRouter.Get("/current", routineHandler.GetCurrentClass)
			routineRouter.Get("/rooms", routineHandler.GetRoomAvailability)
		})

		api.Route("/courses", func(courseRouter chi.Router) {
			courseRouter.Get("/", courseHandler.GetCourses)
			courseRouter.Get("/analytics", courseHandler.GetAnalytics)
			courseRouter.Post("/prerequisites/check", courseHandler.CheckPrerequisites)
			courseRouter.Get("/{code}", courseHandler.GetCourseByCode)
		})

		api.Route("/attendance", func(attRouter chi.Router) {
			attRouter.Get("/", attendanceHandler.GetStudentAttendance)
			attRouter.Get("/safebunk", attendanceHandler.AnalyzeSafeBunk)
		})

		// Usama's API Domain: Student Directory & Academic Results
		api.Route("/students", func(studentRouter chi.Router) {
			studentRouter.Get("/", studentHandler.GetAllStudents)
			studentRouter.Get("/analytics", studentHandler.GetAnalytics)
			studentRouter.Get("/{reg}", studentHandler.GetStudentProfile)
		})

		api.Route("/results", func(resultRouter chi.Router) {
			resultRouter.Get("/transcript/{reg}", resultHandler.GetTranscript)
			resultRouter.Get("/simulate", resultHandler.SimulateTargetGPA)
		})

		// Auth Routes
		api.Route("/auth", func(authRouter chi.Router) {
			if authHandler == nil {
				authRouter.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
					utils.Error(w, http.StatusServiceUnavailable, "Database is not connected. Auth unavailable.", nil)
				})
				return
			}

			// Public auth endpoints with STRICT rate limit (5 req/min)
			authRouter.Group(func(public chi.Router) {
				public.Use(customMiddleware.RateLimitMiddleware(authRateLimiter))
				public.Post("/login", authHandler.Login)
				public.Post("/refresh", authHandler.Refresh)
				public.Post("/logout", authHandler.Logout)
			})

			// Protected auth endpoints
			authRouter.Group(func(protected chi.Router) {
				protected.Use(customMiddleware.AuthMiddleware(cfg))

				// Current user profile
				protected.Get("/me", authHandler.Me)
				protected.Post("/change-password", authHandler.ChangePassword)

				// Admin-only user registration
				protected.With(customMiddleware.RequireRoles("admin")).Post("/register", authHandler.Register)
			})
		})
	})

	// 7. Start HTTP Server with Graceful Shutdown
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🚀 Server running on port %s (http://localhost:%s)", cfg.Port, cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("❌ Server listen error: %v", err)
		}
	}()

	// Wait for terminate signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("❌ Forced server shutdown: %v", err)
	}

	log.Println("✅ Server exited cleanly.")
}
