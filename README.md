# Gantry by MES

Gantry is an event management web application for use by MINDEF/SAF events. This repository hosts the web server application 
part of the application.

### Binaries and Usage

Binaries are stored in cmd, and there are 2 binaries (`checkin` and `uploadguests`). 

#### checkin

`checkin` is the web server, and takes no arguments to start up, but if run locally requires a .env in the cmd/checkin folder, with the keys in the .env.example file in the root directory. Make sure before building a binary that is meant to be run locally that the first line of the main function is uncommented in cmd/checkin/main.go:

```go
func main() {
	loadEnvironmentalVariables() //comment this out for heroku production
  ...
```

Note that .env folders themselves are git ignored for security reasons.

#### uploadguests

`uploadguests` is a command line tool to take a CSV with NRICs and names in the following format:

1234A,LTC Jim Bob

1235B,ME4 John   

and register all the guests in that CSV for a given event. Note that the CSV should have no header line. Try `./uploadguests -h` (on Linux and Mac), or `uploadguests.exe -h` (on Windows) for more information about usage. Example usage to upload from guestlist.csv to an event with a known eventID using the dsd19 user account:

```
./uploadguests -src=guestlist.csv -addr=https://mesgantry-backend.herokuapp.com -event=aa19239f-f9f5-4935-b1f7-0edfdceabba7 -out=output.txt -u=dsd19
```

You will be prompted for a password after hitting enter. Upload will begin shortly after, and output will be dumped into output.txt. Make sure to read the output to check for errors registering any guests (for example, frequently there are some guests who appear twice in the CSV, though this isn't an issue, sometimes two different guests with the same last 5 digits of their NRIC appear in the CSV, and the second guest will fail to register).

If guests have tags, that is to say they are in the following format:

1234A,LTC Jim Bob,"VIP,CONFIRMED"

1235B,ME4 John,VIP

Added the `-tags` argument to ./uploadguests

### Project Layout

#### Guide
`model.go` contains the data types used in the business logic and interfaces for the data access objects and other objects whose implementations require a dependency. Changes to the keys used when the datatypes are marshalled into JSON and changes to the database column names to be read for each field of the datatype are done in this file.

`postgres` contains the data access layer of the project. Changes to how event, guest, user information etc is fetched from the database, or created/updated/deleted should be done here.

`http` contains the http/network layer of the project. Any changes to processing and replying to http requests (for example, registering new API endpoints to handle, or changing the format of a http response for one of the endpoints) should be done here.

`http/jwt` contains the Javascript Web Token handling/generation logic.

`http/cors` contains the CORS handling logic.

`bcrypt` contains the hashing method used in the project.

`cmd` contains the executables.

#### Concept

The project uses a dependency injection model, 
[as detailed by this blog post, which is an essential read before making changes to the code](https://medium.com/@benbjohnson/standard-package-layout-7cdbc8391fc1). 
In essence, the root directory contains the abstract data structures that are used in the business logic of the program (such as an Event,
or a Guest). The root directory also specifies various interfaces (such as for the data access objects) of objects which encapsulate a 
functionality provided by an external dependency (like the postgres library). The specific implementations of these interfaces which rely
on an external dependency are encapsulated within their own folders (all code using net/http is in the http folder, etc); the root folder
must have no external dependencies whatsoever (though it can important "minor" standard library dependencies, like time, which do not
make sense to encapsulate).

The implementations of dependencies communicate with one another using the interfaces in the root package, and so dependencies can be
plugged in and out. As an example, the postgres folder contains postgres implementations of data access objects like the EventService,
which provides methods like .CreateEvent(), .EventsBy(username), .DeleteEvent() which allow you to create/retrieve/update/delete objects.
But there could also be a blockchain folder containing an implementation of this DAO using a blockchain + a websocket connection (which is
slated to be added in future releases).

The main file, under the cmd folder, initializes the implementations to be used, and ties everything together, before starting the server.

### Branch Management

The `master` branch should always contain the latest released version + hotfixes possibly. Releases, which are tagged, may also have their own branches, so hotfixes can be applied to a release but not to the project going forward. This latter method is recommended if the changes are spaghetti code/untested and meant for an urgent situation which shouldn't be part of the codebase.

When a new version begins development, start a new version development "master" branch, such as `v1.3dev` (in that naming convention). To begin making a change, create a new branch first, make your changes, test them, and merge them (or make a pull request if the team lead wants to review your code) with the development branch. WARNING: Trying to merge a pull request, by default, will merge into master. Carefully select your development branch instead.

The "change" branches should be worked on by one person at one time, but many people can work on the same version development branch, making different changes, in different branches (The general principle is a branch should only have one person committing to it, and changes should be integrated into the larger branch). The change branches should follow the following naming convention `category/changename`. The categories commonly used are `feat` for new features (like a new API endpoint), `bug` for bugfixes, or `optimize` for optimization changes.

For example `feat/websitebuilder` could be a branch creating the website builder back end, `bug/utcfix` could be the branch fixing times not being in UTC on the data access layer, `optimize/parallelfind` could be parallelizing a findGuests function to make it faster. These branches should be deleted after they have been merged back into the version, or after they've been abandoned.

Once all the changes for the release have been merged with the version development branch, merge it with master, and deploy the new version of heroku. Before merging the branch, make sure the docs are updated to the new release.

`git checkout -b <branch_name>` to create a new branch from the current commit

`git checkout -d <branch_name>` deletes a branch locally

`git push --delete <remote_name> <branch_name>` deletes a branch in the remote (GitHub). Remote name is usually origin

`git merge <branch_name>` merges the commits of the branch named with the branch the user is currently on. Use this to merge a branch's changes with the development branch, for example:

```
git checkout v1.3dev
git pull
git merge feat/timezones
git push -u origin v1.3dev
```

### Heroku Deployment

From Scratch:

1. Go to the [heroku dashboard](https://www.heroku.com), navigate to/create (via the New button >> Create New App) the mesgantry-backend app. Click on it, and then go to the Settings tab, and click "Reveal Config Vars". 
2. Set the environmental variables, according to the keys in `env.example`.
3. Go to the Deploy tab, and choose GitHub as the Deployment method. Under the connect to github section, type `gantry-backend`.
3. Click connect next to the right repository.
4. Select the branch you want to deploy from. This is usually master. A test back end may deploy off the development branches.
5. Press 'Enable Automatic Deploys' for test back ends. Now any commits to the specified branch will automatically cause heroku to rebuild, and be deployed. This is probably not desirable for the real backend, which can manually be deployed through 'Manual Deploy' upon new version releases.
6. The current branch has not been deployed yet, however. Go to the 'Manual Deploy' section , select the right branch, and press Deploy.
7. Give heroku a few seconds, and it will finish building the app. Use the Heroku CLI to monitor the progress of the app

#### Heroku CLI

Download the Heroku CLI from https://devcenter.heroku.com/articles/heroku-cli#download-and-install

Go to the command line, and type
```
heroku login
```

to login. You must do this before making any other heroku command.

To see the logs of the gantry backend app (in particular to check if builds proceeded smoothly):

```
heroku logs --tail -a mesgantry-backend
```

The `--tail` option makes the logs appear in live time; if you just want a snapshot at a  particular time, leave it out.

To make changes to the postgres database, do

```
heroku pg:psql -a mesgantry-backend
```

Now navigate the database/make changes like how you already do using psql for a local database.

### Testing

To test a particular package, navigate to that package in the command line, and then run `go test`. To generate a coverage report, do 

```
go test -coverprofile fmt
```

Which will produce output as such:

```
...logging statements...
PASS
coverage: 41.9% of statements
ok      checkin/http    0.811s
```

In order to view exactly which lines in each file were tested and which were not, run 

```
go tool cover -html=fmt
```

In order to test all packages, from the root folder run

```
go test ./...
```

In order to test race conditions (very important for websocket/database access layer!), add the `-race` argument while testing a package. You can add the `-race` argument while testing all packages like above, make sure to do this before merging to master or a main development branch.

To run a specific test, run

```
go test -run "^(TestFunctionName)$"
```

