package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// var scheduledJobLogger = logger.New("ScheduledJob")

type FnJob func(ctx context.Context) error

// Create interface
type SchedulerJob interface {
	Cancel()
	Start(ctx context.Context, maxTried int)
}

type scheduledJob struct {
	name   string
	period time.Duration
	job    FnJob

	mu           *sync.Mutex
	isStarted    bool
	cancelSignal chan interface{}
	isCancelled  bool
	numError     int
}

// NewScheduleJob returns new instance of SchedulerJob
func NewScheduleJob(name string, period time.Duration, fn FnJob) SchedulerJob {
	return &scheduledJob{
		name:   name,
		period: period,
		job:    fn,

		mu:           &sync.Mutex{},
		isStarted:    false,
		cancelSignal: make(chan interface{}),
		isCancelled:  false,
	}
}

func (s *scheduledJob) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isCancelled {
		return
	}

	log.Println("Cancelling job", "name", s.name, "period", s.period)
	close(s.cancelSignal)
	s.isCancelled = true
}

func (s *scheduledJob) Start(ctx context.Context, maxTried int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isStarted {
		return
	}

	log.Println("Starting job", "name", s.name, "period", s.period)
	s.isStarted = true

	go func(ctx context.Context) {
		tick := time.NewTicker(s.period)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("receive done signal, exit")
				return
			case <-s.cancelSignal:
				log.Println("Job cancelled", "name", s.name, "period", s.period)
				return
			case <-tick.C:
				t := time.Now()
				s.executeWithRecovery(ctx, maxTried)
				log.Println("Job execution done", "name", s.name, "duration", time.Since(t), "start", t, "period", s.period)
			}
		}
	}(ctx)
}

func (s *scheduledJob) executeWithRecovery(ctx context.Context, maxTried int) {
	defer func() {
		if s.numError >= maxTried {
			log.Println("max tried reached")
			close(s.cancelSignal)
		}
	}()
	defer func() {
		p := recover()
		if p == nil {
			return
		}

		s.numError = s.numError + 1
		switch v := p.(type) {
		case error:
			log.Println("ScheduledJob paniced", v, "name", s.name, "period", s.period)
		default:
			log.Println("ScheduledJob paniced", fmt.Errorf("unknown panic type %T", v), "panic", v)
		}
	}()

	if err := s.job(ctx); err != nil {
		log.Println("job error", err)
		s.numError = s.numError + 1
	}
}
