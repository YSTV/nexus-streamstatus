# Nexus StreamStatus

Small telemetry program which reports the online status (and in future, metadata) of a video stream into a nexus stream distribution server.

## Example Nginx-rtmp config
```
application yourapp {
    exec_push /path/to/nexus-streamstatus -server your.nexusserver.host -name $name -clientaddr $addr
}
```
