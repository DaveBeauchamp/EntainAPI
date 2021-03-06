package db

import (
	"database/sql"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"sync"
	"time"

	"git.neds.sh/matty/entain/racing/proto/racing"
)

// RacesRepo provides repository access to races.
type RacesRepo interface {
	// Init will initialise our races repository.
	Init() error

	// List will return a list of races.
	List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error)
	// GetRaceById returns a race from it's ID.
	GetRace(filter *racing.GetRaceByIdRequestFilter) ([]*racing.Race, error)
}

type racesRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewRacesRepo creates a new races repository.
func NewRacesRepo(db *sql.DB) RacesRepo {
	return &racesRepo{db: db}
}

// Init prepares the race repository dummy data.
func (r *racesRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy races.
		err = r.seed()
	})

	return err
}

func (r *racesRepo) List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getRaceQueries()[racesList]

	query, args = r.applyFilter(query, filter)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanRaces(rows)
}

func (r *racesRepo) applyFilter(query string, filter *racing.ListRacesRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	if len(filter.MeetingIds) > 0 {
		clauses = append(clauses, "meeting_id IN ("+strings.Repeat("?,", len(filter.MeetingIds)-1)+"?)")
		for _, meetingID := range filter.MeetingIds {
			args = append(args, meetingID)
		}
	}
	if filter.VisibleRaces == true {
		clauses = append(clauses, " visible = 1; ")
	}

	if strings.ToUpper(filter.OrderBy) == "DESC" {
		clauses = append(clauses, " order by advertised_start_time desc; ")
	} else if strings.ToUpper(filter.OrderBy) == "ASC" {
		clauses = append(clauses, " order by advertised_start_time asc; ")
	}

	if len(clauses) != 0 {
		if filter.OrderBy != "" {
			query += strings.Join(clauses, " AND ")
		} else {
			query += " WHERE " + strings.Join(clauses, " AND ")
		}
	}

	return query, args
}

func (r *racesRepo) GetRace(filter *racing.GetRaceByIdRequestFilter) ([]*racing.Race, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getRaceQueries()[racesList]

	query, args = r.getRaceFilter(query, filter)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanRaces(rows)
}

func (r *racesRepo) getRaceFilter(query string, filter *racing.GetRaceByIdRequestFilter) (string, []interface{}) {
	var (
		args []interface{}
	)

	if filter == nil {
		return query, args
	}

	if &filter.Id != nil {
		var clause = fmt.Sprintf(" where id = %d", filter.Id)
		query += clause
	}

	return query, args
}

func (m *racesRepo) scanRaces(
	rows *sql.Rows,
) ([]*racing.Race, error) {
	var races []*racing.Race

	for rows.Next() {
		var race racing.Race
		var advertisedStart time.Time

		if err := rows.Scan(&race.Id, &race.MeetingId, &race.Name, &race.Number, &race.Visible, &advertisedStart); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		race.AdvertisedStartTime = ts

		var now = time.Now()
		if now.Before(ts.AsTime()) {
			race.Status = "OPEN"
		} else {
			race.Status = "CLOSED"
		}

		races = append(races, &race)
	}

	return races, nil
}
