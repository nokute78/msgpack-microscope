## msgpack2json

A command line tool to convert MessagePack to JSON.

### Option
```
Usage of ./msgpack2json:
  -f	print file name
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