package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/autopull"
	"github.com/parkr/auto-reply/comments"
	"github.com/parkr/auto-reply/common"
	"github.com/parkr/auto-reply/deprecate"
	"github.com/parkr/auto-reply/labeler"
)

var (
	client *github.Client

	deprecatedRepos = []deprecate.RepoDeprecation{
		deprecate.RepoDeprecation{
			Nwo:     "jekyll/jekyll-help",
			Message: `This repository is no longer maintained. If you're still experiencing this problem, please search for your issue on [Jekyll Talk](https://talk.jekyllrb.com/), our new community forum. If it isn't there, feel free to post to the Help category and someone will assist you. Thanks!`,
		},
	}
)

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "The port to serve to")
	flag.Parse()
	client = common.NewClient()

	deprecationHandler := deprecate.NewHandler(client, deprecatedRepos)
	http.Handle("/_github/repos/deprecated", deprecationHandler)

	autoPullHandler := autopull.NewHandler(client, []string{"jekyll/jekyll"})
	http.Handle("/_github/repos/autopull", autoPullHandler)

	commentsHandler := comments.NewHandler(client,
		[]comments.CommentHandler{
			comments.HandlerPendingFeedbackLabel,
		},
		[]comments.CommentHandler{
			comments.HandlerPendingFeedbackLabel,
			comments.HandlerMergeAndLabel,
		},
	)
	http.Handle("/_github/repos/comments", commentsHandler)

	labelerHandler := labeler.NewHandler(client,
		[]labeler.PushHandler{},
		[]labeler.PullRequestHandler{
			labeler.PendingRebasePRLabeler,
		},
	)
	http.Handle("/_github/repos/labeler", labelerHandler)

	log.Printf("Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func requireSecret(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Our middleware logic goes here...
		next.ServeHTTP(w, r)
	})
}
