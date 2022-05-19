package main

import (
	"database/sql"
	"flag"
	"git.neds.sh/entain/sports/db"
	"git.neds.sh/entain/sports/proto/sports"
	"git.neds.sh/entain/sports/service"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

var (
	grpcEndpoint = flag.String("grpc-endpoint", "localhost:8500", "gRPC server endpoint")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatalf("failed running grpc server: %s", err)
	}
}

func run() error {
	conn, err := net.Listen("tcp", ":8500")
	if err != nil {
		return err
	}

	eventDB, err := sql.Open("sqlite3", "./db/events.db")
	if err != nil {
		return err
	}

	sportsRepo := db.NewSportsRepo(eventDB)
	if err := sportsRepo.Init(); err != nil {
		return err
	}

	grpcServer := grpc.NewServer()

	sports.RegisterEventsServer(
		grpcServer,
		service.NewSportsService(
			sportsRepo,
		),
	)

	log.Infof("gRPC server listening on: %s", *grpcEndpoint)

	if err := grpcServer.Serve(conn); err != nil {
		return err
	}

	return nil
}
