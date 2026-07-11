package assistant

import (
	"context"
	"fmt"

	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/domain"
	"gorm.io/gorm"
)

// DeviceWakeWordProvider mengimplementasikan WakeWordProvider lewat query
// devices -> user_settings (GORM).
//
// ASUMSI yang perlu kamu koreksi kalau beda:
//   - Kolom di tabel "devices" yang menyimpan device_id dari URL WS aku
//     tebak namanya "device_id" (bukan "id"). Ganti WHERE clause kalau beda.
//   - Import path domain package aku tebak dari pola module di service.go.
type DeviceWakeWordProvider struct {
	db *gorm.DB
}

func NewDeviceWakeWordProvider(db *gorm.DB) *DeviceWakeWordProvider {
	return &DeviceWakeWordProvider{db: db}
}

func (p *DeviceWakeWordProvider) GetWakeWords(ctx context.Context, deviceID string) (map[string]Personality, error) {
	var userID string
	if err := p.db.WithContext(ctx).
		Table("devices").
		Select("user_id").
		Where("id = ?", deviceID).
		Scan(&userID).Error; err != nil {
		return nil, fmt.Errorf("lookup device %q: %w", deviceID, err)
	}
	if userID == "" {
		return nil, fmt.Errorf("device %q belum terdaftar", deviceID)
	}

	var settings domain.UserSettings
	if err := p.db.WithContext(ctx).First(&settings, "user_id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("load user_settings utk user %s: %w", userID, err)
	}

	out := make(map[string]Personality, len(settings.WakeWords))
	for _, w := range settings.WakeWords {
		out[w.Word] = Personality(w.Personality)
	}
	return out, nil
}
