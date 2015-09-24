# repobench

An ipfs repo benchmarking tool.

## Usage

Make sure you have an ipfs repo initialized, no daemon running and a few hundred MB at least of free disk space.

Simply run `repobench` and it will give you some performance numbers back.

## Example

```
whyrusleeping@idril ~/g/s/g/w/repobench (master)> repobench 
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
```
