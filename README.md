![example usage](./assets/cyops.png)
# dd-inserter - Simple emitter of process data
**cyops-se**: *This application is part of the cyops.se community and use the same language and terminology. If there are acronyms or descriptions here that are unknown or ambiguous, please visit the [documentations](https://github.com/cyops-se/docs) site to see if it is explained there. You are welcome to help us improve the content regardless if you find what you are looking for or not*.

## Introduction
**NOTE! In this version, the application is dependent on a TimescaleDB emitter as some meta information like description, location, unit, min, max, is stored there. Please refer to section below on how to set it up for version 0.2.0**

This application (```dd-inserter```) receives one way UDP messages produced by [dd-opcda](https://github.com/cyops-se/dd-opcda) and inserts the data into a timescale database or/and forwards them on the RabbitMQ channel.

Typical usage is as the outer end of a data diode used to replicate real time data from a sensitive system to a potentially hostile network in order to maintain full network isolation of the sensitive network.

![example usage](./assets/diode-1.png)

## Overview
```dd-inserter``` listens by default at UDP port 4357 (configurable) for messages sent by ```dd-inserter``` through a data diode. The messages are interpreted and stored in a specified Timescale database (localhost by default).

***This is a very basic solution with a lot of opportunities for improvements. Suggestions are welcome!***

## Sequence checks
To detect data loss, a simple sequence number is sent with every message. Data loss is reported as soon as the sequence is broken which is not fully reliable as UDP packets are not guaranteed to come in the order they were sent.

# Build
The application has been successfully been tested on Windows (x86_64) and Linux (x86_64) but should work on any platform that is supported by the third party modules that this application depends on.

Run the following commands to build a ```dd-inserter``` executable (Windows)

```
> go get github.com/cyops-se/dd-inserter
> cd %GOPATH%\src\github.com\cyops-se\dd-inserter
> go build
```

# Install as Windows service

Start a command prompt (cmd) or powershell terminal as Administrator (or as a user with privileges to install new services) and navigate to the directory where the dd-opcda executable is located.

Run the following command:

```
.\dd-inserter -cmd install
```

This will create a new service entry. Now open the services management console (services.msc) and find 'dd-inserter from cyops-se'. Change start method and account information if needed. Start method must be changed to **Automatic** for the service to start automatically after a system reboot, and the account under which the service runs must be permitted to access the emitters (TimescaleDB and/or RabbitMQ).

To remove the service, just run the following command as Administrator:

```
.\dd-inserter -cmd remove
```

# TimescaleDB emitter
The ```dd-inserter``` TiemscaleDB emitter is dependent on a TimescaleDB timeseries database with specific table structures. Instructions on how to install the Timescale timeseries database for use with ```dd-inserter``` (and Grafana) can be found [here!](./TIMESCALE.md)

## Set up a timescale database for use with ```dd-inserter``` version 0.1.0

The application is currently hardcoded to use a specific hyper table and structure according to the following PostgreSQL commands as shown below.

***You must now use the password entered during the installation of the PostgreSQL database engine***

```
> psql -U postgres
Password for user postgres: XXXX
```
```
psql# CREATE DATABASE processdata;
psql# \c processdata
psql# CREATE EXTENSION IF NOT EXISTS timescaledb;
psql# CREATE TABLE measurements (time TIMESTAMPTZ NOT NULL, name TEXT NOT NULL, value DOUBLE PRECISION NOT NULL, quality NUMERIC NOT NULL);
psql# SELECT create_hypertable('measurements','time');
```

If you get an error when creating the extension, it is probably because you didn't restart the PostgreSQL windows service after installing TimescaleDB. See [this instruction](./TIMESCALE.md) for more information.



## Set up a timescale database for use with ```dd-inserter``` version 0.2.0

The application is currently hardcoded to use a specific hyper table and structure according to the following PostgreSQL commands as shown below.

***You must now use the password entered during the installation of the PostgreSQL database engine***

```
> psql -U postgres
Password for user postgres: XXXX
```
```
psql# \c postgres
psql# create table measurements.raw_measurements (time TIMESTAMPTZ NOT NULL, tag INTEGER NOT NULL, value DOUBLE PRECISION, quality INTEGER);
psql# select create_hypertable('measurements.raw_measurements','time');
psql# create sequence measurements.tags_tag_id_seq;
psql# create table measurements.tags (tag_id integer NOT NULL DEFAULT nextval('measurements.tags_tag_id_seq'), name text, description text, location text, type text, unit text, min double precision, max double precision);

psql# alter table measurements.tags add primary key(tag_id);
psql# alter table measurements.tags add constraint tags_name_key unique (name);

GRANT USAGE ON SEQUENCE tags_tag_id_seq TO [user];
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA measurements TO [user];
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA measurements TO [user];
```

If you get an error when creating the extension (select create_hypertable()), it is probably because you didn't restart the PostgreSQL windows service after installing TimescaleDB. See [this instruction](./TIMESCALE.md) for more information.

# RabbitMQ emitter
To be defined

# User interface
A simple user interface is provided to configure, operate and monitor the application health. See [the user interface section](./USERINTERFACE.md) for more information.