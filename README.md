# msgpack-microscope

A library and tool to analyze MessagePack

## Installation
TODO: go get hoge

## Usage
TODO :add code

## Tool
* [msgpack2json](cmd/msgpack2json/README.md)

A tool to analyze MessagePack. It is inspired by [msgpack-inspect](https://github.com/tagomoris/msgpack-inspect).
```
$ printf "\x82\xa7compact\xc3\xa6schema\x00" | ./msgpack2json
{"format":"fixmap", "header":"0x82", "length":2, "raw":"0x82a7636f6d70616374c3a6736368656d6100", "value":
    [
        {"key":
            {"format":"fixstr", "header":"0xa7", "raw":"0xa7636f6d70616374", "value":"compact"},
         "value":
            {"format":"true", "header":"0xc3", "raw":"0xc3", "value":true}
        },
        {"key":
            {"format":"fixstr", "header":"0xa6", "raw":"0xa6736368656d61", "value":"schema"},
         "value":
            {"format":"positive fixint", "header":"0x00", "raw":"0x00", "value":0}
        }
    ]
}
```

## License

[Apache License v2.0](https://www.apache.org/licenses/LICENSE-2.0)
