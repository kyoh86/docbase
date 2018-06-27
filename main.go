package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/kyoh86/go-docbase/docbase"
)

// nolint
var (
	version = "snapshot"
	commit  = "snapshot"
	date    = "snapshot"
)

func main() {
	var token string
	var domainFlag string

	app := kingpin.New("docbase", "A CLI tool to make the docbase more convenience!").Version(version).Author("kyoh86")
	app.Flag("token", "a token to access docbase API. more: https://help.docbase.io/posts/45703#%E3%82%A2%E3%82%AF%E3%82%BB%E3%82%B9%E3%83%88%E3%83%BC%E3%82%AF%E3%83%B3").Envar("DOCBASE_API_TOKEN").Required().StringVar(&token)
	app.Flag("domain", "a domain of docbase team.").Envar("DOCBASE_DOMAIN").Required().StringVar(&domainFlag)

	tagsCMD := app.Command("tag", "manipulate tags").Alias("tags")
	tagsListCMD := tagsCMD.Command("list", "listup tags").Alias("ls")
	tagsEditCMD := tagsCMD.Command("edit", "edit a tag")
	var tagsEditOpt struct {
		Tags map[string]string
	}
	tagsEditOpt.Tags = map[string]string{}
	tagsEditCMD.Arg("tags", "tags map to edit").Required().StringMapVar(&tagsEditOpt.Tags)

	fullCommand := kingpin.MustParse(app.Parse(os.Args[1:]))
	domain := docbase.Domain(domainFlag)
	dbtrans := &docbase.TokenTransport{Token: token}
	dbclient := docbase.NewClient(dbtrans.Client())
	ctx := context.Background()
	switch fullCommand {
	case tagsListCMD.FullCommand():
		tags, _, err := dbclient.Tag.List(ctx, domain, nil)
		if err != nil {
			panic(err)
		}
		for _, t := range tags {
			fmt.Println(t.Name)
		}
	case tagsEditCMD.FullCommand():
		for old, new := range tagsEditOpt.Tags {
			posts, _, err := dbclient.Post.List(ctx, domain, &docbase.PostListOptions{Query: fmt.Sprintf("tag:%s", old)})
			if err != nil {
				panic(err)
			}
			for _, p := range posts {
				edit := docbase.PostEditRequest{}
				for _, t := range p.Tags {
					if old == t.Name {
						edit.Tags = append(edit.Tags, new)
					} else {
						edit.Tags = append(edit.Tags, t.Name)
					}
				}
				_, _, err := dbclient.Post.Edit(ctx, domain, p.ID, &edit)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
