package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	addr := "localhost:9000"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	serverReader := bufio.NewReader(conn)
	serverWriter := bufio.NewWriter(conn)

	go func() {
		for {
			line, err := serverReader.ReadString('\n')
			if err != nil {
				fmt.Println("\nDisconnected from server")
				os.Exit(0)
			}
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "[game.") {
				fmt.Printf("\n [EVENT] %s\n", line)
			} else {
				fmt.Printf("%s\n", line)
			}
			fmt.Print("> ")
		}
	}()

	input := bufio.NewScanner(os.Stdin)
	fmt.Println("Type commands (GET_STATS, GET_AVG_SCORE, GET_HIGHSCORE, GET_PANGRAMS, QUIT)")
	fmt.Print("> ")
	for input.Scan() {
		cmd := strings.TrimSpace(input.Text())
		if cmd == "" {
			fmt.Print("> ")
			continue
		}
		if strings.EqualFold(cmd, "quit") {
			return
		}
		if _, err := fmt.Fprintln(serverWriter, cmd); err != nil {
			log.Fatalf("Write: %v", err)
		}
		if err := serverWriter.Flush(); err != nil {
			log.Fatalf("Flush: %v", err)
		}

		fmt.Print("> ")
	}
	if err := input.Err(); err != nil {
		log.Fatalf("stdin: %v", err)
	}
}
