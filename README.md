# sensiword

Sensitive words filtering component based on Aho-Corasick algorithm

## Usage

Use Sensitive after Init words.

```go
    sensi := NewSensitive(true) // true means case-insensitive
	sensi.Init("versionString", sensiWordTestDataS...)
    // use it after Init
	sensi.FindAll("hello, gcd, GDB,nanaive哈哈哈，敏感词, caicai我是谁")
	sensi.Replace("hello, gcd, GDB,nanaive哈哈哈，敏感词, caicai我是谁", []rune("*"[0]))
```

If you want rebuild, you can.

```go
    sensi := NewSensitive(true) // true means case-insensitive
    sensi.Init("versionString", sensiWordTestDataS...)
    // rebuild
    sensi.Rebuild("versionString1", words...)
```


If you want part rebuild that only add/delete some word.

```go
    sensi := NewSensitive(true) // true means case-insensitive
    sensi.Init("versionString", sensiWordTestDataS...)
    // part rebuild
    sensi.PartRebuild("versionString1", addWords,  delWords)
```

## Test Result

### 1. Load about 200,000 words.
      
```text
=== RUN   TestLoadSensitive
load 204291 words 10 times, total cost time : 5.051958401s, avg time : 505.19584ms
--- PASS: TestLoadSensitive (5.06s)
```

### 2. In 200,000 words case, rebuild
      


```text
=== RUN   TestRebuildSensitive
endLoad  529.436739ms
rebuild 204291 words 10 times, total cost time : 4.969092815s, avg time : 496.909281ms
--- PASS: TestRebuildSensitive (5.51s)
PASS
 
=== Benchmark  BenchmarkSensitive_Rebuild-8
BenchmarkSensitive_Rebuild-8                   1        1014956599 ns/op        171448040 B/op   2080883 allocs/op
```

### 3. In 200,000 words case, part rebuild
      


```text
=== RUN   TestPartRebuildSensitive
endLoad  550.175636ms
part rebuild 204291 words 10 times, total cost time : 4.213755152s, avg time : 421.375515ms
--- PASS: TestPartRebuildSensitive (4.78s)
PASS
 
=== Benchmark BenchmarkSensitive_PartRebuild-8
BenchmarkSensitive_PartRebuild-8               2         566693237 ns/op        69287340 B/op     782536 allocs/op
```

### 4. In 200,000 words case, find all sensitive word.
```text
=== Benchmark BenchmarkSensitive_FindAll
BenchmarkSensitive_FindAll-8               90987             11369 ns/op            1844 B/op         29 allocs/op
```



### 5. In 200,000 words case, replace the sensitive word to *.
```text
=== Benchmark BenchmarkSensitive_ReplaceAll
BenchmarkSensitive_ReplaceAll-8            91120             11885 ns/op            1290 B/op         13 allocs/op
```

