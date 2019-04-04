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

### Deployment Branches

#### Quick Guide

```
git checkout tags/v1.1 -b heroku-deploy-v1.1
git cherry-pick 5c77797f7a8f6c3f1860b1cc072d94355ce0581f
git push -u origin heroku-deploy-v1.1
```

Cherry-pick picks a commit which adds the procfile and comments out loadEnvironmentalVariables() in the main.

#### In-Depth Guide

All deployments should be off a separate branch from master, and should be tied to a specific release of the project. Start by switching to a new branch at the tag of the release you want to deploy (the example below assumes you are trying to deploy v1.1; try to follow the naming convention)

```
git checkout tags/v1.1 -b heroku-deploy-v1.1
```

You need a file called Procfile in your root directory, which tells heroku where the main function is, effectively, and what commands to run to run the go program (after the go buildpack actually compiles the program). Add the following line in the Procfile,

```
web: cd cmd/checkin && checkin
```

You can automate creating the file by just running the following command from the command line, in the root directory:

```
echo "web: cd cmd/checkin && checkin" >> Procfile
```

This Procfile tells heroku to navigate to the checkin folder, and then run the checkin executable (note that heroku runs on Ubuntu 18.04 as of the time of writing - google heroku stacks to learn more - so the executable is a linux binary). If the main function is in a file in another directory, update the cd half of the command accordingly.

Go to `cmd/checkin/main.go` and comment out the first line: 

```go
func main() {
	//loadEnvironmentalVariables() //comment this out for heroku production
  ...
```

Commit the new main.go and Procfile to the deployment branch. 

```
git push -u origin heroku-deploy-v1.1
```

### Heroku Deployment

From Scratch:

1. Go to the [heroku dashboard](https://www.heroku.com), navigate to/create (via the New button >> Create New App) the mesgantry-backend app. Click on it, and then go to the Settings tab, and click "Reveal Config Vars". 
2. Set the environmental variables, according to the keys in `env.example`.
3. Go to the Deploy tab, and choose GitHub as the Deployment method. Under the connect to github section, type `gantry-backend`.
3. Click connect next to the right repository.
4. Select the branch you want to deploy from.
5. Press 'Enable Automatic Deploys'. Now any commits to the deployment branch will automatically cause heroku to rebuild, and be deployed.
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

