require('dotenv').config({path: '../.env'});

const path = require('path');
const http = require('http');
const https = require('https');

const mongodb = require('mongodb');
const _axios = require('axios');
const randomPointsOnPolygon = require('random-points-on-polygon');
const turf = require('@turf/turf');
const _ = require('lodash');
const { v4: uuidv4 } = require('uuid');

const {
    PORT,
    MONGO_URL,
    MONGO_DB_NAME
} = process.env;

const axios = _axios.create({
  //60 sec timeout
  timeout: 60000,

  //keepAlive pools and reuses TCP connections, so it's faster
  httpAgent: new http.Agent({ keepAlive: true }),
  httpsAgent: new https.Agent({ keepAlive: true }),

  //follow up to 10 HTTP 3xx redirects
  maxRedirects: 10,

  //cap the maximum content length we'll accept to 50MBs, just in case
  maxContentLength: 50 * 1000 * 1000
});

const configPath = process.argv[2];

const {
    boundingBox,
    floorBounds,
    totalDevices,
    venue,
} = require(path.join(__dirname, configPath));

const levelPolygon = turf.polygon([boundingBox]);

const suffixes = [
    'Enterprise',
    'Supervisor',
    'Manager',
    'Security',
    'Maintenance',
];

const deviceTypes = [
    'iPhone XS',
    'iPhone XS Max',
    'iPhone XR',
    'iPhone 11 Pro',
    'iPhone 11 Pro Max',
    'iPhone 11 SE',
    'iPhone 11',
    'iPhone 8',
    'iPhone 8 Plus',
    'iPhone 7 Plus',
    'iPhone 7',
    'iPhone 6S',
    'iPad',
];

class Device {
    constructor({
        device,
        timeOffset,
        venue,
    } = {}) {
        this.device = device;
        this.timeOffset = timeOffset + Math.floor(Math.random() * 30) - 15;
        this.floor = Math.floor(Math.random() * (floorBounds[1] - floorBounds[0])) + floorBounds[0];
        this.venue = venue;
        this.acc = (Math.random() * 2) + 2.5;
        this.position = randomPointsOnPolygon(1, levelPolygon)[0];
        this.previousPosition = this.position;
        this.targetPosition = this.position;
        this.bearing = 0;
        this.time = Date.now();

        // average walking speed is 1.4 m/s
        // so we vary our walking speed per device from 1.2 to 1.6
        this.speed = Math.round(Math.random() * .4) + 1.2

        // chance for device to stay still once at target position
        const chanceToStayStill = 95;
        this.mobility = (Math.random() + chanceToStayStill) / 100

        this.batch = [];
        this.lastBatchPush = Date.now();
        this.pushToBatch();
    }

    async tick(time) {
        this.autoMove(time - this.time);
        this.time = time;

        if (this.batch.length > 20) {
            return this.report();
        }
    }

    async getToken() {
        const {data} = await axios({
            method: 'get',
            url: `http://localhost:${PORT}/device/${this.device}/token`,
        });
        
        this.token = data.token;
        this.tokenExpiresAt = new Date(data.expiresAt);
    }

    calcBearing() {
        this.bearing = turf.bearing(this.position, this.targetPosition);
    }

    moveTo(newPosition) {
        this.previousPosition = this.position;
        this.position = newPosition;
    }

    autoMove(duration) {
        const targetDistance = turf.distance(
            this.position.geometry.coordinates,
            this.targetPosition.geometry.coordinates,
            {units:'meters'}
        );

        this.acc += (Math.random() * 0.2) - 0.10;
        this.acc = Math.max(Math.min(this.acc, 4.5), 2);

        const distanceTraveled = this.speed * duration / 1000;

        // if we are at our target then use a 10% chance to start
        // towards a new target position
        if (turf.booleanEqual(this.position, this.targetPosition)) {
            if (Math.random() > this.mobility) {
                this.targetPosition = randomPointsOnPolygon(1, levelPolygon)[0];
                this.calcBearing();
                console.log(`${this.device} moving`);
            }
        } else if (targetDistance <= distanceTraveled) {
            this.moveTo(this.targetPosition);
            console.log(`${this.device} arrived`);
        } else {
            this.moveTo(turf.transformTranslate(this.position, distanceTraveled, this.bearing, {
                units: 'meters',
            }))
        }

        if (turf.booleanEqual(this.position, this.previousPosition)) {
            if (this.time - this.lastBatchPush > 30 * 1000) {
                this.pushToBatch();
            }
        } else {
            this.pushToBatch();
        }
    }

    pushToBatch() {
        this.batch.push(this.getData());
        this.lastBatchPush = this.time;
    }

    getData() {
        const time = this.time + this.timeOffset;
        return {
            acc: this.acc,
            device: this.device,
            floor: this.floor,
            lonlat: this.position.geometry.coordinates,
            time: time,
            userConsent: true,
            venue: this.venue,
        };
    }

    async report() {
        const batch = this.batch;
        console.log(`${this.device} reporting`)
        this.batch = [];

        try {
            if (!this.token || this.tokenExpiresAt <= new Date()) {
                await this.getToken();
            }

            if (!this.token) {
                console.log('device has no token after attempting to get one');
                return;
            }

            const result = await axios({
                method: 'post',
                url: `http://localhost:${PORT}/positions`,
                headers: {
                    Authorization: `Bearer ${this.token}`
                },
                data: batch,
            });
            
            console.log(`${this.device} report result: ${result.data && result.data.map(({status}) => status)}`);
        } catch (err) {
            if (err.response) {
              // Request made and server responded
              console.log(err.response.status, err.response.data, err.response.headers);              
            } else if (err.request) {
              // The request was made but no response was received
              console.log(err.request);
            } else {
              // Something happened in setting up the request that triggered an Error
              console.log('Error', err.message);
            }
        }
    }
}

async function main(){
    const mongourl = `${MONGO_URL}/${MONGO_DB_NAME}`
    console.log('connecting to mongodb', mongourl);
    console.time('connected to mongodb');
    const client = new mongodb.MongoClient(mongourl, {
		useNewUrlParser: true,
		useUnifiedTopology: true,
	});
	await client.connect();
    console.timeEnd('connected to mongodb');
    const db = client.db();

    const eventsCol = db.collection('position-event');
    const results = await eventsCol
        .find()
        .sort({time: -1})
        .limit(1)
        .toArray();

    let timeOffset = -100 * 60 * 1000;
    if (results[0] && results[0].time) {
        timeOffset = results[0].time - Date.now()
    }
    console.log(timeOffset);

    // get list of devices for our venue
    const devicesCol = db.collection('device');
    let devices = await devicesCol
        .find({ venue })
        .limit(totalDevices)
        .toArray();

    devices = devices.map((device) => new Device({
        device: device._id + '',
        timeOffset,
        venue,
    }));

    // make new devices if we don't have enough devices
    // already in the DB
    const totalToMake = totalDevices - devices.length;
    console.log(`total devices: ${devices.length} need to make ${totalToMake} more`)
    for (let i = 0; i < totalToMake; i++) {
        const type = _.sample(deviceTypes);
        const newDevice = {
            _id: uuidv4(),
            venue,
            type,
            name: `${_.sample(suffixes)} ${type}`,
        };
        await devicesCol.insertOne(newDevice);
        devices.push(new Device({
            device: newDevice._id + '',
            timeOffset,
            venue,
        }))
    }

    client.close();

    const sleep = (t) => new Promise(r => setTimeout(r, t));

    console.time('total')
    let currentTime = Date.now();
    for (let i = 0; i < 6000; i++) {
        await sleep(10);
        currentTime += 20 * 1000;

        const timer = `iteration ${i}`;
        console.time(timer)
        await Promise.all(
            devices.map((device) => {
                return device.tick(currentTime)
                    .catch((err) => err.code && console.log(err.code, err.response))
            })
        );

        console.timeEnd(timer);
    }
    console.timeEnd('total');
}

main();
