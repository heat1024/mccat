# mccat
terminal base memcached cli client by golang

## How to use

### Run CLI mode

#### run without build

```Shell
$ go run main.go
connect to localhost:11211
localhost:11211> 
```

#### run after build

```Shell
$ go build -o {name}
$ ./{name} [example.memcached.com:11211] (default localhost:11211)
connect to example.memcached.com:11211
example.memcached.com:11211> 
```

#### show usage

```Shell
# before run mccat

$ ./mccat help
How to use mccat(memcached cat)
mccat [URL:PORT] (default server : localhost:11211)

  --help [-h]               : show usage
```

#### show command manual

```Shell
# in mccat terminal

localhost:11211> help 
Command list
> get key [key2] [key3] ...                                             : get data from server
> set key ttl                                                           : set data (overwrite when exist)
> add key ttl                                                           : add new data (error when key exist)
> append key ttl                                                        : append data from exist data
> prepend key ttl                                                       : prepend data from exist data
> incr[increase] key number                                             : increase numeric value
> decr[decrease] key number                                             : decrease numeric value
> del[delete|rm|remove] key                                             : remove key item from server
> getall[allitems] [--name namespace] [--grep grep_words] --verbose     : get all items from server (can grep by namespace or key words)
> help                                                                  : show usage
```

#### command examples

<details>

<summary>get, set, del commands</summary>

```Shell
localhost:11211> get test 
no values
localhost:11211> set test 3600 
Test data
key test set complate
localhost:11211> get test 
test: Test data
localhost:11211> del test 
key test deleted
localhost:11211> get test 
no values
```

</details>

<details>

<summary>getall command</summary>

- test data
```Shell
localhost:11211> getall 
Key counts: 0

localhost:11211> set test:test1 3600 
namespace test
key test:test1 set complate
localhost:11211> set test:2nd 3600   
namespace test 2nd
key test:2nd set complate
localhost:11211> set test:3rd 3600 
namespace test 3rd
key test:3rd set complate
localhost:11211> get test:test1 
test:test1: namespace test
localhost:11211> get test:2nd   
test:2nd: namespace test 2nd
localhost:11211> get test:3rd 
test:3rd: namespace test 3rd
```

- select namespace

```Shell
localhost:11211> getall -v 
Key counts: 3
  - test:3rd : namespace test 3rd
  - test:2nd : namespace test 2nd
  - test:test1 : namespace test

localhost:11211> getall --name test -v 
Key counts: 3
  - test:2nd : namespace test 2nd
  - test:test1 : namespace test
  - test:3rd : namespace test 3rd
```

- grep word in key

```Shell
localhost:11211> getall --name test --grep 2nd -v 
Key counts: 1
  - test:2nd : namespace test 2nd
```

- select namespace not match

```Shell
localhost:11211> getall --vname test -v 
Key counts: 0
```

- select key word not match

```Shell
localhost:11211> getall --vgrep 2nd -v 
Key counts: 2
  - test:3rd : namespace test 3rd
  - test:test1 : namespace test
```

</details>
