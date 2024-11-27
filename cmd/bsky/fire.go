package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path"
	"strconv"
	"strings"

	// "github.com/gorilla/websocket"
	// "github.com/reiver/go-bsky/firehose"
	// "bsky/pkg/firehose"
	"github.com/sweetbbak/bsky-cli/firehose"

	"github.com/bluesky-social/indigo/api/atproto"
	comatproto "github.com/bluesky-social/indigo/api/atproto"
	appbsky "github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/repomgr"
	"github.com/whyrusleeping/cbor-gen"
	// "github.com/bluesky-social/indigo/xrpc"
)

func hose() error {
	hose := firehose.New()
	hooks := firehose.Hook{
		Name:      "posts",
		Predicate: firehose.IsPost(),
		Action: func(ctx context.Context, commit *atproto.SyncSubscribeRepos_Commit, op *atproto.SyncSubscribeRepos_RepoOp, record typegen.CBORMarshaler) {

			// var buf bytes.Buffer
			// buf := bytes.NewBuffer()
			// record.MarshalCBOR(&buf)
			// fmt.Println(op.Action, buf.String())
			// fmt.Printf("%s %s %s\n", commit.Time, commit.Repo, buf.String())

			var rep *repo.Repo
			var err error

			if len(commit.Blocks) > 0 {
				rep, err = repo.ReadRepoFromCar(ctx, bytes.NewReader(commit.Blocks))
				if err != nil {
					fmt.Printf("ReadRepoFromCar: %v\n", err)
				}
			}

			ek := repomgr.EventKind(op.Action)
			switch ek {
			case repomgr.EvtKindCreateRecord, repomgr.EvtKindUpdateRecord:
				rc, rec, err := rep.GetRecord(ctx, op.Path)
				if err != nil {
					e := fmt.Errorf("getting record %s (%s) within seq %d for %s: %w", op.Path, *op.Cid, commit.Seq, commit.Repo, err)
					log.Println(e)
				}

				if lexutil.LexLink(rc) != *op.Cid {
					fmt.Printf("mismatch in record and op cid: %s != %s", rc, *op.Cid)
				}

				//fmt.Println("got record", rc, rec)
				banana := lexutil.LexiconTypeDecoder{
					Val: rec,
				}

				var pst = appbsky.FeedPost{}
				b, err := banana.MarshalJSON()
				if err != nil {
					fmt.Println(err)
				}

				err = json.Unmarshal(b, &pst)
				if err != nil {
					fmt.Println(err)
				}

				var userProfile *appbsky.ActorDefs_ProfileViewDetailed
				var replyUserProfile *appbsky.ActorDefs_ProfileViewDetailed

				// userProfile, err = appbsky.ActorGetProfile(context.TODO(), nil, commit.Repo)
				// if err != nil {
				// 	fmt.Println(err)
				// }

				// if pst.Reply != nil {
				// 	replyUserProfile, err = appbsky.ActorGetProfile(context.TODO(), nil, strings.Split(pst.Reply.Parent.Uri, "/")[2])
				// 	if err != nil {
				// 		fmt.Println(err)
				// 	}
				// }

				//Handle if its a post
				if pst.LexiconTypeID == "app.bsky.feed.post" {
					PrintPost(pst, userProfile, replyUserProfile, nil, op.Path)
				} else if pst.LexiconTypeID == "app.bsky.feed.like" && 1 == 0 {
					var like = appbsky.FeedLike{}
					err = json.Unmarshal(b, &like)
					if err != nil {
						fmt.Println(err)
					}

					likedDid := strings.Split(like.Subject.Uri, "/")[2]

					rrb, err := comatproto.SyncGetRepo(ctx, nil, likedDid, "")
					if err != nil {
						fmt.Println(err)
					}

					rr, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(rrb))
					if err != nil {
						fmt.Println(err)
					}

					_, rec, err := rr.GetRecord(ctx, like.Subject.Uri[strings.LastIndex(like.Subject.Uri[:strings.LastIndex(like.Subject.Uri, "/")], "/")+1:])
					if err != nil {
						log.Println(err)
					}

					//fmt.Println("got record", rc, rec)
					banana := lexutil.LexiconTypeDecoder{
						Val: rec,
					}

					var pst = appbsky.FeedPost{}
					b, err := banana.MarshalJSON()
					if err != nil {
						fmt.Println(err)
					}

					err = json.Unmarshal(b, &pst)
					if err != nil {
						fmt.Println(err)
					}

					likedUserProfile, err := appbsky.ActorGetProfile(context.TODO(), nil, likedDid)
					if err != nil {
						fmt.Println(err)
					}

					PrintPost(pst, likedUserProfile, nil, userProfile, like.Subject.Uri[strings.LastIndex(like.Subject.Uri, "/")+1:])
				}
			}
		},
	}

	hose.Hooks = append(hose.Hooks, hooks)
	return hose.Run(context.Background())
}

func PrintPost(pst appbsky.FeedPost, userProfile, replyUserProfile, likingUserProfile *appbsky.ActorDefs_ProfileViewDetailed, postPath string) {
	const minFollowers = 1

	//Try to use the display name and follower count if we can get it
	if userProfile != nil && userProfile.FollowersCount != nil {
		var enoughfollowers bool

		if *userProfile.FollowersCount >= minFollowers {
			enoughfollowers = true
		}

		if likingUserProfile != nil {
			if *likingUserProfile.FollowersCount >= minFollowers {
				enoughfollowers = true
			}

		}

		if enoughfollowers {
			var rply, likedTxt string
			if pst.Reply != nil && replyUserProfile != nil && replyUserProfile.FollowersCount != nil {
				rply = " ➡️ " + replyUserProfile.Handle + ":" + strconv.Itoa(int(*userProfile.FollowersCount)) + "\n" //+ "https://staging.bsky.app/profile/" + strings.Split(pst.Reply.Parent.Uri, "/")[2] + "/post/" + path.Base(pst.Reply.Parent.Uri) + "\n"
			} else if likingUserProfile != nil {
				likedTxt = likingUserProfile.Handle + ":" + strconv.Itoa(int(*likingUserProfile.FollowersCount)) + " ❤️ "
				rply = ":\n"
			} else {
				rply = ":\n"
			}

			url := "https://bsky.app/profile/" + userProfile.Handle + "/post/" + path.Base(postPath)
			fmtdstring := likedTxt + userProfile.Handle + ":" + strconv.Itoa(int(*userProfile.FollowersCount)) + rply + pst.Text + "\n" + url + "\n"
			fmt.Println(fmtdstring)
		}
	} else {
		// fmt.Println(pst.Text)
		// fmt.Printf("\x1b[90m%s:\x1b[0m \x1b[38;2;255;255;255m%s\x1b[0m\n", pst.LexiconTypeID, pst.Text)
		fmt.Printf("\x1b[90m%s:\x1b[0m \x1b[38;2;255;255;255m%s\x1b[0m\n", pst.CreatedAt, pst.Text)
	}
}

// func _hose() error {
//
// 	// The Bluesky Firehose API use WebSockets.
// 	//
// 	// This is the URL that we will connect to it.
// 	const uri string = firehose.WebSocketURI
//
// 	// Connect to the WebSocket.
// 	conn, _, err := websocket.DefaultDialer.Dial(uri, http.Header{})
// 	if nil != err {
// 		fmt.Fprintf(os.Stderr, "ERROR: could not connect to Bluesky Firehose API at %q: %s \n", uri, err)
// 		return err
// 	}
// 	defer conn.Close() // <-- we need to eventually close the WebSocket, so that we don't have a resource leak.
//
// 	// A WebSocket returns a series of messages.
// 	//
// 	// Here we loop, read each message from the WebSocket one-by-one.
// 	for {
// 		// Here we are just getting the raw binary data.
// 		// Later we decode it.
// 		wsMessageType, wsMessage, err := conn.ReadMessage()
// 		if err != nil {
// 			fmt.Fprintf(os.Stderr, "ERROR: problem reading from WebSocket for the connection to the Bluesky Firehose API at %q: %s \n", uri, err)
// 			return err
// 		}
//
// 		// Technically a WebSocket message can either be 'text' message, a 'binary' message, or a few other control messages.
// 		// We expect the Bluesky Firehose API to only return binary messages.
// 		//
// 		// If we receive a message from the WebSocket that is not binary, then we will just ignore it.
// 		if websocket.BinaryMessage != wsMessageType {
// 			continue
// 		}
//
// 		// Here we turn the WebSocket message into a Firehose Message.
// 		var message firehose.Message = firehose.Message(wsMessage)
//
// 		// Here we decode the message.
// 		var header firehose.MessageHeader
// 		var payload firehose.MessagePayload = firehose.MessagePayload{}
//
// 		err = message.Decode(&header, &payload)
// 		if nil != err {
// 			fmt.Fprintf(os.Stderr, "ERROR: problem decoding WebSocket message from the connection to the Bluesky Firehose API at %q: %s \n", uri, err)
// 			continue
// 		}
//
// 		fmt.Printf("%s %s\n", header.Type, payload)
//
// 		//@TODO: Do whatever you want to do with decode message-header and message-payload..
// 	}
// }
