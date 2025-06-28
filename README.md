# Speediot CLI

A command-line interface (CLI) typing test application written in Go.

## Features

*   Typing tests with different difficulty levels (Easy, Medium, Hard, Dynamic).
*   Leaderboard to track WPM (Words Per Minute) and Accuracy.
*   Username input for personalized scores.
*   Dynamic text content from a SQLite database, including facts with numbers and special characters.

## Prerequisites

To compile and run this application, you need to have Go installed on your system.

*   **Go:** [Download and install Go](https://golang.org/doc/install) (version 1.16 or higher recommended).

## Compilation

To compile the application for your respective operating system, navigate to the project root directory in your terminal and run the appropriate command:

### Windows

```bash
go build -o speediot_cli_windows_amd64.exe
```

### macOS (Intel)

```bash
GOOS=darwin GOARCH=amd64 go build -o speediot_cli_darwin_amd64
```

### macOS (Apple Silicon / ARM64)

```bash
GOOS=darwin GOARCH=arm64 go build -o speediot_cli_darwin_arm64
```

### Linux

```bash
GOOS=linux GOARCH=amd64 go build -o speediot_cli_linux_amd64
```

## Execution

After successful compilation, you can run the executable from your terminal. Ensure you are in the directory where the executable was created.

### Windows

```bash
.\speediot_cli_windows_amd64.exe
```

### macOS

```bash
./speediot_cli_darwin_amd64
# or for Apple Silicon
./speediot_cli_darwin_arm64
```

### Linux

```bash
./speediot_cli_linux_amd64
```

## Database

The application uses a SQLite database (`texts.db`) to store dynamic text content and leaderboard scores. This file will be created automatically when you run the application for the first time.

## Contributing

Feel free to fork the repository, make improvements, and submit pull requests.
