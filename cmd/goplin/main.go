package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/alecthomas/kong"
	"github.com/spf13/viper"
	"goplin"
)

type Context struct {
	Debug bool
}

type TagsListCmd struct {
}

var (
	client *goplin.Client
)

func (t *TagsListCmd) Run(ctx *Context) error {
	tags, err := client.GetTags()
	if err != nil {
		return err
	}

	fmt.Println("Tags:")

	for _, tag := range tags {
		fmt.Printf("ID: '%s', Parent ID: '%s', Title: '%s'\n", tag.ID, tag.ParentID, tag.Title)
	}

	return nil
}

var cli struct {
	Debug bool `help:"Enable debug mode."`

	Tags struct {
		List TagsListCmd `cmd:"" requires:"" help:"List tags"`
	} `cmd:"" help:"Tag commands"`
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
