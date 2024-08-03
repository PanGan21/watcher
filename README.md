# watcher

Watcher is a Go application that watches for changes in files based on your specified patterns and restarts a given command when a modification occurs.

## Features

- Watches directories and files for modifications.
- Triggers command restart based on user-defined patterns.
- Supports delays between restarts to avoid unnecessary executions.
- Offers options for ignoring specific file changes.
- Provides verbose output for debugging purposes.

## Instalation

This project requires Go to be installed on your system. You can download and install Go from the official website: https://go.dev/doc/install

Once Go is installed, navigate to the project directory and run the following command to build the executable:

```bash
go build -o watcher cmd/main.go
```

This will create an executable named `watcher` in your current directory.

## Usage

Run the COMMAND and restart when a file matching the pattern has been modified.

```bash
Usage:
  watcher [COMMAND] [flags]

Flags:
  -d, --delay duration   duration to delay the restart of the command (default 1s)
  -f, --filter event     filter file system event (CREATE|WRITE|REMOVE|RENAME|CHMOD)
  -h, --help             help for watcher
  -i, --ignore glob      ignore pathname glob pattern
  -p, --pattern glob     trigger pathname glob pattern (default "**") (default [**])
  -r, --restart          restart the command on exit
  -s, --signal signal    signal used to stop the command (default "SIGTERM")
  -t, --target path      observation target path (default "./") (default [./])
  -v, --verbose          verbose output
```

### Example

This command watches all Go files (\*\*.go) in the current directory and its subdirectories and restarts the command go run ./bin/main.go whenever a Go file is modified. It waits for 5 seconds before restarting the command.

`go run cmd/main.go -p '**/*.go' -t "./" -- go run ./bin/main.go -d 5s`

### Filtering Events

You can use the -f or --filter option to specify which file system events you want to watch for. This allows you to fine-tune the application's behavior. Here are the valid options for the filter:

- `CREATE`: Watch for new files being created.
- `WRITE`: Watch for existing files being modified.
- `REMOVE`: Watch for files being deleted.
- `RENAME`: Watch for files being renamed.
- `CHMOD`: Watch for file permission changes.

You can combine multiple filter options separated by commas. For example, to watch for file creation and deletion events only, you would use:

`watcher [COMMAND] -f CREATE,REMOVE`

### Verbose Mode

The `-v` or `--verbose` flag enables verbose output during runtime. This provides detailed information about watched targets, detected changes, and command execution.

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
