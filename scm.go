package discordgo_scm

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/gobwas/glob"
)

// Feature is a handler for various events
type Feature struct {
	Type discordgo.InteractionType

	// Handler function for feature interactions
	Handler func(*discordgo.Session, *discordgo.InteractionCreate)

	// ApplicationCommand if Type is discordgo.InteractionApplicationCommand
	// or discordgo.InteractionApplicationCommandAutocomplete.
	// Not needed for Type discordgo.InteractionMessageComponent.
	ApplicationCommand *discordgo.ApplicationCommand

	// CustomID if Type is discordgo.InteractionMessageComponent.
	// It is in glob format. Use "" (zero string) or "*" to match all CustomIDs
	CustomID string

	customIDGlob *glob.Glob
}

type SCM struct {
	Features      []*Feature
	botCommandIDs map[string][]string
}

func NewSCM() *SCM {
	return &SCM{
		Features:      []*Feature{},
		botCommandIDs: map[string][]string{},
	}
}

// AddFeature adds a Feature to the SCM.
func (s *SCM) AddFeature(f *Feature) error {
	g, err := glob.Compile(f.CustomID)
	if err != nil {
		return err
	}
	f.customIDGlob = &g
	s.Features = append(s.Features, f)
	return nil
}

func (s *SCM) AddFeatures(ff []*Feature) error {
	for _, f := range ff {
		if err := s.AddFeature(f); err != nil {
			return err
		}
	}
	return nil
}

// CreateCommands registers any commands (Features with Type discordgo.InteractionApplicationCommand or discordgo.InteractionApplicationCommandAutocomplete) with the API.
// Leave guildID as empty string for global commands.
// Session must already be connected beforehand.
func (s *SCM) CreateCommands(c *discordgo.Session, guildID string) error {
	appID := c.State.User.ID

	if _, ok := s.botCommandIDs[appID]; ok {
		return errors.New("this application has already registered commands")
	}

	var applicationCommands []*discordgo.ApplicationCommand

	for _, f := range s.Features {
		if f.Type == discordgo.InteractionApplicationCommand || f.Type == discordgo.InteractionApplicationCommandAutocomplete {
			applicationCommands = append(applicationCommands, f.ApplicationCommand)
		}
	}

	createdCommands, err := c.ApplicationCommandBulkOverwrite(appID, guildID, applicationCommands)

	if err != nil {
		return err
	}

	createdCommandIDs := []string{}

	for _, cc := range createdCommands {
		createdCommandIDs = append(createdCommandIDs, cc.ID)
	}

	s.botCommandIDs[appID] = createdCommandIDs

	return nil
}

// DeleteCommands deregisters any commands registered using CreateCommands with the API.
func (s *SCM) DeleteCommands(c *discordgo.Session, guildID string) error {
	appID := c.State.User.ID

	for _, ccID := range s.botCommandIDs[appID] {
		if err := c.ApplicationCommandDelete(appID, guildID, ccID); err != nil {
			return err
		}
	}

	return nil
}

func (s *SCM) HandleInteractionCreate(c *discordgo.Session, i *discordgo.InteractionCreate) {
	// Call relevant Feature handlers

	for _, f := range s.Features {
		if f.Type != i.Type {
			continue
		}
		isCustomIDMatch := func() bool {
			customID := ""
			switch i.Type {
			case discordgo.InteractionMessageComponent:
				customID = i.MessageComponentData().CustomID
			case discordgo.InteractionModalSubmit:
				customID = i.ModalSubmitData().CustomID
			default:
				return false
			}
			return f.CustomID == "" || f.customIDGlob != nil && (*f.customIDGlob).Match(customID)
		}
		match := false
		if i.Type == discordgo.InteractionMessageComponent {
			// If this is a MessageComponent interaction, check that the CustomID matches
			if isCustomIDMatch() {
				match = true
			}
		} else if i.Type == discordgo.InteractionApplicationCommand || i.Type == discordgo.InteractionApplicationCommandAutocomplete {
			// If it's a command interaction such as an ApplicationCommand or ApplicationCommandAutocomplete, check name.
			if f.ApplicationCommand.Name == i.ApplicationCommandData().Name {
				match = true
			}
		} else if i.Type == discordgo.InteractionModalSubmit {
			// Modal compares by CustomID
			if isCustomIDMatch() {
				match = true
			}
		}
		if !match {
			continue
		}
		f.Handler(c, i)
	}
}
