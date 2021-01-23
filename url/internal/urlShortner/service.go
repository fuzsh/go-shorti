package urlShortner

import (
	"context"
	"net/http"
	"net/url"
	"url/internal/config"
	"url/internal/store"
	"url/internal/track"
	"url/pkg/log"
	"url/pkg/validators"
)

// Service encapsulates use case logic.
type Service interface {
	EnCode(ctx context.Context, dto InputDTO, userID int) (string, error)
	Load(r *http.Request, url string) (string, error)
}

type InputDTO struct {
	URL       string `json:"url" validate:"required,url"`
	SimilarTo string `json:"similar_to"`
}

type service struct {
	repo    Repository
	store   Store
	logger  log.Logger
	tracker *track.Tracker
}

// NewService creates a new service.
func NewService(trackerStore *store.PostgresStore, store Store, repo Repository, logger log.Logger) Service {
	tracker := track.NewTracker(trackerStore, "salt", &track.TrackerConfig{Logger: logger})
	return service{repo, store, logger, tracker}
}

func (s service) EnCode(ctx context.Context, req InputDTO, userID int) (string, error) {
	if ok, err := validators.Validate(req); !ok {
		return "", err
	}
	URI, err := url.ParseRequestURI(req.URL)
	if err != nil {
		return "", err
	}
	// generate link and save
	path, err := s.repo.Create(ctx, URI.String(), req.SimilarTo)
	if err != nil {
		return "", err
	}
	u := url.URL{
		Scheme: config.Cfg.Options.Schema,
		Host:   config.Cfg.Options.BaseURL,
		Path:   path,
	}
	if err := s.createLink(userID, req.URL, path); err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s service) Load(request *http.Request, url string) (string, error) {
	uri, err := s.repo.FindOne(request.Context(), url)
	if err != nil {
		return "", err
	}
	go s.track(request)
	return uri, nil
}

func (s service) track(r *http.Request) {
	s.tracker.Hit(r, nil)
}

func (s service) createLink(userID int, url, path string) error {
	tx := s.store.NewTx()
	linkID, err := s.store.CreateLink(tx, url, path)
	if err != nil {
		s.store.Rollback(tx)
		return err
	}
	if err := s.store.CreateUserLinkRelation(tx, userID, linkID); err != nil {
		s.store.Rollback(tx)
		return err
	}
	s.store.Commit(tx)
	return nil
}
