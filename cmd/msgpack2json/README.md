# msgpack2json

A command line tool to convert [MessagePack](https://msgpack.org/) to JSON.
Read MessagePack from STDIN/File/Http POST.

It is inspired by [msgpack-inspect](https://github.com/tagomoris/msgpack-inspect).

## Quick Start
```shell
$ printf "\x82\xa7compact\xc3\xa6schema\x00"|./msgpack2json
$ printf "\x82\xa7compact\xc3\xa6schema\x00"|./msgpack2json -r
```

## Options
```
Usage of ./msgpack2json:
  -e	enable Fluentd event time ext format
  -f	show data source (e.g. stdin, filename)
  -p uint
    	port number for server mode (default 8080)
  -r	raw JSON mode
  -s	http server mode
  -v	show version
```

### -r: raw JSON mode
To output plane JSON.

```shell
$ printf "\x82\xa7compact\xc3\xa6schema\x00"|./msgpack2json -r
```
```json
{"compact":true,"schema":0}
```

### -f: show data source (e.g. stdin, filename)
Append data source as header.

### -s: http server mode
Waiting Messagepack data from port 8080 with http.

### -p uint: port number for server mode
Change port number which http server uses.

### -e: enable Fluentd event time ext format
If set, msgpack2json can analyze Fluentd Event Time format.

With option
```shell
$ printf "\xd7\x00\x5c\xda\x05\x00\x00\x00\x00\x00"| ./msgpack2json -e
```
```json
{"format":"event time", "header":"0xd7", "type":0, "raw":"0xd7005cda050000000000", "value":"2019-05-14 09:00:00 +0900 JST"}
```

Without option
```shell
$ printf "\xd7\x00\x5c\xda\x05\x00\x00\x00\x00\x00"| ./msgpack2json 
```
```json
{"format":"fixext 8", "header":"0xd7", "type":0, "raw":"0xd7005cda050000000000", "value":"0x5cda050000000000"}
```


## Example (STDIN)

```shell
$ printf "\x82\xa7compact\xc3\xa6schema\x00" | ./msgpack2json
```
```json
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

## Example (File)

```shell
$ printf "\x82\xa7compact\xc3\xa6schema\x00" > b.msgp
$ ./msgpack2json b.msgp 
```
```json
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

## Example (HTTP Server)

```shell
$ printf "\x82\xa7compact\xc3\xa6schema\x00" > b.msgp
$ ./msgpack2json -s &
$ curl -sS localhost:8080 -X POST --data-binary "@b.msgp"
```
```json
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