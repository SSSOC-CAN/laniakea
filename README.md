# Laniakea
Repository for the Laniakea Daemon (laniakea)

# Installation

Laniakea runs on Go 1.17

On Linux:

```
$ wget https://golang.org/dl/go1.17.1.linux-amd64.tar.gz
$ sha256sum go1.17.1.linux-amd64.tar.gz | awk -F " " '{ print $1 }'
dab7d9c34361dc21ec237d584590d72500652e7c909bf082758fb63064fca0ef
```
`dab7d9c34361dc21ec237d584590d72500652e7c909bf082758fb63064fca0ef` is what you should see to verify the SHA256 checksum
```
$ sudo tar -C /usr/local -xzf go1.17.1.linux-amd64.tar.gz
```
Now modify your `.bashrc` or `.bash_profile` and add the following lines:
```
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/Path/To/Your/Working/Directory
export PATH=$PATH:$GOPATH/bin
```
If you type `$ go version` you should see `go1.17 linux/amd64`
## Installing Laniakea from source

In your working directory (which is your `GOPATH`), create a `/src/github.com/SSSOC-CAN` directory
```
$ mkdir src/github.com/SSSOC-CAN
$ cd src/github.com/SSSOC-CAN
```
Then clone the repository
```
$ git clone git@github.com:SSSOC-CAN/fmtd.git
$ cd fmtd
```

## Installing Protoc

In order to compile protos in Go, you need `protoc`.

On Linux:

```
$uname -m
x86_64
```
To check your OS architecture. Once you find that, go to https://github.com/protocolbuffers/protobuf/releases and find the releases that matches your OS and architecture. In your $HOME directory:
```
$ curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v3.17.3/protoc-3.17.3-linux-x86_64.zip
$ unzip protoc-3.17.3-linux-x86_64.zip -d $HOME/.local
```
Add the following line to `.bashrc` or `.bash_profile`
```
export PATH=$PATH:$HOME/.local/bin
```
Then
```
$ source .bashrc            #or source .bash_profile
$ protoc --version
libprotoc 3.17.3
```

# Contribution Guide

## Coding Style
Our aim is to write the most legible code we can within the constraints of the language. [This blog post](https://golang.org/doc/effective_go) is a good start for understanding how to write effective Go code. Important things to retain would include:
- using `switch` and `case` instead of `if` and `else if` when appropriate
- declaring global variables and constants as well as import statements in one block instead of many seperate lines. Ex:
```
import (
    "fmt"
    "os"
)

const (
    pi = 3.14
)

var (
    foo string
    bar bool
)
```

## Comments
Ideally, we'd like to use Go's native automated documentation tool `godoc`. For that, we must comment functions, methods, structs, interfaces, packages, etc. Using the following:
- a single or multi line comment above the declaration statement of what it is we are documenting starting with the exact name of that object. Ex:
```
// foo returns the product of two integers
func foo(a, b int) int {
    return a+b
}
```
- This rule also applies to packages but package documentation starts with the word `Package` and then is followed by the name of the package. Ex:
```
// Package bar is the place you want to be 
package bar
```

For effective work with teams, work on a go file that is merged into the `main` branch, but has future work that remains to be completed, please add `TODO:<name> -` comments in areas where future work will be performed. For example:
```
func sum(a, b int) int {
    // TODO:SSSOCPaulCote - Add error handling for invalid types
    return a+b
}
```
In general, any useful comments to understand the context of your code are encouraged.

## Git Workflow

When making any changes to the codebase, please create a feature branch and send frequent commits on the feature branch. The name of the new feature branch should be short and contain `iss<number_of_the_issue>` for example `iss28_new_command`. When the new feature is ready to be merged, create a merge request. In the creation of the merge request, please add any of the organization administrators as reviewers and await their approval before merging any changes into `main`. Additionally, a reference to the issue in Github **MUST** be included in the merge request. Which means any and all work is going to be documented using issues.
```
$ git checkout -b <name_of_new_branch>          # Creates a feature branch with specified name and carries over any uncommited changes from the previous branch
$ git add -A                                    # Adds all modified files to the commit
$ git commit -m "useful comments"               # Creates a commit with useful comments attached
$ git push origin <name_of_new_branch>          # pushes the commit on the new branch to GitHub
```
In the future, Continuous Integration (CI) and automated testing will be a part of every commit

## Testing

[This blog post](https://blog.alexellis.io/golang-writing-unit-tests/) covers how Go handles unit tests natively. A strategy for integration testing will be devised at a later time. For now, remember that for every method/function you create, a minimum of one unit test will be created to ensure that this function meets the requirements defined for it.

To perform unit tests, run the following command from the projects root directory:
```
$ go test -v ./...
```

### Compiling Binaries

#### Linux
A Makefile has been created to simplify the process. It's important that this project be in the expected location (i.e. $GOPATH/src/github.com/SSSOC-CAN/fmtd). Simply run the following from the root directory to compile binaries
```
$ make install
```

#### Windows
```
# In Gitbash
$ GO111MODULE=on go install -v github.com/SSSOC-CAN/fmtd/cmd/laniakea
$ GO111MODULE=on go install -v github.com/SSSOC-CAN/fmtd/cmd/lanicli
```
To manually test Laniakea and lanicli, have one terminal window open and run
```
$ laniakea
```
And then in another, run `lanicli` followed by a supported command. For example,
```
$ lanicli stop
```

### Adding New Commands and Compiling Protos
In the `lanirpc` directory are the proto files for the RPC server. When new commands are created, the protos must be recompiled:
```
$ protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative lani.proto
```
The REST reverse proxy and OpenAPI file are compiled as follows:
```
$ protoc --grpc-gateway_out=. --grpc-gateway_opt=logtostderr=true --grpc-gateway_opt=paths=source_relative --grpc-gateway_opt=grpc_api_configuration=lani.yaml lani.proto
$ protoc --openapiv2_out=. --openapiv2_opt=logtostderr=true --openapiv2_opt=grpc_api_configuration=lani.yaml --openapiv2_opt=json_names_for_fields=false lani.proto
```

For REST proxy and OpenAPI commands to work, some dependencies are needed and are installed as follows:
```
$ go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
     github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
     google.golang.org/protobuf/cmd/protoc-gen-go \
     google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

Write the proxy command in the `rpcserver.go` file and then define it's permissions in `grpc_intercept.go`. For lanicli access, write a proxy in `commands.go`.

## Release Strategy

*TBD*


# NGINX Reverse Proxy

## gRPC

To setup an NGINX reverse proxy for the gRPC API, create a file in `/etc/nginx/sites-enabled/lani-proxy.conf`:
```nginx
upstream lani_grpc {
    server localhost:7777;
}

server {
    listen 7070 ssl http2;
    ssl_certificate /path/to/lani_tls.cert; # Make sure this is the TLS certificate generated by Laniakea
    ssl_certificate_key /path/to/lani_tls.key; # Make sure this is the TLS certificate key generated by Laniakea
    location / {
        grpc_pass grpcs://lani_grpc;
    }
}
```

## REST

To setup an NGINX reverse proxy for the REST API, create a file in `/etc/nginx/sites-enabled/lani-rest-proxy.conf`:
```nginx
upstream lani_rest {
    server localhost:8080;
}

server {
    listen 8081 ssl;
    ssl_certificate /path/to.lani_tls.cert; # Make sure this is the TLS certificate generated by Laniakea
    ssl_certificate_key /path/to/lani_tls.key; # Make sure this is the TLS certificate key generated by Laniakea
    location / {
        proxy_pass https://lani_rest;
        proxy_buffering off; # To make sure Subscribe responses aren't truncated
    }
}
```
