The benchmarks presented are intended to demonstrate nothing more than that the
design of `kid.New()` stands up to heavy use, maintaining low and predictable
cost under concurrent use on both amd64 and arm64 platforms.

Beyond that, there is a fair amount of apples-to-oranges comparison going on
here. The generators use different mechanisms and make different trade-offs, so
these results should not be interpreted as a general performance ranking. They
are included primarily to provide context for the performance of `kid`.

amd64:

```
❯ # set machine in performance mode
❯ echo "performance" | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
❯ go1.27.1 test -cpu 1,2,4,8,16,32 -test.benchmem -bench .
goos: linux
goarch: amd64
pkg: github.com/mwyvr/kid/v2/eval/bench
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkKid                    39246956                30.24 ns/op            0 B/op          0 allocs/op
BenchmarkKid-2                  40488458                28.21 ns/op            0 B/op          0 allocs/op
BenchmarkKid-4                  36920178                32.40 ns/op            0 B/op          0 allocs/op
BenchmarkKid-8                  36412285                32.19 ns/op            0 B/op          0 allocs/op
BenchmarkKid-16                 34016721                35.04 ns/op            0 B/op          0 allocs/op
BenchmarkKid-32                 52991698                21.57 ns/op            0 B/op          0 allocs/op
BenchmarkXid                    43516267                27.69 ns/op            0 B/op          0 allocs/op
BenchmarkXid-2                  48154291                26.66 ns/op            0 B/op          0 allocs/op
BenchmarkXid-4                  38111800                26.68 ns/op            0 B/op          0 allocs/op
BenchmarkXid-8                  39837633                30.46 ns/op            0 B/op          0 allocs/op
BenchmarkXid-16                 44162358                32.34 ns/op            0 B/op          0 allocs/op
BenchmarkXid-32                 56189682                20.83 ns/op            0 B/op          0 allocs/op
BenchmarkKsuid                  16280052                73.79 ns/op            0 B/op          0 allocs/op
BenchmarkKsuid-2                14535638                81.47 ns/op            0 B/op          0 allocs/op
BenchmarkKsuid-4                12211960                98.39 ns/op            0 B/op          0 allocs/op
BenchmarkKsuid-8                10352878               114.1 ns/op             0 B/op          0 allocs/op
BenchmarkKsuid-16                8125656               144.4 ns/op             0 B/op          0 allocs/op
BenchmarkKsuid-32                6773056               181.6 ns/op             0 B/op          0 allocs/op
BenchmarkUuidV4                 27705572                41.64 ns/op            0 B/op          0 allocs/op
BenchmarkUuidV4-2               45230329                24.05 ns/op            0 B/op          0 allocs/op
BenchmarkUuidV4-4               37373485                31.79 ns/op            0 B/op          0 allocs/op
BenchmarkUuidV4-8               39761662                31.20 ns/op            0 B/op          0 allocs/op
BenchmarkUuidV4-16              40580817                32.80 ns/op            0 B/op          0 allocs/op
BenchmarkUuidV4-32              53659066                21.67 ns/op            0 B/op          0 allocs/op
BenchmarkUuidV7                 19121680                53.92 ns/op            0 B/op          0 allocs/op
BenchmarkUuidV7-2               19397617                58.87 ns/op            0 B/op          0 allocs/op
BenchmarkUuidV7-4               14587982                78.62 ns/op            0 B/op          0 allocs/op
BenchmarkUuidV7-8               12456988                98.07 ns/op            0 B/op          0 allocs/op
BenchmarkUuidV7-16               9941598               122.5 ns/op             0 B/op          0 allocs/op
BenchmarkUuidV7-32               8870276               134.6 ns/op             0 B/op          0 allocs/op
BenchmarkUlid                   20184441                59.22 ns/op           16 B/op          1 allocs/op
BenchmarkUlid-2                 30992666                38.65 ns/op           16 B/op          1 allocs/op
BenchmarkUlid-4                 38232565                31.56 ns/op           16 B/op          1 allocs/op
BenchmarkUlid-8                 40451748                31.68 ns/op           16 B/op          1 allocs/op
BenchmarkUlid-16                44647472                33.07 ns/op           16 B/op          1 allocs/op
BenchmarkUlid-32                47914470                24.42 ns/op           16 B/op          1 allocs/op
BenchmarkSonyflake                 31126             38879 ns/op               0 B/op          0 allocs/op
BenchmarkSonyflake-2               30860             38880 ns/op               0 B/op          0 allocs/op
BenchmarkSonyflake-4               30924             38808 ns/op               0 B/op          0 allocs/op
BenchmarkSonyflake-8               30895             38821 ns/op               0 B/op          0 allocs/op
BenchmarkSonyflake-16              30830             38893 ns/op               0 B/op          0 allocs/op
BenchmarkSonyflake-32              30854             38879 ns/op               0 B/op          0 allocs/op
BenchmarkGoShortUniqueID         5252468               227.3 ns/op            87 B/op          5 allocs/op
BenchmarkGoShortUniqueID-2       6168906               202.3 ns/op            87 B/op          5 allocs/op
BenchmarkGoShortUniqueID-4       6072679               201.5 ns/op            87 B/op          5 allocs/op
BenchmarkGoShortUniqueID-8       5526255               208.1 ns/op            87 B/op          5 allocs/op
BenchmarkGoShortUniqueID-16      4957240               235.1 ns/op            87 B/op          5 allocs/op
BenchmarkGoShortUniqueID-32      4289436               279.8 ns/op            88 B/op          5 allocs/op
PASS
ok      github.com/mwyvr/kid/eval/bench 65.226s
```

Generators utilizing `crypto/rand` show a marked difference when run on MacOS
arm64 platforms rather than Linux amd64 platforms, highlighting the additional
fixed overhead of obtaining those bytes through the MacOS CSPRNG interface.

arm64, on mains power, performance set to "Hi Power":

```
❯ go1.27.1 test -cpu 1,2,4,8,16 -test.benchmem -bench .
goos: darwin
goarch: arm64
pkg: github.com/mwyvr/kid/v2/eval/bench
cpu: Apple M4 Max
BenchmarkKid                    37839336                31.70 ns/op            0 B/op          0 allocs/op
BenchmarkKid-2                  38356630                32.38 ns/op            0 B/op          0 allocs/op
BenchmarkKid-4                  37081099                31.75 ns/op            0 B/op          0 allocs/op
BenchmarkKid-8                  22871845                53.33 ns/op            0 B/op          0 allocs/op
BenchmarkKid-16                 18078981                66.86 ns/op            0 B/op          0 allocs/op
BenchmarkXid                    39083120                30.87 ns/op            0 B/op          0 allocs/op
BenchmarkXid-2                  39633554                29.63 ns/op            0 B/op          0 allocs/op
BenchmarkXid-4                  40535119                28.29 ns/op            0 B/op          0 allocs/op
BenchmarkXid-8                  25732113                47.59 ns/op            0 B/op          0 allocs/op
BenchmarkXid-16                 18500341                64.79 ns/op            0 B/op          0 allocs/op
BenchmarkKsuid                   5715820               211.2 ns/op             0 B/op          0 allocs/op
BenchmarkKsuid-2                 4328422               276.0 ns/op             0 B/op          0 allocs/op
BenchmarkKsuid-4                 3495243               342.7 ns/op             0 B/op          0 allocs/op
BenchmarkKsuid-8                 3501904               343.4 ns/op             0 B/op          0 allocs/op
BenchmarkKsuid-16                3431340               350.1 ns/op             0 B/op          0 allocs/op
BenchmarkUuidV4                  7120689               168.2 ns/op             0 B/op          0 allocs/op
BenchmarkUuidV4-2                4898661               245.4 ns/op             0 B/op          0 allocs/op
BenchmarkUuidV4-4                1739146               683.7 ns/op             0 B/op          0 allocs/op
BenchmarkUuidV4-8                2706901               439.8 ns/op             0 B/op          0 allocs/op
BenchmarkUuidV4-16               2781525               433.8 ns/op             0 B/op          0 allocs/op
BenchmarkUuidV7                 13893277                86.37 ns/op            0 B/op          0 allocs/op
BenchmarkUuidV7-2                8902195               133.6 ns/op             0 B/op          0 allocs/op
BenchmarkUuidV7-4                6094590               196.7 ns/op             0 B/op          0 allocs/op
BenchmarkUuidV7-8                3846578               310.9 ns/op             0 B/op          0 allocs/op
BenchmarkUuidV7-16               5023530               238.5 ns/op             0 B/op          0 allocs/op
BenchmarkUlid                   14247670                84.86 ns/op           16 B/op          1 allocs/op
BenchmarkUlid-2                  8885481               142.0 ns/op            16 B/op          1 allocs/op
BenchmarkUlid-4                  6190344               193.4 ns/op            16 B/op          1 allocs/op
BenchmarkUlid-8                  5116689               232.1 ns/op            16 B/op          1 allocs/op
BenchmarkUlid-16                 5344026               219.7 ns/op            16 B/op          1 allocs/op
BenchmarkSonyflake                 30775             38967 ns/op               0 B/op          0 allocs/op
BenchmarkSonyflake-2               30961             38740 ns/op               0 B/op          0 allocs/op
BenchmarkSonyflake-4               30948             38760 ns/op               0 B/op          0 allocs/op
BenchmarkSonyflake-8               30951             38755 ns/op               0 B/op          0 allocs/op
BenchmarkSonyflake-16              30957             38745 ns/op               1 B/op          0 allocs/op
BenchmarkGoShortUniqueID         5457844               199.6 ns/op            87 B/op          5 allocs/op
BenchmarkGoShortUniqueID-2       9173799               128.6 ns/op            87 B/op          5 allocs/op
BenchmarkGoShortUniqueID-4       8025865               147.2 ns/op            87 B/op          5 allocs/op
BenchmarkGoShortUniqueID-8       6283911               191.5 ns/op            87 B/op          5 allocs/op
BenchmarkGoShortUniqueID-16      4095393               294.8 ns/op            87 B/op          5 allocs/op
PASS
ok      github.com/mwyvr/kid/eval/bench 57.883s
```
