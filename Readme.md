# Nstrm

##Example
fizzbuzz
```
seq(100) | {x ->
  if x % 15 == 0 {
    "FizzBuzz"
  }
  else if x % 3 == 0 {
    "Fizz"
  }
  else if x % 5 == 0 {
    "Buzz"
  }
  else {
    x
  }
} | STDOUT
```
chat
```
broadcast = chan()
tcp_server(8008) | { s ->
  broadcast | s
  s | broadcast
} | null
```
There are more examples in _examples directory.
#Dependencies

https://github.com/pointlander/peg

#Compile
```
$ make
```
