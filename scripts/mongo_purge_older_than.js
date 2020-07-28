/**
 * use this to purge data older than the purgeDate
 */

var purgeDate = Date.now() - 1000 * 60 * 60 * 24 * 30; // 30 days ago
var bucketSize = 60 * 1000;
var bucket = Math.floor(purgeDate / bucketSize);

db.getCollection('position-events').deleteMany({
    timeBucket: {
        $lt: bucket
    }
})

db.getCollection('minute-aggregation').deleteMany({
    timeBucket: {
        $lt: bucket
    }
});

db.getCollection('contact-event').deleteMany({
    end: {
        $lt: bucket
    }
});