create DATABASE registrationapp;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
\connect registrationapp

create table app_user(
	username text PRIMARY KEY NOT NULL,
	passwordHash text NOT NULL,
	name text NOT NULL,
	createdAt TIMESTAMP NOT NULL,
	updatedAt TIMESTAMP NOT NULL,
	lastLoggedIn TIMESTAMP
);

create table app_admin(
	username text PRIMARY KEY NOT NULL,
	passwordHash text NOT NULL,
	name text NOT NULL
);

create table event(
	ID UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
	name text NOT NULL,
	url text UNIQUE NOT NULL,
	"start" TIMESTAMP,
	"end" TIMESTAMP,
	lat float8,
	long float8,
	radius float8, --in km
	createdAt TIMESTAMP NOT NULL DEFAULT NOW(),
	updatedAt TIMESTAMP NOT NULL DEFAULT NOW()
);

create table guest(
	nricHash text NOT NULL,
	eventID UUID NOT NULL REFERENCES event(ID) ON UPDATE CASCADE ON DELETE CASCADE,
	name text NOT NULL,
	checkedIn BOOLEAN NOT NULL DEFAULT FALSE,
	checkInTime TIMESTAMP,
	PRIMARY KEY(nricHash, eventID)
);

create table hosts(
	username text NOT NULL REFERENCES app_user(username) ON UPDATE CASCADE ON DELETE CASCADE,
	eventID UUID NOT NULL REFERENCES event(ID) ON UPDATE CASCADE ON DELETE CASCADE,
	PRIMARY KEY(username, eventID)
);

create USER server_access with password 'LongNightShortDay';
grant CONNECT on DATABASE registrationapp to server_access;
grant SELECT, INSERT, UPDATE, DELETE on app_user to server_access;
grant SELECT, INSERT, UPDATE on app_admin to server_access;
grant SELECT, INSERT, UPDATE, DELETE on event to server_access;
grant SELECT, INSERT, UPDATE, DELETE on hosts to server_access;
grant SELECT, INSERT, UPDATE, DELETE on guest to server_access;