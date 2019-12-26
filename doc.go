/*
kaspad is a full-node kaspa implementation written in Go.

The default options are sane for most users. This means kaspad will work 'out of
the box' for most users. However, there are also a wide variety of flags that
can be used to control it.

Usage:
  kaspad [OPTIONS]

For an up-to-date help message:
  kaspad --help

An interesting point to note is that the long form of all option flags
(except -C) can be specified in a configuration file that is automatically
parsed when kaspad starts up. By default, the configuration file is located at
~/.kaspad/kaspad.conf on POSIX-style operating systems and %LOCALAPPDATA%\kaspad\kaspad.conf
on Windows. The -C (--configfile) flag can be used to override this location.
*/
package main
