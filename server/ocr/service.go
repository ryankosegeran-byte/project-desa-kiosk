package ocr

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog/log"
)

type Service struct {
	providers []OCRProvider
	strategy  string // failover, round_robin
	index     uint32
	mu        sync.RWMutex
}

func NewService(providers []OCRProvider, strategy string) *Service {
	if strategy == "" {
		strategy = "failover"
	}
	return &Service{
		providers: providers,
		strategy:  strategy,
	}
}

func (s *Service) SetProviders(providers []OCRProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers = providers
}

func (s *Service) SetStrategy(strategy string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.strategy = strategy
}

// ExtractKTP runs the OCR extraction using configured providers.
func (s *Service) ExtractKTP(ctx context.Context, imageData []byte) (*KTPData, error) {
	s.mu.RLock()
	provs := s.providers
	strat := s.strategy
	s.mu.RUnlock()

	if len(provs) == 0 {
		mock := &MockProvider{}
		return mock.ExtractKTP(ctx, imageData)
	}

	if strat == "round_robin" {
		idx := atomic.AddUint32(&s.index, 1) - 1
		p := provs[idx%uint32(len(provs))]
		log.Info().Str("provider", p.Name()).Msg("Invoking AI OCR (Round Robin)")
		return p.ExtractKTP(ctx, imageData)
	}

	// Default: Failover strategy
	var lastErr error
	for _, p := range provs {
		log.Info().Str("provider", p.Name()).Msg("Attempting AI OCR extraction...")
		res, err := p.ExtractKTP(ctx, imageData)
		if err == nil {
			log.Info().Str("provider", p.Name()).Msg("AI OCR extraction successful")
			return res, nil
		}
		log.Warn().Str("provider", p.Name()).Err(err).Msg("AI OCR provider failed, trying failover...")
		lastErr = err
	}

	return nil, fmt.Errorf("all AI OCR providers failed: %w", lastErr)
}

// MockProvider is a fallback provider for testing and offline development
type MockProvider struct{}

func (m *MockProvider) Name() string { return "mock" }
func (m *MockProvider) ExtractKTP(ctx context.Context, imageData []byte) (*KTPData, error) {
	log.Info().Msg("[MOCK OCR] Executing simulated KTP text extraction")
	return &KTPData{
		NIK:             "3201234567890099",
		Nama:            "Ahmad Faisal",
		TempatLahir:     "Garut",
		TanggalLahir:    "1988-10-05",
		JenisKelamin:    "L",
		Alamat:          "Kampung Cibunar No. 22",
		RT:              "02",
		RW:              "01",
		Kelurahan:       "Cibunar",
		Kecamatan:       "Cibatu",
		Agama:           "Islam",
		StatusKawin:     "Kawin",
		Pekerjaan:       "Petani",
		Kewarganegaraan: "WNI",
		Confidence:      0.95,
	}, nil
}
