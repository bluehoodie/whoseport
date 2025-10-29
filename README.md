# whoseport

A tiny utility to determine which process is currently using a given port.

### Prerequisites

This will run on Unix machines (anywhere that `lsof` exists)

### Installing

#### Using Go Install (Recommended)

If you have Go installed:

```bash
go install github.com/bluehoodie/whoseport@latest
```

This will install the binary to your `$GOBIN` directory (or `$GOPATH/bin`, or `$HOME/go/bin` by default).

#### From Source

Clone the repository and install:

```bash
git clone https://github.com/bluehoodie/whoseport.git
cd whoseport
make install
```

Or just build locally for testing:

```bash
make build
./whoseport 8080
```

### Usage

Basic usage:

```bash
$ whoseport 8080
```

Will output an interactive, colorized display:

```
╔════════════════════════════════════════════════════════════════╗
║          Process Information for Port 8080                     ║
╚════════════════════════════════════════════════════════════════╝

  Command:     node
  PID:         325
  User:        colin
  Type:        IPv6
  Node:        TCP
  Name:        *:http-alt (LISTEN)
```

#### Options

- `-k, --kill` - Kill the process using the port
- `-i, --interactive` - Prompt before killing the process
- `--json` - Output in JSON format (original format)

#### Examples

**Find process on a port:**
```bash
$ whoseport 8080
```

**Kill process immediately:**
```bash
$ whoseport --kill 8080
# or
$ whoseport -k 8080
```

**Interactive mode (prompt before killing):**
```bash
$ whoseport --interactive 8080
# or
$ whoseport -i 8080
```

**JSON output:**
```bash
$ whoseport --json 8080
```

Output:
```json
{
	"command": "foo",
	"id": 325,
	"user": "colin",
	"fd": "7u",
	"type": "IPv6",
	"device": "0x5e4b104643390241",
	"size_offset": "0t0",
	"node": "TCP",
	"name": "*:http-alt (LISTEN)"
}
```

## License

See the [LICENSE](LICENSE) file for details

## Acknowledgements

This project is certainly not the first of its kind.  Many before have made utilities such as this, or have created aliases which give roughly the same output.  
This was borne of a simple idea to write a version in Go with a pretty output.  
