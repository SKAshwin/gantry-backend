INSERT INTO app_user (username, passwordHash, name, createdAt, updatedAt, lastLoggedIn) VALUES --DONT PUT TAB AFTER VALUES, IT WILL FUCK UP ON LINUX BY ADDING EXTRA (
    ('ME5Bob', '$2a$05$xuLyzrAaW7Y4mXGAXzPjOOIdTim2BVGzePV73H.Vsy8gKCUxXqRB2', 'Robert Lim', NOW(), NOW(), NULL), --DarkHorseBigCat
	('ME6Alice', '$2a$05$KlEAswrxBhrgXr1Fh7Io0ecBHB182FeJAuP8BppK8BVRujEPlavkK', 'Alice Ng', NOW(), NOW(), NULL), --abcdef12345!@#$%
	('safosscholar', '$2a$05$a6TJMr8YD4jUese5NKOL0.bDNvNMgV5.aDEhWpomDCYtz6jLn7GNe', 'Jonathan Tan', NOW(), NOW(), NOW()), --pas5W0rD!!!
	('AirForceMan', '$2a$05$hAzW2PGlX4ig/oUChUxyA.1R0QqrKMCuxnlifS9qFbQQCivy6OvSq', 'Benjamin Net', NOW(), NOW(), NOW()), --WhyNotUseAnEasyPasswordLikeThis?
	('TestUser', '$2a$05$D/nbFy9utEDFgg.Jfsl39epqO2Yx2nIRClYFGVMw9fnLZlXFFnP5u', 'User McUserson', NOW(), NOW(), NULL); --WhatAreDictionaryAttacks

INSERT into app_admin(username, passwordHash, name) VALUES 
    ('Hackerman','$2a$05$YNWHk.7Su/St644J1BAX7.G8KP3t5ts16bcAPApXSw.yc4hHrgwNi','Drop Table'); --BlackFireFoodHorse

INSERT into event(name, url, ID) VALUES
    ('Data Science CoP','cop2018', '2c59b54d-3422-4bdb-824c-4125775b44c8'),
	('SDB Cohesion','sdbcohesionnovember', '3820a980-a207-4738-b82b-45808fe7aba8'),
	('CSSCOM Planning Seminar','csscom', '03293b3b-df83-407e-b836-fb7d4a3c4966');

INSERT into event(name, url, start, "end", lat, long, radius, ID, createdAt, updatedAt) VALUES
    ('Data Science Department Talk', 'dsdjan2019', '2019-01-10 15:00:00', '2019-01-10 18:00:00', 1.335932, 103.744708, 0.5, 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', '2019-04-01 04:05:36', '2019-04-10 03:02:11');

INSERT into hosts(username, eventID) VALUES
    ('ME5Bob', '2c59b54d-3422-4bdb-824c-4125775b44c8'),
	('TestUser', '3820a980-a207-4738-b82b-45808fe7aba8'),
	('TestUser', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7'),
	('ME5Bob', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7');

INSERT into guest(nricHash, eventID, name, tags, checkedIn, checkInTime) VALUES
    ('A1234', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'A', '{}', FALSE, NULL),
	('B5678', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'B', '{}', FALSE, NULL),
	('C2346', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'C', '{"VIP", "ATTENDING"}', FALSE, NULL),
    ('T4916', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'D', '{"VIP"}', FALSE, NULL),
	('G4433', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'E', '{"ATTENDING"}', FALSE, NULL),
    ('B1234', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'F', '{}', TRUE, NOW()),
	('C8648', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'G', '{}', TRUE, NOW()),
	('D2146', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'H', '{"VIP", "ATTENDING"}', TRUE, NOW()),
    ('S4215', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'I', '{"VIP"}', TRUE, NOW()),
	('Q4333', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'J', '{"ATTENDING"}', TRUE, NOW()),
	('A2234', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'K', '{}', FALSE, NULL),
	('B3678', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'L', '{}', FALSE, NULL),
	('C4346', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'M', '{"VIP", "ATTENDING"}', FALSE, NULL),
    ('T5916', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'N', '{"VIP"}', FALSE, NULL),
	('G6433', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'O', '{"ATTENDING"}', FALSE, NULL),
    ('B7234', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'P', '{}', TRUE, NOW()),
	('C9648', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'Q', '{}', TRUE, NOW()),
	('D8146', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'R', '{"VIP", "ATTENDING"}', TRUE, NOW()),
    ('S1215', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'S', '{"VIP"}', TRUE, NOW()),
	('Q3333', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'T', '{"ATTENDING"}', TRUE, NOW()),
	('A1234', '03293b3b-df83-407e-b836-fb7d4a3c4966', 'A', '{}', FALSE, NOW());