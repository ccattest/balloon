package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"tubr/config"
	"tubr/store"
	"tubr/twitch"
)

var (
	BLACKLIST_PATH   = "./blacklist.txt"
	GLOBAL_BLACKLIST = []string{}
	GAME_LIST_PATH   = "./game-list.json"

	tc = twitch.NewClient(http.DefaultClient)
)

func readBlacklist() {
	f, err := os.Open(BLACKLIST_PATH)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		GLOBAL_BLACKLIST = append(GLOBAL_BLACKLIST, strings.Trim(scanner.Text(), " "))
	}
}

func main() {
	cfg := config.FromENV()

	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt)

	ctx := context.Background()

	_, err := store.Open(ctx, cfg.PostgresConfig)
	if err != nil {
		log.Printf("error: failed to connect to database %+v", err)
	}

	_, cancel := context.WithCancel(ctx)

	twitch.ImportGameList()
	readBlacklist()

	var port = ":8080"
	srv := &http.Server{Addr: port}

	go func() {
		<-kill
		fmt.Println("exiting gracefully")
		cancel()
		if err := srv.Shutdown(ctx); err != nil {
			fmt.Printf("error: failed to exit server gracefully %s\n", err)
		}
	}()

	http.HandleFunc("/processHandler", processHandler)
	http.HandleFunc("/topgames", updateGames)
	http.HandleFunc("/", statusCheck)

	fmt.Printf("Starting server on port %s\n", port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		fmt.Println(err)
		return
	}
}
