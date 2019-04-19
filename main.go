package main

import (
	"context"
	"fmt"

	"log"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/alecthomas/template"
	"github.com/kyoh86/go-docbase/docbase"
)

// nolint
var (
	version = "snapshot"
	commit  = "snapshot"
	date    = "snapshot"
)

func setConfigFlag(cmd *kingpin.CmdClause, token, domainFlag *string) {
	cmd.Flag("token", "a token to access docbase API. more: https://help.docbase.io/posts/45703#%E3%82%A2%E3%82%AF%E3%82%BB%E3%82%B9%E3%83%88%E3%83%BC%E3%82%AF%E3%83%B3").Envar("DOCBASE_API_TOKEN").Required().StringVar(token)
	cmd.Flag("domain", "a domain of docbase team.").Envar("DOCBASE_DOMAIN").Required().StringVar(domainFlag)
}

func wrapCommand(cmd *kingpin.CmdClause, f func(context.Context, docbase.Domain, *docbase.Client) error) (string, func() error) {
	var token string
	var domain string
	setConfigFlag(cmd, &token, &domain)

	return cmd.FullCommand(), func() error {
		dbtrans := &docbase.TokenTransport{Token: token}
		client := docbase.NewClient(dbtrans.Client())
		return f(context.Background(), docbase.Domain(domain), client)
	}
}

func main() {
	app := kingpin.New("docbase", "A CLI tool to make the docbase more convenience!").Version(version).Author("kyoh86")
	app.Command("post", "manipulate posts").Alias("posts")
	app.Command("tag", "manipulate tags").Alias("tags")

	cmds := map[string]func() error{}

	for _, f := range []func(*kingpin.Application) (string, func() error){} {
		key, run := f(app)
		cmds[key] = run
	}
	if err := cmds[kingpin.MustParse(app.Parse(os.Args[1:]))](); err != nil {
		log.Fatalf("error: %s", err)
	}
}

func postList(app *kingpin.Application) (string, func() error) {
	cmd := app.GetCommand("post").Command("list", "listup posts").Alias("search")
	var opt struct {
		Query   string
		Format  string
		Page    int
		PerPage int
	}
	cmd.Flag("page", "page number").Default("1").IntVar(&opt.Page)
	cmd.Flag("per-page", "number to get per page").Default("20").IntVar(&opt.PerPage)
	cmd.Flag("format", "format to show a post").Default("{{.Title}}").StringVar(&opt.Format)
	cmd.Arg("query", "searching query").Default("*").StringVar(&opt.Query)

	return wrapCommand(cmd, func(ctx context.Context, domain docbase.Domain, client *docbase.Client) error {
		format, err := template.New("post").Parse(opt.Format)
		if err != nil {
			return err
		}
		posts, _, err := client.Post.List(ctx, domain, &docbase.PostListOptions{
			Query: opt.Query,
			ListOptions: docbase.ListOptions{
				Page:    opt.Page,
				PerPage: opt.PerPage,
			},
		})
		if err != nil {
			return err
		}
		for _, p := range posts {
			if err := format.Execute(os.Stdout, p); err != nil {

				return err
			}
			fmt.Println()
		}
		return nil
	})
}

func postGet(app *kingpin.Application) (string, func() error) {
	cmd := app.GetCommand("post").Command("get", "get a post")
	var opt struct {
		ID     int64
		Format string
	}
	cmd.Flag("format", "format to show a post").Default("{{.Title}}").StringVar(&opt.Format)
	cmd.Arg("id", "post id").Required().Int64Var(&opt.ID)

	return wrapCommand(cmd, func(ctx context.Context, domain docbase.Domain, client *docbase.Client) error {
		format, err := template.New("post").Parse(opt.Format)
		if err != nil {
			return err
		}
		post, _, err := client.Post.Get(ctx, domain, docbase.PostID(opt.ID))
		if err != nil {
			return err
		}
		if err := format.Execute(os.Stdout, post); err != nil {
			return err
		}
		fmt.Println()
		return nil
	})
}

func tagList(app *kingpin.Application) (string, func() error) {
	cmd := app.GetCommand("tag").Command("list", "listup tags").Alias("ls")
	return wrapCommand(cmd, func(ctx context.Context, domain docbase.Domain, client *docbase.Client) error {
		tags, _, err := client.Tag.List(ctx, domain, nil)
		if err != nil {
			return err
		}
		for _, t := range tags {
			fmt.Println(t.Name)
		}
		return nil
	})
}

func tagEdit(app *kingpin.Application) (string, func() error) {
	cmd := app.GetCommand("tag").Command("edit", "edit a tag")
	var opt struct {
		Tags map[string]string
	}
	opt.Tags = map[string]string{}
	cmd.Arg("tags", "tags map to edit").Required().StringMapVar(&opt.Tags)

	return wrapCommand(cmd, func(ctx context.Context, domain docbase.Domain, client *docbase.Client) error {
		for old, new := range opt.Tags {
			posts, _, err := client.Post.List(ctx, domain, &docbase.PostListOptions{Query: fmt.Sprintf("tag:%s", old)})
			if err != nil {
				return err
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
				_, _, err := client.Post.Edit(ctx, domain, p.ID, &edit)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}
