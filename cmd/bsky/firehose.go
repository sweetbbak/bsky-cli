package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/watzon/lining/client"
)

func Firehose(cmd *cli.Command) error {
	// get credentials, CLI flags shadow env vars (and config) conf -> env -> flags
	handle := os.Getenv("BSKY_HANDLE")
	if len(cmd.String("user")) != 0 {
		handle = cmd.String("user")
	}

	apikey := os.Getenv("BSKY_PASS")
	if len(cmd.String("pass")) != 0 {
		apikey = cmd.String("pass")
	}

	server := "https://bsky.social"
	if len(cmd.String("server")) != 0 {
		server = cmd.String("server")
	}

	fireurl := "wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos"
	// server := "wss://bsky.network"

	cfg := &client.Config{
		Handle:            handle,
		APIKey:            apikey,
		ServerURL:         server,
		Timeout:           30 * time.Second,
		RequestsPerMinute: 60,
		FirehoseURL:       fireurl,
		BurstSize:         5,
	}

	// Create a new client
	c, err := client.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Connect to the server
	err = c.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	err = c.SubscribeToFirehose(context.TODO(), &client.FirehoseCallbacks{
		OnCommit: func(evt *client.CommitEvent) error {
			fmt.Println("Event from ", evt.Repo)
			for _, op := range evt.Ops {
				fmt.Printf(" - %s record %s\n", op.Action, op.Path)
			}
			return nil
		},
		OnHandle: func(evt *client.HandleEvent) error {
			fmt.Printf("%s\n", evt)
			return nil
		},
		// OnInfo: func(evt *client.InfoEvent) error {
		// 	fmt.Printf("%s: %s\n", evt.Name, evt.Message)
		// 	return nil
		// },
		// OnMigrate: func(evt *client.MigrateEvent) error {
		// 	fmt.Printf("%s\n", evt)
		// 	return nil
		// },
		// OnTombstone: func(evt *client.TombstoneEvent) error {
		// 	fmt.Printf("%s\n", evt)
		// 	return nil
		// },
	})
	if err != nil {
		log.Fatal(err)
	}

	defer c.CloseFirehose()

	ch := make(chan bool)
	<-ch

	return nil
}
