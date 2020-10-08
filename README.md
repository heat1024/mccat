# mccat

terminal base memcached cli client

## How to use

### Run CLI mode

#### show usage

```Shell
# before run mccat

$ ./pkg/mccat_for_mac help[-h]
How to use mccat(memcached cat)
--------------------------------------------------------------------
- when connect to tcp server (default)
   $ mccat [tcp://]URL:PORT (default : localhost:11211)
- when connect to unix socket
   $ mccat [unix://]PATH

  --help [-h]               : show usage
```

#### connect to memcached server

- connect with tcp address

```Shell
$ ./pkg/mccat_for_mac [tcp://example.memcached.com:11211]
connect to tcp://example.memcached.com:11211
tcp://example.memcached.com:11211>
```

- connect with unix socket

```Shell
$ ./pkg/mccat_for_mac [sock://]/var/run/memcached/memcached.sock
connect to memcached server [sock:///var/run/memcached/memcached.sock]
sock:///var/run/memcached/memcached.sock> 
```

#### show command manual

```Shell
# in mccat terminal
localhost:11211> help
Command list
> get key [key2] [key3] ...                                             : Get data from server
> set key ttl                                                           : Set data (overwrite when exist)
> add key ttl                                                           : Add new data (error when key exist)
> append key ttl                                                        : Append data from exist data
> prepend key ttl                                                       : Prepend data from exist data
> replace key ttl                                                       : Replace data from exist data
> incr[increase] key number                                             : Increase numeric value
> decr[decrease] key number                                             : Decrease numeric value
> del[delete|rm|remove] key [key2] [key3] ...                           : Remove key item from server
> key_counts                                                            : Get key counts
> get_all [--name namespace] [--grep grep_words] --verbose              : Get "almost" all items from server (can grep by namespace or key words)
> flush_all                                                             : Get key counts
> help                                                                  : Show usage
```

#### command examples

<details open=true><summary>get, set, del commands</summary>

Default operations.

```Shell
localhost:11211> get test
test : got error! (cache missed)
localhost:11211> set test 3600
input value> Test data
key test set complate
localhost:11211> get test
test: Test data
localhost:11211> del test
key test deleted
localhost:11211> get test
test : got error! (cache missed)
```

</details>

<details open=true><summary>key_counts[keycounts]</summary>

`key_counts[keycounts]` command display whole keys count in memcached server.

```Shell
localhost:11211> key_counts
Key counts: 3
```

</details>

<details open=true><summary>get, del multi</summary>

`get` and `del` commands are support multi key operation.

```Shell
localhost:11211> getall
  - test3
  - test2
  - test1
localhost:11211> get test2 test1 test3
test2 : test2
test1 : test1
test3 : test3
localhost:11211> get test2 test1 test4 test3
test2 : test2
test1 : test1
test4 : got error! (cache missed)
test3 : test3
localhost:11211> del test3 test1 test2
key test3 deleted
key test1 deleted
key test2 deleted
localhost:11211> get test1 test2 test3
test1 : got error! (cache missed)
test2 : got error! (cache missed)
test3 : got error! (cache missed)
```

</details>

<details open=true><summary>get_all[getall] command details</summary>

`get_all[getall]` is get **almost all** keys from server.
(Memcached not support get all keys mechanism because performance problem)

- test data

```Shell
localhost:11211> set test:test1 3600
input value> namespace test
key test:test1 set complate
localhost:11211> set test:2nd 3600
input value> namespace test 2nd
key test:2nd set complate
localhost:11211> set test:3rd
input value> namespace test 3rd
key test:3rd set complate
localhost:11211> getall
  - test:3rd
  - test:2nd
  - test:test1
```

- select namespace

```Shell
localhost:11211> getall -v
  - test:3rd : namespace test 3rd
  - test:2nd : namespace test 2nd
  - test:test1 : namespace test

localhost:11211> getall -n test -v
  - test:3rd : namespace test 3rd
  - test:2nd : namespace test 2nd
  - test:test1 : namespace test
```

- grep word in key

```Shell
localhost:11211> getall -n test -g 2nd -v
  - test:2nd : namespace test 2nd
```

- select namespace not match

```Shell
localhost:11211> getall --vn test -v
```

- select key word not match

```Shell
localhost:11211> getall -vg 2nd -v
  - test:3rd : namespace test 3rd
  - test:test1 : namespace test
```

</details>

<details open=false><summary>set, add, append, prepend, replace commands</summary>

support memcached operations

- set : create data. if key exist, overwrite it

```Shell
localhost:11211> set test
input value> hello
localhost:11211> get test
test : hello
```

- append : append behind exist data (key must exist)

```Shell
localhost:11211> append test
input value> , world!
key test append complate
localhost:11211> get test
test : hello, world!
```

- prepend : prepend before exist data (key must exist)

```Shell
localhost:11211> prepend test
input value> mccat? 
key test prepend complate
localhost:11211> get test
test : mccat? hello, world!
```

- replace : replace exist data (key must exist)

```Shell
localhost:11211> replace test
input value> new data
key test replace complate
localhost:11211> get test
test : new data
localhost:11211> replace not_exist
input value> test
failed to replace: key does not exist
```

- add : add new data (key must not exist)

```Shell
localhost:11211> add test 3600
input value> some data
failed to add: key exist
localhost:11211> add test_add 3600
input value> add command test
key test_add add complate
localhost:11211> get test_add
test_add : add command test
```

</details>

<details open=true><summary>flush_all</summary>

`flush_all` remove all keys in memcached server.

```Shell
localhost:11211> flushall
All keys deleted
localhost:11211> keycounts
Key counts: 0
```

</details>
