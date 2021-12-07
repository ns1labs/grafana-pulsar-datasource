# Running Grafana in Docker

Two files are provided to run Grafana in Docker. This makes things easier for developing plugins under Linux.

* Grafana.ini. This is a very basic configuration file that enables loading of the `ns1labs-pulsar-datsource`locally and
without checking its signature.
* Grafana.yml. Docker-compose file for creating a container for Grafana.

## Requirements
You need to have installed docker and docker-compose.

## Running
1. Set up the environment variable `GRAFANA_PLUGINS_ROOT_DEV` with the location of your plugins. I.e.
```shell
export GRAFANA_PLUGINS_ROOT_DEV="${HOME}/Workspace/grafana_plugins"
```
This is the directory that Grafana will scan for user installed plugins.
2. The first time you can execute
```shell
docker-compose -f grafana.yml up
```
This command will download the Grafana image if not present, and create a container. Port mapped is the same, 3000.
If you already executed `docker-compose up`, you just can execute
```shell
docker-compose -f grafana.yml start
```

Every time you create a new container with `docker-compose up`, the password for user `admin` will be reset to its default
and the user will be asked to change it.

Don't forget that on every change made to the `grafana-pulsar-datasource` you have to restart docker to have Grafana load
the changes
```shell
docker-compose -f grafana.yml restart
```

