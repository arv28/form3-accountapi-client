
## Overview
form3-accountapi-client is a Go client library for accessing the fake account api. It supports the following operations on the accounts resource:
- Create - create a new account resource
- Fetch - Fetch an existing account based on id
- Delete - Delete an existing account resource

## Tests
To run the tests, simply run the following commands in terminal:
```
docker-compose build  # to build the image
docker-compose up # start services and run the tests
```

## Example Usage
```
client := api.NewClient("http://localhost:8080")

// create a new account
acc, err := client.Create(&accounts.AccountData{...})

// Fetch an existing account
acc, err := client.Fetch("eb0bd6f5-c3f5-44b2-b677-acd23cdde73c")

// Delete an existing account
err := client.Delete("eb0bd6f5-c3f5-44b2-b677-acd23cdde73c", 0)

```
## Design Decisions

Following best practices were used during the development
- Project structured into different packages for model types and actual library code.
- A common method to send and retry http requests
- Client HTTP requests has a timeout period for blocking call
- Unit tests and coverage for Client libraries
- Supports different error objects based on the response status code (we could extend this more in future)
- Minimal dependency on external or third party libraries
