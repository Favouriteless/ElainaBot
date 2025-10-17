package discord

const ( // All guild permission bits as constants. These can be ORed
	PermCreateInvite              = 1 << 0
	PermKick                      = 1 << 1
	PermBan                       = 1 << 2
	PermAdministrator             = 1 << 3
	PermManageChannels            = 1 << 4
	PermManageGuilds              = 1 << 5
	PermAddReactions              = 1 << 6
	PermViewAuditLogs             = 1 << 7
	PermPrioritySpeaker           = 1 << 8
	PermStream                    = 1 << 9
	PermViewChannel               = 1 << 10
	PermSendMessages              = 1 << 11
	PermSendTtsMessages           = 1 << 12
	PermManageMessages            = 1 << 13
	PermEmbedLinks                = 1 << 14
	PermAttachFiles               = 1 << 15
	PermReadMessageHistory        = 1 << 16
	PermMentionEveryone           = 1 << 17
	PermExternalEmojis            = 1 << 18
	PermViewGuildInsights         = 1 << 19
	PermConnect                   = 1 << 20
	PermSpeak                     = 1 << 21
	PermMuteMembers               = 1 << 22
	PermDeafenMembers             = 1 << 23
	PermMoveMembers               = 1 << 24
	PermUseVad                    = 1 << 25 // Voice activity detection
	PermChangeNickname            = 1 << 26
	PermManageNicknames           = 1 << 27
	PermManageRoles               = 1 << 28
	PermManageWebhooks            = 1 << 29
	PermManageGuildExpressions    = 1 << 30 // Emojis, stickers and soundboards
	PermUseApplicationCommands    = 1 << 31
	PermRequestToSpeak            = 1 << 32
	PermManageEvents              = 1 << 33
	PermManageThreads             = 1 << 34
	PermCreatePublicThreads       = 1 << 35
	PermCreatePrivateThreads      = 1 << 36
	PermExternalStickers          = 1 << 37
	PermSendInThreads             = 1 << 38
	PermUseEmbeddedActivities     = 1 << 39
	PermModerateMembers           = 1 << 40
	PermViewMonetizationAnalytics = 1 << 41
	PermUseSoundboard             = 1 << 42
	PermCreateGuildExpressions    = 1 << 43 // Emojis, stickers and soundboards
	PermCreateEvents              = 1 << 44
	PermUseExternalSoundboard     = 1 << 45
	PermSendVoiceMessages         = 1 << 46
	PermSendPolls                 = 1 << 49
	PermUseExternalApps           = 1 << 50
	PermPinMessages               = 1 << 51
)

const ( // All gateway intent bits expressed as constants. These can be ORed
	IntentGuilds                 = 1 << 0
	IntentGuildMembers           = 1 << 1
	IntentGuildModeration        = 1 << 2
	IntentGuildExpressions       = 1 << 3
	IntentGuildIntegrations      = 1 << 4
	IntentGuildWebhooks          = 1 << 5
	IntentGuildInvites           = 1 << 6
	IntentGuildVoiceStates       = 1 << 7
	IntentGuildPresences         = 1 << 8
	IntentGuildMessages          = 1 << 9
	IntentGuildMessageReactions  = 1 << 10
	IntentGuildMessageTyping     = 1 << 11
	IntentDirectMessages         = 1 << 12
	IntentDirectMessageReactions = 1 << 13
	IntentDirectMessageTyping    = 1 << 14
	IntentMessageContent         = 1 << 15
	IntentGuildScheduledEvents   = 1 << 16
	IntentAutoModConfig          = 1 << 20
	IntentAutoModExec            = 1 << 21
	IntentGuildMessagePolls      = 1 << 24
	IntentDirectMessagePolls     = 1 << 25
)
