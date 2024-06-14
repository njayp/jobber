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

The manager runs jobs, stops jobs, and gets the status of jobs. The `user`, `id` combination of each job is used as its unique identifier. The manager has several linux-only features, so the manager is guarded with build tags.

#### Environment

The manager uses cgroups v2. To keep its environment consistent, testing and running is done in an alpine container.

#### Binary

usage | description | notes |
|-|-|-|
|`start [flags]`|main command, starts the manager|
|`spawn [flags] <cmd> [<cmd args...>]`|starts a child process in a target cgroup | started as a child process by the manager. target cgroup is specified with flags

#### CGroups

Before controllers can be added to the cgroups hierarchy, the init process must be moved to a leaf cgroup [(docs)](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html#no-internal-process-constraint). `jobber` is designed to be the init process; it moves itself by making a new folder `cgroups/jobber`, then writing `1` to `cgroups/jobber/cgroup.procs`.

```
                 ┌───────┐
                 │cgroups│
                 └┬────┬─┘
                  │    │
                  │    │
            ┌─────┴─┐ ┌┴─────┐
            │ jobs  │ │jobber│
            └┬────┬─┘ └──────┘
             │    │
             │    │
 ┌-──────────┴─┐ ┌┴────────────┐
 │   <user>    │ │   <user>    │
 └┬────────────┘ └┬───────────┬┘
  │               │           │
  │               │           │
 ┌┴────────┐    ┌─┴───────┐  ┌┴────────┐
 │  <job>  │    │  <job>  │  │  <job>  │
 └─────────┘    └─────────┘  └─────────┘
```

This cgroups hierarchy provides fine-grained control at all levels. For example, setting limits at the jobber level prevents the sum of all processes from overwhelming the machine. To enable the necessary controllers in all cgroups, `+cpu +memory +io +pids` is written to `cgroup.subtree_control` in all non-leaf cgroups starting from the root. Before starting a job, the manager creates a child cgroup of the appropriate user cgroup, and sets the following limits by writing the limit to the corresponding file. Setting a maximum number of processes prevents fork bomb attacks.

| limit | file | data |
|-|-|-|
| maximum percentage usage of one cpu core (10%) | cpu.max | 10000 100000 | 
| memory usage maximum (10M) | memory.max | 10M |
| io read and write maximums (1MB/s) | io.max | 1:0 wbps=1048576 rbps=1048576 |
| maximum number of processes (100) | pids.max | 100 |

The manager can kill all processes in a cgroup by writing `1` to `cgroup.freeze`. The manager can get the status of a job by reading `cgroup.events`; `populated` shows whether there are processes running, and `frozen` shows whether those processes are frozen.

#### Process

In order to protect the system files from jobs, job processes are run as a linux user. This is accomplished by modifying processes before they are started using `syscall.SysProcAttr`. For this iteration, the only valid user is `nobody`, but this could be expanded in the future for security and convenience.

To start a process in the appropriate cgroup, `jobber` will call itself with the subcommand `spawn`. This child process moves itself to the target cgroup and starts the job process so that the job process is automatically added to the target cgroup. The stdout of the job process is stored in a file at `/tmp/jobber/<user>/<job>/out.txt`, and the stderr at `/tmp/jobber/<user>/<job>/err.txt`. The server monitors the stderr of its child process for errors.

### API

The job runner is meant to run in a remote linux machine, so it has an accompanying client. The client sends commands over gRPC.

#### Authentication

The server and client authorize each other's certificates (mTLS). For this iteration, the CAs and certs are pre-generated. The client certificate functions as the client secret. Below is the server and client tls configuration. 

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

#### Authorization

Both a `UnaryInterceptor` and a `StreamInterceptor` are added to the server. These interceptors both extract the peer certificates from the context. They use the `Subject Alternate Name`, `EmailAddresses`, to identify the user.

There are two hard-coded roles.

| role | permissions |
| - | - |
| writer | start, stop, status, stream |
| reader | status, stream |

Each user has an assigned role. If a user tries to call an rpc that its role does not have permission to call, the request is rejected. For this iteration, there are two hard-coded users, `reader-user` and `writer-user`.

#### Proto

```proto
service Jobber {
    // Start throws error if process does not start.
    rpc Start(StartRequest) returns (StartResponse);
    // Stop does not wait for the cgroup to exit. Status should be used 
    // to check whether a process has exited.
    rpc Stop(StopRequest) returns (StopResponse);
    // IDEA watching functionality should be added to this rpc.
    rpc Status(StatusRequest) returns (StatusResponse);
    // Stream copies and follows one file for neatness and control. 
    // It is currently called twice, once for "out.txt", and once for "err.txt". 
    // Reusing the same client will reuse the same connection for both calls. 
    rpc Stream(StreamRequest) returns (stream StreamResponse);
}

message StartRequest {
    // used in exec.Cmd
    string cmd = 1;
    repeated string args = 2;
}

message StartResponse {
    string id = 1;
}

message StopRequest {
    string id = 1;
}

message StopResponse {}

message StatusRequest {
    string id = 1;
}

message StatusResponse {
    State state = 1;
}

message StreamRequest {
    string id = 1;
    StreamIndicator streamIndicator = 2;
}

message StreamResponse {
    bytes data = 1;
}

////

// If State is unknown, the rpc throws error
enum State {
    State_Unspecified = 0;
    Running = 1;
    Exited = 2;
    Frozen = 3;
}

enum StreamIndicator {
    StreamIndicator_Unspecified = 0;
    StdOut = 1;
    StdErr = 2;
}
```

#### Concurrency

The only race condition/collision occurs if two clients create a job with the same id. To prevent this from happening, the server generates a UUID for each job id.

#### Job Output Stream

```go
func CopyFollow(ctx context.Context, filepath string, w io.Writer) error
```

`CopyFollow` uses `io.Copy` to read from file `/tmp/jobber/<user>/<job>/<filepath>` to `w`. When EOF is reached, the function will use `syscall.InotifyAddWatch` to watch the file. Once the watcher receives a `syscall.IN_MODIFY` event, `io.Copy` is run again with the same reader, so that file offset is retained. This continues until `ctx` is cancelled.

```go
type StreamWriter struct {
    send func(*pb.StreamResponse) error
}

func (s *StreamWriter) Write(p []byte) (int, error)
```

`StreamWriter` wraps `Jobber_StreamServer.Send`. It is used as the `io.Writer` parameter for `CopyFollow` so that `io.Copy` writes to the gRPC stream.

### UX

The client uses `cobra` client framework. There are four subcommands:

| usage | description | notes |
|-|-|-|
|`start [flags] <cmd> [<cmd args...>]`|Starts a job and prints the generated job id|Throws error if the command fails to start
|`stop [flags] <job id>`|Stops a job via cgroups, which will send a SIGKILL|
|`status [flags] <job id>`|Prints the status of a job|
|`stream [flags] <job id>`|Streams stdout and stderr of a job|Stream calls the Stream rpc twice, once for `stdout`, and once for `stderr`.

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

* add watch to Status rpc
* input validation/sanitization
* automatically stop output streams when the job exits
* have a linux user for each authn user, for resource limits and usage
* ability to delete jobs
* job isolation
* testing with mocks