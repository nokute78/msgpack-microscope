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
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/nokute78/msgpack-microscope/pkg/msgpack"
)

const version string = "0.0.1"

type config struct {
	showSource bool
	serverMode bool
	eventTime  bool
	serverPort uint
	rawmode    bool
}

type serverHandler struct {
	cnf *config
}

func decodeAndOutput(in io.Reader, out io.Writer, file string, cnf *config) int {
	b, err := ioutil.ReadAll(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ioutil.ReadAll :%v\n", err)
		return 1
	}

	buf := bytes.NewBuffer(b)
	for buf.Len() > 0 {
		ret, err := msgpack.Decode(buf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error(%s) detected. Incoming data may be broken.\n", err)
			if ret == nil {
				return 1
			}
			/* ret is broken, but try to output as much as possible. */
		}
		if cnf.showSource {
			fmt.Fprintf(out, "%s: ", file)
		}
		if cnf.rawmode {
			outputJSON(ret, out, 0)
		} else {
			outputVerboseJSON(ret, out, 0)
		}
		fmt.Fprintf(out, "\n")
	}

	return 0
}

func (h *serverHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		decodeAndOutput(req.Body, os.Stdout, time.Now().Format(time.UnixDate), h.cnf)
	}
}

func readHTTP(cnf *config) int {
	handler := &serverHandler{cnf: cnf}
	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", cnf.serverPort),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	fmt.Fprintf(os.Stderr, "readHTTP error:%s", s.ListenAndServe())

	return 0
}

func readStdin(cnf *config) int {
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		return decodeAndOutput(os.Stdin, os.Stdout, "(stdin)", cnf)
	}
	return 0
}

func readFiles(files []string, cnf *config) {
	if len(files) > 0 {
		for _, v := range files {
			file, err := os.Open(v)
			if err != nil {
				fmt.Fprintf(os.Stderr, "os.Open :%v\n", err)
				continue
			}
			defer file.Close()
			if decodeAndOutput(file, os.Stdout, v, cnf) != 0 {
				continue
			}
		}
	}
}

func outputVerboseKV(obj *msgpack.MPObject, i uint32, out io.Writer, nest int) {
	spaces := strings.Repeat("    ", nest)

	fmt.Fprintf(out, "%s{\"key\":\n", spaces)
	outputVerboseJSON(obj.Child[i*2], out, nest+1)
	fmt.Fprint(out, ",\n")
	fmt.Fprintf(out, "%s \"value\":\n", spaces)
	outputVerboseJSON(obj.Child[i*2+1], out, nest+1)
	fmt.Fprintf(out, "\n%s}", spaces)
}

func outputVerboseJSON(obj *msgpack.MPObject, out io.Writer, nest int) {
	if obj == nil {
		return
	}
	spaces := strings.Repeat("    ", nest)

	switch {
	case msgpack.IsArray(obj.FirstByte):
		spaces2 := strings.Repeat("    ", nest+1)

		// array header info
		fmt.Fprintf(out, `%s{"format":"%s", "header":"0x%02x", "length":%d, "raw":"0x%0x", "value":`, spaces, obj.FormatName, obj.FirstByte, obj.Length, obj.Raw)

		if int(obj.Length) != len(obj.Child) {
			fmt.Fprintf(os.Stderr, "Error: size mismatch. length is %d, buf %d children.\n", obj.Length, len(obj.Child))
			return
		}

		// array body info
		fmt.Fprintf(out, "\n%s[\n", spaces2)
		if obj.Length > 0 {
			var i uint32
			for i = 0; i < obj.Length-1; i++ {
				outputVerboseJSON(obj.Child[i], out, nest+2)
				fmt.Fprintf(out, ",\n")
			}
			outputVerboseJSON(obj.Child[obj.Length-1], out, nest+2)
		}
		fmt.Fprintf(out, "\n%s]\n%s}\n", spaces2, spaces)
	case msgpack.IsMap(obj.FirstByte):
		spaces2 := strings.Repeat("    ", nest+1)
		// map header info
		fmt.Fprintf(out, `%s{"format":"%s", "header":"0x%02x", "length":%d, "raw":"0x%0x", "value":`, spaces, obj.FormatName, obj.FirstByte, obj.Length, obj.Raw)

		if int(obj.Length*2) != len(obj.Child) {
			fmt.Fprintf(os.Stderr, "Error: size mismatch. length is %d, buf %d(!=length*2) children.\n", obj.Length, len(obj.Child))
			return
		}

		// map body info
		fmt.Fprintf(out, "\n%s[\n", spaces2)
		var i uint32
		if obj.Length > 0 {
			for i = 0; i < obj.Length-1; i++ {
				outputVerboseKV(obj, i, out, nest+2)
				fmt.Fprint(out, ",\n")
			}
			outputVerboseKV(obj, obj.Length-1, out, nest+2)
		}
		fmt.Fprintf(out, "\n%s]\n%s}", spaces2, spaces)

	case msgpack.IsString(obj.FirstByte) || msgpack.IsBin(obj.FirstByte):
		fmt.Fprintf(out, `%s{"format":"%s", "header":"0x%02x", "raw":"0x%0x", "value":"%s"}`, spaces, obj.FormatName, obj.FirstByte, obj.Raw, obj.DataStr)
	case msgpack.IsExt(obj.FirstByte):
		fmt.Fprintf(out, `%s{"format":"%s", "header":"0x%02x", "type":%d, "raw":"0x%0x", "value":"%s"}`, spaces, obj.FormatName, obj.FirstByte, obj.ExtType, obj.Raw, obj.DataStr)
	case msgpack.NilFormat == obj.FirstByte:
		fmt.Fprintf(out, `%s{"format":"%s", "header":"0x%02x", "raw":"0x%0x", "value":null}`, spaces, obj.FormatName, obj.FirstByte, obj.Raw)
	case msgpack.NeverUsedFormat == obj.FirstByte:
		fmt.Fprintf(out, `%s{"format":"%s", "header":"0x%02x", "raw":"0x%0x", "value":%s}`, spaces, obj.FormatName, obj.FirstByte, obj.Raw, obj.DataStr)
		fmt.Fprintf(os.Stderr, "Error: Never Used Format detected\n")
		return
	default:
		fmt.Fprintf(out, `%s{"format":"%s", "header":"0x%02x", "raw":"0x%0x", "value":%s}`, spaces, obj.FormatName, obj.FirstByte, obj.Raw, obj.DataStr)
	}
}

func outputKV(obj *msgpack.MPObject, i uint32, out io.Writer, nest int) {
	outputJSON(obj.Child[i*2], out, nest)
	fmt.Fprint(out, ":")
	outputJSON(obj.Child[i*2+1], out, nest)
}

func outputJSON(obj *msgpack.MPObject, out io.Writer, nest int) {
	switch {
	case msgpack.IsMap(obj.FirstByte):
		if int(obj.Length*2) != len(obj.Child) {
			fmt.Fprintf(os.Stderr, "Error: size mismatch. length is %d, buf %d(!=length*2) children.\n", obj.Length, len(obj.Child))
			return
		}
		fmt.Fprint(out, "{")
		if obj.Length > 0 {
			var i uint32
			for i = 0; i < obj.Length-1; i++ {
				outputKV(obj, i, out, nest+1)
				fmt.Fprint(out, ",")
			}
			outputKV(obj, obj.Length-1, out, nest+1)
		}
		fmt.Fprint(out, "}")
	case msgpack.IsArray(obj.FirstByte):
		if int(obj.Length) != len(obj.Child) {
			fmt.Fprintf(os.Stderr, "Error: size mismatch. length is %d, buf %d children.\n", obj.Length, len(obj.Child))
			return
		}
		fmt.Fprint(out, "[")
		if obj.Length > 0 {
			var i uint32
			for i = 0; i < obj.Length-1; i++ {
				outputJSON(obj.Child[i], out, nest+1)
				fmt.Fprint(out, ",")
			}
			outputJSON(obj.Child[obj.Length-1], out, nest+1)
		}
		fmt.Fprint(out, "]")
	case msgpack.IsString(obj.FirstByte) || msgpack.IsBin(obj.FirstByte):
		fmt.Fprintf(out, "\"%s\"", obj.DataStr)
	case msgpack.NilFormat == obj.FirstByte:
		fmt.Fprintf(out, "null")
	case msgpack.NeverUsedFormat == obj.FirstByte:
		fmt.Fprintf(os.Stderr, "Error: Never Used Format detected\n")
		return
	default:
		fmt.Fprint(out, obj.DataStr)
	}
}

func cmdMain() int {
	ret := 1
	showVersion := false

	config := config{}

	flag.BoolVar(&config.showSource, "f", false, "show data source (e.g. stdin, filename)")
	flag.BoolVar(&config.serverMode, "s", false, "http server mode")
	flag.BoolVar(&config.rawmode, "r", false, "raw JSON mode")
	flag.BoolVar(&config.eventTime, "e", false, "enable Fluentd event time ext format")
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.UintVar(&config.serverPort, "p", 8080, "port number for server mode")

	flag.Parse()

	if showVersion {
		fmt.Printf("Ver: %s\n", version)
		return 0
	}

	if config.eventTime {
		msgpack.RegisterFluentdEventTime()
	}

	if config.serverMode {
		ret = readHTTP(&config)
	} else {

		/* from STDIN */
		ret = readStdin(&config)

		/* from files */
		readFiles(flag.Args(), &config)
	}

	return ret
}

func main() {
	os.Exit(cmdMain())
}
