package rfid

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
)

// Broker handles multiple Server-Sent Events (SSE) clients and broadcasts RFID scan events to them.
type Broker struct {
	mu          sync.RWMutex
	clients     map[chan string]bool
	newClients  chan chan string
	closeClient chan chan string
	messages    chan string
}

// NewBroker creates and returns a new SSE Broker instance.
func NewBroker() *Broker {
	return &Broker{
		clients:     make(map[chan string]bool),
		newClients:  make(chan chan string),
		closeClient: make(chan chan string),
		messages:    make(chan string, 10),
	}
}

// Start runs the Broker event loop in a background goroutine.
func (b *Broker) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				// Close all client channels
				b.mu.Lock()
				for client := range b.clients {
					close(client)
				}
				b.clients = make(map[chan string]bool)
				b.mu.Unlock()
				log.Info().Msg("RFID SSE Broker stopped")
				return

			case client := <-b.newClients:
				b.mu.Lock()
				b.clients[client] = true
				b.mu.Unlock()
				log.Debug().Msg("Client SSE terhubung ke RFID Broker")

			case client := <-b.closeClient:
				b.mu.Lock()
				if _, ok := b.clients[client]; ok {
					delete(b.clients, client)
					close(client)
					log.Debug().Msg("Client SSE terputus dari RFID Broker")
				}
				b.mu.Unlock()

			case msg := <-b.messages:
				b.mu.RLock()
				for client := range b.clients {
					// Non-blocking write to avoid hanging if client is slow
					select {
					case client <- msg:
					default:
						log.Warn().Msg("Client SSE lambat, event RFID dilewatkan")
					}
				}
				b.mu.RUnlock()
			}
		}
	}()
	log.Info().Msg("RFID SSE Broker dijalankan")
}

// Subscribe creates a client channel and registers it with the broker.
func (b *Broker) Subscribe() chan string {
	client := make(chan string, 5)
	b.newClients <- client
	return client
}

// Unsubscribe de-registers a client channel.
func (b *Broker) Unsubscribe(client chan string) {
	b.closeClient <- client
}

// Publish broadcasts an RFID UID message to all subscribed clients.
func (b *Broker) Publish(uid string) {
	b.messages <- uid
}
