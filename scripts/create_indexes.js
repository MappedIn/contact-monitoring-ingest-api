/**
 * ensure that the mongodb instance has these indexes
 * for adequate performance
 */

db.getCollection('contact-event').createIndex({
    "devices" : 1,
    "end" : -1,
    "start" : -1
});

db.getCollection('contact-event').createIndex({
    "devices.0" : 1,
    "devices.1" : 1,
    "end" : -1,
    "start" : -1

}, {
    "unique" : true,
});

db.getCollection('contact-event').createIndex({
    "duration" : 1
});

db.getCollection('contact-event').createIndex({
    "end" : 1
});

db.getCollection('contact-event').createIndex({
    "mindistance" : 1
});

db.getCollection('device').createIndex({
    "venue" : 1
});

db.getCollection('device').createIndex({
    "_fts" : "text",
    "_ftsx" : 1
}, {
    "weights" : {
        "_id" : 1,
        "deviceType" : 1,
        "name" : 1
    },
    "default_language" : "english",
    "language_override" : "language",
    "textIndexVersion" : 3
});

db.getCollection('minute-aggregation').createIndex({
    "events.0.device" : 1,
    "events.1.device" : 1,
    "timeBucket" : -1
}, {
    "unique": true
});

db.getCollection('minute-aggregation').createIndex({
    "events.device" : 1,
    "timeBucket" : 1
});

db.getCollection('minute-aggregation').createIndex({
    "timeBucket" : 1
});

db.getCollection('position-event').createIndex({
    "device" : 1,
    "timeBucket" : 1
}, {
    "unique": true
});

db.getCollection('position-event').createIndex({
        "lonlat" : "2dsphere",
        "floor" : 1,
        "timeBucket" : -1
}, {
    "2dsphereIndexVersion" : 3
});

db.getCollection('position-event').createIndex({
    "timeBucket" : 1,
    "venue" : 1
});

db.getCollection('position-event').createIndex({
    "venue" : 1
});