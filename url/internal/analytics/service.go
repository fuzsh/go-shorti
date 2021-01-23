package analytics

import (
	"fmt"
	"strconv"
	"url/pkg/log"
)

var modeTypes = []string{
	"all",
	"platform",
	"browser",
}

var dateTypes = []string{
	"daily",
	"yesterday",
	"weekly",
	"monthly",
}

// Service encapsulates use case logic.
type Service interface {
	Analytic(queries queries, userID int) (interface{}, error)
}
type service struct {
	repo   Store
	logger log.Logger
}

type queries struct {
	Unique string
	Date   string
	Mode   string
}

// NewService creates a new service.
func NewService(repo Store, logger log.Logger) Service {
	return service{repo, logger}
}

func (s service) Analytic(queries queries, userID int) (interface{}, error) {
	conf, err := s.queriesValidator(queries)
	if err != nil {
		return nil, err
	}
	return s.repo.GetAnalytics(nil, conf, userID)
}

func (s service) queriesValidator(queries queries) (Config, error) {
	uniq, err := strconv.ParseBool(queries.Unique)
	if err != nil {
		return Config{}, fmt.Errorf("enter the correct boolean value, %s is not boolean", queries.Unique)
	}
	if !contains(modeTypes, queries.Mode) {
		return Config{}, fmt.Errorf("enter the correct mode, %s is not contain %s", modeTypes, queries.Mode)
	}
	if !contains(dateTypes, queries.Date) {
		return Config{}, fmt.Errorf("enter the correct date type, %s is not contain %s", dateTypes, queries.Date)
	}
	return Config{
		Unique: uniq,
		Date:   queries.Date,
		Mode:   queries.Mode,
	}, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
