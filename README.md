# ALE
[![Build Status](https://travis-ci.org/alde/ale.svg?branch=master)](https://travis-ci.org/alde/ale)
[![Coverage Status](https://coveralls.io/repos/github/alde/ale/badge.svg?branch=master)](https://coveralls.io/github/alde/ale?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/alde/ale)](https://goreportcard.com/report/github.com/alde/ale)

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
response (sample):
```json
200 OK
{
    "stages": [
        {
            "status": "SUCCESS",
            "name": "Preparation - Delete workspace when build is done",
            "log_text": "\u003cspan class=\"timestamp\"\u003e\u003cb\u003e15:17:10\u003c/b\u003e \u003c/span\u003e\u003cstyle\u003e.timestamper-plain-text {visibility: hidden;}\u003c/style\u003e[WS-CLEANUP] Deleting project workspace...\n\u003cspan class=\"timestamp\"\u003e\u003cb\u003e15:17:10\u003c/b\u003e \u003c/span\u003e\u003cstyle\u003e.timestamper-plain-text {visibility: hidden;}\u003c/style\u003e[WS-CLEANUP] Deferred wipeout is used...\n\u003cspan class=\"timestamp\"\u003e\u003cb\u003e15:17:10\u003c/b\u003e \u003c/span\u003e\u003cstyle\u003e.timestamper-plain-text {visibility: hidden;}\u003c/style\u003e[WS-CLEANUP] done\n",
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


## TODO
* Only crawl entries that were not previously marked as done
* Figure out how to get around jenkins' "hasMore: true" annoyance (although maybe for a separate call)
    * From jenkins plugin doc:
    ```
    Hardcoded API limits that may be overridden by setting the properties at startup (requires restarting Jenkins to see the change):
        * Characters in each step's log entry (default: 10240 or 10kB) - com.cloudbees.workflow.rest.external.FlowNodeLogExt.maxReturnChars
    ```
