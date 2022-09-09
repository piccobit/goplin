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

type CliContext struct {
	Debug    bool
	NoHeader bool
}

type ListTagsCmd struct {
	NoHeader       bool   `help:"Do not print header."`
	Fields         string `help:"Show only the specified fields."`
	DuplicatesOnly bool   `name:"duplicates-only" help:"List only duplicate tags."`
	OrphansOnly    bool   `name:"orphans-only" help:"List only orphan tags."`
	OrderBy        string `name:"order-by" help:"Order by specified field."`
	OrderDir       string `name:"order-dir" help:"Order by specified direction: ASC or DESC."`

	IDs []string `arg:"" optional:"" name:"id" help:"List tags with the specified IDs."`
}

type ListNotesCmd struct {
	NoHeader bool   `help:"Do not print header."`
	Fields   string `help:"Show only the specified fields."`
	By       string `name:"by" help:"Find by ID or tag."`
	OrderBy  string `name:"order-by" help:"Order by specified field."`
	OrderDir string `name:"order-dir" help:"Order by specified direction: ASC or DESC."`

	IDs []string `arg:"" optional:"" name:"id" help:"List notes with the specified IDs or tag IDs."`
}

type DeleteTagsCmd struct {
	IDs []string `arg:"" name:"id" help:"Delete tags with the specified IDs."`
}

type DeleteTagFromNoteCmd struct {
	TagID struct {
		TagID string `arg:""`
		From  struct {
			NoteID struct {
				NoteID string `arg:""`
			} `arg:""`
		} `cmd:""`
	} `arg:""`
}

var cli struct {
	Debug bool `help:"Enable debug mode."`

	List struct {
		Tags  ListTagsCmd  `cmd:"" requires:"" help:"List tags."`
		Notes ListNotesCmd `cmd:"" requires:"" help:"List notes."`
	} `cmd:"" help:"Joplin list commands."`

	Delete struct {
		Tags DeleteTagsCmd        `cmd:"" requires:"" help:"Delete tags."`
		Tag  DeleteTagFromNoteCmd `cmd:"" requires:"" help:"Delete tag from note."`
	} `cmd:"" help:"Joplin delete commands."`
}

var (
	client *goplin.Client
)

func getItemTypes() []string {
	return []string{
		"unknown",
		"note",
		"folder",
		"setting",
		"resource",
		"tag",
		"note_tag",
		"search",
		"alarm",
		"master_key",
		"item_change",
		"note_resource",
		"resource_local_state",
		"revision",
		"migration",
		"smart_filter",
		"command",
	}
}

func (cmd *ListTagsCmd) Run(ctx *CliContext) error {
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
						PrintCells(tag, cmd.Fields, &goplin.TagFormats)
					}
				} else {
					PrintCells(tag, cmd.Fields, &goplin.TagFormats)
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
				PrintCells(tag, cmd.Fields, &goplin.TagFormats)
			}
		}
	}

	return nil
}

func (cmd *ListNotesCmd) Run(ctx *CliContext) error {
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
		notes, err := client.GetAllNotes(cmd.Fields, cmd.OrderBy, cmd.OrderDir)
		if err != nil {
			return err
		}

		for _, note := range notes {
			PrintCells(note, cmd.Fields, &goplin.NoteFormats)
		}
	} else {
		if strings.ToLower(cmd.By) == "tag" {
			for _, id := range cmd.IDs {
				notes, err := client.GetNotesByTag(id, cmd.OrderBy, cmd.OrderDir)
				if err != nil {
					fmt.Printf("%-32s <= ERROR: note not found\n", id)
				} else {
					for _, note := range notes {
						PrintCells(note, cmd.Fields, &goplin.NoteFormats)
					}
				}
			}
		} else {
			for _, id := range cmd.IDs {
				note, err := client.GetNote(id, cmd.Fields)
				if err != nil {
					fmt.Printf("%-32s <= ERROR: note not found\n", id)
				} else {
					PrintCells(note, cmd.Fields, &goplin.NoteFormats)
				}

			}
		}
	}

	return nil
}

func (cmd *DeleteTagsCmd) Run(ctx *CliContext) error {
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

func (cmd *DeleteTagFromNoteCmd) Run(ctx *CliContext) error {
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

func PrintCells(cell interface{}, fields string, format *map[string]goplin.CellFormat) {

	columns := strings.Split(fields, ",")

	for i, column := range columns {
		value := reflect.ValueOf(cell)
		cf := (*format)[column]
		if i == 0 {
			fmt.Printf(cf.Format, value.FieldByName(cf.Field))
		} else {
			fmt.Printf(" \u2502 "+cf.Format, value.FieldByName(cf.Field))
		}
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

	ctx := kong.Parse(&cli)
	err = ctx.Run(&CliContext{Debug: cli.Debug})
	ctx.FatalIfErrorf(err)
}
