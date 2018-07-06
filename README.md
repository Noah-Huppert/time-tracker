# Time Tracker
Open source time tracker.

# Table Of Contents
- [Overview](#overview)
- [Development](#development)
	- [Protocol Buffers](#protocol-buffers)
	- [Database](#database)

# Overview
Provides work time tracking features.

# Development
## Protocol Buffers
Protocol buffers is used with GRPC.  

To compile services and models run:  

```
make proto
```

## Database
PostgreSQL is used to store data.  

To start a development database execute:  

```
make pg
```
