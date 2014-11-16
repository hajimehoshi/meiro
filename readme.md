# Meiro

Software to generate mazes

## Benchmark

Just a note for me.

```
:; go test -bench . -cpuprofile cpu.out -cpu 4 ./field
:; go tool pprof field.test cpu.out
``` 
