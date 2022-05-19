package service

import (
	"git.neds.sh/entain/sports/db"
	"git.neds.sh/entain/sports/proto/sports"
	"golang.org/x/net/context"
)

type Sports interface {
	// ListEvents will return a collection of races.
	ListEvents(ctx context.Context, in *sports.ListEventsRequest) (*sports.ListEventsResponse, error)
}

// sportsService implements the Racing interface.
type sportsService struct {
	sportsRepo db.SportsRepo
}

// NewSportsService instantiates and returns a new racingService.
func NewSportsService(sportsRepo db.SportsRepo) Sports {
	return &sportsService{sportsRepo}
}

func (s *sportsService) ListEvents(ctx context.Context, in *sports.ListEventsRequest) (*sports.ListEventsResponse, error) {
	events, err := s.sportsRepo.List(in.Filter)
	if err != nil {
		return nil, err
	}

	return &sports.ListEventsResponse{Events: events}, nil
}
