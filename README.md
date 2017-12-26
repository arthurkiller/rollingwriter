# RollingWriter
Rolling file IO writer is write in go. Easy to ID nonlock rollig writer

## Features
* Auto rotate
* Parallel safe
* Implement go io.Writer
* Time rotate with corn style task schedual
* Volume rotate

## Quick Start

## What's in it
it contains 2 separate patrs:
* Manager: decide when to rotate the file with policy
* IOWriter: impement the io.Writer and do the io write

## Contribute && TODO
Now I am about to release the v1.0.0-prerelease with redesigned interface

Any new feature inneed pls [put up an issue](https://github.com/arthurkiller/rollingWriter/issues/new)
