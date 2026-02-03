package restapi

import (
	"bytes"
	. "elaina-common"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const maxRestAttempts = 3

var globalUnknownMu = &sync.Mutex{} // Requests with an unknown bucket will be executed synchronously, using a global mutex.
var globalRetryAfter = atomic.Int64{}
var buckets = make(map[string]*routeBucket)

var routeCreateCommand = newApiRoute(http.MethodPost, "/applications/%d/commands", nil)
var routeDeleteCommand = newApiRoute(http.MethodDelete, "/applications/%d/commands/%d", nil)

var routeGetMessage = newApiRoute(http.MethodGet, "/channels/%d/messages/%d", nil)
var routeCreateMessage = newApiRoute(http.MethodPost, "/channels/%d/messages", nil)
var routeDeleteMessage = newApiRoute(http.MethodDelete, "/channels/%d/messages", nil)
var routeCreateReaction = newApiRoute(http.MethodPost, "/channels/%d/messages/%d/reactions/%s/@me", nil)

var routeGetChannel = newApiRoute(http.MethodGet, "/channels/%d", nil)
var routeCreateDM = newApiRoute(http.MethodPost, "/users/@me/channels", nil)

var routeGetGuild = newApiRoute(http.MethodGet, "/guilds/%d", nil)
var routeGetRole = newApiRoute(http.MethodGet, "/guilds/%d/roles/%d", nil)
var routeGetGuildMember = newApiRoute(http.MethodGet, "/guilds/%d/members/%d", nil)
var routeModifyGuildMember = newApiRoute(http.MethodPatch, "/guilds/%d/members/%d", nil)
var routeKickGuildMember = newApiRoute(http.MethodDelete, "/guilds/%d/members/%d", nil)
var routeCreateGuildBan = newApiRoute(http.MethodPost, "/guilds/%d/bans/%d", nil)
var routeDeleteGuildBan = newApiRoute(http.MethodDelete, "/guilds/%d/bans/%d", nil)

func newApiRoute(method string, path string, headers http.Header) *route {
	return &route{method: method, path: path, headers: headers, urlToBucket: make(map[string]string)}
}

type route struct {
	method      string
	path        string
	headers     http.Header
	urlToBucket map[string]string
}

func (route *route) do(body []byte, attempt int, args ...any) (respBody []byte, err error) {
	waitUntil := time.UnixMilli(globalRetryAfter.Load())
	if wait := waitUntil.Sub(time.Now()); wait > 0 { // If we're globally rate limited, wait until it expires.
		time.Sleep(wait)
	}

	url := BaseApiUrl + fmt.Sprintf(route.path, args...)

	bucket := buckets[route.urlToBucket[url]]
	if bucket == nil {
		globalUnknownMu.Lock()
	} else if bucket.consume() {
		defer bucket.Unlock()
	}

	resp, err := SendHttp(route.method, url, bytes.NewReader(body), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if bucket != nil {
		if err := bucket.update(resp.Header); err != nil {
			resp.Body.Close()
			return nil, err
		}
	} else {
		globalUnknownMu.Unlock()
		bucketId := resp.Header.Get("X-RateLimit-Bucket")

		bucket = &routeBucket{id: bucketId}
		if err := bucket.update(resp.Header); err != nil {
			return nil, err
		}

		buckets[bucketId] = bucket
		route.urlToBucket[url] = bucketId
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
	case http.StatusNoContent:
	case http.StatusInternalServerError:
		fallthrough
	case http.StatusServiceUnavailable:
		fallthrough
	case http.StatusBadGateway:
		if attempt < maxRestAttempts {
			slog.Warn(fmt.Sprintf("[REST] Attempt %d failed, retrying...", attempt))
			return route.do(body, attempt+1, args...)
		}
		return nil, fmt.Errorf("exceeded maximum number of retries")
	case http.StatusTooManyRequests:
		hGlobal := resp.Header.Get("X-RateLimit-Global")
		hRetryAfter := resp.Header.Get("Retry-After")

		retryAfter, err := strconv.ParseFloat(hRetryAfter, 64)
		if err != nil {
			return respBody, err // Should never be hit
		}
		retry := time.Now().Add(time.Duration(retryAfter * float64(time.Second))).Add(time.Millisecond * 500) // Weirdly, still get more 429s unless it's offset further

		if hGlobal != "" {
			globalRetryAfter.Store(retry.UnixMilli())
		} else {
			bucket.mu.Lock()
			bucket.reset = retry
			bucket.remaining = 0
			bucket.limit = 0
			bucket.mu.Unlock()
		}

		if attempt < maxRestAttempts {
			return route.do(body, attempt+1, args...)
		}
		return nil, fmt.Errorf("rate limit: exceeded maximum number of retries")
	case http.StatusUnauthorized:
		panic(errors.New("invalid bot token/tried to access something a bot can't")) // Really, really, terribly awfully horrible if this is ever hit.
	default:
		err = RestError{Response: resp, Body: respBody}
	}

	return
}

type RestError struct {
	Response *http.Response
	Body     []byte
}

func (err RestError) Error() string {
	return fmt.Sprintf("rest error (%v): %s", err.Response.Status, string(err.Body))
}

type RestErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func getCacheable[K comparable, T any](cache *LRUCache[K, T], id K, route *route, args ...any) (*T, error) {
	if val := cache.Get(id); val != nil {
		return val, nil
	}

	resp, err := route.do(nil, 1, args)
	if err != nil {
		return nil, err
	}

	var val T
	if err = json.Unmarshal(resp, &val); err != nil {
		return nil, err
	}
	cache.Add(id, val)
	return &val, nil
}

// --------------------------------------------------------------------
// |                             COMMANDS                             |
// --------------------------------------------------------------------

func CreateOrUpdateCommand(command *ApplicationCommand) ([]byte, error) {
	enc, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}
	return routeCreateCommand.do(enc, 1, CommonSecrets.Id)
}

func DeleteCommand(command Snowflake) error {
	_, err := routeDeleteCommand.do(nil, 1, CommonSecrets.Id, command)
	return err
}

// --------------------------------------------------------------------
// |                             MESSAGES                             |
// --------------------------------------------------------------------

func GetMessage(channel Snowflake, message Snowflake) (*Message, error) {
	resp, err := routeGetMessage.do(nil, 1, channel, message)
	if err != nil {
		return nil, err
	}

	var msg Message
	if err = json.Unmarshal(resp, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func CreateMessage(channel Snowflake, content string, tts bool) (*Message, error) {
	enc, err := json.Marshal(struct {
		Content string `json:"content"`
		Tts     bool   `json:"tts"`
	}{content, tts})
	if err != nil {
		return nil, err
	}

	resp, err := routeCreateMessage.do(enc, 1, channel)
	if err != nil {
		return nil, err
	}

	var msg Message
	if err = json.Unmarshal(resp, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func DeleteMessage(channel Snowflake, message Snowflake) error {
	_, err := routeDeleteMessage.do(nil, 1, channel, message)
	return err
}

// CreateReaction creates a reaction to a message using the bot account. emoji must be either a Unicode emoji for
// built-in emojis or a string in the format "name:snowflake" for custom discord emojis.
func CreateReaction(channelId Snowflake, messageId Snowflake, emoji string) error {
	_, err := routeCreateMessage.do(nil, 1, channelId, messageId, emoji)
	return err
}

// --------------------------------------------------------------------
// |                             CHANNELS                             |
// --------------------------------------------------------------------

func GetChannel(channelId Snowflake) (*Channel, error) {
	return getCacheable(ChannelCache, channelId, routeGetChannel, channelId)
}

func CreateDM(recipient Snowflake) (*Channel, error) {
	body, err := json.Marshal(struct {
		Recipient Snowflake `json:"recipient_id"`
	}{recipient})
	if err != nil {
		return nil, err
	}

	resp, err := routeCreateDM.do(body, 1)
	if err != nil {
		return nil, err
	}

	var channel Channel
	if err = json.Unmarshal(resp, &channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

// --------------------------------------------------------------------
// |                              GUILDS                              |
// --------------------------------------------------------------------

func GetGuild(id Snowflake) (*Guild, error) {
	return getCacheable(GuildCache, id, routeGetGuild, id)
}

func GetRole(guildId Snowflake, roleId Snowflake) (*Role, error) {
	return getCacheable(RoleCache, roleId, routeGetRole, guildId, roleId)
}

func GetGuildMember(guild Snowflake, guildMemberId Snowflake) (*GuildMember, error) {
	return getCacheable(GuildMemberCache, guildMemberId, routeGetGuildMember, guild, guildMemberId)
}

func ModifyGuildMember(guildId Snowflake, userId Snowflake, payload ModifyGuildMemberPayload) error {
	enc, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = routeModifyGuildMember.do(enc, 1, guildId, userId)
	return err
}

func KickUser(guildId Snowflake, userId Snowflake) error {
	_, err := routeKickGuildMember.do(nil, 1, guildId, userId)
	return err
}

func CreateBan(guildId Snowflake, userId Snowflake, deleteSeconds int) error {
	enc, err := json.Marshal(struct {
		Seconds int `json:"delete_message_seconds"`
	}{deleteSeconds})
	if err != nil {
		return err
	}
	_, err = routeCreateGuildBan.do(enc, 1, guildId, userId)
	return err
}

func DeleteBan(guildId Snowflake, userId Snowflake) error {
	_, err := routeDeleteGuildBan.do(nil, 1, guildId, userId)
	return err
}

// --------------------------------------------------------------------
// |                              USERS                               |
// --------------------------------------------------------------------
