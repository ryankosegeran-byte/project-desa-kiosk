package rfid

import (
	"sync"
	"time"
)

// Relay is a server-level in-memory hub for real-time RFID coordination,
// scoped per desa. It handles:
//
//  1. UID relay: kiosks push scanned UIDs; admin-panel browsers subscribe.
//  2. Registration session state: when an operator opens the "link card" step
//     in the admin panel, a session is marked active for that desa (with the
//     warga name being registered). Kiosks observe this to switch into
//     "registration scan" mode.
//  3. Kiosk-busy state: a kiosk reports when a resident is mid-flow (creating a
//     letter). While busy, the registration session is "pending" and the kiosk
//     does NOT switch screens, so the resident is not interrupted.
//
// No persistence needed — this is real-time coordination only.
type Relay struct {
	mu      sync.RWMutex
	clients map[string]map[chan string]bool // desa_id -> UID subscribers (admin panel)

	sessions map[string]*session // desa_id -> registration session

	// Subscribers that want the full session state (admin panel + kiosk).
	statusClients map[string]map[chan SessionState]bool

	// Subscribers that want a notification whenever resident data changes for a
	// desa. Used by kiosks to trigger an immediate incremental sync pull instead
	// of waiting for the next polling tick. The payload is the entity kind that
	// changed (e.g. "warga").
	syncClients map[string]map[chan string]bool
}

type session struct {
	active    bool
	name      string // name of the warga being registered (from admin form)
	kioskBusy bool   // a resident is mid-flow on the kiosk
	updatedAt time.Time
}

// SessionState is the snapshot broadcast to subscribers.
type SessionState struct {
	Active    bool   `json:"active"`
	Name      string `json:"name"`
	KioskBusy bool   `json:"kiosk_busy"`
	// Pending is true when a session is requested but the kiosk is busy.
	Pending bool `json:"pending"`
}

// NewRelay creates a new RFID relay hub.
func NewRelay() *Relay {
	return &Relay{
		clients:       make(map[string]map[chan string]bool),
		sessions:      make(map[string]*session),
		statusClients: make(map[string]map[chan SessionState]bool),
		syncClients:   make(map[string]map[chan string]bool),
	}
}

// ---------- UID relay (kiosk -> admin panel) ----------

// Subscribe registers a UID listener (admin panel) for a desa.
func (r *Relay) Subscribe(desaID string) chan string {
	ch := make(chan string, 5)
	r.mu.Lock()
	if r.clients[desaID] == nil {
		r.clients[desaID] = make(map[chan string]bool)
	}
	r.clients[desaID][ch] = true
	r.mu.Unlock()
	return ch
}

// Unsubscribe removes a UID listener.
func (r *Relay) Unsubscribe(desaID string, ch chan string) {
	r.mu.Lock()
	if subs, ok := r.clients[desaID]; ok {
		delete(subs, ch)
		if len(subs) == 0 {
			delete(r.clients, desaID)
		}
	}
	r.mu.Unlock()
	for range ch {
	}
}

// Publish broadcasts a UID to all admin-panel subscribers for the given desa.
func (r *Relay) Publish(desaID string, uid string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for ch := range r.clients[desaID] {
		select {
		case ch <- uid:
		default:
		}
	}
}

// ---------- Registration session ----------

// StartSession marks a session active for a desa, with the warga name.
func (r *Relay) StartSession(desaID, name string) {
	r.mu.Lock()
	s := r.sessions[desaID]
	if s == nil {
		s = &session{}
		r.sessions[desaID] = s
	}
	s.active = true
	s.name = name
	s.updatedAt = time.Now()
	r.mu.Unlock()
	r.broadcastStatus(desaID)
}

// StopSession marks the session inactive for a desa.
func (r *Relay) StopSession(desaID string) {
	r.mu.Lock()
	s := r.sessions[desaID]
	if s == nil {
		s = &session{}
		r.sessions[desaID] = s
	}
	s.active = false
	s.name = ""
	s.updatedAt = time.Now()
	r.mu.Unlock()
	r.broadcastStatus(desaID)
}

// SetKioskBusy updates whether a resident is mid-flow on the kiosk for a desa.
func (r *Relay) SetKioskBusy(desaID string, busy bool) {
	r.mu.Lock()
	s := r.sessions[desaID]
	if s == nil {
		s = &session{}
		r.sessions[desaID] = s
	}
	if s.kioskBusy == busy {
		r.mu.Unlock()
		return
	}
	s.kioskBusy = busy
	s.updatedAt = time.Now()
	r.mu.Unlock()
	r.broadcastStatus(desaID)
}

// State returns the current session snapshot for a desa.
func (r *Relay) State(desaID string) SessionState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stateLocked(desaID)
}

func (r *Relay) stateLocked(desaID string) SessionState {
	s := r.sessions[desaID]
	if s == nil {
		return SessionState{}
	}
	return SessionState{
		Active:    s.active,
		Name:      s.name,
		KioskBusy: s.kioskBusy,
		Pending:   s.active && s.kioskBusy,
	}
}

// SubscribeStatus registers a listener for full session-state changes.
func (r *Relay) SubscribeStatus(desaID string) chan SessionState {
	ch := make(chan SessionState, 4)
	r.mu.Lock()
	if r.statusClients[desaID] == nil {
		r.statusClients[desaID] = make(map[chan SessionState]bool)
	}
	r.statusClients[desaID][ch] = true
	r.mu.Unlock()
	return ch
}

// UnsubscribeStatus removes a status listener.
func (r *Relay) UnsubscribeStatus(desaID string, ch chan SessionState) {
	r.mu.Lock()
	if subs, ok := r.statusClients[desaID]; ok {
		delete(subs, ch)
		if len(subs) == 0 {
			delete(r.statusClients, desaID)
		}
	}
	r.mu.Unlock()
	for range ch {
	}
}

func (r *Relay) broadcastStatus(desaID string) {
	r.mu.RLock()
	state := r.stateLocked(desaID)
	subs := make([]chan SessionState, 0, len(r.statusClients[desaID]))
	for ch := range r.statusClients[desaID] {
		subs = append(subs, ch)
	}
	r.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- state:
		default:
		}
	}
}

// ---------- Sync notifications (server -> kiosk) ----------

// SubscribeSync registers a listener that receives a notification whenever
// resident data changes for the given desa.
func (r *Relay) SubscribeSync(desaID string) chan string {
	ch := make(chan string, 8)
	r.mu.Lock()
	if r.syncClients[desaID] == nil {
		r.syncClients[desaID] = make(map[chan string]bool)
	}
	r.syncClients[desaID][ch] = true
	r.mu.Unlock()
	return ch
}

// UnsubscribeSync removes a sync listener.
func (r *Relay) UnsubscribeSync(desaID string, ch chan string) {
	r.mu.Lock()
	if subs, ok := r.syncClients[desaID]; ok {
		delete(subs, ch)
		if len(subs) == 0 {
			delete(r.syncClients, desaID)
		}
	}
	r.mu.Unlock()
	for range ch {
	}
}

// NotifySync broadcasts a data-change notification to all kiosks subscribed for
// the given desa. `kind` describes what changed (e.g. "warga"). Best-effort and
// non-blocking: slow subscribers simply miss the tick and catch up on the next
// notification or the periodic polling fallback.
func (r *Relay) NotifySync(desaID, kind string) {
	if desaID == "" {
		return
	}
	r.mu.RLock()
	subs := make([]chan string, 0, len(r.syncClients[desaID]))
	for ch := range r.syncClients[desaID] {
		subs = append(subs, ch)
	}
	r.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- kind:
		default:
		}
	}
}
