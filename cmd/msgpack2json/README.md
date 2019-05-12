## msgpack2json

A command line tool to convert MessagePack to JSON.

It is inspired by [msgpack-inspect](https://github.com/tagomoris/msgpack-inspect).

### Option
```
Usage of ./msgpack2json:
  -V	show version
  -e	enable Fluentd event time ext format
  -f	show data source (e.g. stdin, filename)
  -p uint
    	port number for server mode (default 8080)
  -s	http server mode
  -v	verbose mode
```

### Example (STDIN)

```
$ printf "\x82\xa10\x01\xa11\x81\xa1a\x02"|./msgpack2json
{"0":1,"1":{"a":2}}
```

```
$ printf "\x82\xa7compact\xc3\xa6schema\x00"|./msgpack2json
{"compact":true,"schema":0}
```

### Example (File)

```
$ printf "\x82\xa10\x01\xa11\x81\xa1a\x02" > a.msgp
$ printf "\x82\xa7compact\xc3\xa6schema\x00" > b.msgp

$ ./msgpack2json a.msgp b.msgp 
{"0":1,"1":{"a":2}}
{"compact":true,"schema":0}
```

### Example (HTTP Server)

```
$ printf "\x82\xa10\x01\xa11\x81\xa1a\x02" > a.msgp

$ ./msgpack2json -s &
$ curl -sS localhost:8080 -X POST --data-binary "@a.msgp"
{"0":1,"1":{"a":2}}
```