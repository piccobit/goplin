package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/imroc/req/v3"
	"github.com/piccobit/goplin"
	"github.com/spf13/viper"
)

type Globals struct {
	Debug bool `help:"Enable debug output."`
}

type ListTagsCmd struct {
	NoHeader       bool   `help:"Do not print header."`
	Fields         string `help:"Show only the specified fields."`
	DuplicatesOnly bool   `name:"duplicates-only" help:"List only duplicate tags."`
	OrphansOnly    bool   `name:"orphans-only" help:"List only orphan tags."`
	OrderBy        string `name:"order-by" help:"Order by specified field."`
	OrderDir       string `name:"order-dir" help:"Order by specified direction: ASC or DESC."`

	IDs []string `arg optional name:"id" help:"List tags with the specified IDs."`
}

type ListNotesCmd struct {
	NoHeader bool   `help:"Do not print header."`
	Fields   string `help:"Show only the specified fields."`
	By       string `name:"by" help:"Find by ID or tag."`
	In       string `name:"in" help:"Find notes in specified notebook"`
	OrderBy  string `name:"order-by" help:"Order by specified field."`
	OrderDir string `name:"order-dir" help:"Order by specified direction: ASC or DESC."`

	IDs []string `arg optional name:"id" help:"List notes with the specified IDs or tag IDs."`
}

type ListNotebooksCmd struct {
	NoHeader bool   `help:"Do not print header."`
	Fields   string `help:"Show only the specified fields."`
	OrderBy  string `name:"order-by" help:"Order by specified field."`
	OrderDir string `name:"order-dir" help:"Order by specified direction: ASC or DESC."`

	IDs []string `arg optional name:"id" help:"List notebooks with the specified IDs or tag IDs."`
}

type DeleteTagsCmd struct {
	IDs []string `arg name:"id" help:"Delete tags with the specified IDs."`
}

type DeleteTagFromNoteCmd struct {
	TagID struct {
		TagID string `arg`
		From  struct {
			NoteID struct {
				NoteID string `arg`
			} `arg`
		} `cmd`
	} `arg`
}

type SearchCmd struct {
	NoHeader bool   `help:"Do not print header."`
	Fields   string `help:"Show only the specified fields."`
	Type     string `help:"Search for specified type"`

	Query string `arg name:"query" help:"Search query (for details see https://joplinapp.org/help/#searching)."`
}

type CreateNoteCmd struct {
	Format string `help:"Format of the new note: Markdown or HTML"`

	Title    string   `arg name:"title" help:"Title of the new note"`
	Body     string   `arg name:"body" help:"Body of the new note"`
	Notebook string   `arg name:"notebook" help:"Name of the notebook to store the note in"`
	Tags     []string `arg optional name:"tags" help:"Tags to attach to the new note"`
}

type CLI struct {
	Globals

	List struct {
		Tags      ListTagsCmd      `cmd requires help:"List tags."`
		Notes     ListNotesCmd     `cmd requires help:"List notes."`
		Notebooks ListNotebooksCmd `cmd requires help:"List notebooks."`
	} `cmd help:"Joplin list commands."`

	Delete struct {
		Tags DeleteTagsCmd        `cmd requires help:"Delete tags."`
		Tag  DeleteTagFromNoteCmd `cmd requires help:"Delete tag from note."`
	} `cmd help:"Joplin delete commands."`

	Search SearchCmd `cmd help:"Joplin search command."`

	Create struct {
		Note CreateNoteCmd `cmd requires help:"Create note."`
	} `cmd help:"Joplin create commands."`
}

var (
	client *goplin.Client
)

func (cmd *ListTagsCmd) Run(ctx *Globals) error {
	if ctx.Debug {
		req.EnableDumpAll()
		req.EnableDebugLog()
	}

	if len(cmd.Fields) == 0 {
		cmd.Fields = "id,parent_id,title"
	}

	if !cmd.NoHeader {
		if !cmd.DuplicatesOnly {
			PrintHeader("Tags", cmd.Fields, &goplin.TagFormats)
		}
	}

	if len(cmd.IDs) == 0 {
		tags, err := client.GetAllTags(cmd.OrderBy, cmd.OrderDir)
		if err != nil {
			return err
		}

		if cmd.DuplicatesOnly {
			if !cmd.NoHeader {
				fmt.Println("Duplicate tags:")
			}

			tagsFound := make(map[string][]string)

			for _, tag := range tags {
				tagsFound[tag.Title] = append(tagsFound[tag.Title], tag.ID)
			}

			duplicatesFound := 0

			for title, ids := range tagsFound {
				if len(ids) > 1 {
					duplicatesFound++

					fmt.Printf("%s:", title)
					for _, id := range ids {
						fmt.Printf(" %s", id)
					}
					fmt.Println()
				}
			}

			if duplicatesFound == 0 {
				fmt.Println("No duplicates found.")
			}
		} else {
			orphansFound := 0

			for _, tag := range tags {
				if cmd.OrphansOnly {
					var notes []goplin.Note
					notes, err = client.GetNotesByTag(tag.ID, cmd.OrderBy, cmd.OrderDir)
					if err != nil {
						continue
					}

					if len(notes) == 0 {
						orphansFound++
						PrintRow(tag, cmd.Fields, &goplin.TagFormats)
					}
				} else {
					PrintRow(tag, cmd.Fields, &goplin.TagFormats)
				}
			}

			if cmd.OrphansOnly {
				if orphansFound == 0 {
					fmt.Println("No orphans found.")
				}
			}
		}
	} else {
		for _, id := range cmd.IDs {
			tag, err := client.GetTag(id, cmd.Fields)
			if err != nil {
				fmt.Printf("%-32s <= ERROR: tag not found\n", id)
			} else {
				PrintRow(tag, cmd.Fields, &goplin.TagFormats)
			}
		}
	}

	return nil
}

func (cmd *ListNotesCmd) Run(ctx *Globals) error {
	var err error
	var notes []goplin.Note

	if ctx.Debug {
		req.EnableDumpAll()
		req.EnableDebugLog()
	}

	if len(cmd.Fields) == 0 {
		cmd.Fields = "id,parent_id,title"
	}

	if !cmd.NoHeader {
		PrintHeader("Notes", cmd.Fields, &goplin.NoteFormats)
	}

	if len(cmd.IDs) == 0 {
		if len(cmd.In) == 0 {
			notes, err = client.GetAllNotes(cmd.Fields, cmd.OrderBy, cmd.OrderDir)
		} else {
			notes, err = client.GetNotesInNotebook(cmd.In, cmd.Fields, cmd.OrderBy, cmd.OrderDir)
		}

		if err != nil {
			return err
		}

		for _, note := range notes {
			PrintRow(note, cmd.Fields, &goplin.NoteFormats)
		}
	} else {
		if strings.ToLower(cmd.By) == "tag" {
			for _, id := range cmd.IDs {
				notes, err := client.GetNotesByTag(id, cmd.OrderBy, cmd.OrderDir)
				if err != nil {
					fmt.Printf("%-32s <= ERROR: note not found\n", id)
				} else {
					for _, note := range notes {
						PrintRow(note, cmd.Fields, &goplin.NoteFormats)
					}
				}
			}
		} else {
			for _, id := range cmd.IDs {
				note, err := client.GetNote(id, cmd.Fields)
				if err != nil {
					fmt.Printf("%-32s <= ERROR: note not found\n", id)
				} else {
					PrintRow(note, cmd.Fields, &goplin.NoteFormats)
				}

			}
		}
	}

	return nil
}

func (cmd *ListNotebooksCmd) Run(ctx *Globals) error {
	if ctx.Debug {
		req.EnableDumpAll()
		req.EnableDebugLog()
	}

	if len(cmd.Fields) == 0 {
		cmd.Fields = "id,parent_id,title"
	}

	if !cmd.NoHeader {
		PrintHeader("Notebooks", cmd.Fields, &goplin.NotebookFormats)
	}

	if len(cmd.IDs) == 0 {
		notebooks, err := client.GetAllNotebooks(cmd.Fields, cmd.OrderBy, cmd.OrderDir)
		if err != nil {
			return err
		}

		for _, notebook := range notebooks {
			PrintRow(notebook, cmd.Fields, &goplin.NoteFormats)
		}
	} else {
		for _, id := range cmd.IDs {
			note, err := client.GetNotebook(id, cmd.Fields)
			if err != nil {
				fmt.Printf("%-32s <= ERROR: notebook not found\n", id)
			} else {
				PrintRow(note, cmd.Fields, &goplin.NoteFormats)
			}

		}
	}

	return nil
}

func (cmd *DeleteTagsCmd) Run(ctx *Globals) error {
	if ctx.Debug {
		req.EnableDumpAll()
		req.EnableDebugLog()
	}

	for _, id := range cmd.IDs {
		err := client.DeleteTag(id)
		if err != nil {
			fmt.Printf("Could not find tag with ID '%s'\n", id)
		} else {
			fmt.Printf("Tag with ID '%s' deleted'\n", id)
		}
	}

	return nil
}

func (cmd *DeleteTagFromNoteCmd) Run(ctx *Globals) error {
	if ctx.Debug {
		req.EnableDumpAll()
		req.EnableDebugLog()
	}

	err := client.DeleteTagFromNote(cmd.TagID.TagID, cmd.TagID.From.NoteID.NoteID)
	if err != nil {
		fmt.Printf("Could not find tag with ID '%s'\n", cmd.TagID)
	} else {
		fmt.Printf("Tag with ID '%s' deleted'\n", cmd.TagID)
	}

	return nil
}

func (cmd *SearchCmd) Run(ctx *Globals) error {
	if ctx.Debug {
		req.EnableDumpAll()
		req.EnableDebugLog()
	}

	if len(cmd.Fields) == 0 {
		cmd.Fields = "id,parent_id,title"
	}

	if !cmd.NoHeader {
		PrintHeader("Search", cmd.Fields, &goplin.SearchFormats)
	}

	items, err := client.Search(cmd.Query, cmd.Type, cmd.Fields)
	if err != nil {
		return fmt.Errorf("could not execute query '%s'\n", cmd.Query)
	}

	for _, item := range items {
		PrintRow(item, cmd.Fields, &goplin.SearchFormats)
	}

	return nil
}

func (cmd *CreateNoteCmd) Run(ctx *Globals) error {
	if ctx.Debug {
		req.EnableDumpAll()
		req.EnableDebugLog()
	}

	format := goplin.Undefined

	switch strings.ToLower(cmd.Format) {
	default:
		fallthrough
	case "markdown":
		format = goplin.Markdown
	case "html":
		format = goplin.HTML
	}

	return client.CreateNote(cmd.Title, format, cmd.Body, cmd.Notebook, cmd.Tags)
}

func PrintHeader(title string, fields string, format *map[string]goplin.CellFormat) {
	fmt.Printf("%s:\n", title)

	columns := strings.Split(fields, ",")

	for i, column := range columns {
		cf := (*format)[column]
		if i == 0 {
			fmt.Printf(cf.Format, cf.Name)
		} else {
			fmt.Printf(" \u2502 "+cf.Format, cf.Name)
		}
	}

	fmt.Println()
}

func PrintRow(cell interface{}, fields string, format *map[string]goplin.CellFormat) {

	columns := strings.Split(fields, ",")

	for i, column := range columns {
		value := reflect.ValueOf(cell)
		cf := (*format)[column]
		vof := value.FieldByName(cf.Field)

		var s string

		if i == 0 {
			s = fmt.Sprintf(cf.Format, vof)
		} else {
			s = fmt.Sprintf(" \u2502 "+cf.Format, vof)
		}

		if vof.Kind() == reflect.String {
			s = strings.TrimSuffix(s, "\n")
		}

		fmt.Printf("%s", s)
	}

	fmt.Println()
}

func main() {
	var err error

	viper.SetDefault("api_token", "")
	viper.SetConfigName(".goplin") // name of config file (without extension)
	viper.SetConfigType("yaml")    // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("$HOME")   // call multiple times to add many search paths
	err = viper.ReadInConfig()     // Find and read the config file
	if err != nil {                // handle errors reading the config file
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
		} else {
			// Config file was found but another error was produced
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	apiToken := viper.GetString("api_token")

	client, err = goplin.New(apiToken)
	if err != nil {
		log.Fatal(err)
	}

	if len(apiToken) == 0 {
		viper.Set("api_token", client.GetApiToken())
		err = viper.WriteConfigAs(path.Join(os.Getenv("HOME"), ".goplin"))
		if err != nil {
			log.Fatal(err)
		}
	}

	cli := CLI{
		Globals: Globals{},
	}

	ctx := kong.Parse(&cli)
	err = ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}
