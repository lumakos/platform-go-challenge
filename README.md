# Few words about the app
In the root of the app, there is a <b>docker-compose.yml</b> file that uses a <b>Dockerfile</b> for the main app which run on port 8088 and a <b>Dockerfile.test</b> for the test suite.

### Steps
1. Get a user's favorite assets
2. Add an asset to favorites
3. Remove an asset from favorites
4. Edit an asset's description

## API Endpoints
> <b>Endpoint:</b> GET /api/v1/users/{userID}/favorites

> <b>Query Parameters:</b> page,limit

> <b>Description:</b> Get a user's favorite assets

```
curl -X GET http://localhost:8088/api/v1/users/{userID}/favorites
```
> Example
```
curl -X GET http://localhost:8088/api/v1/users/1/favorites?page=2&limit=1
```

> Response
```
{
    "data": [
        {
            "id": 1,
            "type": "Insight",
            "Description": "insight-data-text",
            "data": {
                "text": "insight-data-text"
            }
        }
    ],
    "meta": {
        "current_page": 1,
        "per_page": 10,
        "total_count": 1,
        "total_pages": 1
    }
}
```
OR
```
{
    "data": [],
    "meta": {
        "current_page": 1,
        "per_page": 10,
        "total_count": 0,
        "total_pages": 0
    }
}
```
> Error Responses:
```
429: Rate limit exceeded.
```

---

> <b>Endpoint:</b> POST /api/v1/users/{userID}/favorites


> <b>Description:</b> Add an asset to favorites

```
curl -X POST http://localhost:8088/api/v1/users/{userID}/favorites -H "Content-Type: application/json" -d '{
  "type": "AssteType",
  "data": {}                      
}'
```
> Example: Insight
```
curl -X POST http://localhost:8088/api/v1/users/1/favorites -H "Content-Type: application/json" -d '{
  "type": "Insight",
  "data": {                                                 
    "text": "insight-data-text"
  }                      
}'
```
> Example: Chart
```
curl -X POST http://localhost:8088/api/v1/users/1/favorites -H "Content-Type: application/json" -d '{
  "type": "Chart",
  "data": {                                                 
    "title": "Target Audience Analysis",
    "x_axis": "x-axis-test",
    "y_axis": "y-axis-test",
    "data": [12.2]
  }                      
}'
```
> Example: Audience
```
curl -X POST http://localhost:8088/api/v1/users/1/favorites -H "Content-Type: application/json" -d '{
  "type": "Audience",
  "data": {                                                 
    "gender": "female", 
    "birth_country": "Greece",
    "age_range": "25-34",
    "social_media_usage": 12.4,
    "purchases_last_month": 15
  }
}'
```
> Response
```
{
    "data": [
        {
            "Description": "insight-data-text",
            "data": {
                "text": "insight-data-text"
            },
            "id": 1,
            "type": "Insight"
        }
    ]
}
```
> Error Responses:
```
400: Invalid input or asset type.
429: Rate limit exceeded.
```
---
> <b>Endpoint:</b> DELETE /api/v1/users/{userID}/favorites/{assetID}

> <b>Description:</b> Remove an asset from favorites

```
curl -X DELETE http://localhost:8088/api/v1/users/{userID}/favorites/{assetID}
```
> Example
```
curl -X DELETE http://localhost:8088/api/v1/users/1/favorites/1
```
> Repsonse
```
{
    "data": [
        {
            "id": 1,
            "message": {
                "text": "Asset removed from favorites"
            }
        }
    ]
}
```
> Error Responses:
```
404: Asset not found.
```

---
> <b>Endpoint:</b> PUT /api/v1/users/{userID}/favorites/{assetID}

> <b>Description:</b> Edit an asset's description

```
curl -X PUT http://localhost:8088/api/v1/users/{userID}/favorites/{assetID} -H "Content-Type: application/json" -d '{
  "description": "string-text"
}'
```

> Example
```
curl -X PUT http://localhost:8088/api/v1/users/1/favorites/1 -H "Content-Type: application/json" -d '{
  "description": "Updated description: Social media usage trends"
}'
```
> Response
```
{
    "data": [
        {
            "Description": "Updated description: Social media usage trends",
            "id": 1,
            "message": {
                "text": "Asset description updated"
            },
            "type": "Insight"
        }
    ]
}
```
> Error Responses:
```
404: Asset not found.
429: Rate limit exceeded.
```

* <u>You can find the file "GWI.postman_collection.json" in the project root, which contains all the HTTP requests for this assignment.</u>
---

## Build and start the app
```
docker-compose run up -d --build
```

## Stop the app container
```
docker-compose stop
```

## Run tests
```
docker-compose run --rm app-test
```

## Running host
```
http://localhost:8088
```

## Rate Limiting
- Maximum requests: 20 requests per minute per user.
- Time window: 1 minute.
- Response on limit exceeded: 429 - Rate limit exceeded, try again later.

The rate limiting is handled using timestamps stored for each user and is checked every time an action is performed (adding a favorite, editing descriptions, etc.).

## Data Models
### Asset Model

Represents an asset that a user can add to their favorites. It contains the asset type, description, and data (specific to the type of asset).

```
type Asset struct {
  ID          uint      `json:"id"`
  Type        AssetType `json:"type"`
  Description interface{}
  Data        json.RawMessage `json:"data,omitempty"`
}
```
### Asset Types

- <b>Chart</b>: Represents chart data.
- <b>Insight</b>: Represents insight information.
- <b>Audience</b>: Represents demographic information.

Each asset type is validated using the ValidateAsset function which ensures the data format and required fields are correct.

### Validation
Each asset type has its own validation rules:

- <b>Chart</b>: Title, axes, and data fields must be non-empty.
- <b>Insight</b>: Text must not be empty or exceed 500 characters.
- <b>Audience</b>: Gender must be "male" or "female", age range and purchases must be valid.

Invalid input results in a 400 Bad Request error with a specific error message.

## How it works
- The UserStore keeps the favorites for each user, using a sync.Map to store the data with the user ID as the key.
- For rate limiting, the RateLimitStore holds timestamps of each request and ensures users cannot exceed the maximum number of requests.
- The Asset data types are validated upon receiving a request to ensure correctness before adding to a user's favorites list.