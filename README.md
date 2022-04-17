# latency-monitor
A Golang CLI service to poll and monitor latencies of REST APIs

# setting up the config file

Please do ensure that "timings" node is maintained. "endpoints" array can have any number of monitors - only method, name and url are mandatory. within the endpoint object - basicAuth and headers array is optional.

```
{
  "timings":{
    "intervalSeconds": 10,
    "runDurationHours": 1
  },
  "endpoints": [
    {
      "method": "GET",
      "name": "Google APIs Service Direcrory",
      "url": "https://servicedirectory.googleapis.com/$discovery/rest?version=v1"
    },
    {
      "method": "GET",
      "name": "ReqRes Slow 2sec minimum API",
      "url": "https://reqres.in/api/users?delay=2"
    },
    {
      "method": "GET",
      "name": "Apigee Mock API - Echo",
      "url": "https://mocktarget.apigee.net/echo",
      "basicAuth": {
        "userName": "test",
        "password": "cat"  
      },
      "headers": [
        {
          "name": "client_id",
          "value": "5387hkjdhfkj34h5"
        },
        {
          "name": "client_secret",
          "value": "hj3hk465h4k6jh456"
        }
      ]    
    }
  ]
}

```

# Running the service

go run .\latencyMonitor.go

A successful startup and run looks as below -

```
Reading config from config/config.json
Logging to file - logs/latencies-17-April-PID_13964.log
Endpoint 1 "Google APIs Service Direcrory"       https://servicedirectory.googleapis.com/$discovery/rest?version=v1
Endpoint 2 "ReqRes Slow 2sec minimum API"        https://reqres.in/api/users?delay=2
Endpoint 3 "Apigee Mock API - Echo"      https://mocktarget.apigee.net/echo
Using config as 10 seconds interval and 1 hour run duration
Starting up ... 
Started up 3 pollers ...

System: Google APIs Service Direcrory    HTTP Status 200         2.04 seconds
System: ReqRes Slow 2sec minimum API     HTTP Status 200         2.50 seconds
System: Apigee Mock API - Echo           HTTP Status 200         280 ms

System: Google APIs Service Direcrory    HTTP Status 200         343 ms
System: ReqRes Slow 2sec minimum API     HTTP Status 200         2.19 seconds
System: Apigee Mock API - Echo           HTTP Status 200         241 ms

System: Google APIs Service Direcrory    HTTP Status 200         343 ms
System: ReqRes Slow 2sec minimum API     HTTP Status 200         2.20 seconds
System: Apigee Mock API - Echo           HTTP Status 200         238 ms
```
