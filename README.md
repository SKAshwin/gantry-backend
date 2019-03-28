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

In order to make a deployment on heroku, as the Go buildpack doesn't play nice with multiple main functions in the cmd folder, first switch to a new branch. All deployment branches should start at a release of the project (see releases), and be named `heroku-deploy-v1.1` for example. In this branch, remove the exceptions in the .gitignore which prevent binaries from being uploaded to the repository, and add a file in the root directory called Procfile which contains:

```
web: cmd/checkin/checkin
```

Go to `cmd/checkin/main.go` and comment out the first line: 

```go
func main() {
	//loadEnvironmentalVariables() //comment this out for heroku production
  ...
```

Compile `cmd/checkin/main.go` by running `go build .` when in that folder. Commit the new main.go, executable, new .gitignore and Procfile to the deployment branch.

Any changes in the deployment branch (hotfixes during an event; by right, a version release should be tested and not require further changes once deployed) require a recompiling of the executable and committing the new executable to the deployment branch in order for changes to actually go through.

Link heroku to auto deploy from the deployment branch. Never commit a Procfile or executable to the master branch.

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

