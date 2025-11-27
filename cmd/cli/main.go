package main

import (
	spellingbeepb "Assignment-1/api/spellingbee/v1"
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	client := spellingbeepb.NewGameManagerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cr, err := client.CreateGame(ctx, &spellingbeepb.CreateGameRequest{Kind: "spellingbee"})
	if err != nil {
		panic(err)
	}

	fmt.Println("Spelling Bee!")
	fmt.Println("(Type exit to leave)")

	letters := cr.GetLetters()

	//m := manager.Get()
	//_, g, err := m.Create("spellingbee")
	//if err != nil {
	//	fmt.Println("error:", err)
	//	return
	//}

	//sb, ok := g.(*games.SpellingBee)
	//if !ok {
	//	fmt.Println("Error: couldn't start SpellingBee game", err)
	//}
	score := 0

	for {
		fmt.Printf("\n%s \nEnter Word > ", letters)
		var word string
		if _, err := fmt.Scanln(&word); err != nil {
			fmt.Println(err)
			continue
		}
		word = strings.TrimSpace(word)

		if strings.EqualFold(word, "exit") {
			fmt.Printf("Final score: %d\n", score)
			fmt.Println("Bye!")
			break
		}

		resp, err := client.SubmitWord(context.Background(), &spellingbeepb.SubmitWordRequest{
			Id:   cr.GetId(),
			Word: word,
		})
		if err != nil {
			fmt.Println("Error: ", err)
			continue
		}

		fmt.Println(resp.GetMessage())
		letters = resp.GetLetters()
		score = int(resp.GetScore())
		//valid, _, msg := sb.ValidateWord(word)
		//if !valid {
		//	fmt.Printf("Invalid word!: %s\n", msg)
		//} else {
		//	fmt.Printf(msg)
		//}
	}
}
