CREATE TABLE macro (
    guild BIGINT UNSIGNED,
    keyword VARCHAR(25),
    response VARCHAR(280),
    PRIMARY KEY(guild, keyword)
);

CREATE TABLE ban (
    guild BIGINT UNSIGNED,
    user_id BIGINT UNSIGNED,
    expires BIGINT UNSIGNED,
    reason VARCHAR(280),
    PRIMARY KEY(guild, user_id)
);