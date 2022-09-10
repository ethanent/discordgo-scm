# discordgo-scm

> Slash commands manager for discordgo

[![Go Reference](https://pkg.go.dev/badge/github.com/ethanent/discordgo-scm.svg)](https://pkg.go.dev/github.com/ethanent/discordgo-scm)

## Install

```sh
go get github.com/ethanent/discordgo-scm/v2
```

## Usage

SCM is based around the concept of a [Feature](https://pkg.go.dev/github.com/ethanent/discordgo-scm#Feature). It's meant
to be a somewhat futureproof way to handle all kinds of Discord interactions.

You may create Features for a number of different interaction types, including ApplicationCommand,
ApplicationCommandAutocomplete, and MessageComponent.

| Interaction Type                                    | Relevant Feature Properties       |
|-----------------------------------------------------|-----------------------------------|
| discordgo.InteractionApplicationCommand             | Type, Handler, ApplicationCommand |
| discordgo.InteractionApplicationCommandAutocomplete | Type, Handler, ApplicationCommand |
| discordgo.InteractionMessageComponent               | Type, Handler, CustomID           |

Now, to actually use the library, you must create an SCM and add Features.

Create an SCM:

```go
m := scm.NewSCM()
```

Add a Feature to your SCM:

```go
m.AddFeature(myFeature)
```

Have your SCM handle interactions with a bot:

```go
s.AddHandler(m.HandleInteractionCreate)
```

Register ApplicationCommands with your bot:

```go
// Where s is your discordgo session

err := m.CreateCommands(s, "")
// Please handle your errors :)
```

Delete ApplicationCommands once bot shuts down:

```go
m.DeleteCommands(s, "")
```

See the [godoc](https://pkg.go.dev/github.com/ethanent/discordgo-scm) for full details.
