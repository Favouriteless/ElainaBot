package discord

// Snowflake is the identifier type used by discord. It is structured as specified by
// https://discord.com/developers/docs/reference#snowflakes
type Snowflake = string

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
	// TODO: A lot more fields. Only bothered adding the ones needed for early partial objs
}

// User represents https://discord.com/developers/docs/resources/guild#guild-object
type User struct {
	Id               Snowflake         `json:"id"`
	Username         string            `json:"username"`
	Discriminator    string            `json:"discriminator"`
	GlobalName       string            `json:"global_name"` // Nullable
	Avatar           string            `json:"avatar"`
	Bot              *bool             `json:"bot"`                    // Optional
	System           *bool             `json:"system"`                 // Optional
	MfaEnabled       *bool             `json:"mfa_enabled"`            // Optional
	Banner           string            `json:"banner"`                 // Optional, nullable
	AccentColor      *int              `json:"accent_color"`           // Optional
	Locale           string            `json:"locale"`                 // Optional
	Verified         *bool             `json:"verified"`               // Optional
	Email            string            `json:"email"`                  // Optional
	Flags            *int              `json:"flags"`                  // Optional
	PremiumType      *int              `json:"premium_type"`           // Optional
	PublicFlags      *int              `json:"public_flags"`           // Optional
	AvatarDecoration *AvatarDecoration `json:"avatar_decoration_data"` // Optional
	Collectibles     *Collectibles     `json:"collectibles"`           // Optional
	PrimaryGuild     *UserPrimaryGuild `json:"primary_guild"`          // Optional, nullable
}

// AvatarDecoration represents https://discord.com/developers/docs/resources/user#avatar-decoration-data-object
type AvatarDecoration struct {
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

type Application struct {
	Id          Snowflake `json:"id"`
	Name        string    `json:"name"`
	Icon        string    `json:"icon"` // Nullable
	Description string    `json:"description"`
	Flags       *int      `json:"flags,omitempty"` // Optional, not nullable
	// TODO: A lot more fields. Only bothered adding the ones needed for early partial objs
}

// Message represents https://discord.com/developers/docs/resources/message#message-object
type Message struct {
	Id              Snowflake `json:"id"`
	ChannelId       Snowflake `json:"channel_id"`
	Author          User      `json:"author"`
	Content         string    `json:"content"` // Requires MESSAGE_CONTENT intent
	Timestamp       string    `json:"timestamp"`
	EditedTimestamp string    `json:"edited_timestamp"` // Nullable
	Tts             bool      `json:"tts"`
	MentionEveryone bool      `json:"mention_everyone"`
	Mentions        []User    `json:"mentions"`
	// MentionRoles
	// MentionChannels
	// Attachments
	// Embeds
	// Reactions
	// Nonce (? lol what)
	Pinned    bool       `json:"pinned"`
	WebhookId *Snowflake `json:"webhook_id"` // Optional
	Type      int        `json:"type"`
	// Activity
	Application   *Application `json:"application"`    // Optional
	ApplicationId *Snowflake   `json:"application_id"` // Optional
	Flags         *int         `json:"flags"`          // Optional
	// MessageReference
	// MessageSnapshots
	// ReferencedMessage
	// InteractionMetadata
	// Interaction
	// Thread
	// Components
	// StickerItems
	// Stickers
	// Position
	// RoleSubscriptionData
	// Resolved
	// Poll
	// Call
}
