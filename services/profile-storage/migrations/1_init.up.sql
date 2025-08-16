CREATE TABLE player_credentials
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    login_token UUID NOT NULL UNIQUE
);
