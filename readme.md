# Myrmica Gallienii - Keep Forks Synchronized

[![Build Status](https://travis-ci.org/containous/gallienii.svg?branch=master)](https://travis-ci.org/containous/gallienii)
[![Docker Build Status](https://img.shields.io/docker/build/containous/gallienii.svg)](https://hub.docker.com/r/containous/gallienii/builds/)


## Description

Keep forks synchronized by making PR on forks.


## CLI

```
Myrmica gallienii: Keep forks synchronized.

Usage: gallienii [--flag=flag_argument] [-f[flag_argument]] ...     set flag_argument to flag(s)
   or: gallienii [--flag[=true|false| ]] [-f[true|false| ]] ...     set true/false to boolean flag(s)

Available Commands:
        gen                                                Generate configuration file.
        sync                                               Synchronize forks.
        version                                            Display the version.
Use "gallienii [command] --help" for more information about a command.

Flags:
-h, --help Print Help (this message) and exit
```

### Generate Configuration File

```
Generate configuration file.

Usage: gen [--flag=flag_argument] [-f[flag_argument]] ...     set flag_argument to flag(s)
   or: gen [--flag[=true|false| ]] [-f[true|false| ]] ...     set true/false to boolean flag(s)

Flags:
    --org    Generate a default configuration file for an organization name. 
    --sample Generate a sample configuration file.                           (default "true")
-t, --token  GitHub Token.                                                   
    --user   Generate a default configuration file for a user name.          
-h, --help   Print Help (this message) and exit
```

### Synchronize Forks

```
Synchronize forks.

Usage: sync [--flag=flag_argument] [-f[flag_argument]] ...     set flag_argument to flag(s)
   or: sync [--flag[=true|false| ]] [-f[true|false| ]] ...     set true/false to boolean flag(s)

Flags:
    --dry-run    Dry run mode.                      (default "true")
    --port       Server port.                       (default "80")
    --rules-path Path to the rules file.            (default "./gallienii.toml")
    --server     Server mode.                       (default "false")
-t, --token      GitHub Token.                      
    --verbose    Verbose mode.                      (default "false")
-h, --help       Print Help (this message) and exit
```