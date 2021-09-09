# fmtd
Repository for the Facility Management Tool Daemon (fmtd)

# Installation

fmtd runs on Go 1.17

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
If you type `$ go version` you should see
```
go version go1.17 linux/amd64
```
## Installing fmtd from source

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

For effective work with teams, if a work on a go file is merged into the `main` branch, but future work remains to be completed, please add `TODO:<name> -` comments in areas where future work will be performed. For example:
```
func sum(a, b int) int {
    // TODO:SSSOCPaulCote - Add error handling for invalid types
    return a+b
}
```
In general, any useful comments to understand the context of your code are encouraged.

## Git Workflow





