package track

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"url/pkg/log"
)

const (
	defaultWorkerBufferSize = 100
	defaultWorkerTimeout    = time.Second * 10
	maxWorkerTimeout        = time.Second * 60
)

// TrackerConfig is the optional configuration for the Tracker.
type TrackerConfig struct {
	// Worker sets the number of workers that are used to store hits.
	Worker int

	// WorkerBufferSize is the size of the buffer used to store hits.
	WorkerBufferSize int

	// WorkerTimeout sets the timeout used to store hits.
	// This is used to allow the workers to store hits even if the buffer is not full yet.
	WorkerTimeout time.Duration

	// ReferrerDomainBlacklist see HitOptions.ReferrerDomainBlacklist.
	ReferrerDomainBlacklist []string

	// ReferrerDomainBlacklistIncludesSubdomains see HitOptions.ReferrerDomainBlacklistIncludesSubdomains.
	ReferrerDomainBlacklistIncludesSubdomains bool

	// Sessions enables/disables session tracking.
	// It's enabled by default.
	Sessions bool

	// SessionMaxAge is used to define how long a session runs at maximum.
	// Set to two hours by default.
	SessionMaxAge time.Duration

	// SessionCleanupInterval sets the session cache lifetime.
	// If not passed, the default will be used.
	SessionCleanupInterval time.Duration


	// Logger is the log.Logger used for logging.
	Logger log.Logger
}

// The TrackerConfig just passes on the values and overwrites them if required.
func (config *TrackerConfig) validate() {
	if config.Worker < 1 {
		config.Worker = runtime.NumCPU()
	}

	if config.WorkerBufferSize < 1 {
		config.WorkerBufferSize = defaultWorkerBufferSize
	}

	if config.WorkerTimeout <= 0 {
		config.WorkerTimeout = defaultWorkerTimeout
	} else if config.WorkerTimeout > maxWorkerTimeout {
		config.WorkerTimeout = maxWorkerTimeout
	}

	if config.Logger == nil {
		config.Logger = log.New()
	}
}

// Tracker.
// Make sure you call Stop to make sure the hits get stored before shutting down the server.
type Tracker struct {
	store                                     Store
	salt                                      string
	hits                                      chan Hit
	stopped                                   int32
	worker                                    int
	workerBufferSize                          int
	workerTimeout                             time.Duration
	workerCancel                              context.CancelFunc
	workerDone                                chan bool
	referrerDomainBlacklist                   []string
	referrerDomainBlacklistIncludesSubdomains bool
	geoDBMutex                                sync.RWMutex
	logger                                    log.Logger
}

// NewTracker creates a new tracker for given store, salt and config.
// Pass nil for the config to use the defaults.
// The salt is mandatory.
func NewTracker(store Store, salt string, config *TrackerConfig) *Tracker {
	if config == nil {
		// the other default values are set by validate
		config = &TrackerConfig{
			Sessions: true,
		}
	}

	config.validate()

	tracker := &Tracker{
		store:                   store,
		salt:                    salt,
		hits:                    make(chan Hit, config.Worker*config.WorkerBufferSize),
		worker:                  config.Worker,
		workerBufferSize:        config.WorkerBufferSize,
		workerTimeout:           config.WorkerTimeout,
		workerDone:              make(chan bool),
		referrerDomainBlacklist: config.ReferrerDomainBlacklist,
		referrerDomainBlacklistIncludesSubdomains: config.ReferrerDomainBlacklistIncludesSubdomains,
		logger:       config.Logger,
	}
	tracker.startWorker()
	return tracker
}

// Hit stores the given request.
// The request might be ignored if it meets certain conditions. The HitOptions, if passed, will overwrite the Tracker configuration.
func (tracker *Tracker) Hit(r *http.Request, options *HitOptions) {
	if atomic.LoadInt32(&tracker.stopped) > 0 {
		return
	}

	if !IgnoreHit(r) {
		if options == nil {
			options = &HitOptions{
				ReferrerDomainBlacklist:                   tracker.referrerDomainBlacklist,
				ReferrerDomainBlacklistIncludesSubdomains: tracker.referrerDomainBlacklistIncludesSubdomains,
			}
		}

		tracker.hits <- HitFromRequest(r, tracker.salt, options)
	}
}

// Flush flushes all hits to store that are currently buffered by the workers.
// Call Tracker.Stop to also save hits that are in the queue.
func (tracker *Tracker) Flush() {
	tracker.stopWorker()
	tracker.startWorker()
}

// Stop flushes and stops all workers.
func (tracker *Tracker) Stop() {
	if atomic.LoadInt32(&tracker.stopped) == 0 {
		atomic.StoreInt32(&tracker.stopped, 1)
		tracker.stopWorker()
		tracker.flushHits()
	}
}

func (tracker *Tracker) startWorker() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	tracker.workerCancel = cancelFunc

	for i := 0; i < tracker.worker; i++ {
		go tracker.aggregate(ctx)
	}
}

func (tracker *Tracker) stopWorker() {
	tracker.workerCancel()

	for i := 0; i < tracker.worker; i++ {
		<-tracker.workerDone
	}
}

func (tracker *Tracker) flushHits() {
	// this function will make sure all dangling hits will be saved in database before shutdown
	hits := make([]Hit, 0, tracker.workerBufferSize)

	for {
		stop := false

		select {
		case hit := <-tracker.hits:
			hits = append(hits, hit)

			if len(hits) == tracker.workerBufferSize {
				tracker.saveHits(hits)
				hits = hits[:0]
			}
		default:
			stop = true
		}

		if stop {
			break
		}
	}

	tracker.saveHits(hits)
}

func (tracker *Tracker) aggregate(ctx context.Context) {
	hits := make([]Hit, 0, tracker.workerBufferSize)
	timer := time.NewTimer(tracker.workerTimeout)
	defer timer.Stop()

	for {
		timer.Reset(tracker.workerTimeout)

		select {
		case hit := <-tracker.hits:
			hits = append(hits, hit)
			tracker.saveHits(hits)
			hits = hits[:0]
		case <-timer.C:
			tracker.saveHits(hits)
			hits = hits[:0]
		case <-ctx.Done():
			tracker.saveHits(hits)
			tracker.workerDone <- true
			return
		}
	}
}

func (tracker *Tracker) saveHits(hits []Hit) {
	if len(hits) > 0 {
		fmt.Println(hits)
		if err := tracker.store.SaveHits(hits); err != nil {
			tracker.logger.Infof("error saving hits: %s", err)
		}
	}
}
