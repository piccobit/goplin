# Goplin - Golang module to access the Joplin Data API

`Goplin` allows you to access the notebooks, notes, tags & resources stored in your Joplin instance. It is a Golang module and a command line utility at the same time. 

`Goplin` needs a running Joplin instance on your machine to access the data. It does not work directly with your Joplin cloud or self-hosted instance.

## WIP - Work In Progress

`Goplin` is still a work in progress, it doesn't support the complete Joplin API at the moment and the Golang API might be also changed in the future.

## Authorisation

Running `Goplin` the first time it will try to get an authorisation token from your running local Joplin instance. Switching to your local Joplin instance you will see a dialog asking you to grant or deny access to your data. Granting access will return the authorisation token back to `Goplin` and stored in a file called `.goplin` in your home directory. Please keep in mind that the authorisation token is stored unencrypted and anybody with access to this file can retrieve the authorisation token.

## Commands

### Help

Calling `Goplin` with the command line option `-h` or `--help` will provide you with a brief description of the available commands and options:

```shell
$ goplin --help
Usage: goplin <command>

Flags:
  -h, --help     Show context-sensitive help.
      --debug    Enable debug output.

Commands:
  list tags [<id> ...]
    List tags.

  list notes [<id> ...]
    List notes.

  list notebooks [<id> ...]
    List notebooks.

  list resources [<id> ...]
    List resources.

  delete tags <id> ...
    Delete tags.

  delete tag <tag-id> from <note-id>

  search <query>
    Joplin search command.

  create note <title> <body> <notebook> [<tags> ...]
    Create note.

Run "goplin <command> --help" for more information on a command.
```

