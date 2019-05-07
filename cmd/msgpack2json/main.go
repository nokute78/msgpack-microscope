/*
   Copyright 2019 Takahiro Yamashita

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"../../msgpack" /* TODO */
	"bytes"
	"flag"
	"fmt"
	"github.com/mattn/go-isatty"
	"io"
	"io/ioutil"
	"os"
)

func decodeAndOutput(in io.Reader, out io.Writer, file string, b_filename *bool) int {
	b, err := ioutil.ReadAll(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ioutil.ReadAll :%v\n", err)
		return 1
	}
	ret, err := msgpack.Decode(bytes.NewBuffer(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return 1
	}
	if *b_filename {
		fmt.Fprintf(out, "%s: ", file)
	}
	outputJSON(ret, out, 0)
	fmt.Fprintf(out, "\n")

	return 0
}

func readStdin(b_filename *bool) int {
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		return decodeAndOutput(os.Stdin, os.Stdout, "(stdin)", b_filename)
	}
	return 0
}

func readFiles(b_filename *bool) {
	if flag.NArg() > 0 {
		for _, v := range flag.Args() {
			file, err := os.Open(v)
			if err != nil {
				fmt.Fprintf(os.Stderr, "os.Open :%v\n", err)
				continue
			}
			defer file.Close()
			if decodeAndOutput(file, os.Stdout, v, b_filename) != 0 {
				continue
			}
		}
	}
}

func outputJSON(obj *msgpack.MPObject, out io.Writer, nest int) {
	switch {
	case msgpack.IsMap(obj.FirstByte):
		fmt.Fprintf(out, "{")
		var i uint32
		for i = 0; i < obj.Length-1; i++ {
			/* key */
			outputJSON(obj.Child[i*2], out, nest+1)
			fmt.Fprintf(out, ":")
			outputJSON(obj.Child[i*2+1], out, nest+1)
			fmt.Fprintf(out, ",")
		}
		outputJSON(obj.Child[(obj.Length-1)*2], out, nest+1)
		fmt.Fprintf(out, ":")
		outputJSON(obj.Child[(obj.Length-1)*2+1], out, nest+1)
		fmt.Fprintf(out, "}")
	case msgpack.IsArray(obj.FirstByte):
		fmt.Fprintf(out, "[")
		var i uint32
		for i = 0; i < obj.Length-1; i++ {
			outputJSON(obj.Child[i], out, nest+1)
			fmt.Fprintf(out, ",")
		}
		outputJSON(obj.Child[obj.Length-1], out, nest+1)
		fmt.Fprintf(out, "]")
	case msgpack.IsString(obj.FirstByte):
		fmt.Fprintf(out, "\"%s\"", obj.DataStr)
	default:
		fmt.Fprintf(out, obj.DataStr)
	}
}

func cmdMain() int {
	b_filename := flag.Bool("f", false, "print file name")
	flag.Parse()

	msgpack.Init()

	/* from STDIN */
	ret := readStdin(b_filename)

	/* from files */
	readFiles(b_filename)

	return ret
}

func main() {
	os.Exit(cmdMain())
}
