package discordgo_scm

import (
	"errors"
	"github.com/bwmarrin/discordgo"
)

type Feature struct {
	Type    discordgo.InteractionType
	Handler func(*discordgo.Session, *discordgo.InteractionCreate)

	// ApplicationCommand if Type is discordgo.InteractionApplicationCommand
	// or discordgo.InteractionApplicationCommandAutocomplete.
	// Not needed for Type discordgo.InteractionMessageComponent.
	ApplicationCommand *discordgo.ApplicationCommand

	// CustomID if Type is discordgo.InteractionMessageComponent.
	// Leave as zero value "" to receive all InteractionMessageComponent
	// interactions.
	CustomID string
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
func (s *SCM) AddFeature(f *Feature) {
	s.Features = append(s.Features, f)
}

func (s *SCM) AddFeatures(ff []*Feature) {
	for _, f := range ff {
		s.AddFeature(f)
	}
}

// CreateCommands registers any commands (Features with Type discordgo.InteractionApplicationCommand or discordgo.InteractionApplicationCommandAutocomplete) with the API.
// Leave guildID as empty string for global commands.
// NOTE: Bot must already be started beforehand.
func (s *SCM) CreateCommands(c *discordgo.Session, guildID string) error {
	appID := c.State.User.ID

	if _, ok := s.botCommandIDs[appID]; ok {
		return errors.New("this application already has registered commands once")
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
	// Find relevant Feature
	var relevantFeature *Feature

	for _, f := range s.Features {
		if f.Type == i.Type {
			if i.Type == discordgo.InteractionMessageComponent ||{
				// If this is a MessageComponent interaction, check that the CustomID matches
				if f.CustomID == i.MessageComponentData().CustomID || f.CustomID == "" {
					relevantFeature = f
					break
				}
			} else if i.Type == discordgo.InteractionApplicationCommand || i.Type == discordgo.InteractionApplicationCommandAutocomplete {
				// If it's a command interaction such as an ApplicationCommand or ApplicationCommandAutocomplete, check name.
				if f.ApplicationCommand.Name == i.ApplicationCommandData().Name {
					relevantFeature = f
					break
				}
			} else if i.Type == discordgo.InteractionModalSubmit {
				// Modal compares by CustomID
				if i.ModalSubmitData().CustomID == f.CustomID || f.CustomID == "" {
					relevantFeature = f
					break
				}
			} else {
				// not sure what to do w this
			}
		}
	}

	// Handle if we have identified a relevant Feature
	if relevantFeature != nil {
		relevantFeature.Handler(c, i)
	}
}
