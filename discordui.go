package cormorant

/*
	TODO:
	• Strip leading @s from join and leave commands
	• Make sure there isn't a role with the specified name before creating a new one, since right now it can do this if you name a role that it can't add.
	• > Searching roles for Admin.
	  > Error calling join command InteractionRespond: HTTP 400 Bad Request, {"message": "Cannot send an empty message", "code": 50006}
	• Consider setting the names of the color roles to the value of the color instead of the user ID?
	• Don't show @everyone in the autocomplete and don't allow people to try to join or leave it
*/

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const enableRecover = true

type MsgSource struct {
	channel string
	userID  string
}

func Recover() {
	if enableRecover {
		if r := recover(); r != nil {
			fmt.Println("Error in main(): ", r)
		}
	}
}

const (
	ColorCmd    = "color"
	JoinCmd     = "join"
	LeaveCmd    = "leave"
	ShutdownCmd = "shutdown"

	ColorOption = "color"
	GroupOption = "group"
)

type DiscordUI struct {
	session                        *discordgo.Session
	appID                          string
	botToken                       string
	runningChan                    chan int
	readyHandlerRemove             func()
	interactionCreateHandlerRemove func()
	botID                          string
	guildChannels                  map[string]string
	rebootRequested                int
}

func (this *DiscordUI) RebootRequested() int {
	return this.rebootRequested
}

func NewDiscordUI(appID string, botToken string, runningChan chan int) DiscordUI {
	d := DiscordUI{session: nil, appID: appID, botToken: botToken, runningChan: runningChan, readyHandlerRemove: nil,
		botID: "", guildChannels: make(map[string]string)}
	return d
}

func (this *DiscordUI) Run() {
	discord, err := discordgo.New("Bot " + this.botToken)
	if err != nil {
		panic(fmt.Sprintf("Error creating Discord session: %s", err.Error()))
	}
	this.session = discord

	this.readyHandlerRemove = discord.AddHandler(this.ready)
	this.interactionCreateHandlerRemove = discord.AddHandler(this.interactionCreate)

	var minColorLen, maxColorLen int = 3, 7
	var adminPermission int64 = 0x08

	err = discord.Open()
	this.session.ShouldReconnectOnError = true

	if err != nil {
		panic(fmt.Sprintf("Error opening Discord session: %s", err.Error()))
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
	shutdownCmd := &discordgo.ApplicationCommand{ID: ShutdownCmd, ApplicationID: this.appID, Type: discordgo.ChatApplicationCommand, Name: ShutdownCmd,
		Description: "Shut the bot down", DefaultMemberPermissions: &adminPermission} // require administrator permission
	if _, err := discord.ApplicationCommandCreate(this.appID, "", colorCmd); err != nil {
		panic(fmt.Sprintf("Error registering color command: %s", err.Error()))
	}
	if _, err := discord.ApplicationCommandCreate(this.appID, "", joinCmd); err != nil {
		panic(fmt.Sprintf("Error registering join command: %s", err.Error()))
	}
	if _, err := discord.ApplicationCommandCreate(this.appID, "", leaveCmd); err != nil {
		panic(fmt.Sprintf("Error registering leave command: %s", err.Error()))
	}
	if _, err := discord.ApplicationCommandCreate(this.appID, "", shutdownCmd); err != nil {
		panic(fmt.Sprintf("Error registering shutdown command: %s", err.Error()))
	}
	fmt.Println("Commands registered.")
}

func (this *DiscordUI) ready(s *discordgo.Session, m *discordgo.Ready) {
	this.botID = m.User.ID
}

func (this *DiscordUI) hasRolePtr(m *discordgo.Member, role *discordgo.Role) bool {
	for i := 0; i < len(m.Roles); i++ {
		if m.Roles[i] == role.ID {
			return true
		}
	}
	return false
}

func (this *DiscordUI) hasRole(guild, userID, roleID string) bool {
	m, err := this.session.GuildMember(guild, userID)
	if err != nil {
		//this.Error(fmt.Sprintf("hasRole(%s, %s, %s) = %s", guild, userID, roleID, err.Error()))
		return false
	} else {
		for i := 0; i < len(m.Roles); i++ {
			if m.Roles[i] == roleID {
				//this.Error(fmt.Sprintf("hasRole(%s, %s, %s) = true", guild, userID, roleID))
				return true
			}
		}
	}
	//this.Error(fmt.Sprintf("hasRole(%s, %s, %s) = false", guild, userID, roleID))
	return false
}

func (this *DiscordUI) interactionCreate(s *discordgo.Session, ic *discordgo.InteractionCreate) {
	if ic.Type == discordgo.InteractionApplicationCommand {
		cmdData := ic.ApplicationCommandData()
		options := cmdData.Options
		fmt.Printf("InteractionCreate event received. Command is \"%v\".\n", cmdData.Name)
		switch cmdData.Name {
		case ColorCmd:
			if len(options) == 1 && options[0].Name == ColorOption && options[0].Type == discordgo.ApplicationCommandOptionString {
				hexCode := options[0].StringValue()
				// parse the color code which can be 3 ("FFF"), 6 ("FFFFFF"), or 7 ("#FFFFFF") characters
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
		case ShutdownCmd:
			respData := &discordgo.InteractionResponseData{Content: "Shutting down."}
			resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
			if err := this.session.InteractionRespond(ic.Interaction, resp); err != nil {
				fmt.Printf("Error calling ShutdownCmd InteractionRespond: %v\n", err.Error())
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
				if groles, err := s.GuildRoles(ic.GuildID); err != nil {
					fmt.Printf("Error calling GuildRoles(%v): %v\n", ic.GuildID, err.Error())
				} else {
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
				}
				respData := &discordgo.InteractionResponseData{Content: "",
					TTS: false, Embeds: []*discordgo.MessageEmbed{}, Components: []discordgo.MessageComponent{}, Choices: choices}
				resp := &discordgo.InteractionResponse{Type: discordgo.InteractionApplicationCommandAutocompleteResult, Data: respData}
				if err := this.session.InteractionRespond(ic.Interaction, resp); err != nil {
					fmt.Printf("Error calling autocomplete InteractionRespond: %v\n", err.Error())
				}
			}
		}
	} else {
		fmt.Printf("ic.Type was %v instead of InteractionApplicationCommand.\n", ic.Type.String())
		respData := &discordgo.InteractionResponseData{Content: fmt.Sprintf("ic.Type was %v instead of InteractionApplicationCommand.", ic.Type.String()),
			TTS: false, Embeds: []*discordgo.MessageEmbed{}, Components: []discordgo.MessageComponent{}}
		resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
		if err := this.session.InteractionRespond(ic.Interaction, resp); err != nil {
			fmt.Printf("Error calling default InteractionRespond: %v\n", err.Error())
		}
	}
}

func (this *DiscordUI) handleJoinCommand(interaction *discordgo.Interaction, targGuild string, member *discordgo.Member, groupParam string) {
	s := this.session
	response := ""
	mroles := member.Roles
	groles, err := s.GuildRoles(targGuild)
	groupParam = strings.ReplaceAll(groupParam, "@", "")
	if err != nil {
		response = "Failed to retrieve list of guild roles: " + err.Error()
	} else {
		botm, err := s.GuildMember(targGuild, this.botID)
		if err != nil {
			response = fmt.Sprintf("Unable to find guild member %s (the bot) in guild %s", this.botID, targGuild)
		} else {
			botHighRole := this.findBotRole(groles, botm)
			var roleFound *discordgo.Role = nil
			fmt.Printf("Searching roles for %v.\n", groupParam)
			for i := 0; i < len(groles); i++ {
				if groles[i].Name == groupParam {
					roleFound = groles[i]
					break
				}
			}
			var name string
			if len(member.Nick) > 0 {
				name = member.Nick
			} else {
				name = member.User.Username
			}
			if roleFound != nil {
				if this.assignableRole(roleFound, botHighRole) {
					// make sure they don't already have the role
					hasRole := false
					for i := 0; i < len(mroles); i++ {
						if mroles[i] == groupParam {
							hasRole = true
							break
						}
					}
					if !hasRole {
						// assign the user to the specified role
						err = s.GuildMemberRoleAdd(targGuild, member.User.ID, roleFound.ID)
						if err != nil {
							response = fmt.Sprintf("Failed to assign new role %s: %s", roleFound.Name, err.Error())
						} else {
							response = fmt.Sprintf("Added @%v to `@%v`", name, roleFound.Name)
						}
					}
				}
			} else if this.assignableRoleName(groupParam) {
				// create a new role with the specified name
				newColor := 0xe67e22
				hoist := false
				var permissions int64 = 0
				mentionable := true
				roleParams := &discordgo.RoleParams{Name: groupParam, Color: &newColor, Hoist: &hoist, Permissions: &permissions, Mentionable: &mentionable}
				if role, err := s.GuildRoleCreate(targGuild, roleParams); err != nil {
					response = fmt.Sprintf("Failed to create new role %v: %v", groupParam, err.Error())
				} else {
					err = s.GuildMemberRoleAdd(targGuild, member.User.ID, role.ID)
					if err != nil {
						response = fmt.Sprintf("Failed to assign new role %s: %s", role.Name, err.Error())
					} else {
						response = this.sortGuildRoles(groles, s, targGuild, role, botm, true, fmt.Sprintf("Added @%v to new role `@%v`.", name, role.Name))
					}
				}
			} else {
				response = fmt.Sprintf("Unable to assign role: %s", groupParam)
			}
		}
	}
	respData := &discordgo.InteractionResponseData{Content: response,
		TTS: false, Embeds: []*discordgo.MessageEmbed{}, Components: []discordgo.MessageComponent{}}
	resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
	if err := this.session.InteractionRespond(interaction, resp); err != nil {
		fmt.Printf("Error calling join command InteractionRespond: %v\n", err.Error())
	}
}

func (this *DiscordUI) handleLeaveCommand(interaction *discordgo.Interaction, targGuild string, member *discordgo.Member, groupParam string) {
	s := this.session
	response := ""
	mroles := member.Roles
	groles, err := s.GuildRoles(targGuild)
	if err != nil {
		response = "Failed to retrieve list of guild roles: " + err.Error()
	} else {
		botm, err := s.GuildMember(targGuild, this.botID)
		if err != nil {
			response = fmt.Sprintf("Unable to find guild member %s (the bot) in guild %s", this.botID, targGuild)
		} else {
			botHighRole := this.findBotRole(groles, botm)
			var roleFound *discordgo.Role = nil
			fmt.Printf("Searching roles for %v.\n", groupParam)
			for i := 0; i < len(groles); i++ {
				if groles[i].Name == groupParam {
					roleFound = groles[i]
					break
				}
			}
			var name string
			if len(member.Nick) > 0 {
				name = member.Nick
			} else {
				name = member.User.Username
			}
			if roleFound != nil {
				if this.assignableRole(roleFound, botHighRole) {
					// make sure they already have the role
					hasRole := false
					for i := 0; i < len(mroles); i++ {
						if mroles[i] == roleFound.ID {
							hasRole = true
							break
						}
					}
					if hasRole {
						// remove the user from the specified role
						err = s.GuildMemberRoleRemove(targGuild, member.User.ID, roleFound.ID)
						if err != nil {
							response = fmt.Sprintf("Failed to remove role %s: %s", roleFound.Name, err.Error())
						} else {
							response = fmt.Sprintf("Removed @%v from `@%v`", name, roleFound.Name)
						}
					} else {
						response = fmt.Sprintf("You aren't in the role %v!", groupParam)
					}
				}
			} else if this.assignableRoleName(groupParam) {
				response = fmt.Sprintf("Role %v not found!", groupParam)
			}
		}
	}
	respData := &discordgo.InteractionResponseData{Content: response,
		TTS: false, Embeds: []*discordgo.MessageEmbed{}, Components: []discordgo.MessageComponent{}}
	resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
	if err := this.session.InteractionRespond(interaction, resp); err != nil {
		fmt.Printf("Error calling join command InteractionRespond: %v\n", err.Error())
	}
}

func (this *DiscordUI) handleColorCommand(interaction *discordgo.Interaction, targGuild string, member *discordgo.Member, colorParam string) {
	s := this.session
	colorParam, _ = strings.CutPrefix(colorParam, "#")
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
			groles, err := s.GuildRoles(targGuild)
			if err != nil {
				response = "Failed to retrieve list of guild roles: " + err.Error()
			} else {
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
								response = fmt.Sprintf("Failed to change role color: %s", err.Error())
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
								response = fmt.Sprintf("Failed to assign new role %s: %s", role.Name, err.Error())
							} else {
								response = this.sortGuildRoles(groles, s, targGuild, role, botm, false, "Role color set successfully.")
							}
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
			fmt.Printf("Error calling handleColorCommand InteractionRespond: %v\n", err.Error())
		}
	}
}

func (this *DiscordUI) sortGuildRoles(groles []*discordgo.Role, s *discordgo.Session, targGuild string, role *discordgo.Role, botm *discordgo.Member, afterColors bool, success string) string {
	//Sorting guild roles is far more complicated than removing people from them,
	//but also safer in that if another important role is below the bot's role, it
	//won't take people out of that role when they request a color change.
	var err error
	var response string
	fmt.Printf("Sorting guild roles.\n")
	groles, err = s.GuildRoles(targGuild)
	if err != nil {
		response = fmt.Sprintf("Failed to assign new role %s: %s", role.Name, err.Error())
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
			response = fmt.Sprintf("Failed to reorder roles: %s", err.Error())
		} else {
			response = success
		}

	}
	return response
}

func (this *DiscordUI) FindRole(groles []*discordgo.Role, mrole string) *discordgo.Role {
	for i := 0; i < len(groles); i++ {
		if groles[i].ID == mrole {
			return groles[i]
		}
	}
	return nil
}

func (this *DiscordUI) isValidChannel(guild string, ch string) bool {
	xs, err := this.session.GuildChannels(guild)
	if err != nil {
		return false
	}
	for _, x := range xs {
		if x.ID == ch && !(x.Type == discordgo.ChannelTypeDM || x.Type == discordgo.ChannelTypeGroupDM) {
			return true
		}
	}
	return false
}

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

func (this *DiscordUI) send(channel string, message string) {
	_, err := this.session.ChannelMessageSend(channel, message)
	if err != nil {
		fmt.Printf("Error calling ChannelMessageSend(\"%s\", \"%s\"): %s\n", channel, message, err.Error())
	}
}

func (this *DiscordUI) announce(message string) {
	for _, channel := range this.guildChannels {
		this.send(channel, message)
	}
}

func (this *DiscordUI) ExtractUserID(s string) string {
	if strings.HasPrefix(s, "<@") && strings.HasSuffix(s, ">") {
		return strings.TrimSuffix(strings.TrimPrefix(s, "<@"), ">")
	} else {
		return ""
	}
}

func (this *DiscordUI) assignableRole(role *discordgo.Role, botHighRole int) (assignable bool) {
	assignable = false
	if role.Position < botHighRole {
		assignable = this.assignableRoleName(role.Name)
	}
	return
}

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
