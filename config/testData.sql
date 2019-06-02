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
	('CSSCOM Planning Seminar','csscom', '03293b3b-df83-407e-b836-fb7d4a3c4966'),
    ('Supply Rally','supply','c14a592c-950d-44ba-b173-bbb9e4f5c8b4');

INSERT into event(name, url, start, "end", lat, long, radius, ID, createdAt, updatedAt) VALUES
    ('Data Science Department Talk', 'dsdjan2019', '2019-01-10 15:00:00', '2019-01-10 18:00:00', 1.335932, 103.744708, 0.5, 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', '2019-04-01 04:05:36', '2019-04-10 03:02:11');

INSERT into form (ID, name, eventID, survey, submitTime) VALUES
    ('ec5c5f6f-5384-4406-9beb-73b9effbdf50','Alice', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', '[{"question":"A","answer":"AA1"},{"question":"B","answer":"BB1"},{"question":"C","answer":"CC1"}]', '2019-04-11 08:18:14'),
    ('663fd6e1-b781-49e7-b1ed-dd0e3c6ff28e','Bob', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', '[{"question":"A","answer":"AA2"},{"question":"B","answer":"BB2"},{"question":"C","answer":"CC2"}]', '2019-04-11 09:32:04'),
    ('a6db3963-5389-4dbe-8fc6-bbd7f7ce66b8','Jonathan', '2c59b54d-3422-4bdb-824c-4125775b44c8', '[{"question":"D","answer":"DD"},{"question":"E","answer":"EE"},{"question":"C","answer":"CC3"}]', '2019-02-17 13:18:53');

INSERT into hosts(username, eventID) VALUES
    ('ME5Bob', '2c59b54d-3422-4bdb-824c-4125775b44c8'),
	('TestUser', '3820a980-a207-4738-b82b-45808fe7aba8'),
	('TestUser', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7'),
	('ME5Bob', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7');

INSERT into guest(nricHash, eventID, name, tags, checkedIn, checkInTime) VALUES
    ('$2a$05$0KDoAOK32Z7Dw3t5l8mLfeh9k4XdQ2rPLPXVmV0NMOxjMFmJgXfMG', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'A', '{}', FALSE, NULL), --1234A
	('$2a$05$KJedrPxRji7H9PK6qC.peOJN.YKN0byARYggk2NXFSfjrT1XIx7SW', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'B', '{}', FALSE, NULL), --5678B
	('$2a$05$yV1oi210wpqziULusDmPXukJ13Da8RUeY/vASFXAlLsN/DytFYw.u', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'C', '{"VIP", "ATTENDING"}', FALSE, NULL), --2346C
    ('$2a$05$pUq6Q5IYnNcbWqeeudc02eeBjt1iUewWLd1lPfZH4EPI6p/A1TgP6', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'D', '{"VIP"}', FALSE, NULL), --4916T
	('$2a$05$YWlI8xZ17iNwTbh8kGeG7OL4yLCVD7HmFVUGPgsyR/Uuwh0leMRgi', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'E', '{"ATTENDING"}', FALSE, NULL), --4433G
    ('$2a$05$jlZfVRqXJMEkf7.VR/whpuQ2BHKdYvJVN1LPKbLb92ZHCDN3uKw6K', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'F', '{}', TRUE, NOW()), --1234B
	('$2a$05$XsA.v2McOSjtd8I1oP.hE.2qhBnHxnb7ePEnPtdn.NULaLbVAqpNO', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'G', '{}', TRUE, NOW()), --8648C
	('$2a$05$2vPMWiaglJAAUQJo5lQVGufOYLPPx2VeBzuNV6e38LiohrXLwhztq', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'H', '{"VIP", "ATTENDING"}', TRUE, NOW()), --2146D
    ('$2a$05$xo.VBwQbURoel8RJPK0AVeGmsSWobC12cyIkJJ/.fP2RJxP8wZHVu', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'I', '{"VIP"}', TRUE, NOW()), --4215S
	('$2a$05$S47CHNZvm6.m4EHUuS6YkurCifg0pOzmU6ft3WghPWpUoqt6acDt6', 'aa19239f-f9f5-4935-b1f7-0edfdceabba7', 'J', '{"ATTENDING"}', TRUE, NOW()), --4333Q
	('$2a$05$x64.3poyVqwy4NwE22uTZuP1AjfHhIZIs2lfNZJVtRtNW46aID.ee', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'K', '{}', FALSE, NULL), --2234A
	('$2a$05$lel.jqGzav6PTxgN9t7klOEfVrf5wCWhcy13Oc9MoKZK6BzUWDyiO', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'L', '{}', FALSE, NULL), --3678B
	('$2a$05$f8LJ5oBCth1IC7K.hHcr8uCttpIcCqt29O2DJZQmmhWruDIktzzOe', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'M', '{"VIP", "ATTENDING"}', FALSE, NULL), --4346C
    ('$2a$05$KNbrT044utli70VMzSRqv.DU7b/eXWpHaXdJzVL3UnGmqvgGXiWaC', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'N', '{"VIP"}', FALSE, NULL), -- 5916T
	('$2a$05$5RkG8P3fgX5.9kmuL0rRDex42ZEybZwU.rS2JwoUQmTwtFyOrAwlu', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'O', '{"ATTENDING"}', FALSE, NULL), --6433G
    ('$2a$05$1ZDPchZ6SfdKvh22QNmoNuUg4GpvtbOhHga32I3NgM5Aw6vnE0j3K', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'P', '{}', TRUE, NOW()), --7234B
	('$2a$05$ugJ5KdgL8UBGqS8gshtYMOZoxAF9IsjFrEoPDY6c9KR3.q9r.61aq', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'Q', '{}', TRUE, NOW()), --9648C
	('$2a$05$627D548fpcSjhnUjQTEgr.Cdf3Fx6XdSKUznTLlcREDKXQcNBw0R.', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'R', '{"VIP", "ATTENDING"}', TRUE, NOW()), --8146D
    ('$2a$05$VyIbSlqzOrH/DQnHyT3EzupEef59BiLyILqwHueVIcPu/sEotHaFm', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'S', '{"VIP"}', TRUE, NOW()), --1215S
	('$2a$05$VI006dcHVD2QVDygGGjlLuFdg8yYeLfhiwT6RT4F9pSBLDvOJZo46', '2c59b54d-3422-4bdb-824c-4125775b44c8', 'T', '{"ATTENDING"}', TRUE, NOW()), --3333Q
	('$2a$05$Nt8.5PjdkFJGgQdrKRjbXeqNTQcGwLXrMCyXXyhJPH4fFnc0Br4Xa', '03293b3b-df83-407e-b836-fb7d4a3c4966', 'A', '{}', FALSE, NOW()), --1234A
    ('$2a$05$0jc0dU6smqP/DzDAptlisuJfrKFfv4Y/i7pXYgcGjjRgApa0lSlx2', 'c14a592c-950d-44ba-b173-bbb9e4f5c8b4', 'A', '{"OFFICER"}', FALSE, NULL), --1234A
    ('$2a$05$HS0BoQiagbVZDVXkl5c1PuaCVB6wEL5xgSM.covh3jFjoyVMub1Ge', 'c14a592c-950d-44ba-b173-bbb9e4f5c8b4', 'B', '{"VIP", "ATTENDING"}', TRUE, NOW()), --2834B
    ('$2a$05$buWqmWSyLOGcaLb4PzVAne7TrHEZdPvtTE6g2x8ObHlSTYuQt/uB6', 'c14a592c-950d-44ba-b173-bbb9e4f5c8b4', 'C', '{}', TRUE, NOW()), -- 1212C
    ('$2a$05$yOfIuWrSfRKQCHpXqEbELOgxm7nUPjn4xgWfYcux6t6R8R5OOuy3y', 'c14a592c-950d-44ba-b173-bbb9e4f5c8b4', 'D', '{"VIP", "ATTENDING"}', TRUE, NOW()), -- 1132B
    ('$2a$05$3DDeeRWbDCN../1lhYIUbeqDcPX0IPTwUQJfyXmYW9JIFOimdzJbS', 'c14a592c-950d-44ba-b173-bbb9e4f5c8b4', 'E', '{"VIP"}', TRUE, NOW()), -- 4432Z
    ('$2a$05$ieRGDQQQAkCd7c9dOxz1r.W3iOf8iDXG2Iu4DvZyFl9QNHBldVOn2', 'c14a592c-950d-44ba-b173-bbb9e4f5c8b4', 'F', '{"OFFICER"}', FALSE, NULL); -- 2482D