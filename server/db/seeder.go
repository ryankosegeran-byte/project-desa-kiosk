package db

import (
	"context"
	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/project-desa-kiosk/internal/models"
)

// SeedServerData seeds initial default datasets to PostgreSQL if empty.
func SeedServerData(db *DB) error {
	ctx := context.Background()

	// 1. Check if Desa is empty
	var desaCount int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM desa").Scan(&desaCount)
	if err != nil {
		return err
	}

	desaID := "d3b07384-d113-4ec5-a55e-2e06c783c180"

	if desaCount == 0 {
		log.Info().Msg("Seeding default Desa Mekarjaya...")
		desaRepo := NewDesaRepository(db)
		d := &models.Desa{
			ID:            desaID,
			Nama:          "Desa Mekarjaya",
			KodeDesa:      "3201012001",
			Kecamatan:     "Sukasari",
			Kabupaten:     "Bandung",
			Provinsi:      "Jawa Barat",
			KepalaDesa:    "Ujang Hermawan, S.Sos",
			NIPKepalaDesa: "19750812 200312 1 002",
			AlamatKantor:  "Alamat Kantor Desa Mekarjaya No. 1",
		}
		if err := desaRepo.Create(ctx, d); err != nil {
			return err
		}
	}

	// 2. Check if Users is empty
	var userCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		return err
	}

	if userCount == 0 {
		log.Info().Msg("Seeding default admin accounts...")
		userRepo := NewUserRepository(db)

		// Hashed password: "password"
		pwdHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

		// Superadmin
		sa := &models.User{
			ID:           uuid.New().String(),
			Username:     "admin",
			PasswordHash: string(pwdHash),
			Nama:         "Super Administrator",
			Role:         models.RoleSuperAdmin,
			Active:       true,
		}
		if err := userRepo.Create(ctx, sa); err != nil {
			return err
		}

		// PIC Desa
		pic := &models.User{
			ID:           uuid.New().String(),
			Username:     "pic_mekarjaya",
			PasswordHash: string(pwdHash),
			Nama:         "Ujang Hermawan",
			Role:         models.RolePICDesa,
			Jabatan:      "Kepala Desa",
			DesaID:       desaID,
			Active:       true,
		}
		if err := userRepo.Create(ctx, pic); err != nil {
			return err
		}
	}

	// 3. Check if Kiosks is empty
	var kioskCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM kiosks").Scan(&kioskCount)
	if err != nil {
		return err
	}

	if kioskCount == 0 {
		log.Info().Msg("Seeding default kiosk configuration...")
		desaRepo := NewDesaRepository(db)
		k := &models.Kiosk{
			ID:     "a93b4f62-38d7-4638-b715-8fa9074a38f3",
			DesaID: desaID,
			Nama:   "Kiosk Utama",
			APIKey: "mock_kiosk_key_123",
			Status: "active",
		}
		if err := desaRepo.RegisterKiosk(ctx, k); err != nil {
			return err
		}
	}

	return nil
}
