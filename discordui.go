package cormorant

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const enableRecover = true

// regexExtractErr contains the regex used to match the HTTP errors returned by discord.
// It's compiled and stored in this var the first time extractErrorMessage is called.
var regexExtractErr *regexp.Regexp

// Constants related to the bot's discord commands and their options
const (
	ColorCmd    = "color"
	JoinCmd     = "join"
	LeaveCmd    = "leave"
	WeatherCmd  = "weather"
	ShutdownCmd = "shutdown"

	ColorOption    = "color"
	GroupOption    = "group"
	LocationOption = "location"
	ForecastOption = "forecast"
)

// DiscordUI holds data needed to interact with Discord.
type DiscordUI struct {
	session     *discordgo.Session           // session is the currently open Discord session
	appID       string                       // appID is the bot's application ID, needed to connect to Discord
	botToken    string                       // botToken is the bot's auth token, needed to connect to Discord
	runningChan chan int                     // runningChan is a channel opened in Main to which we send a number (the exit code to be returned) when we are ready to exit.
	botID       string                       // the bot's User ID, used to look up member information for the bot
	guildRoles  map[string][]*discordgo.Role // The roles for each guild the bot is in
}

// NewDiscordUI returns a new DiscordUI object.
func NewDiscordUI(appID string, botToken string, runningChan chan int) *DiscordUI {
	return &DiscordUI{session: nil, appID: appID, botToken: botToken, runningChan: runningChan,
		botID: ""}
}

// Run registers things and connects to discord.
func (this *DiscordUI) Run() {
	discord, err := discordgo.New("Bot " + this.botToken)
	if err != nil {
		panic(fmt.Sprintf("Error creating Discord session: %s", extractErrorMessage(err)))
	}
	this.session = discord

	discord.AddHandler(this.ready)
	discord.AddHandler(this.interactionCreate)
	discord.AddHandler(this.guildRoleCreateHandler)
	discord.AddHandler(this.guildRoleDeleteHandler)
	discord.AddHandler(this.guildRoleUpdateHandler)

	var minColorLen, maxColorLen int = 3, 7
	var adminPermission int64 = 0x08

	err = discord.Open()
	this.session.ShouldReconnectOnError = true

	if err != nil {
		panic(fmt.Sprintf("Error opening Discord session: %s", extractErrorMessage(err)))
	}
	fmt.Println("Cormorant is now running. Press ctrl-c to exit.")

	fmt.Println("Registering commands.")

	// Options has to be a []*ApplicationCommandOption instead of []ApplicationCommandOption, possibly in case multiple commands have some of the same parameters.
	// There's a 100 character length limit on descriptions.
	colorCmdOptions := []*discordgo.ApplicationCommandOption{{Type: discordgo.ApplicationCommandOptionString,
		Name: ColorOption, Description: "A three or six digit hexadecimal color code", ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText},
		Required: true, Autocomplete: false, MinLength: &minColorLen, MaxLength: maxColorLen}}
	colorCmd := &discordgo.ApplicationCommand{ID: ColorCmd, ApplicationID: this.appID, Type: discordgo.ChatApplicationCommand, Name: ColorCmd,
		Description: "Change your name's color to a 3 or 6 digit hex color code, such as FF0 or FFFF00 for yellow.", Options: colorCmdOptions}
	roleOption := &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionString,
		Name: GroupOption, Description: "The group you want to join or leave", ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText},
		Required: true, Autocomplete: true}
	joinCmdOptions := []*discordgo.ApplicationCommandOption{roleOption}
	joinCmd := &discordgo.ApplicationCommand{ID: JoinCmd, ApplicationID: this.appID, Type: discordgo.ChatApplicationCommand, Name: JoinCmd,
		Description: "Join a group (give yourself a pingable role)", Options: joinCmdOptions}
	leaveCmdOptions := []*discordgo.ApplicationCommandOption{roleOption}
	leaveCmd := &discordgo.ApplicationCommand{ID: LeaveCmd, ApplicationID: this.appID, Type: discordgo.ChatApplicationCommand, Name: LeaveCmd,
		Description: "Leave a group (remove a pingable role from yourself)", Options: leaveCmdOptions}
	weatherCmdOptions := []*discordgo.ApplicationCommandOption{
		{
			Type: discordgo.ApplicationCommandOptionString, Name: LocationOption,
			Description: "A postal code or location name, e.g. Paris, France", ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText},
			Required: true, Autocomplete: false,
		}, {
			Type: discordgo.ApplicationCommandOptionInteger, Name: ForecastOption,
			Description: "Current conditions, today's forecast, or forecast for the next week?", ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText},
			Required: false, Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  "Current",
					Value: 0,
				}, {
					Name:  "Today",
					Value: 1,
				}, {
					Name:  "Week",
					Value: 2,
				},
			},
		},
	}
	weatherCmd := &discordgo.ApplicationCommand{ID: WeatherCmd, ApplicationID: this.appID, Type: discordgo.ChatApplicationCommand, Name: WeatherCmd,
		Description: "Check the weather", Options: weatherCmdOptions}
	shutdownCmd := &discordgo.ApplicationCommand{ID: ShutdownCmd, ApplicationID: this.appID, Type: discordgo.ChatApplicationCommand, Name: ShutdownCmd,
		Description: "Shut the bot down", DefaultMemberPermissions: &adminPermission} // require administrator permission
	if _, err := discord.ApplicationCommandCreate(this.appID, "", colorCmd); err != nil {
		panic(fmt.Sprintf("Error registering color command: %s", extractErrorMessage(err)))
	}
	if _, err := discord.ApplicationCommandCreate(this.appID, "", joinCmd); err != nil {
		panic(fmt.Sprintf("Error registering join command: %s", extractErrorMessage(err)))
	}
	if _, err := discord.ApplicationCommandCreate(this.appID, "", leaveCmd); err != nil {
		panic(fmt.Sprintf("Error registering leave command: %s", extractErrorMessage(err)))
	}
	if _, err := discord.ApplicationCommandCreate(this.appID, "", weatherCmd); err != nil {
		panic(fmt.Sprintf("Error registering weather command: %s", extractErrorMessage(err)))
	}
	if _, err := discord.ApplicationCommandCreate(this.appID, "", shutdownCmd); err != nil {
		panic(fmt.Sprintf("Error registering shutdown command: %s", extractErrorMessage(err)))
	}
	fmt.Println("Commands registered.")
}

// ready is our ready event handler. It retrieves and records all the guild roles in memory so they can be looked up faster.
func (this *DiscordUI) ready(s *discordgo.Session, m *discordgo.Ready) {
	this.botID = m.User.ID
	fmt.Println("Getting all guild roles.")
	this.getAllGuildRoles()
	fmt.Println("Guild roles retrieved.")
}

// hasRolePtr returns true if the specified member has the specified role, or false otherwise.
func (this *DiscordUI) hasRolePtr(m *discordgo.Member, role *discordgo.Role) bool {
	for _, r := range m.Roles {
		if r == role.ID {
			return true
		}
	}
	return false
}

// hasRole looks up the guild member with the specified user ID, and if it is found in the specified guild, returns true if it has the specified role.
// Otherwise, it returns false.
func (this *DiscordUI) hasRole(guild, userID, roleID string) bool {
	m, err := this.session.GuildMember(guild, userID)
	if err != nil {
		return false
	} else {
		for _, r := range m.Roles {
			if r == roleID {
				return true
			}
		}
	}
	return false
}

// interactionCreate is our handler for interactionCreate events. It is called by discordgo when a user uses a command, and for autocomplete on command options.
func (this *DiscordUI) interactionCreate(s *discordgo.Session, ic *discordgo.InteractionCreate) {
	if ic.Type == discordgo.InteractionApplicationCommand {
		cmdData := ic.ApplicationCommandData()
		options := cmdData.Options
		fmt.Printf("InteractionCreate event received. Command is \"%v\".\n", cmdData.Name)
		switch cmdData.Name {
		case ColorCmd:
			if len(options) == 1 && options[0].Name == ColorOption && options[0].Type == discordgo.ApplicationCommandOptionString {
				hexCode := options[0].StringValue()
				// parse the color code which can be 3 ("FFF"), 4 ("#FFF"), 6 ("FFFFFF"), or 7 ("#FFFFFF") characters
				// find user's color role
				// if it doesn't exist, create it and assign the color to it
				// if it does exist, assign the color to the role
				this.handleColorCommand(ic.Interaction, ic.GuildID, ic.Member, hexCode)
			}
		case JoinCmd:
			if len(options) == 1 && options[0].Name == GroupOption && options[0].Type == discordgo.ApplicationCommandOptionString {
				this.handleJoinCommand(ic.Interaction, ic.GuildID, ic.Member, options[0].StringValue())
			}
		case LeaveCmd:
			if len(options) == 1 && options[0].Name == GroupOption && options[0].Type == discordgo.ApplicationCommandOptionString {
				this.handleLeaveCommand(ic.Interaction, ic.GuildID, ic.Member, options[0].StringValue())
			}
		case WeatherCmd:
			if len(options) >= 1 {
				forecast := CurrentForecast
				location := ""
				for _, opt := range options {
					if opt.Name == LocationOption {
						location = opt.StringValue()
					} else if opt.Name == ForecastOption {
						forecast = ForecastType(opt.IntValue())
					}
				}
				var responseStr string
				var err error
				if responseStr, err = Forecast(location, forecast); err != nil {
					responseStr = extractErrorMessage(err)
				}
				respData := &discordgo.InteractionResponseData{Content: responseStr}
				resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
				if err := this.session.InteractionRespond(ic.Interaction, resp); err != nil {
					fmt.Printf("Error calling WeatherCmd InteractionRespond: %v\n", extractErrorMessage(err))
				}
			}
		case ShutdownCmd:
			respData := &discordgo.InteractionResponseData{Content: "Shutting down."}
			resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
			if err := this.session.InteractionRespond(ic.Interaction, resp); err != nil {
				fmt.Printf("Error calling ShutdownCmd InteractionRespond: %v\n", extractErrorMessage(err))
			}
			s.Close()
			this.runningChan <- 0xdeadbeef
		}
	} else if ic.Type == discordgo.InteractionApplicationCommandAutocomplete {
		// when the user is typing in a field that's flagged autocomplete we receive these events and can respond to them with InteractionApplicationCommandAutocompleteResult
		cmdData := ic.ApplicationCommandData()
		options := cmdData.Options
		fmt.Printf("InteractionCreate event received for autocompletion. Command is \"%v\".\n", cmdData.Name)
		switch cmdData.Name {
		case JoinCmd, LeaveCmd:
			if len(options) == 1 && options[0].Name == GroupOption && options[0].Type == discordgo.ApplicationCommandOptionString {
				leaving := cmdData.Name == LeaveCmd
				targGuild := ic.GuildID
				partialName := strings.ToLower(options[0].StringValue())
				// search for partialName in roles which the bot can assign which don't look like a user ID
				var choices []*discordgo.ApplicationCommandOptionChoice = make([]*discordgo.ApplicationCommandOptionChoice, 0)
				groles := this.guildRoles[targGuild]
				if botm, err := s.GuildMember(targGuild, this.botID); err == nil {
					botHighRole := this.findBotRole(groles, botm)
					for _, role := range groles {
						if this.assignableRole(role, botHighRole) && strings.Contains(strings.ToLower(role.Name), partialName) {
							// this is a possible choice
							if !leaving || (leaving && this.hasRolePtr(ic.Member, role)) {
								choices = append(choices, &discordgo.ApplicationCommandOptionChoice{Name: role.Name, Value: role.Name})
							}
						}
					}
				}
				respData := &discordgo.InteractionResponseData{Content: "",
					TTS: false, Embeds: []*discordgo.MessageEmbed{}, Components: []discordgo.MessageComponent{}, Choices: choices}
				resp := &discordgo.InteractionResponse{Type: discordgo.InteractionApplicationCommandAutocompleteResult, Data: respData}
				if err := this.session.InteractionRespond(ic.Interaction, resp); err != nil {
					fmt.Printf("Error calling autocomplete InteractionRespond: %v\n", extractErrorMessage(err))
				}
			}
		}
	} else {
		fmt.Printf("ic.Type was %v instead of InteractionApplicationCommand.\n", ic.Type.String())
		respData := &discordgo.InteractionResponseData{Content: fmt.Sprintf("ic.Type was %v instead of InteractionApplicationCommand.", ic.Type.String()),
			TTS: false, Embeds: []*discordgo.MessageEmbed{}, Components: []discordgo.MessageComponent{}}
		resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
		if err := this.session.InteractionRespond(ic.Interaction, resp); err != nil {
			fmt.Printf("Error calling default InteractionRespond: %v\n", extractErrorMessage(err))
		}
	}
}

// handleJoinCommand processes the /join command. It determines whether groupParam is a role that exists, determines whether the user should be able to join it, and
// adds them to it (creating it if necessary) if they should be able to join it. It sends back an error message if they shouldn't be able to.
func (this *DiscordUI) handleJoinCommand(interaction *discordgo.Interaction, targGuild string, member *discordgo.Member, groupParam string) {
	s := this.session
	response := ""
	// mroles := member.Roles
	groles := this.guildRoles[targGuild]
	groupParam = strings.TrimPrefix(groupParam, "@")
	botm, err := s.GuildMember(targGuild, this.botID)
	if err != nil {
		response = fmt.Sprintf("Unable to find guild member %s (the bot) in guild %s", this.botID, targGuild)
	} else {
		botHighRole := this.findBotRole(groles, botm)
		var roleFound *discordgo.Role = nil
		fmt.Printf("Searching roles for %v.\n", groupParam)
		roleFound = this.guildHasRole(groupParam, groles)
		if roleFound != nil {
			if this.assignableRole(roleFound, botHighRole) {
				err = s.GuildMemberRoleAdd(targGuild, member.User.ID, roleFound.ID)
				if err != nil {
					response = fmt.Sprintf("Failed to assign new role %s: %s", roleFound.Name, extractErrorMessage(err))
				} else {
					response = fmt.Sprintf("Added %v to `@%v`", member.DisplayName(), roleFound.Name)
				}
			} else {
				response = fmt.Sprintf("Unable to assign role: %s", groupParam)
			}
		} else if this.assignableRoleName(groupParam) {
			// create a new role with the specified name
			newColor := 0xe67e22
			hoist := false
			var permissions int64 = 0
			mentionable := true
			roleParams := &discordgo.RoleParams{Name: groupParam, Color: &newColor, Hoist: &hoist, Permissions: &permissions, Mentionable: &mentionable}
			if role, err := s.GuildRoleCreate(targGuild, roleParams); err != nil {
				response = fmt.Sprintf("Failed to create new role %v: %v", groupParam, extractErrorMessage(err))
			} else {
				err = s.GuildMemberRoleAdd(targGuild, member.User.ID, role.ID)
				if err != nil {
					response = fmt.Sprintf("Failed to assign new role %s: %s", role.Name, extractErrorMessage(err))
				} else {
					response = this.sortGuildRoles(groles, s, targGuild, role, botm, true, fmt.Sprintf("Added %v to new role `@%v`.", member.DisplayName(), role.Name))
				}
			}
		} else {
			response = fmt.Sprintf("Unable to assign role: %s", groupParam)
		}
	}
	respData := &discordgo.InteractionResponseData{Content: response,
		TTS: false, Embeds: []*discordgo.MessageEmbed{}, Components: []discordgo.MessageComponent{}}
	resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
	if err := this.session.InteractionRespond(interaction, resp); err != nil {
		fmt.Printf("Error calling join command InteractionRespond: %v\n", extractErrorMessage(err))
	}
}

// handleLeaveCommand processes the /leave command. It checks whether the member is in a role named groupParam, determines if it can remove them from it,
// and then does so if it can.
func (this *DiscordUI) handleLeaveCommand(interaction *discordgo.Interaction, targGuild string, member *discordgo.Member, groupParam string) {
	s := this.session
	response := ""
	mroles := member.Roles
	groles := this.guildRoles[targGuild]
	groupParam = strings.TrimPrefix(groupParam, "@")
	botm, err := s.GuildMember(targGuild, this.botID)
	if err != nil {
		response = fmt.Sprintf("Unable to find guild member %s (the bot) in guild %s", this.botID, targGuild)
	} else {
		botHighRole := this.findBotRole(groles, botm)
		var roleFound *discordgo.Role = nil
		fmt.Printf("Searching roles for %v.\n", groupParam)
		for _, r := range groles {
			if r.Name == groupParam {
				roleFound = r
				break
			}
		}
		if roleFound != nil {
			if this.assignableRole(roleFound, botHighRole) {
				// make sure they already have the role
				hasRole := false
				for _, r := range mroles {
					if r == roleFound.ID {
						hasRole = true
						break
					}
				}
				if hasRole {
					// remove the user from the specified role
					err = s.GuildMemberRoleRemove(targGuild, member.User.ID, roleFound.ID)
					if err != nil {
						response = fmt.Sprintf("Failed to remove role %s: %s", roleFound.Name, extractErrorMessage(err))
					} else {
						response = fmt.Sprintf("Removed %v from `@%v`", member.DisplayName(), roleFound.Name)
					}
				} else {
					response = fmt.Sprintf("You aren't in the role %v!", groupParam)
				}
			}
		} else if this.assignableRoleName(groupParam) {
			response = fmt.Sprintf("Role %v not found!", groupParam)
		}
	}
	respData := &discordgo.InteractionResponseData{Content: response,
		TTS: false, Embeds: []*discordgo.MessageEmbed{}, Components: []discordgo.MessageComponent{}}
	resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
	if err := this.session.InteractionRespond(interaction, resp); err != nil {
		fmt.Printf("Error calling join command InteractionRespond: %v\n", extractErrorMessage(err))
	}
}

// handleColorCommand processes the /color command. It makes sure colorParam is a valid color code
// (either 3 or 6 characters, all numbers, after removing the # if there is one), and then takes care of the role management stuff
// involved in setting the user's color.
func (this *DiscordUI) handleColorCommand(interaction *discordgo.Interaction, targGuild string, member *discordgo.Member, colorParam string) {
	s := this.session
	colorParam = strings.TrimPrefix(colorParam, "#")
	if len(colorParam) == 3 {
		cpc := []rune(colorParam)
		cpcDbl := []rune{cpc[0], cpc[0], cpc[1], cpc[1], cpc[2], cpc[2]}
		colorParam = string(cpcDbl)
	}
	newColor64, err := strconv.ParseInt(colorParam, 16, 32)
	fmt.Printf("Parsed color param '%v' to %x\n", colorParam, newColor64)
	response := ""
	if err != nil {
		response = "Failed to parse color code."
	} else if len(colorParam) != 6 {
		response = fmt.Sprintf("Color code needs to be either 3 or 6 numbers. We received %v.", len(colorParam))
	} else {
		newColor := int(newColor64)
		b := newColor & 0xff
		g := (newColor >> 8) & 0xff
		r := (newColor >> 16) & 0xff
		y := (r + r + r + b + g + g + g + g) >> 3
		fmt.Printf("Split %x to %x, %x, %x\n", newColor, r, g, b)
		if y < 72 {
			response = "Sorry, that's too dark."
		} else {
			groles := this.guildRoles[targGuild]
			botm, err := s.GuildMember(targGuild, this.botID)
			if err != nil {
				response = fmt.Sprintf("Unable to find guild member %s (the bot) in guild %s", this.botID, targGuild)
			} else {
				var roleFound *discordgo.Role = nil
				fmt.Printf("Searching roles for %v.\n", member.User.ID)
				for i := 0; i < len(groles); i++ {
					if strings.Contains(groles[i].Name, member.User.ID) {
						roleFound = groles[i]
						//change color of existing role
						hoist := groles[i].Hoist
						permissions := groles[i].Permissions
						mentionable := groles[i].Mentionable
						fmt.Printf("Found it.")
						roleParams := &discordgo.RoleParams{Name: groles[i].Name, Color: &newColor, Hoist: &hoist, Permissions: &permissions, Mentionable: &mentionable}
						var newRole *discordgo.Role
						newRole, err = s.GuildRoleEdit(targGuild, groles[i].ID, roleParams)
						if err != nil {
							response = fmt.Sprintf("Failed to change role color: %s", extractErrorMessage(err))
							break
						} else {
							groles[i] = newRole
							response = "Role color changed successfully."
						}
					}
				}
				if roleFound == nil {
					fmt.Printf("Did not find it. Creating new role.\n")
					hoist := false
					var permissions int64 = 0
					mentionable := false
					roleParams := &discordgo.RoleParams{Name: member.User.ID, Color: &newColor, Hoist: &hoist, Permissions: &permissions, Mentionable: &mentionable}
					role, err := s.GuildRoleCreate(targGuild, roleParams)
					roleFound = role
					if err != nil {
						response = "I don't think we have the manage roles permission."
					} else {
						fmt.Printf("Role created. Assigning it to the user now.\n")
						err = s.GuildMemberRoleAdd(targGuild, member.User.ID, role.ID)
						if err != nil {
							response = fmt.Sprintf("Failed to assign new role %s: %s", role.Name, extractErrorMessage(err))
						} else {
							response = this.sortGuildRoles(groles, s, targGuild, role, botm, false, "Role color set successfully.")
						}
					}
				}
			}
		}
	}
	if len(response) > 0 {
		respData := &discordgo.InteractionResponseData{Content: response}
		resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
		if err := this.session.InteractionRespond(interaction, resp); err != nil {
			fmt.Printf("Error calling handleColorCommand InteractionRespond: %v\n", extractErrorMessage(err))
		}
	}
}

// sortGuildRoles sorts the new role into the specified guild's roles. botm is the bot's *Member,
// afterColors indicates whether the role should go after the colors (if false it goes after the bot's role instead),
// success is the string to return as a response if the guild roles are sorted successfully.
func (this *DiscordUI) sortGuildRoles(groles []*discordgo.Role, s *discordgo.Session, targGuild string, role *discordgo.Role, botm *discordgo.Member, afterColors bool, success string) string {
	//Sorting guild roles is far more complicated than removing people from them,
	//but also safer in that if another important role is below the bot's role, it
	//won't take people out of that role when they request a color change.
	var err error
	var response string
	fmt.Printf("Sorting guild roles.\n")
	groles, err = s.GuildRoles(targGuild)
	if err != nil {
		response = fmt.Sprintf("Error calling s.GuildRoles(\"%s\"): %s", targGuild, extractErrorMessage(err))
	} else {
		var setRoleTo int
		if afterColors {
			setRoleTo = 1
		} else {
			setRoleTo = this.findBotRole(groles, botm)
		}
		for i := 0; i < len(groles); i++ {
			if groles[i].ID == role.ID {
				groles[i].Position = setRoleTo
			} else if groles[i].Position >= setRoleTo {
				groles[i].Position += 1
			}
		}
		sort.Sort(discordgo.Roles(groles))

		groles, err = s.GuildRoleReorder(targGuild, groles)
		if err != nil {
			response = fmt.Sprintf("Failed to reorder roles: %s", extractErrorMessage(err))
		} else {
			response = success
		}
	}
	return response
}

// FindRole finds the role with the ID mrole in groles and returns it, or nil if it isn't found.
func (this *DiscordUI) FindRole(groles []*discordgo.Role, mrole string) *discordgo.Role {
	for i := 0; i < len(groles); i++ {
		if groles[i].ID == mrole {
			return groles[i]
		}
	}
	return nil
}

// findBotRole returns the index in the roles list of the bot's highest role (in terms of position numbers), or -1 if the bot has no roles.
func (this *DiscordUI) findBotRole(groles []*discordgo.Role, botm *discordgo.Member) (botHighRole int) {
	botHighRole = -1
	for i := 0; i < len(botm.Roles); i++ {
		role := this.FindRole(groles, botm.Roles[i])
		if role != nil {
			if role.Position > botHighRole {
				botHighRole = role.Position
			}
		}
	}
	return
}

// assignableRole returns whether the role should be assignable by the bot.
// This is done by checking if the role is both below the bot in the role list and has no permissions granted, and by calling assignableRoleName.
func (this *DiscordUI) assignableRole(role *discordgo.Role, botHighRole int) (assignable bool) {
	assignable = false
	if (role.Position < botHighRole) && (role.Permissions == 0) {
		assignable = this.assignableRoleName(role.Name)
	}
	return
}

// assignableRoleName checks whether the specified roleName is a name which can be used for a group role. It makes sure the role name isn't only digits,
// because that's used for color roles, and that the name isn't "everyone". It doesn't make sure the role doesn't already exist - that is checked elsewhere.
func (this *DiscordUI) assignableRoleName(roleName string) (assignable bool) {
	assignable = false
	onlyDigits := true
	for _, r := range roleName {
		if r < '0' || r > '9' {
			onlyDigits = false
		}
	}
	if !onlyDigits {
		assignable = true
	}
	if strings.ToLower(roleName) == "everyone" {
		assignable = false
	}
	return
}

// getAllGuildRoles gets all of the roles from the servers the bot is part of.
// This will need to be updated if the bot is ever in more than 200 servers,
// because that is the max.
func (this *DiscordUI) getAllGuildRoles() {
	joinedGuilds := this.session.State.Guilds
	guildRoles := make(map[string][]*discordgo.Role)
	for _, g := range joinedGuilds {
		guild, err := this.session.Guild(g.ID)
		if err != nil {
			fmt.Printf("Error calling getGuildRoles: %s", extractErrorMessage(err))
			break
		} else {
			roles := guild.Roles
			guildRoles[g.ID] = roles
		}
	}
	this.guildRoles = guildRoles
}

// updateGuildRoles updates the guild roles for the guild with the specified guild ID.
func (this *DiscordUI) updateGuildRoles(guildID string) {
	guild, err := this.session.Guild(guildID)
	if err == nil {
		this.guildRoles[guildID] = guild.Roles
	} else {
		fmt.Printf("Error calling updateGuildRoles: %s\n", extractErrorMessage(err))
	}
}

// guildRoleCreateHandler is called when a new role is created in a guild the bot is in, and calls updateGuildRoles.
func (this *DiscordUI) guildRoleCreateHandler(_ *discordgo.Session, event *discordgo.GuildRoleCreate) {
	gid := event.GuildRole.GuildID
	this.updateGuildRoles(gid)
	fmt.Printf("Role creation detected. Updated roles for ID: %s\n", gid)
}

// guildRoleDeleteHandler is called when an existing role is deleted in a guild the bot is in, and calls updateGuildRoles.
func (this *DiscordUI) guildRoleDeleteHandler(_ *discordgo.Session, event *discordgo.GuildRoleDelete) {
	gid := event.GuildID
	this.updateGuildRoles(gid)
	fmt.Printf("Role delete detected. Updated roles for ID: %s\n", gid)
}

// guildRoleUpdateHandler is called when an existing role is updated in a guild the bot is in, and calls updateGuildRoles.
func (this *DiscordUI) guildRoleUpdateHandler(_ *discordgo.Session, event *discordgo.GuildRoleUpdate) {
	gid := event.GuildRole.GuildID
	this.updateGuildRoles(gid)
	fmt.Printf("Role update detected. Updated roles for ID: %s\n", gid)
}

// guildHasRole checks if there is a role in roles whose name matches roleName case-insensitively, and if so, it returns that role.
// If not, it returns nil.
func (this *DiscordUI) guildHasRole(roleName string, roles []*discordgo.Role) (role *discordgo.Role) {
	role = nil
	roleName = strings.ToLower(roleName)
	for _, r := range roles {
		if strings.ToLower(r.Name) == roleName {
			role = r
			break
		}
	}
	return
}

// extractErrorMessage extracts the error message from a HTTP error returned by discord, or, if it's not a HTTP error, just returns the error string.
func extractErrorMessage(err error) (ret string) {
	ret = err.Error()
	if regexExtractErr == nil {
		var errCompile error
		if regexExtractErr, errCompile = regexp.Compile("HTTP \\d+ [\\w ]+, {\"message\": \"(.+)\","); errCompile != nil {
			ret = fmt.Sprintf("Error compiling error-parsing regex %v to parse error %v.", errCompile.Error(), ret)
		}
	}
	matches := regexExtractErr.FindStringSubmatch(ret)
	if matches != nil && len(matches) > 1 {
		ret = matches[1]
	}
	return
}
