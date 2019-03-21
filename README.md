# Gantry by MES

Gantry is an event management web application for use by MINDEF/SAF events. This repository hosts the web server application 
part of the application.

### Binaries and Usage

Binaries are stored in cmd, and there are 2 binaries (checkin and uploadguests). checkin is the web server, and takes no arguments to
start up, but if run locally requires a .env in the cmd/checkin folder, with the keys in the .env.example file in the root directory.
Note that .env folders themselves are git ignored for security reasons.

### Project Layout

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
