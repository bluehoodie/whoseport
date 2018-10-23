# whoseport

A tiny utility to determine which process is currently using a given port.

### Prerequisites

This will run on Unix machines (anywhere that `lsof` exists)

### Installing

If you have go, you can run

```go get -u github.com/bluehoodie/whoseport```

Or you can download and compile from source by downloading and using

```make install```

If you do not have go, downloads will be available shortly.

### Usage

Example:

```$ whoseport 8080```

Will output something like 

```
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

See the [LICENSE.md](LICENSE.md) file for details

## Acknowledgements

This project is certainly not the first of its kind.  Many before have made utilities such as this, or have created aliases which give roughly the same output.  
This was borne of a simple idea to write a version in Go with a pretty output.  
