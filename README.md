# repobench

An ipfs repo benchmarking tool.

## Usage

Make sure you have an ipfs repo initialized, no daemon running and a few hundred MB at least of free disk space.

```
whyrusleeping@idril ~> repobench 
whyrusleeping@idril ~> repobench --blocksize=524288
```

### Options

#### blocksize
Size to use for chunking input during benchmarks.



## Example

```
whyrusleeping@idril ~> repobench 
Bench Config:
	Blocksize: 262144

BlockRewrites:		 3000000	       497 ns/op
RandomBlockWrites:	     200	   9792484 ns/op	  26.77 MB/s
Add File (1.0KiB):	     200	   6868517 ns/op	   0.15 MB/s
Add File (256KiB):	     200	   9824129 ns/op	  26.68 MB/s
Add File (512KiB):	      50	  33289816 ns/op	  15.75 MB/s
Add File (1.0MiB):	      50	  30676414 ns/op	  34.18 MB/s
Add File (2.0MiB):	      20	  50882739 ns/op	  41.22 MB/s
Add File (16MiB):	       3	 339807645 ns/op	  49.37 MB/s
DiskWrite (1.0KiB):	   50000	     30809 ns/op	  33.24 MB/s
DiskWrite (256KiB):	    2000	    641900 ns/op	 408.39 MB/s
DiskWrite (512KiB):	    1000	   1218635 ns/op	 430.23 MB/s
DiskWrite (1.0MiB):	    1000	   2321069 ns/op	 451.76 MB/s
DiskWrite (2.0MiB):	     300	   4616830 ns/op	 454.24 MB/s
DiskWrite (16MiB):	      50	  37242923 ns/op	 450.48 MB/s
```
