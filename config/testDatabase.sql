--run this in psql before running any of the tests in the data access layer (postgres) to set up the test database

create DATABASE registrationapp_test;
\connect registrationapp_test

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE user regapp_test with encrypted password 'henry';
GRANT connect on database registrationapp_test to regapp_test;
GRANT ALL PRIVILEGES on schema public to regapp_test;
