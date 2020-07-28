# Contact Monitoring Data Generator and Visualizer

This sub project includes a service that imitates devices moving in a bounding polygon and sends their data to the CT Ingest API. This sub project also includes another service that reads the data from the DB and displays it using leaflet and shows the contact lines between devices.

The data generator (aka faker) has some parameters that are currently hardcoded, like the variety of device types and the date range for data generation as well as the speed at which data is generated. These parameters could be pulled out into the config to make experimentation easier.

The data visualizer is very rough around the edges and a lot of the UI does not work very well. However it does let you scan through time using a slider at the bottom and displays each time bucket of device events on the map. It displays each device using a deterministically chosen color based on the device ID and a circle representing the accuracy of the device at the time of the event.

## Dependencies

- Mongo v4+
- Node v10+


## Deployment

These projects should not be deployed. They are intended for developer use only for debugging and benchmarking the CT Ingest API.


## Development

### Installation

1. Make sure your mongodb instance for CT Ingest API is running

2. Make sure the CT Ingest API is running

3. Run `yarn` or `npm i` to install node_modules

### Running Data Generator

Run `yarn faker <relative path to config>` or `npm run faker <relative path to config>`

### Running Data Visualizer

Run `yarn vis <relative path to config>` or `npm run vis <relative path to config>`

### Config

The config file specified is json and expects these fields:

| Key           | Type                  | Description   |
|---            |---                    |---            |
| venue         | string                | The venue key/slug |
| totalDevices  | integer               | Number of devices to simulate |
| floorBounds   | [integer, integer]    | A tuple representing the min and max floor numbers |
| boundingBox   | [<lat, lon>,...]      | A polygon with lat lon pairs as vertices |
