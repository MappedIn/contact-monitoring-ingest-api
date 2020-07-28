require('dotenv').config({path: '../.env'});
const path = require('path');
const mongodb = require('mongodb');
const app = require('express')();
const http = require('http').createServer(app);

const {
	MONGO_URL,
    MONGO_DB_NAME
} = process.env;

const configPath = process.argv[2];

const {
	boundingBox,
	floorBounds,
} = require(path.join(__dirname, configPath));

async function main() {
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

	const POSITION_EVENT = 'position-event'
	const MINUTE_AGGREGATION = 'minute-aggregation';

	app.get('/', (req, res) => {
		res.sendFile(__dirname + '/visualizer.html');
	});

	app.get('/init', async (req, res) => {
		const [latestPosition] = await db.collection(POSITION_EVENT)
			.find({})
			.sort({timeBucket: -1})
			.project({timeBucket: 1})
			.limit(1)
			.toArray();
		const [earliestPosition] = await db.collection(POSITION_EVENT)
			.find({})
			.sort({timeBucket: 1})
			.project({timeBucket: 1})
			.limit(1)
			.toArray();

		res.send({
			boundingBox,
			floorBounds,
			timeRange: [earliestPosition.timeBucket, latestPosition.timeBucket]
		});
	});

	app.get('/at/:floor/:timebucket', async (req, res) => {
		const timeBucket = parseInt(req.params.timebucket);
		const floor = parseInt(req.params.floor);
		const devices = req.query.devices && req.query.devices.split(',');
		console.log('timebucket', timeBucket, 'floor', floor, 'devices', devices);
		const positionEvents = await db.collection(POSITION_EVENT)
			.find({ timeBucket, floor })
			.project({ device: 1, timeBucket: 1, lonlat: 1, acc: 1, floor: 1 })
			.toArray();

		const query = {
			timeBucket,
			floor,
		};

		if (devices && devices.length > 0) {
			query['events.device'] = { $in: devices };
		}

		const minuteAggregations = await db.collection(MINUTE_AGGREGATION)
			.find(query)
			.project({ events: 1, timeBucket: 1, distance: 1, floor: 1 })
			.toArray();

		const data = {
			positionEvents: positionEvents,
			minuteAggregations: minuteAggregations,
		};

		res.send(data);
	});

	http.listen(3000, () => {
		console.log('listening on *:3000');
	});
}

main();


