package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Print("missing port number")
		os.Exit(1)
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Print("port must be an integer")
		os.Exit(1)
	}

	lsof(port)
}

func lsof(port int) {
	c1 := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port))
	c2 := exec.Command("grep", "LISTEN")

	r, w := io.Pipe()
	c1.Stdout = w
	c2.Stdin = r

	var b2 bytes.Buffer
	c2.Stdout = &b2

	c1.Start()
	c2.Start()
	c1.Wait()
	w.Close()
	c2.Wait()

	io.Copy(os.Stdout, &b2)
	r.Close()
}
