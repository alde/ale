# ALE

Automated Log Extractor

## Flow

```
POST
user     ALE     Jenkins      Database
-+--------+---------+------------+----
 |        |         |            |
 +------->|         |            |
 |        +--------------------->|
 |<-------+         |            |
 |        +-------->| poll       |
 |        |<--------+ !done      |
 |        +--------------------->|

GET
 user     ALE     Jenkins      Database
-+--------+---------+------------+----
 |        |         |            |
 +------->|         |            |
 |        +--------------------->|
 |        |<---------------------+
 |<-------+         |            |
```

## Usage

Watch/process a Build:
```bash
curl -XPOST http://ale-server:port/api/v1/process \
    -H "Content-Type: application/json" \
    -d @- << EOF
{
    "org": "github-org",
    "repo": "github-repo",
    "buildId": "unique-id-of-build",
    "buildUrl": "http://jenkins.local:8080/job/jobId/262"
}
EOF
```
response:
```json
 201 Created
{
    "location": "http://ale-server:port/api/v1/build/unique-id-of-build"
}
```

Query for build information
```bash
curl http://ale-server:port/api/v1/build/unique-id-of-build \
    -H "Accept: application/json"
```
response:
```json
200 OK
{
    // to be determined
}
```
