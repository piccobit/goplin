package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/alecthomas/kong"
	"github.com/imroc/req/v3"
	"github.com/piccobit/goplin"
	"github.com/spf13/viper"
)

type Context struct {
	Debug bool
}

type ListTagsCmd struct {
	ID       string `arg:"" optional:"" name:"id" help:"Tags tag with specified ID"`
	OrderBy  string `name:"order-by" help:"order by specified field"`
	OrderDir string `name:"order-dir" help:"order by specified direction: ASC or DESC"`
}

type ListNotesCmd struct {
	ID       string `arg:"" optional:"" name:"id" help:"Tags notes with specified tag"`
	OrderBy  string `name:"order-by" help:"order by specified field"`
	OrderDir string `name:"order-dir" help:"order by specified direction: ASC or DESC"`
}

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

var cli struct {
	Debug bool `help:"Enable debug mode."`

	List struct {
		Tags  ListTagsCmd  `cmd:"" requires:"" help:"List Joplin tags"`
		Notes ListNotesCmd `cmd:"" requires:"" help:"List notes for specified Joplin tag"`
	} `cmd:"" help:"Joplin list commands"`
}

var (
	client *goplin.Client
)

func (ltc *ListTagsCmd) Run(ctx *Context) error {
	if ctx.Debug {
		req.EnableDumpAll()
		req.EnableDebugLog()
	}

	if len(ltc.ID) == 0 {
		tags, err := client.GetTags(ltc.OrderBy, ltc.OrderDir)
		if err != nil {
			return err
		}

		fmt.Println("Tags:")

		for _, tag := range tags {
			fmt.Printf("ID: '%s', Parent ID: '%s', Title: '%s'\n",
				tag.ID, tag.ParentID, tag.Title)
		}
	} else {
		tag, err := client.GetTag(ltc.ID)
		if err != nil {
			return err
		}

		fmt.Println("Tag:")

		fmt.Printf("ID: '%s', Parent ID: '%s', Title: '%s'\n",
			tag.ID, tag.ParentID, tag.Title)
	}

	return nil
}

func (lnc *ListNotesCmd) Run(ctx *Context) error {
	if ctx.Debug {
		req.EnableDumpAll()
		req.EnableDebugLog()
	}

	if len(lnc.ID) != 0 {
		notes, err := client.GetNote(lnc.ID, lnc.OrderBy, lnc.OrderDir)
		if err != nil {
			return err
		}

		fmt.Println("Notes:")

		for _, note := range notes {
			fmt.Printf("ID: '%s', Parent ID: '%s', Title: '%s'\n",
				note.ID, note.ParentID, note.Title)
		}
	} else {
		notes, err := client.GetNotes(lnc.OrderBy, lnc.OrderDir)
		if err != nil {
			return err
		}

		fmt.Println("Notes:")

		for _, note := range notes {
			fmt.Printf("ID: '%s', Parent ID: '%s', Title: '%s'\n",
				note.ID, note.ParentID, note.Title)
		}
	}

	return nil
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
	err = ctx.Run(&Context{Debug: cli.Debug})
	ctx.FatalIfErrorf(err)
}
