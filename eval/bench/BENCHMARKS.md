The benchmarks presented are intended to demonstrate nothing more than that the
design of `kid.New()` stands up to heavy use, maintaining low and predictable
cost under concurrent use on both amd64 and arm64 platforms.

Beyond that, there is a fair amount of apples-to-oranges comparison going on
here. The generators use different mechanisms and make different trade-offs, so
these results should not be interpreted as a general performance ranking. They
are included primarily to provide context for the performance of `kid`.

amd64:

    ❯ echo "performance" | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor

    ❯ go1.27.1 test -cpu 1,2,4,8,16,32 -test.benchmem -bench .
    goos: linux
    goarch: amd64
    pkg: github.com/mwyvr/kid/eval/bench
    cpu: Intel(R) Core(TM) i9-14900K
    BenchmarkKid                    33479890                29.90 ns/op            0 B/op          0 allocs/op
    BenchmarkKid-2                  37299408                30.42 ns/op            0 B/op          0 allocs/op
    BenchmarkKid-4                  39092158                34.42 ns/op            0 B/op          0 allocs/op
    BenchmarkKid-8                  37107082                35.16 ns/op            0 B/op          0 allocs/op
    BenchmarkKid-16                 32040319                37.40 ns/op            0 B/op          0 allocs/op
    BenchmarkKid-32                 48334761                24.43 ns/op            0 B/op          0 allocs/op
    BenchmarkXid                    42041840                27.39 ns/op            0 B/op          0 allocs/op
    BenchmarkXid-2                  40561254                29.56 ns/op            0 B/op          0 allocs/op
    BenchmarkXid-4                  37046432                32.02 ns/op            0 B/op          0 allocs/op
    BenchmarkXid-8                  36434624                32.06 ns/op            0 B/op          0 allocs/op
    BenchmarkXid-16                 42056515                33.27 ns/op            0 B/op          0 allocs/op
    BenchmarkXid-32                 57809797                22.64 ns/op            0 B/op          0 allocs/op
    BenchmarkKsuid                  16296384                74.23 ns/op            0 B/op          0 allocs/op
    BenchmarkKsuid-2                13288528                83.34 ns/op            0 B/op          0 allocs/op
    BenchmarkKsuid-4                11992980                98.58 ns/op            0 B/op          0 allocs/op
    BenchmarkKsuid-8                10171071               112.0 ns/op             0 B/op          0 allocs/op
    BenchmarkKsuid-16                8037324               144.3 ns/op             0 B/op          0 allocs/op
    BenchmarkKsuid-32                6632499               180.2 ns/op             0 B/op          0 allocs/op
    BenchmarkUuidV4                 32103150                36.21 ns/op            0 B/op          0 allocs/op
    BenchmarkUuidV4-2               52652762                26.75 ns/op            0 B/op          0 allocs/op
    BenchmarkUuidV4-4               33758822                33.66 ns/op            0 B/op          0 allocs/op
    BenchmarkUuidV4-8               36221518                33.82 ns/op            0 B/op          0 allocs/op
    BenchmarkUuidV4-16              42978082                33.54 ns/op            0 B/op          0 allocs/op
    BenchmarkUuidV4-32              55658079                21.20 ns/op            0 B/op          0 allocs/op
    BenchmarkUuidV7                 21967980                54.71 ns/op            0 B/op          0 allocs/op
    BenchmarkUuidV7-2               19782433                59.61 ns/op            0 B/op          0 allocs/op
    BenchmarkUuidV7-4               14816305                77.07 ns/op            0 B/op          0 allocs/op
    BenchmarkUuidV7-8               12235728                92.53 ns/op            0 B/op          0 allocs/op
    BenchmarkUuidV7-16              11204155               112.6 ns/op             0 B/op          0 allocs/op
    BenchmarkUuidV7-32               8841495               135.9 ns/op             0 B/op          0 allocs/op
    BenchmarkUlid                   20344285                58.53 ns/op           16 B/op          1 allocs/op
    BenchmarkUlid-2                 29226808                40.74 ns/op           16 B/op          1 allocs/op
    BenchmarkUlid-4                 33935920                32.94 ns/op           16 B/op          1 allocs/op
    BenchmarkUlid-8                 44736744                32.57 ns/op           16 B/op          1 allocs/op
    BenchmarkUlid-16                48807732                33.71 ns/op           16 B/op          1 allocs/op
    BenchmarkUlid-32                48737070                24.79 ns/op           16 B/op          1 allocs/op
    BenchmarkSonyflake                 30884             38837 ns/op               0 B/op          0 allocs/op
    BenchmarkSonyflake-2               30804             38924 ns/op               0 B/op          0 allocs/op
    BenchmarkSonyflake-4               30796             38923 ns/op               0 B/op          0 allocs/op
    BenchmarkSonyflake-8               30864             38884 ns/op               0 B/op          0 allocs/op
    BenchmarkSonyflake-16              30885             38833 ns/op               0 B/op          0 allocs/op
    BenchmarkSonyflake-32              30927             38811 ns/op               0 B/op          0 allocs/op
    BenchmarkGoShortUniqueID         5278425               228.0 ns/op            87 B/op          5 allocs/op
    BenchmarkGoShortUniqueID-2       5559020               212.4 ns/op            87 B/op          5 allocs/op
    BenchmarkGoShortUniqueID-4       5518480               207.4 ns/op            87 B/op          5 allocs/op
    BenchmarkGoShortUniqueID-8       5414343               225.5 ns/op            87 B/op          5 allocs/op
    BenchmarkGoShortUniqueID-16      4611174               257.1 ns/op            87 B/op          5 allocs/op
    BenchmarkGoShortUniqueID-32      4029286               298.6 ns/op            88 B/op          5 allocs/op
    PASS
    ok      github.com/mwyvr/kid/eval/bench 66.164s

Generators utilizing `crypto/rand` show a marked difference when run on MacOS
arm64 platforms rather than Linux amd64 platforms, highlighting the additional
fixed overhead of obtaining those bytes through the MacOS CSPRNG interface.

arm64, on mains power, performance set to "Hi Power":

    ❯ go1.27.1 test -cpu 1,2,4,8,16 -test.benchmem -bench .
    goos: darwin
    goarch: arm64
    pkg: github.com/mwyvr/kid/eval/bench
    cpu: Apple M4 Max
    BenchmarkKid                    38147826                31.37 ns/op            0 B/op          0 allocs/op
    BenchmarkKid-2                  36859046                32.22 ns/op            0 B/op          0 allocs/op
    BenchmarkKid-4                  35832750                31.53 ns/op            0 B/op          0 allocs/op
    BenchmarkKid-8                  23950760                52.28 ns/op            0 B/op          0 allocs/op
    BenchmarkKid-16                 17955895                67.20 ns/op            0 B/op          0 allocs/op
    BenchmarkXid                    39871525                30.26 ns/op            0 B/op          0 allocs/op
    BenchmarkXid-2                  39114384                30.04 ns/op            0 B/op          0 allocs/op
    BenchmarkXid-4                  44777858                28.29 ns/op            0 B/op          0 allocs/op
    BenchmarkXid-8                  26317808                47.35 ns/op            0 B/op          0 allocs/op
    BenchmarkXid-16                 18420999                65.57 ns/op            0 B/op          0 allocs/op
    BenchmarkKsuid                   5700789               210.6 ns/op             0 B/op          0 allocs/op
    BenchmarkKsuid-2                 4427120               270.5 ns/op             0 B/op          0 allocs/op
    BenchmarkKsuid-4                 3504735               342.9 ns/op             0 B/op          0 allocs/op
    BenchmarkKsuid-8                 3501440               342.7 ns/op             0 B/op          0 allocs/op
    BenchmarkKsuid-16                3434904               350.8 ns/op             0 B/op          0 allocs/op
    BenchmarkUuidV4                  7117567               168.8 ns/op             0 B/op          0 allocs/op
    BenchmarkUuidV4-2                4823792               250.5 ns/op             0 B/op          0 allocs/op
    BenchmarkUuidV4-4                1737500               698.7 ns/op             0 B/op          0 allocs/op
    BenchmarkUuidV4-8                2699523               452.6 ns/op             0 B/op          0 allocs/op
    BenchmarkUuidV4-16               2767894               436.1 ns/op             0 B/op          0 allocs/op
    BenchmarkUuidV7                 14865351                85.49 ns/op            0 B/op          0 allocs/op
    BenchmarkUuidV7-2                8879036               134.8 ns/op             0 B/op          0 allocs/op
    BenchmarkUuidV7-4                6082287               197.7 ns/op             0 B/op          0 allocs/op
    BenchmarkUuidV7-8                3909390               299.2 ns/op             0 B/op          0 allocs/op
    BenchmarkUuidV7-16               4933089               242.3 ns/op             0 B/op          0 allocs/op
    BenchmarkUlid                   14177623                84.08 ns/op           16 B/op          1 allocs/op
    BenchmarkUlid-2                  8931943               141.4 ns/op            16 B/op          1 allocs/op
    BenchmarkUlid-4                  6166581               195.1 ns/op            16 B/op          1 allocs/op
    BenchmarkUlid-8                  5184649               234.3 ns/op            16 B/op          1 allocs/op
    BenchmarkUlid-16                 5373464               224.2 ns/op            16 B/op          1 allocs/op
    BenchmarkSonyflake                 31234             39032 ns/op               0 B/op          0 allocs/op
    BenchmarkSonyflake-2               30950             38757 ns/op               0 B/op          0 allocs/op
    BenchmarkSonyflake-4               30940             38757 ns/op               0 B/op          0 allocs/op
    BenchmarkSonyflake-8               30949             38755 ns/op               0 B/op          0 allocs/op
    BenchmarkSonyflake-16              31006             38969 ns/op               1 B/op          0 allocs/op
    BenchmarkGoShortUniqueID         5489059               199.6 ns/op            87 B/op          5 allocs/op
    BenchmarkGoShortUniqueID-2       9448108               126.7 ns/op            87 B/op          5 allocs/op
    BenchmarkGoShortUniqueID-4       7998060               146.4 ns/op            87 B/op          5 allocs/op
    BenchmarkGoShortUniqueID-8       6234685               191.4 ns/op            87 B/op          5 allocs/op
    BenchmarkGoShortUniqueID-16      4045058               293.8 ns/op            87 B/op          5 allocs/op
    PASS
    ok      github.com/mwyvr/kid/eval/bench 58.234s
