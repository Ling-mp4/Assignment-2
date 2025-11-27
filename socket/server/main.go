package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	//"Assignment-1/internal/mq"
	"Assignment-1/internal/stats"

	"github.com/streadway/amqp"
)

const (
	amqpURL = "amqp://guest:guest@localhost:5672/"
	tcpAddr = ":9000"
)

var (
	clientsMu sync.Mutex
	client    = make(map[net.Conn]struct{})
)

func broadcast(line string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for c := range client {
		_, _ = fmt.Fprintln(c, line)
	}
}

func main() {
	store, err := stats.NewStore("./data/stats.json")
	if err != nil {
		log.Fatalf("stats.NewStore: %v", err)
	}

	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	if err = ch.ExchangeDeclare(
		"game.events",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		log.Fatalf("Failed to declare an exchange: %v", err)
	}

	q, err := ch.QueueDeclare(
		"",
		false,
		true,
		true,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	if err := ch.QueueBind(q.Name, "player.*", "game.events", false, nil); err != nil {
		log.Fatalf("Failed to bind a queue: %v", err)
	}
	if err := ch.QueueBind(q.Name, "game.*", "game.events", false, nil); err != nil {
		log.Fatalf("Failed to bind a queue: %v", err)
	}

	deliveries, err := ch.Consume(q.Name, "", true, true, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}
	ln, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Listening on tcp://%s", tcpAddr)

	go func() {
		for d := range deliveries {
			payload := strings.TrimSpace(string(d.Body))
			line := fmt.Sprintf("[%s] %s", d.RoutingKey, payload)
			log.Println(line)

			if strings.HasPrefix(d.RoutingKey, "game.pangram") || strings.Contains(strings.ToLower(line), "pangram") {
				parts := strings.Fields(payload)
				word := ""
				if len(parts) > 1 {
					word = parts[1]
				}
				_ = store.AddPangram(word)
			}
			broadcast(line)
		}
		log.Printf("RabbitMQ consumer ended")
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go handleClient(conn, store)
	}
}

func handleClient(conn net.Conn, store *stats.Store) {
	//defer conn.Close()
	//clientsMu.Lock()
	//client[conn] = struct{}{}
	//clientsMu.Unlock()
	//defer func() {
	//	clientsMu.Lock()
	//	delete(client, conn)
	//	clientsMu.Unlock()
	//}()
	defer conn.Close()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		cmd := strings.TrimSpace(line)
		switch strings.ToUpper(cmd) {
		case "HELP":
			_, _ = fmt.Fprintln(w, "Commands: GET_STATS, GET_AVG_SCORE, GET_HIGHSCORE, GET_PANGRAMS, QUIT")
			_ = w.Flush()

		case "GET_STATS":
			s := store.Snapshot()
			_, _ = fmt.Fprintf(w, "TotalGames:%d TotalPangrams:%d HighScore:%d GamesWithScores:%d\n", s.TotalGames, s.TotalPangrams, s.HighScore, s.GamesWithScore)
			_ = w.Flush()
		case "GET_AVG_SCORE":
			s := store.Snapshot()
			avg := 0.0
			if s.GamesWithScore > 0 {
				avg = float64(s.TotalScore) / float64(s.GamesWithScore)
			}
			_, _ = fmt.Fprintf(w, "AVG_SCORE: %.2f\n", avg)
			_ = w.Flush()
		case "GET_HIGHSCORE":
			s := store.Snapshot()
			_, _ = fmt.Fprintf(w, "HIGHSCORE: %d\n", s.HighScore)
			_ = w.Flush()
		case "GET_PANGRAMS":
			s := store.Snapshot()
			if len(s.PangramList) == 0 {
				_, _ = fmt.Fprintln(w, "No pangrams yet")
			} else {
				for i, p := range s.PangramList {
					_, _ = fmt.Fprintf(w, "%d: %s\n", i+1, p)
				}
			}
			_ = w.Flush()
		case "QUIT":
			_, _ = fmt.Fprintln(w, "Goodbye")
			_ = w.Flush()
		case "":

		default:
			_, _ = fmt.Fprintln(w, "UNKNOWN_COMMAND - try HELP")
			_ = w.Flush()
		}
	}
}
