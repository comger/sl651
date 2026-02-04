package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sl651-platform/internal/config"
	"sl651-platform/internal/device"
	"sl651-platform/internal/forward"
	"sl651-platform/internal/heartbeat"
	"sl651-platform/internal/model"

	_ "sl651-platform/docs" // Import swagger docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title SL651 Platform API
// @version 1.0
// @description REST API for SL651 Device Management Platform
// @BasePath /api/v1
type Server struct {
	config           *config.Config
	deviceManager    *device.Manager
	forwardService   *forward.Service
	heartbeatManager *heartbeat.HeartbeatManager
	router           *gin.Engine
	server           *http.Server
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestLogger middleware to debug routing issues
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log results
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		log.Printf("[HTTP] %s | %3d | %13v | %15s | %-7s %s\n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
	}
}

func NewServer(cfg *config.Config, deviceManager *device.Manager, forwardService *forward.Service, heartbeatManager *heartbeat.HeartbeatManager) *Server {
	// Force Debug Mode for now to see verbose output
	gin.SetMode(gin.DebugMode)

	router := gin.New()
	router.Use(gin.Recovery())
	// Use our custom logger instead of/in addition to default
	router.Use(RequestLogger())
	router.Use(CORSMiddleware())

	s := &Server{
		config:           cfg,
		deviceManager:    deviceManager,
		forwardService:   forwardService,
		heartbeatManager: heartbeatManager,
		router:           router,
		server: &http.Server{
			Handler:      router,
			Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
	}

	s.setupRoutes()

	return s
}

func (s *Server) setupRoutes() {
	// Root route for connectivity check
	s.router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "SL651 Platform API Online"})
	})

	// Swagger route
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := s.router.Group("/api/v1")
	{
		// Global Data Endpoint
		api.GET("/data", s.getGlobalDeviceData)

		devices := api.Group("/devices")
		{
			devices.GET("", s.listDevices)
			devices.POST("", s.createDevice)
			devices.GET("/:id", s.getDevice)
			devices.PUT("/:id", s.updateDevice)
			devices.DELETE("/:id", s.deleteDevice)
			devices.GET("/:id/data", s.getDeviceData)
			devices.GET("/:id/data/latest", s.getLatestDeviceData)
		}

		statistics := api.Group("/statistics")
		{
			statistics.GET("/devices", s.getDeviceStatistics)
			statistics.GET("/forward", s.getForwardStatistics)
			statistics.GET("/online", s.getOnlineStatistics)
			statistics.GET("/trend", s.getTrendStatistics)
		}

		heartbeat := api.Group("/heartbeat")
		{
			heartbeat.GET("/status", s.getHeartbeatStatus)
			heartbeat.GET("/config", s.getHeartbeatConfig)
			heartbeat.PUT("/config", s.updateHeartbeatConfig)
		}

		health := api.Group("/health")
		{
			health.GET("", s.healthCheck)
		}
	}
}

func (s *Server) Start() {
	log.Printf("Starting HTTP server on %s", s.server.Addr)
	// Print all registered routes for debugging
	for _, route := range s.router.Routes() {
		log.Printf("Route: %s %s", route.Method, route.Path)
	}

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("Failed to shutdown server: %v", err)
	}
}

// ... handlers (getGlobalDeviceData, etc) ...

// @Summary Get global device data
// @Description Retrieve aggregated historical data for multiple devices or all devices
// @Tags data
// @Produce json
// @Param station_ids query string false "Comma-separated list of Station IDs"
// @Param limit query int false "Items per page"
// @Param offset query int false "Offset"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /data [get]
func (s *Server) getGlobalDeviceData(c *gin.Context) {
	limit := 100
	offset := 0
	var stationIDs []string

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	if ids := c.Query("station_ids"); ids != "" {
		stationIDs = strings.Split(ids, ",")
	}

	if len(stationIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":      0,
			"message":   "success",
			"data":      []interface{}{}, // Return empty list
			"timestamp": time.Now().UnixMilli(),
		})
		return
	}

	data, err := s.deviceManager.GetAllDeviceData(c.Request.Context(), stationIDs, limit, offset)
	if err != nil {
		log.Printf("[API] getGlobalDeviceData error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      500,
			"message":   err.Error(),
			"timestamp": time.Now().UnixMilli(),
		})
		return
	}

	log.Printf("[API] getGlobalDeviceData found %d records", len(data))

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "success",
		"data":      data,
		"timestamp": time.Now().UnixMilli(),
	})
}

// @Summary List devices
// @Description Get a list of all registered devices
// @Tags devices
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /devices [get]
func (s *Server) listDevices(c *gin.Context) {
	devices, err := s.deviceManager.ListDevices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      500,
			"message":   err.Error(),
			"timestamp": time.Now().UnixMilli(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "success",
		"data":      devices,
		"timestamp": time.Now().UnixMilli(),
	})
}

// @Summary Create device
// @Description Register a new device
// @Tags devices
// @Accept json
// @Produce json
// @Param device body object true "Device info"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /devices [post]
func (s *Server) createDevice(c *gin.Context) {
	var req struct {
		ID       string             `json:"id" binding:"required"`
		Name     string             `json:"name" binding:"required"`
		Protocol string             `json:"protocol" binding:"required"`
		Address  string             `json:"address" binding:"required"`
		Port     int                `json:"port" binding:"required"`
		Config   model.DeviceConfig `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      400,
			"message":   err.Error(),
			"timestamp": time.Now().UnixMilli(),
		})
		return
	}

	device := &model.Device{
		ID:         req.ID,
		Name:       req.Name,
		DeviceType: model.DeviceTypeRTU,
		Protocol:   model.Protocol(req.Protocol),
		Address:    req.Address,
		Port:       req.Port,
		Credentials: model.Credentials{
			AuthType: model.AuthTypeNone,
		},
		Status:   model.StatusOffline,
		LastSeen: time.Now(),
		Config:   req.Config,
	}

	if err := s.deviceManager.RegisterDevice(c.Request.Context(), device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      500,
			"message":   err.Error(),
			"timestamp": time.Now().UnixMilli(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "success",
		"data":      device,
		"timestamp": time.Now().UnixMilli(),
	})
}

// @Summary Get device
// @Description Get detailed information of a device
// @Tags devices
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /devices/{id} [get]
func (s *Server) getDevice(c *gin.Context) {
	id := c.Param("id")

	device, err := s.deviceManager.GetDevice(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":      404,
			"message":   "Device not found",
			"timestamp": time.Now().UnixMilli(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "success",
		"data":      device,
		"timestamp": time.Now().UnixMilli(),
	})
}

// @Summary Update device
// @Description Update device information
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Param device body object true "Device updates"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /devices/{id} [put]
func (s *Server) updateDevice(c *gin.Context) {
	id := c.Param("id")

	device, err := s.deviceManager.GetDevice(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":      404,
			"message":   "Device not found",
			"timestamp": time.Now().UnixMilli(),
		})
		return
	}

	var req struct {
		Name   *string             `json:"name"`
		Config *model.DeviceConfig `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      400,
			"message":   err.Error(),
			"timestamp": time.Now().UnixMilli(),
		})
		return
	}

	if req.Name != nil {
		device.Name = *req.Name
	}

	if req.Config != nil {
		device.Config = *req.Config
	}

	if err := s.deviceManager.UpdateDevice(c.Request.Context(), device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      500,
			"message":   err.Error(),
			"timestamp": time.Now().UnixMilli(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "success",
		"data":      device,
		"timestamp": time.Now().UnixMilli(),
	})
}

// @Summary Delete device
// @Description Remove a device
// @Tags devices
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /devices/{id} [delete]
func (s *Server) deleteDevice(c *gin.Context) {
	id := c.Param("id")

	if err := s.deviceManager.DeleteDevice(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      500,
			"message":   err.Error(),
			"timestamp": time.Now().UnixMilli(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "success",
		"timestamp": time.Now().UnixMilli(),
	})
}

// @Summary Get device data
// @Description Retrieve historical data for a device
// @Tags devices
// @Produce json
// @Param id path string true "Device ID"
// @Param limit query int false "Items per page"
// @Param offset query int false "Offset"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /devices/{id}/data [get]
func (s *Server) getDeviceData(c *gin.Context) {
	id := c.Param("id")

	limit := 100
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	data, err := s.deviceManager.GetDeviceData(c.Request.Context(), id, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      500,
			"message":   err.Error(),
			"timestamp": time.Now().UnixMilli(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "success",
		"data":      data,
		"timestamp": time.Now().UnixMilli(),
	})
}

// @Summary Get latest device data
// @Description Retrieve the most recent data point for a device
// @Tags devices
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /devices/{id}/data/latest [get]
func (s *Server) getLatestDeviceData(c *gin.Context) {
	id := c.Param("id")

	data, err := s.deviceManager.GetLatestDeviceData(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":      404,
			"message":   "Device data not found",
			"timestamp": time.Now().UnixMilli(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "success",
		"data":      data,
		"timestamp": time.Now().UnixMilli(),
	})
}

// @Summary Get device statistics
// @Description Get platform-wide device statistics
// @Tags statistics
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /statistics/devices [get]
func (s *Server) getDeviceStatistics(c *gin.Context) {
	stats, err := s.deviceManager.GetDeviceStatistics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      500,
			"message":   err.Error(),
			"timestamp": time.Now().UnixMilli(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "success",
		"data":      stats,
		"timestamp": time.Now().UnixMilli(),
	})
}

// @Summary Get forward statistics
// @Description Get forwarding service statistics
// @Tags statistics
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /statistics/forward [get]
func (s *Server) getForwardStatistics(c *gin.Context) {
	stats, err := s.forwardService.GetForwardStatistics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      500,
			"message":   err.Error(),
			"timestamp": time.Now().UnixMilli(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "success",
		"data":      stats,
		"timestamp": time.Now().UnixMilli(),
	})
}

// @Summary Health Check
// @Description Check server health status
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UnixMilli(),
	})
}

// @Summary Get heartbeat status
// @Description Get summary of device heartbeat status
// @Tags heartbeat
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /heartbeat/status [get]
func (s *Server) getHeartbeatStatus(c *gin.Context) {
	devices, err := s.deviceManager.ListDevices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	status := make(map[string]interface{})
	for _, dev := range devices {
		status[dev.ID] = gin.H{
			"status":    dev.Status,
			"last_seen": dev.LastSeen,
			"online":    dev.Status == model.StatusOnline,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": status,
	})
}

// @Summary Get heartbeat config
// @Description Get current heartbeat configuration
// @Tags heartbeat
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /heartbeat/config [get]
func (s *Server) getHeartbeatConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": s.heartbeatManager.GetConfig(),
	})
}

// @Summary Update heartbeat config
// @Description Update heartbeat configuration
// @Tags heartbeat
// @Accept json
// @Produce json
// @Param config body config.HeartbeatConfig true "New Config"
// @Success 200 {object} map[string]interface{}
// @Router /heartbeat/config [put]
func (s *Server) updateHeartbeatConfig(c *gin.Context) {
	var cfg config.HeartbeatConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	s.heartbeatManager.UpdateConfig(cfg)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// @Summary Get online statistics
// @Description Get current online/total device counts
// @Tags statistics
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /statistics/online [get]
func (s *Server) getOnlineStatistics(c *gin.Context) {
	stats, err := s.deviceManager.GetDeviceStatistics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": stats,
	})
}

// @Summary Get trend statistics
// @Description Get historical online trend data
// @Tags statistics
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /statistics/trend [get]
func (s *Server) getTrendStatistics(c *gin.Context) {
	end := time.Now()
	start := end.Add(-24 * time.Hour) // Default to last 24 hours

	stats, err := s.deviceManager.GetTrendStatistics(c.Request.Context(), start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": stats,
	})
}
