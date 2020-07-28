# How do position events get processed into contact events?

This is a step by step explaination of how position events enter the system and are processed.

1. Devices send `positionEvent`s in batches.
2. For each `positionEvent` in a batch we filter out any that have accuracy that is above our allowed threshold or that reference a venue that the device is not signed up for.
3. Within the batch of `positionEvent`s sent by the device their are usually going to be more than 1 per minute, but we are processing the data in 1 minute buckets so we only store 1 `positionEvent` per minute per device and skip the rest. The 1 minute bucket was chosen because it seemed to fit the right balance of deduplicated position data without leaving too much of a gap between movement.
4. An `positionEvent` being processed is processed in 5 stages.

    1. Store the `positionEvent` in our DB with a geo-spatial index. Once we are certain the `positionEvent` is in the DB we can go to stage 2.
    2. Perform a geo-spatial query on the `positionEvent` where we want all `positionEvent`s in a (`ma` + `n` + `da`) radius where `ma` is the maximum accuracy allowed for any `positionEvent` and `n` is the maxmimum distance between devices to determine a contact `positionEvent` and `da` is the accuracy of the `positionEvent` being processed.
    3. The returned results are then further filtered in our application by removing `positionEvent`s that are actually outside of the maximum distance between devices.
    4. For each of these results we insert a aggregate document into our `minuteAggregate` table with id set to a tuple of the 2 device ids involved and the minute as an epoch timestamp. Due to everything being grouped in time buckets, this can in the worst case end up with (`n` (`n` + 1) / 2) - 1 minute aggregates for a given minute where `n` is the number of devices and `n` >= 2.
    5. For each of these `minuteAggregate` events we insert a `contactEvent` with similar information but we will also query for other `contactEvent`s that are one minute before or one minute after and merge those `contactEvent`s together, removing extras.

5. This leaves us with a collection of `contactEvent`s that can be queried by device, venue, time range, and event length of contact very quickly with no processing at query time.
