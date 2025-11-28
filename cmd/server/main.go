package main

import (
	spellingbeepb "Assignment-1/api/spellingbee/v1"
	"Assignment-1/internal/manager"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"

	"Assignment-1/internal/mq"
	"strings"
)

func connectPublisherWithRetry() (*mq.Publisher, error) {
	amqpURL := os.Getenv("RABBITMQ_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@rabbitmq:5672/"
	}
	var pub *mq.Publisher
	var err error
	backoff := 500 * time.Millisecond
	for tries := 0; tries < 10; tries++ {
		pub, err = mq.Connect(amqpURL)
		if err == nil {
			log.Printf("Connected to RabbitMQ at %s", amqpURL)
			return pub, nil
		}
		log.Printf("RabbitMQ connect failed (try %d): %v — retrying in %v", tries+1, err, backoff)
		time.Sleep(backoff)
		backoff *= 2
		if backoff > 5*time.Second {
			backoff = 5 * time.Second
		}
	}
	return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
}

type server struct {
	spellingbeepb.UnimplementedGameManagerServer
	mgr manager.Manager
	pub *mq.Publisher
}

func (s *server) CreateGame(_ context.Context, req *spellingbeepb.CreateGameRequest) (*spellingbeepb.CreateGameResponse, error) {
	id, game, err := s.mgr.Create(req.GetKind())
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %v", err)
	}

	letters := ""
	if lf, ok := game.(interface{ FormatLetters() string }); ok {
		letters = lf.FormatLetters()
	}

	log.Printf("Game created ID=%s, Kind=%s, Letters=%s", id, game.Name(), letters)
	return &spellingbeepb.CreateGameResponse{
		Id:      id,
		Letters: letters,
	}, nil
}

func (s *server) SubmitWord(_ context.Context, req *spellingbeepb.SubmitWordRequest) (*spellingbeepb.SubmitWordResponse, error) {
	game, ok := s.mgr.Get(req.GetId())
	if !ok {
		return &spellingbeepb.SubmitWordResponse{Message: "Game not found"}, nil
	}

	g, ok := game.(interface {
		ValidateWord(string) (bool, int, string)
		FormatLetters() string
		Name() string
	})
	if !ok {
		return &spellingbeepb.SubmitWordResponse{Message: "Invalid game"}, nil
	}

	_, score, msg := g.ValidateWord(req.GetWord())

	if strings.Contains(msg, "Pangram!") && s.pub != nil {
		eventMsg := fmt.Sprintf("PANGRAM %s score=%d", req.GetWord(), score)
		s.pub.Publish("game.pangram", eventMsg)
	}

	letters := ""
	letters = g.FormatLetters()

	return &spellingbeepb.SubmitWordResponse{
		Message: msg,
		Score:   int32(score),
		Letters: letters,
	}, nil
}

func main() {

	var pub *mq.Publisher

	if os.Getenv("ENABLE_RMQ") == "1" {
		p, err := connectPublisherWithRetry()
		if err != nil {
			log.Fatal(err)
		} else {
			pub = p
			defer pub.Close()
		}
	}

	s := &server{
		mgr: manager.Get(),
		pub: pub,
	}

	fmt.Println("Starting gRPC server on port 50051")

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	//s := &server{mgr: manager.Get()}
	//grpcServer := grpc.NewServer()
	//spellingbeepb.RegisterGameManagerServer(grpcServer, s)
	//
	grpcServer := grpc.NewServer()
	//s := &server{mgr: manager.Get()}
	spellingbeepb.RegisterGameManagerServer(grpcServer, s)

	log.Println("gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
