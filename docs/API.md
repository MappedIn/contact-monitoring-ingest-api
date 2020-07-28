FORMAT: 1A
HOST: http://localhost:8090

# Contact Monitoring Ingest API

## Invite Code [/invite-code/{venue_slug}]

An invite code has the following attribues:

+ code - A short randomized alphanumeric string
+ venue - The venue that this invite code is for
+ maxUses - Maximum number of uses for this invite code

+ Parameters
    + venue_slug: my-venue (required, string) - Slug identifier of the venue for the invite code
    + maxUses: 100 (required, number) - Sets the maximum uses for the invite code


### Create Invite Code [POST]

+ Request (application/json)

        {
            "maxUses": 100
        }

+ Response 201 (application/json)

        {
            "code":"YQ4JYDS1",
            "venue":"my-venue-slug",
            "maxUses":100
        }


## Device [/device/{device_id}]

A device has the following attributes:

+ id: (required, string) - The unique identifier for the device
+ venue: my-venue (required, string) - Slug of the venue that this device is allowed to push position events for
+ type: iphone (required, string) - The general make and model of the device

+ Parameters
    + device_id: (required, string) - the identifier for the device

## Device Activation [/device/{device_id}/activate]

+ Parameters
    + device_id: (required, string) - the identifier for the device

### Activate Device [POST]

+ Request (application/json)

        {
            "deviceType": "iphone 8",
            "code": "YQ4JYDS1"
        }

+ Response 200 (application/json)

        {
            "venue": "my-venue"
        }


## Device Token [/device/{device_id}/token]

+ Parameters
    + device_id: (required, string) - the identifier for the device

### Request New Device Token [GET]

+ Response 200 (application/json)

        {
            "expiresAt": "2020-07-24T16:16:23.621796546-04:00",
            "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ2ZW51ZSI6IjQxMC1hbGJlcnQiLCJleHAiOjE1OTU2MjE3ODN9.JWLBMfZTKNGK3bMY7PQpBbc9wMRtXWCd804SW05TqvM"
        }

## Position Events [/positions]

### Push Device Position Events [POST]

+ Request (application/json)

        [
            {
                "device": "choo4Zioc0pengau3obaGuthahPh5oovelohyah5leeshole7ahn0keuqua9ho8oov7ooja3eefoh3ahgeoQuahwaeT4opheishaiNgief7KohtaiRaethie2oozao6l",
                "time": 1595618446073,
                "lonlat": [43.482928, -80.535819],
                "acc": 4.3213,
                "floor": 0,
                "userConsent": true,
                "venue": "my-venue"
            },
            {
                "device": "choo4Zioc0pengau3obaGuthahPh5oovelohyah5leeshole7ahn0keuqua9ho8oov7ooja3eefoh3ahgeoQuahwaeT4opheishaiNgief7KohtaiRaethie2oozao6l",
                "time": 1595618446073,
                "lonlat": [43.482928, -80.535819],
                "acc": 5.123,
                "floor": 0,
                "userConsent": true,
                "venue": "my-venue"
            },
            ...
        ]

+ Response 207 (application/json)

        [
            {
                "status": 200,
                "message": ""
            },
            {
                "status": 400,
                "message": "Accuracy of 5.123 exceeds threshold of 5.0"
            }
        ]
