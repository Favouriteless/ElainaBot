package discord

import (
	"encoding/json"
	"strconv"
)

// StringInt64 is used to cleanly represent the 64-bit ints sent by discord API as they are serialized as strings
type StringInt64 uint64
type Snowflake = StringInt64
type Permissions = StringInt64

func (s *StringInt64) String() string {
	return strconv.FormatUint(uint64(*s), 10)
}

func (s *StringInt64) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	id, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return err
	}
	*s = StringInt64(id)
	return nil
}

func (s *StringInt64) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type ConnectionProperties struct {
	Os      string `json:"os"`
	Browser string `json:"browser"`
	Device  string `json:"device"`
}

// Guild represents https://discord.com/developers/docs/resources/guild#guild-object.
// Owner field is not included as it is impossible for a discord bot to own a guild
type Guild struct {
	Id              Snowflake `json:"id"`
	Name            string    `json:"name"`
	Icon            string    `json:"icon"`             // Nullable
	IconHash        string    `json:"icon_hash"`        // Optional, nullable
	Splash          string    `json:"splash"`           // Nullable
	DiscoverySplash string    `json:"discovery_splash"` // Nullable
	OwnerId         Snowflake `json:"owner_id"`
}

// User represents https://discord.com/developers/docs/resources/guild#guild-object
type User struct {
	Id                   Snowflake             `json:"id"`
	Username             string                `json:"username"`
	Discriminator        string                `json:"discriminator"`
	GlobalName           string                `json:"global_name"` // Nullable
	Avatar               string                `json:"avatar"`
	Bot                  bool                  `json:"bot"`
	System               *bool                 `json:"system"`                 // Optional
	MfaEnabled           *bool                 `json:"mfa_enabled"`            // Optional
	Banner               string                `json:"banner"`                 // Optional, nullable
	AccentColor          *int                  `json:"accent_color"`           // Optional
	Locale               string                `json:"locale"`                 // Optional
	Verified             *bool                 `json:"verified"`               // Optional
	Email                string                `json:"email"`                  // Optional
	PremiumType          *int                  `json:"premium_type"`           // Optional
	AvatarDecorationData *AvatarDecorationData `json:"avatar_decoration_data"` // Optional
	Collectibles         *Collectibles         `json:"collectibles"`           // Optional
	PrimaryGuild         *UserPrimaryGuild     `json:"primary_guild"`          // Optional, nullable
}

// AvatarDecorationData represents https://discord.com/developers/docs/resources/user#avatar-decoration-data-object
type AvatarDecorationData struct {
	Asset string    `json:"asset"`
	SkuId Snowflake `json:"sku_id"`
}

// Collectibles represents https://discord.com/developers/docs/resources/user#collectibles
type Collectibles struct {
	Nameplate *Nameplate `json:"nameplate"`
}

// Nameplate represents https://discord.com/developers/docs/resources/user#nameplate
type Nameplate struct {
	SkuId   Snowflake `json:"sku_id"`
	Asset   string    `json:"asset"`
	Label   string    `json:"label"`
	Palette string    `json:"palette"`
}

// UserPrimaryGuild represents https://discord.com/developers/docs/resources/user#user-object-user-primary-guild
type UserPrimaryGuild struct {
	GuildId *Snowflake `json:"identity_guild_id"` // Nullable
	Enabled *bool      `json:"identity_enabled"`  // Nullable
	Tag     string     `json:"tag"`               // Nullable
	Badge   string     `json:"badge"`             // Nullable
}

type App struct {
	Id          Snowflake `json:"id"`
	Name        string    `json:"name"`
	Icon        string    `json:"icon"` // Nullable
	Description string    `json:"description"`
	Flags       int       `json:"flags,omitempty"` // Optional, not nullable
}

// Message represents https://discord.com/developers/docs/resources/message#message-object
type Message struct {
	Id              Snowflake        `json:"id,omitempty"`
	ChannelId       Snowflake        `json:"channel_id"`
	Author          User             `json:"author,omitempty"`
	Content         string           `json:"content"` // Requires IntentMessageContent
	Timestamp       string           `json:"timestamp,omitempty"`
	EditedTimestamp string           `json:"edited_timestamp,omitempty"` // Nullable
	Tts             bool             `json:"tts"`
	MentionEveryone bool             `json:"mention_everyone"`
	Mentions        []User           `json:"mentions"`
	MentionRoles    []Snowflake      `json:"mention_roles"`
	MentionChannels []ChannelMention `json:"mention_channels,omitempty"`
	Attachments     []Attachment     `json:"attachments"`
	// Embeds
	Reactions []Reaction `json:"reactions,omitempty"` // Optional
	// Nonce (? lol what)
	Pinned    bool      `json:"pinned"`
	WebhookId Snowflake `json:"webhook_id,omitempty"` // Optional
	Type      int       `json:"type"`
	// Activity
	Application       *App              `json:"application,omitempty"`        // Optional
	ApplicationId     Snowflake         `json:"application_id,omitempty"`     // Optional
	Flags             int               `json:"flags,omitempty"`              // Optional
	MessageReference  MessageReference  `json:"message_reference,omitempty"`  // Optional
	MessageSnapshots  []MessageSnapshot `json:"message_snapshots,omitempty"`  // Optional
	ReferencedMessage *Message          `json:"referenced_message,omitempty"` // Optional, Nullable
	// InteractionMetadata
	Thread *Channel `json:"thread,omitempty"` // Optional
	// Components
	StickerItems         []StickerItem         `json:"sticker_items,omitempty"`          // Optional
	Position             *int                  `json:"position,omitempty"`               // Optional
	RoleSubscriptionData *RoleSubscriptionData `json:"role_subscription_data,omitempty"` // Optional
	Resolved             *ResolvedData         `json:"resolved,omitempty"`               // Optional
	Poll                 *Poll                 `json:"poll,omitempty"`                   // Optional
	Call                 *MessageCall          `json:"call,omitempty"`                   // Optional
}

// Role represents https://discord.com/developers/docs/topics/permissions#role-object
type Role struct {
	Id          Snowflake   `json:"id"`
	Name        string      `json:"name"`
	Colors      RoleColors  `json:"colors"`
	Hoist       bool        `json:"hoist"`
	Icon        string      `json:"icon"`          // Optional, nullable
	Emoji       string      `json:"unicode_emoji"` // Optional, nullable
	Position    int         `json:"position"`
	Permissions Permissions `json:"permissions"`
	Managed     bool        `json:"managed"`
	Mentionable bool        `json:"mentionable"`
	Tags        *RoleTags   `json:"tags"` // Optional
	Flags       int         `json:"flags"`
}

// RoleColors represents https://discord.com/developers/docs/topics/permissions#role-object-role-colors-object
type RoleColors struct {
	PrimaryColor   int  `json:"primary_color"`
	SecondaryColor *int `json:"secondary_color"` // Nullable
	TertiaryColor  *int `json:"tertiary_color"`  // Nullable
}

// RoleTags represents https://discord.com/developers/docs/topics/permissions#role-object-role-tags-structure
type RoleTags struct {
	BotId                 *Snowflake `json:"bot_id"`             // Optional
	IntegrationId         *Snowflake `json:"integration_id"`     // Optional
	PremiumSubscriber     *bool      `json:"premium_subscriber"` // This is nil (false) or not nil (true). Stupid API quirk.
	SubscriptionListingId *Snowflake `json:"subscription_listing_id"`
	AvailableForPurchase  *bool      `json:"available_for_purchase"` // This is nil (false) or not nil (true). Stupid API quirk.
	GuildConnections      *bool      `json:"guild_connections"`      // This is nil (false) or not nil (true). Stupid API quirk.
}

// ChannelMention represents https://discord.com/developers/docs/resources/message#channel-mention-object
type ChannelMention struct {
	Id      Snowflake `json:"id"`
	GuildId Snowflake `json:"guild_id"`
	Type    int       `json:"type"`
	Name    string    `json:"name"`
}

// Attachment represents https://discord.com/developers/docs/resources/message#attachment-object
type Attachment struct {
	Id          Snowflake `json:"id"`
	Filename    string    `json:"filename"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ContentType string    `json:"content_type"`
	Size        int       `json:"size"`
	URL         string    `json:"Url"`
	ProxyUrl    string    `json:"proxy_url"`
	Height      *int      `json:"height"`
	Width       *int      `json:"width"`
	Ephemeral   bool      `json:"ephemeral"`
	Duration    *float32  `json:"duration_secs"`
	Waveform    string    `json:"waveform"`
	Flags       int       `json:"flags"`
}

// Reaction represents https://discord.com/developers/docs/resources/message#reaction-object
type Reaction struct {
	Count        int                  `json:"count"`
	CountDetails ReactionCountDetails `json:"count_details"`
	Me           bool                 `json:"me"`
	MeBurst      bool                 `json:"me_burst"`
	Emoji        Emoji                `json:"emoji"`
	BurstColors  []string             `json:"burst_colors"`
}

// ReactionCountDetails represents https://discord.com/developers/docs/resources/message#reaction-count-details-object
type ReactionCountDetails struct {
	Burst  int `json:"burst"`
	Normal int `json:"normal"`
}

// Emoji represents https://discord.com/developers/docs/resources/emoji#emoji-object
type Emoji struct {
	Id            *Snowflake  `json:"id,omitempty"`             // Nullable
	Name          string      `json:"nam,omitempty"`            // Nullable
	Roles         []Snowflake `json:"roles,omitempty"`          // Optional
	User          User        `json:"user,omitempty"`           // Optional
	RequireColons bool        `json:"require_colons,omitempty"` // Optional
	Managed       bool        `json:"managed,omitempty"`        // Optional
	Animated      bool        `json:"animated,omitempty"`       // Optional
	Available     bool        `json:"available,omitempty"`      // Optional
}

// MessageReference represents https://discord.com/developers/docs/resources/message#message-reference-object
type MessageReference struct {
	Type            *int       `json:"type"`                         // Optional. 0 = DEFAULT, 1 = FORWARD, If unset, assume DEFAULT
	MessageId       *Snowflake `json:"message_id"`                   // Optional
	ChannelId       *Snowflake `json:"channel_id"`                   // Optional
	GuildId         *Snowflake `json:"guild_id"`                     // Optional
	FailIfNotExists bool       `json:"fail_if_not_exists,omitempty"` // Optional, send only
}

// MessageSnapshot represents https://discord.com/developers/docs/resources/message#message-snapshot-object
type MessageSnapshot struct {
	Message Message `json:"message"` // Partial obj
}

// Sticker represents https://discord.com/developers/docs/resources/sticker#sticker-object
type Sticker struct {
	Id          Snowflake  `json:"id"`
	PackId      Snowflake  `json:"pack_id"`
	Name        string     `json:"name"` // Optional
	Description string     `json:"description"`
	Tags        string     `json:"tags"`
	Type        int        `json:"type"`        // 1 = STANDARD, 2 = GUILD
	FormatType  int        `json:"format_type"` // 1 = PNG, 2 = APNG, 3 = LOTTIE, 4 = GIF
	Available   *bool      `json:"available"`   // Optional
	GuildId     *Snowflake `json:"guild_id"`    // Optional
	User        *User      `json:"user"`        // Optional
	SortValue   *int       `json:"sort_value"`  // Optional
}

// StickerItem represents https://discord.com/developers/docs/resources/sticker#sticker-item-object
type StickerItem struct {
	Id         Snowflake `json:"id"`
	Name       string    `json:"name"`
	FormatType int       `json:"format_type"` // 1 = PNG, 2 = APNG, 3 = LOTTIE, 4 = GIF
}

// RoleSubscriptionData represents https://discord.com/developers/docs/resources/message#role-subscription-data-object
type RoleSubscriptionData struct {
	RoleSubscriptionListingId Snowflake `json:"role_subscription_listing_id"`
	TierName                  string    `json:"tier_name"`
	TotalMonthsSubscribed     int       `json:"total_months_subscribed"`
	IsRenewal                 bool      `json:"is_renewal"`
}

// MessageCall represents https://discord.com/developers/docs/resources/message#message-call-object
type MessageCall struct {
	Participants   []Snowflake `json:"participants"`
	EndedTimestamp string      `json:"ended_timestamp"` // Optional
}

// Channel represents https://discord.com/developers/docs/resources/channel#channel-object
type Channel struct {
	Id                            Snowflake        `json:"id"`
	Type                          int              `json:"type"`
	GuildId                       Snowflake        `json:"guild_id"`                           // Optional
	Position                      *int             `json:"position"`                           // Optional
	PermissionOverwrites          []Overwrite      `json:"permission_overwrites"`              // Optional
	Name                          string           `json:"name"`                               // Optional, nullable
	Topic                         string           `json:"topic"`                              // Optional, nullable
	Nsfw                          *bool            `json:"nsfw"`                               // Optional
	LastMessageId                 *Snowflake       `json:"last_message_id"`                    // Optional, nullable
	Bitrate                       *int             `json:"bitrate"`                            // Optional
	UserLimit                     *int             `json:"user_limit"`                         // Optional
	RateLimitPerUser              *int             `json:"rate_limit_per_user"`                // Optional
	Recipients                    []User           `json:"recipients"`                         // Optional
	Icon                          string           `json:"icon"`                               // Optional, nullable
	OwnerId                       Snowflake        `json:"owner_id"`                           // Optional
	ApplicationId                 Snowflake        `json:"application_id"`                     // Optional
	Managed                       *bool            `json:"managed"`                            // Optional
	ParentId                      Snowflake        `json:"parent_id"`                          // Optional, nullable
	LastPinTimestamp              string           `json:"last_pin_timestamp"`                 // Optional, nullable
	RtcRegion                     *string          `json:"rtc_region"`                         // Optional, nullable
	VideoQualityMode              int              `json:"video_quality_mode"`                 // Optional
	MessageCount                  int              `json:"message_count"`                      // Optional
	MemberCount                   int              `json:"member_count"`                       // Optional
	ThreadMetadata                *ThreadMetadata  `json:"thread_metadata"`                    // Optional
	Member                        *ThreadMember    `json:"member"`                             // Optional
	DefaultAutoArchiveDuration    *int             `json:"default_auto_archive_duration"`      // Optional
	Permissions                   Permissions      `json:"permissions"`                        // Optional
	Flags                         *int             `json:"flags"`                              // Optional
	TotalMessageSent              *int             `json:"total_message_sent"`                 // Optional
	AvailableTags                 []ForumTag       `json:"available_tags"`                     // Optional
	AppliedTags                   []Snowflake      `json:"applied_tags"`                       // Optional
	DefaultReaction               *DefaultReaction `json:"default_reaction"`                   // Optional, nullable
	DefaultThreadRateLimitPerUser *int             `json:"default_thread_rate_limit_per_user"` // Optional
	DefaultSortOrder              *int             `json:"default_sort_order"`                 // Optional
	DefaultForumLayout            *int             `json:"default_forum_layout"`               // Optional
}

func (channel Channel) GetOverwrite(role Snowflake) *Overwrite {
	for _, overwrite := range channel.PermissionOverwrites {
		if overwrite.Id == role {
			return &overwrite
		}
	}
	return nil
}

// Overwrite represents https://discord.com/developers/docs/resources/channel#overwrite-object
type Overwrite struct {
	Id    Snowflake   `json:"id"`
	Type  int         `json:"type"`
	Allow Permissions `json:"allow"`
	Deny  Permissions `json:"deny"`
}

// ForumTag represents https://discord.com/developers/docs/resources/channel#forum-tag-object
type ForumTag struct {
	Id        Snowflake `json:"id"`
	Name      string    `json:"name"`
	Moderated bool      `json:"moderated"`
	EmojiId   Snowflake `json:"emoji_id"`   // Nullable
	EmojiName string    `json:"emoji_name"` // Nullable
}

// DefaultReaction represents https://discord.com/developers/docs/resources/channel#default-reaction-object
type DefaultReaction struct {
	EmojiId   Snowflake `json:"emoji_id"`   // Optional, nullable
	EmojiName string    `json:"emoji_name"` // Optional, nullable
}

// GuildMember represents https://discord.com/developers/docs/resources/guild#guild-member-object
type GuildMember struct {
	User                       *User                 `json:"user"`   // Optional
	Nick                       string                `json:"nick"`   // Optional, nullable
	Avatar                     string                `json:"avatar"` // Optional, nullable
	Banner                     string                `json:"banner"` // Optional, nullable
	Roles                      []Snowflake           `json:"roles"`
	JoinedAt                   string                `json:"joined_at"`     // Nullable
	PremiumSince               string                `json:"premium_since"` // Optional, nullable
	Deaf                       bool                  `json:"deaf"`
	Mute                       bool                  `json:"mute"`
	Flags                      int                   `json:"flags"`
	Pending                    *bool                 `json:"pending"`                      // Optional
	Permissions                Permissions           `json:"permissions"`                  // Optional
	CommunicationDisabledUntil string                `json:"communication_disabled_until"` // Optional
	AvatarDecorationData       *AvatarDecorationData // Optional, nullable
}

// ThreadMember represents https://discord.com/developers/docs/resources/channel#thread-member-object
type ThreadMember struct {
	Id            Snowflake    `json:"id"`      // Optional (thread ID)
	UserId        Snowflake    `json:"user_id"` // Optional
	JoinTimestamp string       `json:"join_timestamp"`
	Flags         int          `json:"flags"`
	Member        *GuildMember `json:"member"` // Optional
}

// ThreadMetadata represents https://discord.com/developers/docs/resources/channel#thread-metadata-object
type ThreadMetadata struct {
	Archived            bool   `json:"archived"`
	AutoArchiveDuration int    `json:"auto_archive_duration"`
	ArchiveTimestamp    string `json:"archive_timestamp"`
	Locked              bool   `json:"locked"`
	Invitable           *bool  `json:"invitable"`        // Optional
	CreateTimestamp     string `json:"create_timestamp"` // Optional, nullable
}

// Interaction represents https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object
type Interaction struct {
	Id                  Snowflake        `json:"id"`
	ApplicationId       Snowflake        `json:"application_id"`
	Type                int              `json:"type"` // 1 = PING, 2 = APPLICATION_COMMAND, 3 = MESSAGE_COMPONENT, 4 = APPLICATION_COMMAND_AUTOCOMPLETE, 5 = MODAL_SUBMIT
	Data                *json.RawMessage `json:"data"` // Always present on application Commands, message components & model submit. Optional for future proofing.
	Guild               *Guild           `json:"guild"`
	GuildId             Snowflake        `json:"guild_id"`
	Channel             *Channel         `json:"channel"`
	ChannelId           Snowflake        `json:"channel_id"`
	Member              *GuildMember     `json:"member"`
	User                *User            `json:"user"`
	Token               string           `json:"token"` // Continuation token
	Message             *Message         `json:"message"`
	AppPermissions      string           `json:"app_permissions"`
	Locale              string           `json:"locale"`
	GuildLocale         string           `json:"guild_locale"`
	Context             *int             `json:"context"` // 0 = GUILD, 1 = BOT_DM, 2 = PRIVATE_CHANNEL
	AttachmentSizeLimit int              `json:"attachment_size_limit"`
}

// AllowedMentions represents https://discord.com/developers/docs/resources/message#allowed-mentions-object
type AllowedMentions = string

// Poll represents https://discord.com/developers/docs/resources/poll#poll-object-poll-object-structure
type Poll struct {
	Question         PollMedia    `json:"question"` // Only supports text of PollMedia
	Answers          []PollAnswer `json:"answers"`
	Expiry           string       `json:"expiry,omitempty"`
	AllowMultiselect bool         `json:"allow_multiselect"`
	LayoutType       int          `json:"layout_type"` // Must be 1. It's future proofing by discord.
	Results          *PollResults `json:"results,omitempty"`
}

// PollMedia represents https://discord.com/developers/docs/resources/poll#poll-media-object-poll-media-object-structure
type PollMedia struct {
	Text  string `json:"text"`            // Max length: 300 for question, 55 for answer
	Emoji *Emoji `json:"emoji,omitempty"` // PARTIAL: Only ID or name needed
}

// PollAnswer represents https://discord.com/developers/docs/resources/poll#poll-answer-object-poll-answer-object-structure
type PollAnswer struct {
	AnswerId  Snowflake `json:"answer_id,omitempty"` // Only sent by discord. Do not send.
	PollMedia PollMedia `json:"poll_media"`
}

// PollResults represents https://discord.com/developers/docs/resources/poll#poll-results-object-poll-results-object-structure
type PollResults struct {
	Finalized    bool              `json:"is_finalized"`
	AnswerCounts []PollAnswerCount `json:"answer_counts"`
}

// PollAnswerCount represents https://discord.com/developers/docs/resources/poll#poll-results-object-poll-answer-count-object-structure
type PollAnswerCount struct {
	Id    int `json:"id"`
	Count int `json:"count"`
}

// ResolvedData represents https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-object-resolved-data-structure
type ResolvedData struct {
	Users       map[Snowflake]User        `json:"users"`
	Members     map[Snowflake]GuildMember `json:"members"`
	Roles       map[Snowflake]Role        `json:"roles"`
	Channels    map[Snowflake]Channel     `json:"channels"`
	Messages    map[Snowflake]Message     `json:"messages"`
	Attachments map[Snowflake]Attachment  `json:"discord.Attachments"`
}

// UnavailableGuild represents https://discord.com/developers/docs/resources/guild#unavailable-guild-object
type UnavailableGuild struct {
	Id          Snowflake `json:"id"`
	Unavailable bool      `json:"unavailable"`
}

// GuildBan represents https://discord.com/developers/docs/resources/guild#ban-object
type GuildBan struct {
	User   User   `json:"user"`
	Reason string `json:"reason"`
}
