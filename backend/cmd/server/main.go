package main

import (
	"log"
	"sts-backend/internal/config"
	"sts-backend/internal/database"
	"sts-backend/internal/handlers"
	"sts-backend/internal/services"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	database.Init(cfg)

	// Initialize auth handlers with email service
	handlers.InitAuthHandlers(cfg)

	// Initialize complaint service with escalation support
	services.InitComplaintService(database.DB, cfg)

	// Test Database connection
	err := database.DB.Ping()
	if err != nil {
		log.Printf("Database Connection Check: %v", err)
	} else {
		log.Println("Database Connection Check: Successfully connected!")
	}

	// Initialize escalation scheduler (runs every hour)
	escalationScheduler := services.NewEscalationScheduler(database.DB, cfg, 1*time.Hour)
	escalationScheduler.Start()
	log.Println("✅ Complaint escalation scheduler started (runs every 1 hour)")

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Basic health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API Routes
	api := r.Group("/api")
	{
		// Admin authentication routes
		adminAuth := api.Group("/admin/auth")
		{
			adminAuth.POST("/login", handlers.AdminLogin)
			adminAuth.POST("/logout", handlers.AdminLogout)
			adminAuth.POST("/refresh", handlers.RefreshAccessToken)
			adminAuth.GET("/profile", handlers.AdminProfile)
			adminAuth.GET("/bootstrap-status", handlers.GetBootstrapStatus)

			// Password reset routes
			adminAuth.POST("/forgot-password", handlers.RequestPasswordReset)
			adminAuth.POST("/verify-reset-token", handlers.VerifyResetToken)
			adminAuth.POST("/reset-password", handlers.ResetPassword)
		}

		// User management routes (super admin only)
		adminUsers := api.Group("/admin/users")
		adminUsers.Use(handlers.RequireSuperAdmin)
		{
			adminUsers.POST("", handlers.CreateUser)
			adminUsers.GET("", handlers.GetAllUsers)
			adminUsers.PUT("/:id", handlers.UpdateUser)
			adminUsers.DELETE("/:id", handlers.DeleteUser)
		}

		adminAccess := api.Group("/admin/access-controls")
		adminAccess.Use(handlers.RequireSuperAdmin)
		{
			adminAccess.GET("/admin-users/:id/permissions", handlers.GetAdminUserPermissions)
			adminAccess.PUT("/admin-users/:id/permissions", handlers.UpdateAdminUserPermissions)
			adminAccess.GET("/app-users/:id/permissions", handlers.GetAppUserPermissions)
			adminAccess.PUT("/app-users/:id/permissions", handlers.UpdateAppUserPermissions)
		}

		// Lounge routes
		api.GET("/lounges", handlers.RequireAdminPermission("lounges.read"), handlers.GetLounges)
		api.GET("/lounges/pending", handlers.RequireAdminPermission("lounges.verify"), handlers.GetPendingLounges)
		api.GET("/lounges/:id", handlers.RequireAdminPermission("lounges.read"), handlers.GetLoungeById)
		api.POST("/lounges", handlers.RequireAdminPermission("lounges.write"), handlers.CreateLounge)
		api.PUT("/lounges/:id", handlers.RequireAdminPermission("lounges.write"), handlers.UpdateLounge)
		api.PUT("/lounges/:id/verify", handlers.RequireAdminPermission("lounges.verify"), handlers.VerifyLounge)
		api.DELETE("/lounges/:id", handlers.RequireAdminPermission("lounges.write"), handlers.DeleteLounge)

		// Lounge Owner routes
		api.GET("/lounge-owners/pending", handlers.RequireAdminPermission("lounge_owners.verify"), handlers.GetPendingLoungeOwners)
		api.GET("/lounge-owners/:id", handlers.RequireAdminPermission("lounge_owners.read"), handlers.GetLoungeOwnerById)
		api.PUT("/lounge-owners/:id/verify", handlers.RequireAdminPermission("lounge_owners.verify"), handlers.VerifyLoungeOwner)

		// Bus routes
		api.GET("/buses", handlers.RequireAdminPermission("buses.read"), handlers.GetBuses)
		api.GET("/buses/pending", handlers.RequireAdminPermission("buses.verify"), handlers.GetPendingBuses)
		api.GET("/buses/:id", handlers.RequireAdminPermission("buses.read"), handlers.GetBusById)
		api.POST("/buses", handlers.RequireAdminPermission("buses.write"), handlers.CreateBus)
		api.PUT("/buses/:id", handlers.RequireAdminPermission("buses.write"), handlers.UpdateBus)
		api.PUT("/buses/:id/verify", handlers.RequireAdminPermission("buses.verify"), handlers.VerifyBus)

		// Bus Owner routes
		api.GET("/bus-owners", handlers.RequireAdminPermission("bus_owners.read"), handlers.GetBusOwners)
		api.GET("/bus-owners/pending", handlers.RequireAdminPermission("bus_owners.verify"), handlers.GetPendingBusOwners)
		api.GET("/bus-owners/:id", handlers.RequireAdminPermission("bus_owners.read"), handlers.GetBusOwnerById)
		api.POST("/bus-owners", handlers.RequireAdminPermission("bus_owners.write"), handlers.CreateBusOwner)
		api.PUT("/bus-owners/:id", handlers.RequireAdminPermission("bus_owners.write"), handlers.UpdateBusOwner)
		api.PUT("/bus-owners/:id/verify", handlers.RequireAdminPermission("bus_owners.verify"), handlers.VerifyBusOwner)
		api.DELETE("/bus-owners/:id", handlers.RequireAdminPermission("bus_owners.write"), handlers.DeleteBusOwner)

		// Driver routes
		api.GET("/drivers", handlers.RequireAdminPermission("drivers.read"), handlers.GetDrivers)
		api.GET("/drivers/pending", handlers.RequireAdminPermission("drivers.verify"), handlers.GetPendingDrivers)
		api.GET("/drivers/:id", handlers.RequireAdminPermission("drivers.read"), handlers.GetDriverById)
		api.POST("/drivers", handlers.RequireAdminPermission("drivers.write"), handlers.CreateDriver)
		api.PUT("/drivers/:id", handlers.RequireAdminPermission("drivers.write"), handlers.UpdateDriver)
		api.PUT("/drivers/:id/verify", handlers.RequireAdminPermission("drivers.verify"), handlers.VerifyDriver)

		// Conductor routes
		api.GET("/conductors", handlers.RequireAdminPermission("conductors.read"), handlers.GetConductors)
		api.GET("/conductors/pending", handlers.RequireAdminPermission("conductors.verify"), handlers.GetPendingConductors)
		api.GET("/conductors/:id", handlers.RequireAdminPermission("conductors.read"), handlers.GetConductorById)
		api.POST("/conductors", handlers.RequireAdminPermission("conductors.write"), handlers.CreateConductor)
		api.PUT("/conductors/:id", handlers.RequireAdminPermission("conductors.write"), handlers.UpdateConductor)
		api.PUT("/conductors/:id/verify", handlers.RequireAdminPermission("conductors.verify"), handlers.VerifyConductor)

		// Booking routes
		api.GET("/bookings", handlers.RequireAdminPermission("bookings.read"), handlers.GetBookings)
		api.GET("/bookings/status", handlers.RequireAdminPermission("bookings.read"), handlers.GetBookingsByStatus)
		api.GET("/bookings/search", handlers.RequireAdminPermission("bookings.read"), handlers.SearchBookings)
		api.POST("/bookings", handlers.RequireAdminPermission("bookings.write"), handlers.CreateBooking)
		api.GET("/bookings/:id", handlers.RequireAdminPermission("bookings.read"), handlers.GetBookingByID)
		api.PUT("/bookings/:id", handlers.RequireAdminPermission("bookings.write"), handlers.UpdateBooking)
		api.PUT("/bookings/:id/status", handlers.RequireAdminPermission("bookings.write"), handlers.UpdateBookingStatus)
		api.PUT("/bookings/:id/payment", handlers.RequireAdminPermission("bookings.write"), handlers.UpdatePaymentStatus)
		api.DELETE("/bookings/:id", handlers.RequireAdminPermission("bookings.write"), handlers.CancelBooking)

		// Lounge Booking routes
		api.GET("/lounge-bookings", handlers.RequireAdminPermission("lounge_bookings.read"), handlers.GetLoungeBookings)
		api.GET("/lounge-bookings/:id", handlers.RequireAdminPermission("lounge_bookings.read"), handlers.GetLoungeBookingByID)
		api.POST("/lounge-bookings", handlers.RequireAdminPermission("lounge_bookings.write"), handlers.CreateLoungeBooking)
		api.PUT("/lounge-bookings/:id", handlers.RequireAdminPermission("lounge_bookings.write"), handlers.UpdateLoungeBooking)
		api.PATCH("/lounge-bookings/:id/payment-status", handlers.RequireAdminPermission("lounge_bookings.write"), handlers.UpdateLoungeBookingPaymentStatus)
		api.PATCH("/lounge-bookings/:id/booking-status", handlers.RequireAdminPermission("lounge_bookings.write"), handlers.UpdateLoungeBookingStatus)
		api.DELETE("/lounge-bookings/:id", handlers.RequireAdminPermission("lounge_bookings.write"), handlers.DeleteLoungeBooking)

		// Complaint routes
		api.GET("/complaints", handlers.GetComplaints)
		api.GET("/complaints/:id", handlers.GetComplaintById)
		api.PUT("/complaints/:id/status", handlers.UpdateComplaintStatus)
		api.POST("/complaints/:id/escalate", handlers.RequireAdminPermission("complaints.escalate"), handlers.ManualEscalateComplaint)
		api.GET("/complaints/:id/escalation", handlers.GetComplaintEscalation)

		// Initialize escalation service for complaint handlers
		escalationSvc := services.NewEscalationService(database.DB, cfg)
		handlers.SetEscalationService(escalationSvc)

		// Escalation routes
		escalationHandler := handlers.NewEscalationHandler(database.DB, cfg)
		api.GET("/escalation/config", handlers.RequireAdminPermission("escalation.read"), escalationHandler.GetEscalationConfig)
		api.GET("/escalation/config/:category", handlers.RequireAdminPermission("escalation.read"), escalationHandler.GetEscalationConfigForCategory)
		api.GET("/escalation/complaint/:id", handlers.RequireAdminPermission("escalation.read"), escalationHandler.GetComplaintEscalation)
		api.GET("/escalation/complaint/:id/history", handlers.RequireAdminPermission("escalation.read"), escalationHandler.GetEscalationHistory)
		api.POST("/escalation/complaint/:id/escalate", handlers.RequireAdminPermission("escalation.manage"), escalationHandler.EscalateComplaint)
		api.POST("/escalation/complaint/:id/assign", handlers.RequireAdminPermission("escalation.manage"), escalationHandler.AssignComplaint)
		api.POST("/escalation/complaint/:id/initialize", handlers.RequireAdminPermission("escalation.manage"), escalationHandler.InitializeComplaintEscalation)
		api.GET("/escalation/stats", handlers.RequireAdminPermission("escalation.read"), escalationHandler.GetEscalationStats)

		// Placeholder endpoints used by frontend lounge TV widgets.
		api.GET("/tv/broadcasts", func(c *gin.Context) {
			c.JSON(200, gin.H{"items": []interface{}{}})
		})
		api.GET("/lounge-ads/lounge/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"items": []interface{}{}})
		})
		api.GET("/departures/lounge/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"items": []interface{}{}})
		})
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
