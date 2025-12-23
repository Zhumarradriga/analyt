package service

import (
	"context"
	"log"
	"analytservice/internal/domain"
	"analytservice/internal/repository"
	"sync"
	"time"
)

type AnalyticsService struct {
	repo          repository.EventRepo
	eventChan     chan domain.Event
	batchSize     int
	flushInterval time.Duration
	wg            sync.WaitGroup
	quit          chan struct{}
}

func NewAnalyticsService(repo repository.EventRepo, batchSize int, interval time.Duration) *AnalyticsService {
	return &AnalyticsService{
		repo:          repo,
		eventChan:     make(chan domain.Event, batchSize*2), // запас если нужен будет
		batchSize:     batchSize,
		flushInterval: interval,
		quit:          make(chan struct{}),
	}
}

// Track добавляет событие в канал (не блокирует)
func (s *AnalyticsService) Track(e domain.Event) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}

	select {
	case s.eventChan <- e:
		// OK
	default:
		// канал полон, логируем ошибку
		log.Println("ERROR: Event buffer full. Dropping event.")
	}
}

// фоновый воркер для сбора и записи батчами
func (s *AnalyticsService) Start() {
	s.wg.Add(1)
	go s.loop()
}

// завершаем работу после всех задачек
func (s *AnalyticsService) Stop() {
	close(s.quit)
	s.wg.Wait()
	close(s.eventChan) 
}

func (s *AnalyticsService) loop() {
	defer s.wg.Done()
	
	buffer := make([]domain.Event, 0, s.batchSize)
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(buffer) == 0 {
			return
		}
		saveBuffer := make([]domain.Event, len(buffer))
		copy(saveBuffer, buffer)
		buffer = buffer[:0]

		// если за пять секунд не грузит то померло
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.repo.SaveBatch(ctx, saveBuffer); err != nil {
			log.Printf("ERROR: Failed to flush batch to ClickHouse: %v", err)
			// померло, выкид
		} else {
			log.Printf("INFO: Flushed %d events", len(saveBuffer))
		}
	}

	for {
		select {
		case e := <-s.eventChan:
			buffer = append(buffer, e)
			if len(buffer) >= s.batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-s.quit:
			flush()
			return
		}
	}
}