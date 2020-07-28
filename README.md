![CI build](https://github.com/MappedIn/contact-monitoring-ingest-api/workflows/CI/badge.svg)

# Contact Monitoring Ingestion API

1. [Purpose](#purpose)
2. [Scripts](#scripts)
3. [Installation](#installation)
4. [Dependencies](#dependencies)
5. [API Route Documentation](docs/API.md)
6. [How To Provision Devices](docs/ProvisioningDevices.md)
7. [Dataflow Explanation](docs/Dataflow.md)

[](#purpose)

# Purpose

This service is a centralized contact monitoring ingestion API. We designed our solution to monitor contacts within a building like a workplace or campus. The mobile SDK that works with this API sends position events in batches only when the device is inside a geofence. The events are tied to an anonymized device ID (currently a UUID) which is also registered with the API. We believe that by having the monitoring only happen within a small space where contact is most likely to happen, and have the devices be anonymized and opt in helps to prevent privacy concerns. Devices will never be monitored outside of the designated geofence.

As this service receives position events from devices it attempts to determine if any other events in that time frame are near enough to count as a short contact event then it determines if there were at least 5 minutes worth of these short contact events to count as a full contact event.


[](#scripts)

# Scripts

We are using make to manage tooling on this project.

Run `make help` to get a list of available commands:

| command        | description |
| ---            | ---
| install        | installs Go dependencies (but not Go itself)
| run            | runs from source code; dev only
| build          | builds binary for this project
| start          | runs the previously built binary
| docker-build   | builds docker image for this project
| docker-run     | runs previously built docker image for this project
| test           | runs all tests with verbose output
| test           | runs all tests with verbose output and coverage


[](#installation)

# Installation

-   [ ] Start local MongoDB using docker ```docker run --name ct_mongo -p 27017:27017 -v /$(pwd)/data:/data/db -v /$(pwd)/scripts:/scripts -d mongo:4```
-   [ ] Create the indexes in MongoDB ```docker exec -it ct_mongo mongo localhost:27017/contact-monitoring /scripts/create_indexes.js```
-   [ ] Install Go v1.14+
-   [ ] Make a .env file `cp .env.example .env`
-   [ ] Update .env (see below)
-   [ ] Run `make run` to start the server

## Environment Variables

| name                              | description                   |
| ---                               | ---
| PORT                              | Port the ingest API will run on
| MONGO_URL                         | MongoDB connection url
| MONGO_DB_NAME                     | Name of the DB that will be used in MongoDB
| DEVICE_TOKEN_SECRET               | Secret used to sign device JWT
| INVITE_CODE_USER                  | Basic auth user for accessing invite code
| INVITE_CODE_PASS                  | Basic auth pass for accessing invite code
| MAXIMUM_DISTANCE_BETWEEN_DEVICES  | The maximum distance devices can be from one another to determine a contact event in meters
| ACCURACY_THRESHOLD                | The maximum accuracy an event can have to deem it viable for processing in meters


[](#dependencies)

# Dependencies

- Mongo v4+
- Go v1.14+
