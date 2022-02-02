# functional-benchmarker
This is an internal utility based on the project's functional test framework which allows benchmarking
OpenShift Logging's efficiency in processing messages.  The intention is to use this utility to assess changes
to the processing pipelines.

## Requirements

* gnuplot for reporting

## Building
```make bin/functional-benchmarker```

## Results

The benchmark utility will generate results in a directory prefixed with "benchmark" and write similiar
results to stdout.

```
$ ./bin/functional-benchmarker
                                                                               
                                      Mem(Mb)                                  
                                                                               
    103 +------------------------------------------------------------------+   
        |     +  *  +     +     +     +      +     +     +     +     +     |   
  102.9 |-+      *                            'mem.data' using 1:2 *******-|   
  102.8 |-+      *                                                       +-|   
        |        *                                                         |   
  102.7 |-+      *                                                       +-|   
        |        *                                                         |   
  102.6 |-+      *                                                       +-|   
        |        *                                                         |   
  102.5 |-+      *                                                       +-|   
  102.4 |-+      *                                                       +-|   
        |        *                                                         |   
  102.3 |-+      *                                                       +-|   
        |        *                                                         |   
  102.2 |-+      *                                                       +-|   
  102.1 |-+      *                                                       +-|   
        |    *****  +     +     +     +      +     +     +     +     +     |   
    102 +------------------------------------------------------------------+   
      33:30 34:00 34:30 35:00 35:30 36:00  36:30 37:00 37:30 38:00 38:30 39:00 
                                       Time                                    
                                                                               

                                                                               
                                     CPU(Cores)                                
                                                                               
   0.003 +-----------------------------------------------------------------+   
         |     +  *  +     +     +     +     +  *  +  *  +     +     +     |   
         |        *                           'c*u.dat*' using 1:2 ******* |   
         |        *                             *     *                    |   
  0.0025 |-+      *                             *     *                  +-|   
         |        *                             *     *                    |   
         |        *                             *     *                    |   
         |        *                             *     *                    |   
   0.002 |-+      *     *************************     *********************|   
         |        *     *                                                  |   
         |        *     *                                                  |   
         |        *     *                                                  |   
         |        *     *                                                  |   
  0.0015 |-+      *     *                                                +-|   
         |        *     *                                                  |   
         |        *     *                                                  |   
         |     +  *  +  *  +     +     +     +     +     +     +     +     |   
   0.001 +-----------------------------------------------------------------+   
       33:30 34:00 34:30 35:00 35:30 36:00 36:30 37:00 37:30 38:00 38:30 39:00 
                                        Time                                   
                                                                               

      Total      Size   Elapsed      Mean       Min       Max    Median
        Msg   (bytes)                 (s)       (s)       (s)       (s)
   --------  --------  --------  --------  --------  --------  --------
        297      1024      5m0s     3.425     0.874     5.991     2.990

```
## Platform notes
Running on `crc` requires enabling monitoring and adding more memory:
```
 crc config set enable-cluster-monitoring true
 crc start -m16384 
```