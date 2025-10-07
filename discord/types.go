package discord

// Snowflake is the identifier type used by discord. It is structured as specified by
// https://discord.com/developers/docs/reference#snowflakes
type Snowflake = uint64

type ConnectionProperties struct {
	Os      string `json:"os"`
	Browser string `json:"browser"`
	Device  string `json:"device"`
}

// Guild represents a discord guild https://discord.com/developers/docs/resources/guild#guild-object.
// Owner field is not included as it is impossible for a discord bot to own a guild
type Guild struct {
	Id              Snowflake `json:"id"`
	Name            string    `json:"name"`
	Icon            *string   `json:"icon"`                // Nullable
	IconHash        *string   `json:"icon_hash,omitempty"` // Optional, nullable
	Splash          *string   `json:"splash"`              // Nullable
	DiscoverySplash *string   `json:"discovery_splash"`    // Nullable
	OwnerId         Snowflake `json:"owner_id"`
	// TODO: A lot more fields. Only bothered adding the ones needed for early partial objs
}

// User represents a discord user https://discord.com/developers/docs/resources/guild#guild-object
type User struct {
	Id            Snowflake `json:"id"`
	Username      string    `json:"username"`
	Discriminator string    `json:"discriminator"`
	GlobalName    *string   `json:"global_name,omitempty"` // Nullable
	Avatar        *string   `json:"avatar"`
	Bot           *bool     `json:"bot,omitempty"` // Optional, not nullable
	// TODO: A lot more fields. Only bothered adding the ones needed for early partial objs
}

type Application struct {
	Id          Snowflake `json:"id"`
	Name        string    `json:"name"`
	Icon        *string   `json:"icon"` // Nullable
	Description string    `json:"description"`
	Flags       *int      `json:"flags,omitempty"` // Optional, not nullable
	// TODO: A lot more fields. Only bothered adding the ones needed for early partial objs
}
