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

create table hosts(
	username text NOT NULL REFERENCES app_user(username) ON UPDATE CASCADE ON DELETE CASCADE,
	eventID UUID NOT NULL REFERENCES event(ID) ON UPDATE CASCADE ON DELETE CASCADE,
	PRIMARY KEY(username, eventID)
);

INSERT INTO app_user (username, passwordHash, name, createdAt, updatedAt, lastLoggedIn) VALUES --DONT PUT TAB AFTER VALUES, IT WILL FUCK UP ON LINUX BY ADDING EXTRA (
    ('ME5Bob', '$2a$05$xuLyzrAaW7Y4mXGAXzPjOOIdTim2BVGzePV73H.Vsy8gKCUxXqRB2', 'Robert Lim', NOW(), NOW(), NULL), --DarkHorseBigCat
	('ME6Alice', '$2a$05$KlEAswrxBhrgXr1Fh7Io0ecBHB182FeJAuP8BppK8BVRujEPlavkK', 'Alice Ng', NOW(), NOW(), NULL), --abcdef12345!@#$%
	('safosscholar', '$2a$05$a6TJMr8YD4jUese5NKOL0.bDNvNMgV5.aDEhWpomDCYtz6jLn7GNe', 'Jonathan Tan', NOW(), NOW(), NOW()), --pas5W0rD!!!
	('AirForceMan', '$2a$05$hAzW2PGlX4ig/oUChUxyA.1R0QqrKMCuxnlifS9qFbQQCivy6OvSq', 'Benjamin Net', NOW(), NOW(), NOW()), --WhyNotUseAnEasyPasswordLikeThis?
	('TestUser', '$2a$05$D/nbFy9utEDFgg.Jfsl39epqO2Yx2nIRClYFGVMw9fnLZlXFFnP5u', 'User McUserson', NOW(), NOW(), NULL); --WhatAreDictionaryAttacks

INSERT into app_admin(username, passwordHash, name) VALUES 
    ('Hackerman','$2a$05$YNWHk.7Su/St644J1BAX7.G8KP3t5ts16bcAPApXSw.yc4hHrgwNi','Drop Table'); --BlackFireFoodHorse

INSERT into event(name, url) VALUES
    ('Data Science CoP','cop2018'),
	('SDB Cohesion','sdbcohesionnovember');

INSERT into event(name, url, start, "end", lat, long, radius) VALUES
    ('Data Science Department Talk', 'dsdjan2019', '2019-01-10 15:00:00', '2019-01-10 18:00:00', 1.335932, 103.744708, 0.5);

INSERT into hosts(username, eventID) VALUES
    ('ME5Bob', '2c59b54d-3422-4bdb-824c-4125775b44c8'),
	('TestUser', '3820a980-a207-4738-b82b-45808fe7aba8'),
	('TestUser', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7'),
	('ME5Bob', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7');

create USER server_access with password 'LongNightShortDay';
grant CONNECT on DATABASE registrationapp to server_access;
grant SELECT, INSERT, UPDATE, DELETE on app_user to server_access;
grant SELECT, INSERT, UPDATE on app_admin to server_access;
grant SELECT, INSERT, UPDATE, DELETE on event to server_access;
grant SELECT, INSERT, UPDATE, DELETE on hosts to server_access;