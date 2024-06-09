---
authors: Nick Powell (nickjaypowell@gmail.com)
state: draft
---

# RFD - Jobber: Linux Job Runner

## Required Approvers

* Engineering: @rosstimothy && @strideynet && @Tener

## What

Implement a prototype job worker service that provides an API to run arbitrary Linux commands according to this [challenge document](https://github.com/gravitational/careers/blob/main/challenges/systems/challenge-1.md).


## Why

This is a coding challenge designed to test mastery in several skill areas.

* cgroups and linux systems
* remote systems and remote code execution
* gRPC APIs
* authn/authz
* UX
* async work environment and communication

## Details

### Manager

The manager runs jobs, stops jobs, and gets the status of jobs. The `user`, `name` combination of each job is used as its unique identifier. The manager has several linux-only features, so the manager is guarded with build tags.

#### Environment

The manager uses cgroups v2. To keep its environment consistent, testing and running is done in an alpine container.

#### CGroups
```
            ┌───────┐
            │cgroups│
            └┬────┬─┘
             │    │
             │    │
            ┌┴────┴─┐
            │jobber │
            └┬────┬─┘
             │    │
             │    │
 ┌-──────────┴─┐ ┌┴────────────┐
 │    user     │ │    user     │
 └┬────────────┘ └┬───────────┬┘
  │               │           │
  │               │           │
 ┌┴────────┐    ┌─┴───────┐  ┌┴────────┐
 │   job   │    │   job   │  │   job   │
 └─────────┘    └─────────┘  └─────────┘
```

This cgroups hierarchy provides fine-grained control at all levels. For example, setting limits at the jobber level prevents the sum of all processes from overwhelming the machine. Before starting a job, the manager creates a child cgroup at the process level.  If a job specifies resource limits, the limits are set in the cgroup before the process is attached. The manager can set the following limits:

* CPU percentage usage maximum (%)
* Memory usage maximum (MB)
* IO read and write maximums (MB/s)

The manager can attach a process to a cgroup by adding its pid to `cgroup.procs`. Once a job is running, the manager can use cgroups to kill it by writing `1` to `cgroup.kill`. The manager can get the status of a job by reading `cgroup.procs`; the presence or absence of pids conveys whether the job is running or has exited.

#### Process

In order to protect the system files from jobs, job processes are run as a linux user. For this iteration, the only valid user is `nobody`, but this could be expanded in the future for security and convenience. This is accomplished by modifying processes before they are started using `syscall.SysProcAttr`.

Once the appropriate cgroup is created, the process is started and attached to the cgroup. The process is killed if the attach fails. The output of the process is stored in a file at `/tmp/jobber/<user>/<job>/out.txt`.

### API

The job runner is meant to run in a remote linux machine, so it has an accompanying client. The client sends commands over gRPC.

#### Authorization

The server and client authorize each other's certificates (mTLS). For this iteration, the CAs and certs are pre-generated. Below is the server and client tls configuration.

```go
    // Server
    config := &tls.Config{
        Certificates: []tls.Certificate{serverCert},
        // pool of acceptable client CAs
        ClientCAs:    caCertPool,
        // require verify client CA
        ClientAuth:   tls.RequireAndVerifyClientCert,
        // set minimum version to latest
        MinVersion:   tls.VersionTLS13,
        // TLS 1.3 ciphersuites are not configurable
    }
```

```go
    // Client
    config := &tls.Config{
        Certificates: []tls.Certificate{clientCert},
        // pool of acceptable server CAs
        RootCAs:      caCertPool,
        // set minimum version to latest
        MinVersion:   tls.VersionTLS13,
        // TLS 1.3 ciphersuites are not configurable
    }
```

#### Authentication

Both a `UnaryInterceptor` and a `StreamInterceptor` are added to the server. These interceptors both extract the peer certificates from the context. If a certificate's subject contains a matching token, the request is authenticated.

#### Proto

```proto
service Jobber {
    rpc Start(StartRequest) returns (StartResponse);
    rpc Stop(StopRequest) returns (StopResponse);
    rpc Status(StatusRequest) returns (StatusResponse);
    rpc Stream(StreamRequest) returns (stream StreamResponse);
}

message StartRequest {
    Limits limits = 1;
    
    // used in exec.Cmd
    string cmd = 2;
    repeated string args = 3;
}

message StartResponse {
    string name = 1;
}

message StopRequest {
    string name = 1;
}

message StopResponse {}

message StatusRequest {
    string name = 1;
}

message StatusResponse {
    // type manager.Status
    int64 status = 1;
}

message StreamRequest {
    string name = 1;
}

message StreamResponse {
    bytes data = 1;
}

////

message Limits {
    // Percentage of CPU cycle [1-100]
    int64 cpuMax = 1; 
    // MB
    int64 memMax = 2; 
    // MB per second
    int64 readMax = 3; 
    // MB per second
    int64 writeMax = 4; 
}
```

#### Concurrency

The only race condition/collision occurs if two clients create a job with the same name. To prevent this from happening, unique job names are randomly generated by the server.

#### Job Output Stream

```go
func CopyFollow(ctx context.Context, filepath string, w io.Writer) error
```

`CopyFollow` uses `io.Copy` to read from file `/tmp/jobber/<user>/<job>/out.txt` to `w`. When EOF is reached, the function will backoff (sleep), and periodically check the modification time of the file. If the file has been modified, `io.Copy` is run again with the same reader, so that file offset is retained. This continues until `ctx` is cancelled.

```go
type StreamWriter struct {
    send func(*pb.StreamResponse) error
}

func (s *StreamWriter) Write(p []byte) (int, error)
```

`StreamWriter` wraps `Jobber_StreamServer.Send`. It is used as the `io.Writer` parameter for `CopyFollow` so that `io.Copy` writes to the gRPC stream.

### UX

The client uses `cobra` client framework. There are four subcommands:

#### Start
Starts a job and prints the generated job name

`start [flags] <cmd> [<cmd args...>]`

| flag | description | 
|------|-------------|
|`-s, --stream`|also streams the job's output|
| `--cpu-max uint64`|sets maximum percentage of CPU cycle ([1-100]%)|
|`--mem-max uint64`|sets maximum memory usage (MB)|
|`--read-max uint64`|sets maximum disk read speed (MB/s)|
|`--write-max uint64`|sets maximum disk write speed (MB/s)|

If the `--stream` flag is enabled, the start command will call the `Start` rpc followed by the `Stream` rpc. The user is able to easily run short-running jobs like `start -s ls` that will immediately produce the output the user wants.

#### Stop
Stops a job via SIGKILL

`stop [flags] <name>`

| flag | description | 
|------|-------------|
|`-s, --stream`|also streams the job's output|

#### Status
Prints the status of a job

`status [flags] <name>`

#### Stream
Streams the output of a job

`stream [flags] <name>`

### Testing Plan

#### Unit

Core operations of the code are unit tested.

* cgroup operations
* starting a process as `nobody` user
* reading and following a file (`CopyFollow`)
* mtls authn/authz

The manager is designed to function in an alpine environment, so that is where it is tested. `make test` runs `go test -v ./...` in a `golang:alpine` container.

#### Integration

To ensure that all the components play together nicely, an integration test starts a job and streams its output.

### Future Considerations

* input validation/sanitization
* automatically stop output streams when the job exits
* use something like `fsnotify` to watch output files
* have a linux user for each authn user, for resource limits and usage
* ability to delete jobs
* job isolation
* testing with mocks