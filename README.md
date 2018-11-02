# Qbeat

Welcome to Qbeat. This beat is used to get monitoring / statistics data out of IBM MQ.
It is currently under development so there is no guarantee that it is working fine.

Ensure that this folder is at the following location:
`${GOPATH}/src/github.com/felix-lessoer/qbeat`

As this is under development it would be great if you could share Feedback using Github issues.

## Getting Started with Qbeat

### Requirements

* [Golang](https://golang.org/dl/) 1.7
* [IBM MQ](https://www.ibm.com/de-de/marketplace/secure-messaging) Tested with v.9 but should also work with older versions

Make sure that the following folders exists and that your user has sufficient permissions to the files before building the beat.
If necessary it can be changed in the source files (mqi.go)

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

### Init Project
To get running with Qbeat and also install the
dependencies, run the following command:

```
make setup
```

It will create a clean git history for each major step. Note that you can always rewrite the history if you wish before pushing your changes.

To push Qbeat in the git repository, run the following commands:

```
git remote set-url origin https://github.com/felix-lessoer/qbeat
git push origin master
```

For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).

### Build

To build the binary for Qbeat run the command below. This will generate a binary
in the same directory with the name qbeat.

```
make
```


### Run

To run Qbeat with debugging output enabled, run:

```
./qbeat -c qbeat.yml -e -d "*"
```


### Update

Each beat has a template for the mapping in elasticsearch and a documentation for the fields
which is automatically generated based on `fields.yml` by running the following command.

```
make update
```


### Cleanup

To clean  Qbeat source code, run the following commands:

```
make fmt
make simplify
```

To clean up the build directory and generated artifacts, run:

```
make clean
```


### Clone

To clone Qbeat from the git repository, run the following commands:

```
mkdir -p ${GOPATH}/src/github.com/felix-lessoer/qbeat
git clone https://github.com/felix-lessoer/qbeat ${GOPATH}/src/github.com/felix-lessoer/qbeat
```


For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).


## Packaging

The beat frameworks provides tools to crosscompile and package your beat for different platforms. This requires [docker](https://www.docker.com/) and vendoring as described above. To build packages of your beat, run the following command:

```
make package
```

This will fetch and create all images required for the build process. The whole process to finish can take several minutes.
