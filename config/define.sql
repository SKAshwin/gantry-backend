create DATABASE registrationapp;
\connect registrationapp

-- don't modify the test commands - used by the database access layer test suite

--test
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

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
	url text UNIQUE,
	timetags json NOT NULL DEFAULT '{}', -- should just be an empty JSON object instead of null
	"start" TIMESTAMP,
	"end" TIMESTAMP,
	lat float8,
	long float8,
	radius float8, --in km
	createdAt TIMESTAMP NOT NULL DEFAULT (NOW() at time zone 'utc'),
	updatedAt TIMESTAMP NOT NULL DEFAULT (NOW() at time zone 'utc')
);

create table form (
	ID UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
	name text NOT NULL DEFAULT '',
	eventID UUID NOT NULL REFERENCES event(ID) ON UPDATE CASCADE ON DELETE CASCADE,
	survey json NOT NULL,
	submitTime TIMESTAMP NOT NULL DEFAULT (NOW() at time zone 'utc')
);

create table guest(
	nricHash text NOT NULL,
	eventID UUID NOT NULL REFERENCES event(ID) ON UPDATE CASCADE ON DELETE CASCADE,
	name text NOT NULL,
	tags text[] NOT NULL DEFAULT '{}',
	checkedIn BOOLEAN NOT NULL DEFAULT FALSE,
	checkInTime TIMESTAMP,
	PRIMARY KEY(nricHash, eventID)
);

create table hosts(
	username text NOT NULL REFERENCES app_user(username) ON UPDATE CASCADE ON DELETE CASCADE,
	eventID UUID NOT NULL REFERENCES event(ID) ON UPDATE CASCADE ON DELETE CASCADE,
	PRIMARY KEY(username, eventID)
);

create table website(
	eventID UUID NOT NULL PRIMARY KEY REFERENCES event(ID) ON UPDATE CASCADE ON DELETE CASCADE,
	data json NOT NULL
);
--test

create USER server_access with password 'LongNightShortDay';
grant CONNECT on DATABASE registrationapp to server_access;
GRANT ALL PRIVILEGES on schema public to server_access;
