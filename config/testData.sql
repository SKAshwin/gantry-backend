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
	('SDB Cohesion','sdbcohesionnovember', '3820a980-a207-4738-b82b-45808fe7aba8');

INSERT into event(name, url, start, "end", lat, long, radius, ID) VALUES
    ('Data Science Department Talk', 'dsdjan2019', '2019-01-10 15:00:00', '2019-01-10 18:00:00', 1.335932, 103.744708, 0.5, 'aa19239f-f9f5-4935-b1f7-0edfdceabba7');

INSERT into hosts(username, eventID) VALUES
    ('ME5Bob', '2c59b54d-3422-4bdb-824c-4125775b44c8'),
	('TestUser', '3820a980-a207-4738-b82b-45808fe7aba8'),
	('TestUser', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7'),
	('ME5Bob', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7');
