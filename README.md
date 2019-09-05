# Wrapper command for pull-style borgbackup

## Features

- Compiled binary: File system capability bits (e.g. CAP_DAC_READ_SEARCH) can
  be assigned to it, for root-less operation.
- Access control: Configuration restricts which paths can be backed up by
  different clients (via sshd restrictions, similar to how
  [gitolite](https://gitolite.com/gitolite/index.html) operates).

## Rationale, motivation, and bigger picture

This is a wrapper executable around `borg` (currently only `borg create`). It
restricts the subset of arguments allowed and passes a custom `BORG_RSH`
environment variable.

The use-case is to:

- put this executable as forced command into authorized_keys or sshd_config for
  a remote user which will be used for pulling backups only
- use a reverse-unix-socket through SSH to connect to `borg serve`
- assign CAP_DAC_READ_SEARCH on the binary to grant borg read access to the
  entire filesystem without needing to be uid 0

For more information, please see my (upcoming, sorry) blog post about creating
pull-style backups with borg.

## Usage

```
borg-pull-source CONFIG_FILE CLIENT_NAME
```

* `CONFIG_FILE` must be the path to a config file (see below).
* `CLIENT_NAME` is the client on the behalf of which the wrapper acts.

Note that the command line arguments are expected to be passed by SSHD via the
ForceCommand directive or the `command=` directive from the authorized keys
file, and are thus trusted.

In addition, the contents of the environment variable `SSH_ORIGINAL_COMMAND`
follows the following syntax:

```
[-C COMPRESSION] [-p] [-x] create REPO_PATH ARCHIVE_NAME SOURCE_PATH
```

The flags in square brackets are optional.

* `COMPRESSION` is a valid compression specifier for borg. See the
  `borg-create` help or man page for details.
* `REPO_PATH` is the path to the repository *on the client*. It is passed to
  borg serve on the client.
* `ARCHIVE_NAME` is the name of the archive to create. It is passed to borg
  serve on the client.
* `SOURCE_PATH` is a single path to back up from the server.

`SOURCE_PATH` is checked against a per-client whitelist in the configuration
file.

## Configuration

### Syntax

(see also `docs/config.example.toml`)

```toml
[runtime]
socket_wrapper=<path>
socket_dir=<path>

[client.<name>]
paths=<array of paths>
```

### Description

* `runtime.socket_wrapper` must be the path to an executable like
  `contrib/socket-wrap.sh`. It receives `BORGSOCKETWRAP_SOCKET_PATH` as
  environment variable, is expected to ignore all its arguments and connect the
  path to the unix socket which is given in `BORGSOCKETWRAP_SOCKET_PATH` to
  STDIO.

* `runtime.socket_dir` must be the path to a directory where the unix sockets
  for the clients are stored. The naming scheme for sockets in that directory
  is `client_name + ".sock"`.

* `client.<name>`: A table configuring the client-specific settings. Only
  client names which have a table/section in the config file are allowed.

  * `paths`: Array of paths which the client is allowed to back up. Note that
    subdirectories of paths are not automatically allowed; only the verbatim
    paths given in the array are allowed.

## Example use

On the server:

Prepare an account for the borg backups to be run over. This does not need to
be privileged.

`authorized_keys`:

```
restrict,command="/usr/local/bin/borg-pull-source /etc/borgbackup-pull-source/config.toml my_client",port-forwarding ssh-rsa ...
```

`/etc/borgbackup-pull-source/config.toml`:

```toml
[runtime]
socket_wrapper="/usr/local/bin/borg-pull-socket-wrap.sh"
socket_dir="/var/lib/borg-pull-source/remotes/"
home="/var/lib/borg-pull-source/"

[client.my_client]
paths=[
    "/",
]
```

`/usr/local/bin/borg-pull-socket-wrap.sh` is from `contrib/socket-wrap.sh`.

`/usr/local/bin/borg-pull-source` needs to be the compiled binary from this
repository. Ensure to call
`sudo setcap cap_dac_read_search=ep /usr/local/bin/borg-pull-source` on it
**and restrict the execution permissions to the borg backup user**
(since executing it grants you read access to the entire filesystem).

On the client:

* Start a UNIX socket backed by a borg-serve process:

    ```console
    $ socat UNIX-LISTEN:/tmp/serve.sock "EXEC:/usr/bin/borg serve --append-only --restrict-to-path /mnt/backups/repository/ --umask 077"
    ```

    Socat will listen on `/tmp/serve.sock` and spawn a borg serve process once
    a connection comes in, connecting the stdin/stdout of that process to the
    UNIX socket connection.

* Start the pull backup on the server side:

    ```console
    $ ssh -R "/var/lib/borgbackup-pull-source/remotes/my_client.sock:/tmp/serve.sock" my_server -- -p -x create /mnt/backups/repository/ "test-$(date --iso-8601=seconds)" "/"
    ```
