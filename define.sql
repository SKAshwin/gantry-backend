create DATABASE registrationapp;
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
	eventID SERIAL PRIMARY KEY NOT NULL,
	name text NOT NULL,
	startDate DATE NOT NULL,
	endDate DATE NOT NULL,
	geofence CIRCLE NOT NULL
);

create table hosts(
	username text NOT NULL REFERENCES app_user(username) ON UPDATE CASCADE ON DELETE CASCADE,
	eventID SERIAL NOT NULL REFERENCES event(eventID) ON UPDATE CASCADE ON DELETE CASCADE,
	PRIMARY KEY(username, eventID)
);

INSERT INTO app_user (username, passwordHash, name, createdAt, updatedAt, lastLoggedIn) VALUES
	('ME5Bob', '$2a$05$xuLyzrAaW7Y4mXGAXzPjOOIdTim2BVGzePV73H.Vsy8gKCUxXqRB2', 'Robert Lim', NOW(), NOW(), NULL), --DarkHorseBigCat
	('ME6Alice', '$2a$05$KlEAswrxBhrgXr1Fh7Io0ecBHB182FeJAuP8BppK8BVRujEPlavkK', 'Alice Ng', NOW(), NOW(), NULL), --abcdef12345!@#$%
	('safosscholar', '$2a$05$a6TJMr8YD4jUese5NKOL0.bDNvNMgV5.aDEhWpomDCYtz6jLn7GNe', 'Jonathan Tan', NOW(), NOW(), NOW()), --pas5W0rD!!!
	('AirForceMan', '$2a$05$hAzW2PGlX4ig/oUChUxyA.1R0QqrKMCuxnlifS9qFbQQCivy6OvSq', 'Benjamin Net', NOW(), NOW(), NOW()), --WhyNotUseAnEasyPasswordLikeThis?
	('TestUser', '$2a$05$D/nbFy9utEDFgg.Jfsl39epqO2Yx2nIRClYFGVMw9fnLZlXFFnP5u', 'User McUserson', NOW(), NOW(), NULL); --WhatAreDictionaryAttacks

INSERT into app_admin(username, passwordHash, name) VALUES 
	('Hackerman','$2a$05$YNWHk.7Su/St644J1BAX7.G8KP3t5ts16bcAPApXSw.yc4hHrgwNi','Drop Table');

create USER server_access with password 'LongNightShortDay';
grant CONNECT on DATABASE registrationapp to server_access;
grant SELECT, INSERT, UPDATE, DELETE on app_user to server_access;
grant SELECT, INSERT, UPDATE on app_admin to server_access;
