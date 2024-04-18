package cormorant

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
	runningChan                    chan bool
	readyHandlerRemove             func()
	interactionCreateHandlerRemove func()
	botID                          string
	guildChannels                  map[string]string
	rebootRequested                int
}

func (this *DiscordUI) RebootRequested() int {
	return this.rebootRequested
}

func NewDiscordUI(appID string, botToken string, runningChan chan bool) DiscordUI {
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

	// Options has to be a []*ApplicationCommandOption instead of []ApplicationCommandOption, possibly in case multiple commands have some of the same parameters.
	// There's a 100 character length limit on descriptions.
	colorCmdOptions := []*discordgo.ApplicationCommandOption{{Type: discordgo.ApplicationCommandOptionString,
		Name: ColorOption, Description: "A three or six digit hexadecimal color code", ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText},
		Required: true, Autocomplete: true, MinLength: &minColorLen, MaxLength: maxColorLen}}
	colorCmd := &discordgo.ApplicationCommand{ID: ColorCmd, ApplicationID: this.appID, Type: discordgo.ChatApplicationCommand, Name: ColorCmd,
		Description: "Change your name's color to a 3 or 6 digit hex color code, such as FF0 or FFFF00 for yellow.", Options: colorCmdOptions}
	roleOption := &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionRole,
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

	err = discord.Open()
	this.session.ShouldReconnectOnError = true

	if err != nil {
		panic(fmt.Sprintf("Error opening Discord session: %s", err.Error()))
	}
	fmt.Println("Cormorant is now running. Press ctrl-c to exit.")
}

func (this *DiscordUI) ready(s *discordgo.Session, m *discordgo.Ready) {
	this.botID = m.User.ID
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
		switch cmdData.ID {
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
			respData := &discordgo.InteractionResponseData{Content: "Not yet implemented."}
			resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
			if err := this.session.InteractionRespond(ic.Interaction, resp); err != nil {
				fmt.Printf("Error calling InteractionRespond: %v", err.Error())
			}
		case LeaveCmd:
			respData := &discordgo.InteractionResponseData{Content: "Not yet implemented."}
			resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
			if err := this.session.InteractionRespond(ic.Interaction, resp); err != nil {
				fmt.Printf("Error calling InteractionRespond: %v", err.Error())
			}
		case ShutdownCmd:
			respData := &discordgo.InteractionResponseData{Content: "Shutting down."}
			resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
			if err := this.session.InteractionRespond(ic.Interaction, resp); err != nil {
				fmt.Printf("Error calling InteractionRespond: %v", err.Error())
			}
			this.rebootRequested = 0xdeadbeef
			s.Close()
			this.runningChan <- false
		}
	} else {
		respData := &discordgo.InteractionResponseData{Content: fmt.Sprintf("ic.Type was %v instead of InteractionApplicationCommand.", ic.Type)}
		resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
		if err := this.session.InteractionRespond(ic.Interaction, resp); err != nil {
			fmt.Printf("Error calling InteractionRespond: %v", err.Error())
		}
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
	response := ""
	if err != nil {
		response = "Failed to parse color code."
	} else {
		newColor := int(newColor64)
		b := newColor & 0xff
		g := (newColor >> 8) & 0xff
		r := (newColor >> 16) & 0xff
		y := (r + r + r + b + g + g + g + g) >> 3
		if y < 72 {
			response = "Sorry, that's too dark."
		} else {
			mroles := member.Roles
			groles, err := s.GuildRoles(targGuild)
			if err != nil {
				response = "Failed to retrieve list of guild roles: " + err.Error()
			} else {
				botm, err := s.GuildMember(targGuild, this.botID)
				if err != nil {
					response = fmt.Sprintf("Unable to find guild member %s (the bot) in guild %s", this.botID, targGuild)
				} else {
					botHighRole := -1
					userHighRole := -1
					for i := 0; i < len(botm.Roles); i++ {
						role := this.FindRole(groles, botm.Roles[i])
						if role != nil {
							if role.Position > botHighRole {
								botHighRole = role.Position
							}
						}
					}
					roles := make([]*discordgo.Role, len(mroles))
					for i := 0; i < len(mroles); i++ {
						roles[i] = this.FindRole(groles, mroles[i])
						if roles[i].Position > userHighRole {
							userHighRole = roles[i].Position
						}
					}
					var roleFound *discordgo.Role = nil
					///fmt.Printf("Searching roles for %v.\n", who)
					for i := 0; i < len(roles); i++ {
						if strings.Contains(roles[i].Name, member.User.ID) {
							roleFound = roles[i]
							//change color of existing role
							hoist := roles[i].Hoist
							permissions := roles[i].Permissions
							mentionable := roles[i].Mentionable
							roleParams := &discordgo.RoleParams{Name: roles[i].Name, Color: &newColor, Hoist: &hoist, Permissions: &permissions, Mentionable: &mentionable}
							_, err = s.GuildRoleEdit(targGuild, roles[i].ID, roleParams)
							if err != nil {
								response = fmt.Sprintf("Failed to change role color: %s", err.Error())
								break
							} else {
								response = "Role color changed successfully."
							}
						}
					}
					if roleFound == nil {
						//fmt.Printf("Did not find it. Creating new role.\n")
						hoist := false
						var permissions int64 = 0
						mentionable := false
						roleParams := &discordgo.RoleParams{Name: member.User.ID, Color: &newColor, Hoist: &hoist, Permissions: &permissions, Mentionable: &mentionable}
						role, err := s.GuildRoleCreate(targGuild, roleParams)
						roleFound = role
						if err != nil {
							response = "I don't think we have the manage roles permission."
						} else {
							err = s.GuildMemberRoleAdd(targGuild, member.User.ID, role.ID)
							if err != nil {
								response = fmt.Sprintf("Failed to assign new role %s: %s", role.Name, err.Error())
							} else {
								//Sorting guild roles is far more complicated than removing people from them,
								//but also safer in that if another important role is below the bot's role, it
								//won't take people out of that role when they request a color change.

								groles, err = s.GuildRoles(targGuild)
								if err != nil {
									response = fmt.Sprintf("Failed to assign new role %s: %s", role.Name, err.Error())
								} else {
									//for i := 0; i < len(groles); i++ {
									//	fmt.Printf("After adding new role, role %s is at position %v.\n", groles[i].Name, groles[i].Position)
									//}
									botHighRole = -1
									for i := 0; i < len(botm.Roles); i++ {
										r := this.FindRole(groles, botm.Roles[i])
										if r != nil {
											if r.Position > botHighRole {
												botHighRole = r.Position
											}
										}
									}
									for i := 0; i < len(groles); i++ {
										if groles[i].ID == role.ID {
											groles[i].Position = botHighRole
										} else if groles[i].Position >= botHighRole {
											groles[i].Position += 1
										}
									}
									sort.Sort(discordgo.Roles(groles))
									//for i := 0; i < len(groles); i++ {
									//	fmt.Printf("After sorting, %s is at position %v.\n", groles[i].Name, groles[i].Position)
									//}
									groles, err = s.GuildRoleReorder(targGuild, groles)
									if err != nil {
										response = fmt.Sprintf("Failed to reorder roles: %s", err.Error())
									} else {
										response = "Role color set successfully."
									}
									//for i := 0; i < len(groles); i++ {
									//	fmt.Printf("After reordering, %s is at position %v.\n", groles[i].Name, groles[i].Position)
									//}
								}
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
			fmt.Printf("Error calling InteractionRespond: %v", err.Error())
		}
	}
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

// commandParse returns (isMatch, paramsString, params), where paramsString is everything after the command prefix and
// params is a slice of parameters found after the comamnd prefix.
// isMatch will be false if the beginning of the string is not the command prefix.
func (this *DiscordUI) commandParse(msg string, commandPrefix string) (bool, string, []string) {
	msgLower := strings.ToLower(msg)
	if strings.HasPrefix(msgLower, commandPrefix) {
		paramsStr := strings.TrimSpace(strings.TrimPrefix(msg, commandPrefix))
		arr := strings.Split(paramsStr, " ")
		if len(arr) == 1 && len(arr[0]) == 0 {
			arr = []string{}
		}
		if strings.HasPrefix(paramsStr, "\"") && strings.HasSuffix(paramsStr, "\"") {
			paramsStr = strings.TrimPrefix(paramsStr, "\"")
			paramsStr = strings.TrimSuffix(paramsStr, "\"")
		}
		var arr2 []string = make([]string, 0)
		//find where parameters start with " and end with "
		start := -1
		//fmt.Printf("len(arr)=%v. arr=%v\n", len(arr), arr)
		for i, x := range arr {
			//fmt.Printf("i=%v x=%v\n", i, x)
			if strings.HasPrefix(x, "\"") {
				s := strings.TrimPrefix(x, "\"")
				if strings.HasSuffix(x, "\"") {
					s = strings.TrimSuffix(s, "\"")
					start = -1
				} else {
					start = len(arr2)
				}
				//fmt.Printf("Appending %s to arr2\n", s)
				arr2 = append(arr2, s)
			} else if start != -1 {
				//fmt.Printf("Altering arr2[%v]\n", start)
				arr2[start] += " " + arr[i]
				if strings.HasSuffix(x, "\"") {
					arr2[start] = strings.TrimSuffix(arr2[start], "\"")
					start = -1
				}
			} else {
				//fmt.Printf("Appending(B) %s to arr2\n", arr[i])
				arr2 = append(arr2, arr[i])
			}
		}
		return true, paramsStr, arr2
	} else {
		return false, "", []string{}
	}
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
