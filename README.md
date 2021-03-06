# <img src="ale-logo2.png" height="100px" alt="ALE">
[![Build Status](https://dev.azure.com/alde08/ale/_apis/build/status/alde.ale?branchName=master)](https://dev.azure.com/alde08/ale/_build/latest?definitionId=1&branchName=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/alde/ale)](https://goreportcard.com/report/github.com/alde/ale)

Automated Log Extractor

## Purpose
The intent for this project is to crawl the workflow API in Jenkins, and extract a more structured log divided into stages.
It'll use the configured regex to try to extract the timestamp from each log line.


### Configuration

The following is the default config
```toml
[server]
address = "0.0.0.0" # IP address to bind
port = 7654 # The Port to bind

[logging]
level = "debug"
format = "text" # Can be json or text

[metadata] # metadata will be presented in the service-metadata route
owner = "${USER}" # Owner of the service

[crawler]
# Regex used to extract the timestamp from the logs.
# Should have two groups, timestamp and log line.
logpattern = '''.*\[([\d{4}\-\d{2}\-\d{2}T\d{2}:\d{2}:\d{2}.\d*Z]*)\].*?\s(.*)$'''
```

See [config_test.toml](config/config_test.toml) for more configuration options.

#### Postgres SQL
To use psql as a backend, add a config similar to:
```toml
[PostgreSQL]
username = "postgres_user"
passwordfile = "/path/to/file/with/password"
host = "postgres.local"
Port = 5432
database = "ale_database_name"
disablessl = true
```

#### Datastore
To use Google Datastore as a backend, add a config similar to:
```toml
[GoogleCloudDatastore]
namespace = "ale-jenkinslog"
project = "my-gcs-project"
```


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

Process a Build:
```bash
curl -XPOST http://ale-server:port/api/v1/process \
    -H "Content-Type: application/json" \
    -d @- << EOF
{
    "buildId": "unique-id-of-build",
    "buildUrl": "http://jenkins.local:8080/job/jobId/262"
}
EOF
```
response:
```json
201 CREATED
{
    "location": "http://ale-server:port/api/v1/build/unique-id-of-build"
}
```
If it has already been crawled, the response will be
```json
302 FOUND
{
    "location": "http://ale-server:port/api/v1/build/unique-id-of-build"
}
```

Query for build information
```bash
curl http://ale-server:port/api/v1/build/unique-id-of-build \
    -H "Accept: application/json"
```
response (sample):
```json
200 OK
{
    "stages": [
        {
            "status": "SUCCESS",
            "name": "Preparation - Delete workspace when build is done",
            "log": [
                {
                    "timestamp": "09:46:24", // Format will depend on your log and regex
                    "line": "[WS-CLEANUP] Deleting project workspace..."
                },
                {
                    "timestamp": "09:46:24",
                    "line": "[WS-CLEANUP] Deferred wipeout is used..."
                },
                {
                    "timestamp": "09:46:24",
                    "line": "[WS-CLEANUP] done"
                }
            ],
            "log_length": 1119,
            "start_time": 1548083830768
        }
    ],
    "status": "SUCCESS",
    "name": "#502 - org/repo - refs/pull/65/merge",
    "id": "502",
    "build_id": "597bc093-6824-4287-8161-f558f8022ded"
}
```

## API
The POST to start processing takes the following input:

* `buildUrl`
    * **Required** The URL of the build to start crawling. The format should be similar to `http://jenkins.internal:8080/job/jobName/714`, and should end in the build number.
* `buildId`
    * **optional** If provided it will be used as the key of the build.
    * If not provided, a Version 4 UUID will be generated and used as a key.
    * Needs to be unique.
* `forceRecrawl`
    * **optional** If provided, an existing database entry with the same buildId (whether provided or generated), will be deleted before the crawl.
    * Defaults to `false`.

## Getting more logs from Jenkins API

Set the following JAVA_OPTS when you launch your Jenkins
```bash
export JAVA_OPTS="${JAVA_OPTS} -Dfile.encoding=UTF-8 -Dcom.cloudbees.workflow.rest.external.FlowNodeLogExt.maxReturnChars=1048576"
```


## TODO
* Only crawl entries that were not previously marked as done
