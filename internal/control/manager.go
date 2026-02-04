package control

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"sl651-platform/internal/model"
	"sl651-platform/internal/sl651"
	"sl651-platform/internal/storage"
	"strings"
	"sync"
	"time"
)

type Manager struct {
	storage         *storage.Storage
	protocol        *sl651.Protocol
	pendingCommands map[string][]*model.DeviceCommand
	mutex           sync.RWMutex
}

func NewManager(storage *storage.Storage, protocol *sl651.Protocol) *Manager {
	return &Manager{
		storage:         storage,
		protocol:        protocol,
		pendingCommands: make(map[string][]*model.DeviceCommand),
	}
}

// Start loads pending commands from DB during initialization
func (m *Manager) Start(ctx context.Context) error {
	cmds, err := m.storage.ListPendingCommands(ctx)
	if err != nil {
		return err
	}

	m.mutex.Lock()
	for _, cmd := range cmds {
		m.pendingCommands[cmd.DeviceID] = append(m.pendingCommands[cmd.DeviceID], cmd)
	}
	m.mutex.Unlock()

	log.Printf("Control Manager started, loaded %d pending commands", len(cmds))
	return nil
}

func (m *Manager) QueueCommand(ctx context.Context, deviceID string, functionCode string, payload string) (*model.DeviceCommand, error) {
	cmd := &model.DeviceCommand{
		ID:           fmt.Sprintf("CMD-%d", time.Now().UnixNano()),
		DeviceID:     deviceID,
		FunctionCode: functionCode,
		Payload:      payload,
		Status:       model.CommandStatusPending,
		CreatedAt:    time.Now(),
	}

	if err := m.storage.SaveCommand(ctx, cmd); err != nil {
		return nil, err
	}

	m.mutex.Lock()
	m.pendingCommands[deviceID] = append(m.pendingCommands[deviceID], cmd)
	m.mutex.Unlock()

	return cmd, nil
}

// PopPending retrieves the next pending command for a device and marks it as sending
func (m *Manager) PopPending(ctx context.Context, deviceID string) *model.DeviceCommand {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cmds := m.pendingCommands[deviceID]
	if len(cmds) == 0 {
		return nil
	}

	cmd := cmds[0]
	// Remove from memory queue (assuming we try to send it now)
	m.pendingCommands[deviceID] = cmds[1:]

	// Update status in DB
	err := m.storage.UpdateCommandStatus(ctx, cmd.ID, model.CommandStatusSent, "")
	if err != nil {
		log.Printf("[Error] Failed to update command status for %s: %v", cmd.ID, err)
	}
	cmd.Status = model.CommandStatusSent
	sentTime := time.Now()
	cmd.SentAt = &sentTime

	return cmd
}

// HandleResponse processes the response from a device for a sent command
func (m *Manager) HandleResponse(ctx context.Context, deviceID string, functionCode string, payload []byte) {
	// 40H (Read) response is 40H with data
	// 42H (Set) response is 42H (ACK)

	// Find the most recently 'sent' command for this device and function code
	// Simplified: just find the latest 'sent' command for this device
	m.mutex.RLock()
	// This is tricky because we might have multiple pending/sent.
	// Usually one at a time per device.
	m.mutex.RUnlock()

	// Real implementation should query DB for newest 'sent' command of this type
	// For now, let's just log and update if we find a matching one in recent history
	// In next iteration, we'll implement more robust matching using serial number or timestamp

	log.Printf("Received response for device %s: function %s, payload %x", deviceID, functionCode, payload)

	// Update command status in DB (find by deviceID and status=sent)
	// For V1.3, we'll just mark success if we get a valid response
}

func (m *Manager) BuildCommandFrame(cmd *model.DeviceCommand) ([]byte, error) {
	fCode, _ := hex.DecodeString(cmd.FunctionCode)
	if len(fCode) == 0 {
		return nil, fmt.Errorf("invalid function code: %s", cmd.FunctionCode)
	}

	var body []byte
	var err error

	// Interpret payload based on function code
	if strings.ToUpper(cmd.FunctionCode) == "40" {
		// Read parameters - payload should be list of tags in hex
		body, err = hex.DecodeString(cmd.Payload)
	} else if strings.ToUpper(cmd.FunctionCode) == "42" {
		// Set parameters - payload could be hex directly for now
		body, err = hex.DecodeString(cmd.Payload)
	} else {
		body, err = hex.DecodeString(cmd.Payload)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	return m.protocol.BuildMessage(cmd.DeviceID, fCode[0], body)
}

func (m *Manager) GetHistory(ctx context.Context, deviceID string, limit int) ([]*model.DeviceCommand, error) {
	return m.storage.GetCommands(ctx, deviceID, limit)
}
