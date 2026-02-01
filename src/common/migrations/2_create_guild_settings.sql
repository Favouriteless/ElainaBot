CREATE TABLE guild_settings (
    guild_id BIGINT UNSIGNED PRIMARY KEY,
    honeypot_channel BIGINT UNSIGNED,
    hello_enabled BOOL NOT NULL
);