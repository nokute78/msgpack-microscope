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

import(
	"os"
    "io"
    "bytes"
    "fmt"
    "io/ioutil"
    "github.com/mattn/go-isatty"
    "../../msgpack" /* TODO */
)

func outputJSON(obj *msgpack.MPObject, out io.Writer, nest int) {
	switch {
	case msgpack.IsMap(obj.FirstByte):
		fmt.Fprintf(out, "{")
		var i uint32
		for i = 0; i < obj.Length-1; i++{
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
		for i =0 ; i< obj.Length-1; i++ {
			outputJSON(obj.Child[i], out, nest+1)
			fmt.Fprintf(out, ",")
		}
		outputJSON(obj.Child[obj.Length-1], out, nest+1)
		fmt.Fprintf(out, "]")
	case msgpack.IsString(obj.FirstByte):
		fmt.Fprintf(out,"\"%s\"", obj.DataStr)
	default:
		fmt.Fprintf(out, obj.DataStr)
	}
}

func cmdMain() int {
	var b []byte

	if !isatty.IsTerminal(os.Stdin.Fd()) {
		b, _ = ioutil.ReadAll(os.Stdin)
	}
	msgpack.Init()

	ret, err := msgpack.Decode(bytes.NewBuffer(b))
	outputJSON(ret, os.Stdout, 0)

	if err != nil {
		return 1
	}
	return 0
}

func main() {
	os.Exit(cmdMain())
}