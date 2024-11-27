package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	bskya "github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/urfave/cli/v3"
	"github.com/watzon/lining/client"
	"github.com/watzon/lining/models"
	"github.com/watzon/lining/post"
)

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [Y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if response == "" {
			return true
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" || response == "\n" || response == "" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

func verifyPost(post bskya.FeedPost) bool {
	var fmtPost = `
	Text:   %s
	Tags:   %s
`
	fmt.Printf(fmtPost, post.Text, post.Tags)

	for _, img := range post.Embed.EmbedImages.Images {
		fmt.Printf("%s\n", img.Image.Ref.String())
	}

	return askForConfirmation("proceed to post?")
}

func uploadImage(c *client.BskyClient, img string) (lexutil.LexBlob, models.Image, error) {
	b, err := os.ReadFile(img)
	if err != nil {
		return lexutil.LexBlob{}, models.Image{}, err
	}

	image := models.Image{
		Title: "Example Image",
		Data:  b,
	}

	// Upload the image
	uploadedBlob, err := c.UploadImage(context.Background(), image)
	if err != nil {
		log.Fatal(err)
	}
	return *uploadedBlob, image, nil
}

func Post(cmd *cli.Command) error {
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

	// is false when there are no post contents
	var hasContent bool

	cfg := &client.Config{
		Handle:            handle,
		APIKey:            apikey,
		ServerURL:         server,
		Timeout:           30 * time.Second,
		RequestsPerMinute: 60,
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

	// post builder
	builder := client.NewPostBuilder(post.WithAutoLink(true), post.WithAutoHashtag(true), post.WithAutoMention(true))

	// add link
	if len(cmd.String("link")) != 0 {
		myurl := cmd.String("link")

		// title is the same as link unless specified otherwise
		title := cmd.String("link-title")
		if len(title) == 0 {
			title = myurl
		}

		builder.AddLink(title, myurl)
		hasContent = true
	}

	// add image
	images := cmd.StringSlice("image")
	if len(images) != 0 {
		var blobs []lexutil.LexBlob
		var imagedata []models.Image

		for _, img := range images {
			blob, data, err := uploadImage(c, img)
			if err != nil {
				return fmt.Errorf("error uploading image '%s': %s", img, err)
			}

			blobs = append(blobs, blob)
			imagedata = append(imagedata, data)
		}

		// add our images to the client
		builder.WithImages(blobs, imagedata)
		hasContent = true
	}

	// post text
	if len(cmd.String("text")) != 0 {
		builder.AddText(cmd.String("text"))
		hasContent = true
	}

	// check if this post has any content
	if !hasContent {
		return fmt.Errorf("post must have text, images, or links... none are detected")
	}

	// Create a post with text and link
	post, err := builder.Build()
	if err != nil {
		return err
	}

	if cmd.Bool("confirm") {
		if !verifyPost(post) {
			return fmt.Errorf("post aborted")
		}
	}

	// Create the post
	cid, postUri, err := c.PostToFeed(context.Background(), post)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Post created with CID: %s and URI: %s\n", cid, postUri)

	return nil
}
