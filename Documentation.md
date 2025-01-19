# Steps
1. Get a user's favorite assets
2. Add an asset to favorites
3. Remove an asset from favorites
4. Edit an asset's description

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

## cURL commands
```
curl -X GET http://localhost:8088/api/v1/users/{userID}/favorites
curl -X GET http://localhost:8088/api/v1/users/1/favorites?page=2&limit=1
```

```
curl -X POST http://localhost:8088/api/v1/users/{userID}/favorites -H "Content-Type: application/json" -d '{
  "type": "AssteType",
  "data": {}                      
}'
```

```
curl -X POST http://localhost:8088/api/v1/users/1/favorites -H "Content-Type: application/json" -d '{
  "type": "Insight",
  "data": {                                                 
    "text": "insight-data-text"
  }                      
}'
```

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

```
curl -X DELETE http://localhost:8088/api/v1/users/{userID}/favorites/{assetID}
curl -X DELETE http://localhost:8088/api/v1/users/1/favorites/1
```

```
curl -X PUT http://localhost:8088/api/v1/users/{userID}/favorites/{assetID} -H "Content-Type: application/json" -d '{
  "description": "Updated description: Social media usage trends"
}'
curl -X PUT http://localhost:8088/api/v1/users/1/favorites/1 -H "Content-Type: application/json" -d '{
  "description": "Updated description: Social media usage trends"
}'
```

## Rate Limiting
Users are restricted to 5 requests per minute to prevent abuse.
