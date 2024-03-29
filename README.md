# About qBeat

Welcome to qBeat. This beat is used to get monitoring / statistics data out of IBM MQ.

What is a beat? A beat is a lightweight data shipper written in GOLANG developed by Elastic N.V. and the community. It is open source and also implements a framework offered to the community to build their own beats. Those beats may offer special purpose data collection not offered by existing beats. Machinebeat is one of those special purpose beats.

This version of the beat is working fine, but you need to compile it in your environment. You need to have IBM MQ installed in that environment as it is referring to the C library.

It would be great if you could share feedback using Github issues.

## Features
* Collect Q Manager Status
* Collect Q Status and Statistics
* Collect Channel Status
* Collect any response of an Inquire PCF message
* Collect error logs with Filebeat
* Correlate data between Metrics and Logs
* Collect data from remote queue managers
* Collect data from local and remote environments including cloud
* Ready to start using Kibana objects

## Current status
This implementation will be merged into 
* [Metricbeat](https://www.elastic.co/de/products/beats/metricbeat) (also see [PR](https://github.com/elastic/beats/pull/8870))
* [Filebeat](https://www.elastic.co/de/products/beats/fetricbeat) (also see [PR](https://github.com/elastic/beats/pull/8782))

## Getting Started with Qbeat

### Running Qbeat in a docker container

Simply run the provided docker-compose file with `docker compose up -d`

### Requirements for building Qbeat

* [Golang](https://golang.org/dl/) >v1.17
* [IBM MQ](https://www.ibm.com/de-de/marketplace/secure-messaging) Tested with >v8 but should also work with older versions
* [Elastic Stack](https://cloud.elastic.co) >v7.0

Ensure that this folder is at the following location:
`${GOPATH}/src/github.com/felix-lessoer/qbeat`

Make sure that the "Path to MQ lib" folders exists in your env and that the beat user has sufficient permissions to the files before building and running the beat.
If necessary the MQ lib path can be changed in the source files (mqi.go)

#### Linux

##### Path to MQ lib
* /opt/mqm/inc
* /opt/mqm/lib64

##### Step by step guide
* Install the Go runtime and compiler. On Linux, the packaging may vary but a typical directory for the code is `/usr/lib/golang`.

* Create a working directory. For example, ```mkdir $HOME/gowork```

* Set environment variables. Based on the previous lines,

  ```export GOROOT=/usr/lib/golang```

  ```export GOPATH=$HOME/gowork```

* If using a version of Go from after 2017, you must set environment variables to permit some compile/link flags. This is due to a security fix in the compiler.

  ```export CGO_LDFLAGS_ALLOW="-Wl,-rpath.*"```

* Install the git client

#### Windows

##### Path to MQ lib
* C:/Program Files/IBM/MQ/Tools/c/include
* C:/Program Files/IBM/MQ/bin64

##### Step by step guide
* Install the Go runtime and compiler. On Windows, the common directory is `c:\Go`
* Ensure you have a gcc-based compiler, for example from the Cygwin distribution. I use the mingw variation, to ensure compiled code can be used on systems without Cygwin installed
* Create a working directory. For example, ```mkdir c:\Gowork```
* Set environment variables. Based on the previous lines,

  ```set GOROOT=c:\Go```

  ```set GOPATH=c:\Gowork```

  ```set CC=x86_64-w64-mingw32-gcc.exe```

* The `CGO_LDFLAGS_ALLOW` variable is not needed on Windows
* Install the git client

### Kibana dashboards

Download the file to your workstation and upload it via Kibana. It will setup all necessary objects. But not all visualizations will work without changing the default data model of the beat. Also Machine Learning Jobs are not included.
[Download here](https://github.com/felix-lessoer/qbeat/blob/master/Kibana/MQ-Demo-objects.json)

### Configuration

The general configuration file is qbeat.yml . There are multiple configurations that are available:

* period[required]: Defines after which period of time the data should be collected
* bindingQueueManager[required]: Defines the queue manager that is used to collect the metrics. Need to be available in your local setup of via the MQ configured in the connection config 
* targetQueueManager: Array of queue manager names. Defines the remote queue managers that should be used to collect metrics. If not defined the binding queue manager will be used. A remote queue manager need to be accessible from the bindingQueueManager. Multiple hops are possible.

* More details can be found in qbeat.yml

## How to build on your own env

1.) Download all dependencies from go.mod using `go get -u`

2.) You may need to overwrite some modules with the following versions that do not support go.mod in older versions
```
go get k8s.io/client-go@kubernetes-1.14.8
go get k8s.io/api@kubernetes-1.14.8
go get k8s.io/apimachinery@kubernetes-1.14.8
```
3.) Run `go build` in the qbeat repository
