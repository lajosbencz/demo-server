# demo-server

### Creating resources

```bash
curl -k -X POST localhost:8080/ -H "Content-Type: application/json" -d "
{
    \"name\": \"foobar\",
    \"value\": {
        \"some\": {
            \"arbitrary\": [\"data\", 1, 2, 3]
        }
    },
}"

# {"error":false, "payload": {"some":{"arbitrary":["data",1,2,3]}}}
```

### Overwriting resource error

```bash
curl -k -X POST localhost:8080/ -H "Content-Type: application/json" -d "
{
    \"name\": \"foobar\",
    \"value\": {
        \"another\": {
            \"arbitrary\": [\"data\", 1, 2, 3]
        }
    },
}"

# {"error":true, "message": "Resource [foobar] already exists!"}
```

### Overwriting resource forcefully

```bash
curl -k -X POST localhost:8080/ -H "Content-Type: application/json" -d "
{
    \"name\": \"foobar\",
    \"overwrite\": true,
    \"value\": {
        \"another\": {
            \"arbitrary\": [\"data\", 1, 2, 3]
        }
    },
}"

# {"error":false, "payload": {"another":{"arbitrary":["data",1,2,3]}}}
```

### Reading resources

```bash
curl -k localhost:8080/foobar

# {"error":false, "payload": {"another":{"arbitrary":["data",1,2,3]}}}
```

### Deleting resources

```bash
curl -k -X DELETE localhost:8080/foobar

# {"error":false, "payload": {"another":{"arbitrary":["data",1,2,3]}}}
```
