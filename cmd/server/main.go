package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sl651-platform/internal/config"
	"sl651-platform/internal/database"
	"sl651-platform/internal/device"
	"sl651-platform/internal/diagnosis"
	"sl651-platform/internal/forward"
	"sl651-platform/internal/heartbeat"
	"sl651-platform/internal/http"
	"sl651-platform/internal/model"
	"sl651-platform/internal/notifier"
	"sl651-platform/internal/quality"
	"sl651-platform/internal/sl651"
	"sl651-platform/internal/storage"
	"syscall"
)

type TCPServer struct {
	addr             string
	listener         net.Listener
	protocol         *sl651.Protocol
	deviceManager    *device.Manager
	forwardService   *forward.Service
	diagnosisManager *diagnosis.Manager
	qualityManager   *quality.Manager
}

func NewTCPServer(addr string, protocol *sl651.Protocol, deviceManager *device.Manager, forwardService *forward.Service, diagnosisManager *diagnosis.Manager, qualityManager *quality.Manager) *TCPServer {
	return &TCPServer{
		addr:             addr,
		protocol:         protocol,
		deviceManager:    deviceManager,
		forwardService:   forwardService,
		diagnosisManager: diagnosisManager,
		qualityManager:   qualityManager,
	}
}

func (s *TCPServer) Start(ctx context.Context) error {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}

	log.Printf("TCP server started on %s", s.addr)

	go func() {
		<-ctx.Done()
		log.Println("Stopping TCP server...")
		s.listener.Close()
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				log.Printf("Error accepting connection: %v", err)
				continue
			}
		}
		go s.handleConnection(ctx, conn)
	}
}

func (s *TCPServer) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	log.Printf("New connection from %s", remoteAddr)

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Error reading from %s: %v", remoteAddr, err)
		return
	}

	hexData := hex.EncodeToString(buffer[:n])
	log.Printf("Received Hex: %s", hexData)

	msg, err := s.protocol.Parse(hexData)
	if err != nil {
		log.Printf("Protocol Parse Error: %v", err)
		s.diagnosisManager.HandleCommError(ctx, "unknown", model.FaultTypeComm, model.SeverityError, "Protocol Parse Error", err.Error()+" | data: "+hexData)
		s.qualityManager.Evaluate(ctx, "unknown", hexData, nil, err)
		return
	}

	data, err := s.protocol.ConvertToStandardData(msg)
	if err != nil {
		log.Printf("Data Conversion Error: %v", err)
		s.diagnosisManager.HandleCommError(ctx, msg.StationID, model.FaultTypeData, model.SeverityError, "Data Conversion Error", err.Error())
		s.qualityManager.Evaluate(ctx, msg.StationID, hexData, nil, err)
		return
	}

	// Trigger Quality Evaluation
	s.qualityManager.Evaluate(ctx, msg.StationID, hexData, data, nil)

	// Trigger Diagnosis Analysis
	s.diagnosisManager.AnalyzeData(ctx, msg.StationID, data)

	jsonData, _ := json.Marshal(data)
	log.Printf("Parsed Data: %s", jsonData)

	if err := s.deviceManager.ProcessMessage(ctx, msg); err != nil {
		log.Printf("Device Manager Process Error: %v", err)
		s.diagnosisManager.HandleCommError(ctx, msg.StationID, model.FaultTypeData, model.SeverityWarning, "Device Process Error", err.Error())
	}

	// Send Success Response (Function Code - 0x32 for timed report etc)
	response, _ := s.protocol.BuildMessage(msg.StationID, 0xA1, nil) // A1 is often Success Response in some contexts, or just follow standard
	conn.Write(response)
	log.Printf("Sent Response: %x", response)
}

func (s *TCPServer) Stop() {
	if s.listener != nil {
		s.listener.Close()
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Init(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	storage := storage.NewStorage(db)
	protocol := sl651.NewProtocol()
	diagnosisManager := diagnosis.NewManager(storage)
	qualityManager := quality.NewManager(storage)
	deviceManager := device.NewManager(storage, protocol)
	forwardService := forward.NewService(storage)
	notifierService := notifier.NewNotifier()
	heartbeatManager := heartbeat.NewHeartbeatManager(cfg.Heartbeat, deviceManager, notifierService)

	go diagnosisManager.LogSystemEvent(ctx, "INFO", "System", "Platform starting...")
	go diagnosisManager.LogSystemEvent(ctx, "INFO", "Database", "Database initialized and migrated")

	tcpServer := NewTCPServer(":8080", protocol, deviceManager, forwardService, diagnosisManager, qualityManager)
	go func() {
		if err := tcpServer.Start(ctx); err != nil {
			log.Printf("TCP server error: %v", err)
		}
	}()

	go deviceManager.Start(ctx)
	go forwardService.Start(ctx)
	go heartbeatManager.Start(ctx)

	go func() {
		for data := range deviceManager.GetDataChannel() {
			forwardService.GetDataChannel() <- data
		}
	}()

	httpServer := http.NewServer(cfg, deviceManager, forwardService, heartbeatManager, diagnosisManager, qualityManager)
	go httpServer.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	httpServer.Stop()
	tcpServer.Stop()
	cancel()
	log.Println("Server stopped")
}
