# How to provision devices

1.
    Post a new invite code using the basic auth credentials set in your .env

    ```
    curl -XPOST user:pass@localhost:8090/invite-code/my-venue-slug -d '{"maxUses": 100}'
    ```
    
    in this case we are setting the maximum uses to 100 because in our scenario we know we have approximately 100 devices to provision and we would like the invite code to expire after 100 have been provisioned.

    you should get back an invite code in the response:
    ```json
    {
        "code":"YQ4JYDS1",
        "maxUses":100,
        "venue":"my-venue-slug"
    }
    ```

2.
    Our mobile SDK includes example code to allow users to enter this invite code into their device. The SDK internally sends an http request to the ingest API equivalent to

    ```
    curl -XPOST localhost:8090/device/a_long_random_device_id_to_keep_device_anonymous/activate -d '{"deviceType":"iphone 8","code": "YMFW9UHK"}'
    ```

    which will respond with 200 and a body that includes the venue that this device is now provisioned for

    ```json
    {
        "venue":"my-venue-slug"
    }
    ```

3.
    Now that the device is activated for our venue it can request a short lived JWT which will be used with each push of events to verify the device.

    ```
    curl localhost:8090/device/a_long_random_device_id_to_keep_device_anonymous/token
    ```

    which returns a body that includes the token and expiration.
    
    ```json
    {
        "expiresAt":"2020-07-27T15:35:20.666280774-04:00",
        "token":"eyJhbGciOiJIUzI..."
    }
    ```

    When the token expires the device can request a new token using the same endpoint

4.
    Your device is now ready to send position events along with an Authorization header using Bearer <token>

    ```
    curl -XPOST localhost:8090/positions -h 'Authorization=Bearer eyJhbGciOiJIUzI...' -d '[{...}]'
    ```