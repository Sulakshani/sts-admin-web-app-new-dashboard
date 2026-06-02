package services

import (
	"database/sql"
	"log"
	"time"
	
	"sts-backend/internal/config"
)

// EscalationScheduler handles periodic checking and escalation of complaints
type EscalationScheduler struct {
	escalationService *EscalationService
	ticker            *time.Ticker
	stopChan          chan bool
	checkInterval     time.Duration
}

// NewEscalationScheduler creates a new escalation scheduler
func NewEscalationScheduler(db *sql.DB, cfg *config.Config, checkInterval time.Duration) *EscalationScheduler {
	return &EscalationScheduler{
		escalationService: NewEscalationService(db, cfg),
		checkInterval:     checkInterval,
		stopChan:          make(chan bool),
	}
}

// Start begins the periodic escalation check
func (s *EscalationScheduler) Start() {
	log.Printf("Starting complaint escalation scheduler (interval: %v)", s.checkInterval)
	
	s.ticker = time.NewTicker(s.checkInterval)
	
	// Run once immediately on startup
	go s.runEscalationCheck()
	
	// Then run periodically
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.runEscalationCheck()
			case <-s.stopChan:
				log.Println("Stopping complaint escalation scheduler")
				s.ticker.Stop()
				return
			}
		}
	}()
}

// Stop halts the scheduler
func (s *EscalationScheduler) Stop() {
	s.stopChan <- true
}

// runEscalationCheck performs the escalation check
func (s *EscalationScheduler) runEscalationCheck() {
	log.Println("Running scheduled escalation check...")
	
	escalatedCount, err := s.escalationService.CheckAndEscalateComplaints()
	if err != nil {
		log.Printf("Error during escalation check: %v", err)
		return
	}
	
	if escalatedCount > 0 {
		log.Printf("Successfully escalated %d complaint(s)", escalatedCount)
	} else {
		log.Println("No complaints required escalation")
	}
	
	// Log statistics
	stats, err := s.escalationService.GetEscalationStats()
	if err != nil {
		log.Printf("Failed to get escalation stats: %v", err)
	} else {
		log.Printf("Escalation stats: %+v", stats)
	}
}

// RunEscalationCheckNow manually triggers an escalation check (useful for testing)
func (s *EscalationScheduler) RunEscalationCheckNow() (int, error) {
	return s.escalationService.CheckAndEscalateComplaints()
}
