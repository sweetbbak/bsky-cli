package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/urfave/cli/v3"
	// "github.com/watzon/lining/client"
)

const (
	VERSION = "0.1.0"
)

var bsky_app = &cli.Command{
	Name:                  "bsky-cli",
	Usage:                 "a simple cli interface to bsky",
	Description:           "eaily interface with the bsky API on the command line, create and browse posts",
	Version:               fmt.Sprintf("%s %s/%s", VERSION, runtime.GOOS, runtime.GOARCH),
	Copyright:             "Copyright © 2024 sweetbbak MIT",
	Suggest:               true,
	EnableShellCompletion: true,
	// app-global flags
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "color",
			Usage: "enable (on, yes/y, enabled, true) or disable (off, no/n, disabled, false)",
			Value: "toggle",
		},
		&cli.StringFlag{
			Name:  "user",
			Usage: "your bsky username, also use BSKY_HANDLE env var",
		},
		&cli.StringFlag{
			Name:  "pass",
			Usage: "your bsky account password, you can also use BSKY_PASS env var",
		},
		&cli.StringFlag{
			Name:  "server",
			Usage: "bsky server to use, defaults to bsky.social",
		},
	},
	Commands: []*cli.Command{
		{
			Name:        "firehose",
			Description: "bsky firehose API",
			Usage:       "display the bsky firehose, prints every post on the bsky network matching the given pattern",
			ArgsUsage:   " <SEARCH TERM>",
			Aliases:     []string{"fire"},
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "regex",
					Aliases: []string{"r"},
					Usage:   "match posts by regex",
				},
			},
			Action: func(ctx context.Context, c *cli.Command) error {
				// return Firehose(c)
				return hose()
			},
		},
		{
			Name:  "post",
			Usage: "create a post",
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:  "image",
					Usage: "path to a local image (use --link for image links)",
				},
				&cli.StringFlag{
					Name:  "link",
					Usage: "add a link to this post",
				},
				&cli.StringFlag{
					Name:  "link-title",
					Usage: "create a title for a link --link must be specified",
				},
				&cli.StringFlag{
					Name:  "text",
					Usage: "the posts text body",
					// Required: true,
				},
				&cli.BoolFlag{
					Name:  "confirm",
					Usage: "confirm that you want to post after verifying it",
				},
			},
			Action: func(ctx context.Context, c *cli.Command) error {
				return Post(c)
			},
		},
		{
			Name:        "user",
			Description: "retrieve user profiles, follow and unfollow users",
			Usage:       "retrieve user profiles, follow and unfollow users",
			ArgsUsage:   " <USER>",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "follow",
					Usage: "follow the given user",
				},
				&cli.StringFlag{
					Name:  "unfollow",
					Usage: "unfollow the given user",
				},
				&cli.StringFlag{
					Name:  "profile",
					Usage: "retrieve the profile of the given user",
				},
			},
		},
	},
}

func bsky() error {
	if err := bsky_app.Run(context.Background(), os.Args); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := bsky(); err != nil {
		log.Fatal(err)
	}
}
