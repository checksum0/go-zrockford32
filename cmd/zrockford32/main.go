package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/checksum0/go-zrockford32"
)

const generalError = 1

func main() {
	inputFlag := flag.String("input", "", "Input file to read from, defaults to stdin")
	outputFlag := flag.String("output", "", "Output file to write to, defaults to stdout")
	decodeFlag := flag.Bool("decode", false, "Decode input instead of encoding")
	lowercaseFlag := flag.Bool("lowercase", false, "Use lowercase encoding instead of uppercase")

	flag.Parse()

	input, err := getInput(*inputFlag)
	defer input.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open %s for input.\n", *inputFlag)
		os.Exit(generalError)
	}

	output, err := getOutput(*outputFlag)
	defer output.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open %s for output.\n", *outputFlag)
		os.Exit(generalError)
	}

	if *decodeFlag {
		if *lowercaseFlag {
			err = decodeLwr(input, output)
		} else {
			err = decodeStd(input, output)
		}
	} else {
		if *lowercaseFlag {
			err = encodeLwr(input, output)
		} else {
			err = encodeStd(input, output)
		}
	}
	if err != nil {
		os.Exit(generalError)
	}

	return
}

func getInput(path string) (*os.File, error) {
	if len(path) == 0 || path == "-" {
		return os.Stdin, nil
	}

	return os.Open(path)
}

func getOutput(path string) (*os.File, error) {
	if len(path) == 0 || path == "-" {
		return os.Stdout, nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Create(path)
	}

	return os.Open(path)
}

func decodeStd(input io.Reader, output io.Writer) error {
	stream := zrockford32.NewDecoder(zrockford32.StdEncoding, input)
	_, err := io.Copy(output, stream)

	return err
}

func decodeLwr(input io.Reader, output io.Writer) error {
	stream := zrockford32.NewDecoder(zrockford32.LwrEncoding, input)
	_, err := io.Copy(output, stream)

	return err
}

func encodeStd(input io.Reader, output io.Writer) error {
	stream := zrockford32.NewEncoder(zrockford32.StdEncoding, output)
	defer stream.Close()
	_, err := io.Copy(stream, input)

	return err
}

func encodeLwr(input io.Reader, output io.Writer) error {
	stream := zrockford32.NewEncoder(zrockford32.LwrEncoding, output)
	defer stream.Close()
	_, err := io.Copy(stream, input)

	return err
}
