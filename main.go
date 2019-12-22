package main

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"log"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/alecthomas/template"
	"github.com/kyoh86/go-docbase/v2/docbase"
	"github.com/kyoh86/go-docbase/v2/docbase/postquery"
)

// nolint
var (
	version = "snapshot"
	commit  = "snapshot"
	date    = "snapshot"
)

type subCommand func(context.Context, string, *docbase.Client) error

func main() {
	app := kingpin.New("docbase", "A CLI tool to make the docbase more convenience!").Version(version).Author("kyoh86")
	var token string
	var domain string
	app.Flag("token", "a token to access docbase API. more: https://help.docbase.io/posts/45703#%E3%82%A2%E3%82%AF%E3%82%BB%E3%82%B9%E3%83%88%E3%83%BC%E3%82%AF%E3%83%B3").Envar("DOCBASE_API_TOKEN").Required().StringVar(&token)
	app.Flag("domain", "a domain of docbase team.").Envar("DOCBASE_DOMAIN").Required().StringVar(&domain)

	app.Command("post", "manipulate posts").Alias("posts")
	app.Command("tag", "manipulate tags").Alias("tags")

	cmds := map[string]subCommand{}

	for _, f := range []func(*kingpin.Application) (*kingpin.CmdClause, subCommand){
		postList,
		postSearch,
		postGet,
		tagList,
		tagEdit,
	} {
		key, run := f(app)
		cmds[key.FullCommand()] = run
	}

	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))
	client := docbase.NewAuthClient(domain, token)
	if err := cmds[cmd](context.Background(), domain, client); err != nil {
		log.Fatalf("error: %s", err)
	}
}

func postSearch(app *kingpin.Application) (*kingpin.CmdClause, subCommand) {
	cmd := app.GetCommand("post").Command("search", "search words from posts")
	var opt struct {
		Query  string
		Format string
	}
	// cmd.Flag("format", "format to show a post").Default("{{.Title}}").StringVar(&opt.Format)
	cmd.Arg("query", "searching query").Required().StringVar(&opt.Query)

	return cmd, func(ctx context.Context, domain string, client *docbase.Client) error {
		posts, _, err := client.
			Post.
			List().
			Query(postquery.Join(
				postquery.Title(opt.Query),
				postquery.Or(),
				postquery.Body(opt.Query),
			)).
			Do(ctx)
		if err != nil {
			return err
		}
		for _, p := range posts {
			find(domain, p.ID, 0, p.Title, opt.Query)

			scanner := bufio.NewScanner(strings.NewReader(p.Body))
			for i := 1; scanner.Scan(); i++ {
				text := scanner.Text()
				find(domain, p.ID, i, text, opt.Query)
			}
		}
		return nil
	}
}

func find(domain string, postID docbase.PostID, row int, text, query string) {
	index := strings.Index(text, query)
	if index >= 0 {
		fmt.Printf("%s.docbase.io/posts/%d:%d:%d:%s\n", domain, postID, row, index+1, text)
	}
}

func postList(app *kingpin.Application) (*kingpin.CmdClause, subCommand) {
	cmd := app.GetCommand("post").Command("list", "listup posts")
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

	return cmd, func(ctx context.Context, domain string, client *docbase.Client) error {
		format, err := template.New("post").Parse(opt.Format)
		if err != nil {
			return err
		}
		posts, _, err := client.
			Post.
			List().
			Query(opt.Query).
			Page(int64(opt.Page)).
			PerPage(int64(opt.PerPage)).
			Do(ctx)
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
	}
}

func postGet(app *kingpin.Application) (*kingpin.CmdClause, subCommand) {
	cmd := app.GetCommand("post").Command("get", "get a post")
	var opt struct {
		ID     int64
		Format string
	}
	cmd.Flag("format", "format to show a post").Default("{{.Title}}").StringVar(&opt.Format)
	cmd.Arg("id", "post id").Required().Int64Var(&opt.ID)

	return cmd, func(ctx context.Context, domain string, client *docbase.Client) error {
		format, err := template.New("post").Parse(opt.Format)
		if err != nil {
			return err
		}
		post, _, err := client.Post.Get(docbase.PostID(opt.ID)).Do(ctx)
		if err != nil {
			return err
		}
		if err := format.Execute(os.Stdout, post); err != nil {
			return err
		}
		fmt.Println()
		return nil
	}
}

func tagList(app *kingpin.Application) (*kingpin.CmdClause, subCommand) {
	cmd := app.GetCommand("tag").Command("list", "listup tags").Alias("ls")
	return cmd, func(ctx context.Context, domain string, client *docbase.Client) error {
		tags, _, err := client.Tag.List().Do(ctx)
		if err != nil {
			return err
		}
		for _, t := range tags {
			fmt.Println(t.Name)
		}
		return nil
	}
}

func tagEdit(app *kingpin.Application) (*kingpin.CmdClause, subCommand) {
	cmd := app.GetCommand("tag").Command("edit", "edit a tag")
	var opt struct {
		Tags map[string]string
	}
	opt.Tags = map[string]string{}
	cmd.Arg("tags", "tags map to edit").Required().StringMapVar(&opt.Tags)

	return cmd, func(ctx context.Context, domain string, client *docbase.Client) error {
		for oldOne, newOne := range opt.Tags {
			posts, _, err := client.
				Post.
				List().
				Query(postquery.Tag(oldOne)).
				Do(ctx)
			if err != nil {
				return err
			}
			for _, p := range posts {
				tags := make([]string, 0, len(p.Tags))
				for _, t := range p.Tags {
					if oldOne == t.Name {
						tags = append(tags, newOne)
					} else {
						tags = append(tags, t.Name)
					}
				}
				_, _, err := client.
					Post.
					Edit(p.ID).
					Tags(tags).
					Do(ctx)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}
