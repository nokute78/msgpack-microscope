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
	"time"

	"github.com/mattn/go-isatty"
	"github.com/nokute78/msgpack2txt/pkg/msgpack"
)

type config struct {
	showSource bool
	serverMode bool
	eventTime  bool
	serverPort uint
}

func decodeAndOutput(in io.Reader, out io.Writer, file string, cnf *config) int {
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
	if cnf.showSource {
		fmt.Fprintf(out, "%s: ", file)
	}
	outputJSON(ret, out, 0)
	fmt.Fprintf(out, "\n")

	return 0
}

func readHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		decodeAndOutput(req.Body, os.Stdout, time.Now().Format(time.UnixDate), &config{showSource:true})
	}
}

func readHttp(cnf *config) int {
	http.HandleFunc("/", readHTTP)

	fmt.Fprintf(os.Stderr, "%s", http.ListenAndServe(fmt.Sprintf(":%d", cnf.serverPort), nil))
	return 0
}

func readStdin(cnf *config) int {
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		return decodeAndOutput(os.Stdin, os.Stdout, "(stdin)", cnf)
	}
	return 0
}

func readFiles(cnf *config) {
	if flag.NArg() > 0 {
		for _, v := range flag.Args() {
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
	ret := 1

	config := config{}

	flag.BoolVar(&config.showSource, "f", false, "print data source")
	flag.BoolVar(&config.serverMode, "s", false, "http server mode")
	flag.BoolVar(&config.eventTime,"e", false, "support fluentd event time ext format")
	flag.UintVar(&config.serverPort,"p", 8080, "port number for server mode")

	flag.Parse()

	msgpack.Init()

	if config.eventTime {
		msgpack.RegisterFluentdEventTime()
	}

	if config.serverMode {
		ret = readHttp(&config)
	} else {

		/* from STDIN */
		ret = readStdin(&config)

		/* from files */
		readFiles(&config)
	}

	return ret
}

func main() {
	os.Exit(cmdMain())
}
