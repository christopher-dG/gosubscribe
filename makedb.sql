CREATE DATABASE o_subscribe;

\c o_subscribe;

CREATE TABLE users(
  user_disc INTEGER PRIMARY KEY,
  user_id BIGINT,
  user_name VARCHAR(32)
);

CREATE TABLE mappers(
  mapper_id INTEGER PRIMARY KEY,
  mapper_name VARCHAR(15)
);

CREATE TABLE subscriptions(
  user_disc INTEGER REFERENCES users ON DELETE CASCADE,
  mapper_id INTEGER REFERENCES mappers ON DELETE CASCADE,
  PRIMARY KEY(user_disc, mapper_id)
);

CREATE TABLE maps(
  mapper_id  INTEGER REFERENCES mappers ON DELETE CASCADE,
  mapset_id INTEGER,
  status INTEGER,
  PRIMARY KEY(mapper_id, mapset_id)
);
