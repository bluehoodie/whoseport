package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Printf("\t%s {port (int)}\n", os.Args[0])
		fmt.Printf("\tex:\twhoseport 8080\n")
	}
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Printf("error: missing port number\n")
		flag.Usage()
		os.Exit(1)
	}

	port, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		fmt.Printf("error: port must be an integer\n")
		flag.Usage()
		os.Exit(1)
	}

	lsof(port)
}

func lsof(port int) {
	c1 := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port))
	c2 := exec.Command("grep", "LISTEN")

	r, w := io.Pipe()
	defer r.Close()

	c1.Stdout = w
	c2.Stdin = r

	var b2 bytes.Buffer
	c2.Stdout = &b2

	c1.Start()
	c2.Start()

	c1.Wait()
	w.Close()
	c2.Wait()

	d := &data{&b2}
	j, err := json.MarshalIndent(d, "", "\t")
	if err != nil {
		fmt.Fprintf(os.Stderr, "no service found on port %d", port)
	}

	fmt.Fprintf(os.Stdout, "%s\n", j)
}

type data struct {
	values *bytes.Buffer
}

func (i *data) MarshalJSON() ([]byte, error) {
	var values []string

	spl := bytes.Split(i.values.Bytes(), []byte(" "))
	for i, v := range spl {
		if len(v) <= 0 {
			continue
		}

		if len(values) == 8 {
			values = append(values, strings.TrimSpace(string(bytes.Join(spl[i:], []byte(" ")))))
			break
		}

		values = append(values, strings.TrimSpace(string(v)))
	}

	if len(values) != 9 {
		return nil, fmt.Errorf("incorrect number of values returned")
	}

	pid, err := strconv.Atoi(values[1])
	if err != nil {
		return nil, fmt.Errorf("could not convert process id to int: %v", err)
	}

	p := struct {
		Command    string `json:"command"`
		ID         int    `json:"id"`
		User       string `json:"user"`
		FD         string `json:"fd"`
		Type       string `json:"type"`
		Device     string `json:"device"`
		SizeOffset string `json:"size_offset"`
		Node       string `json:"node"`
		Name       string `json:"name"`
	}{
		Command:    values[0],
		ID:         pid,
		User:       values[2],
		FD:         values[3],
		Type:       values[4],
		Device:     values[5],
		SizeOffset: values[6],
		Node:       values[7],
		Name:       values[8],
	}
	return json.Marshal(p)
}
