package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	blocks "github.com/ipfs/go-ipfs/blocks"
	key "github.com/ipfs/go-ipfs/blocks/key"
	core "github.com/ipfs/go-ipfs/core"
	cr "github.com/ipfs/go-ipfs/core/corerepo"
	importer "github.com/ipfs/go-ipfs/importer"
	chunk "github.com/ipfs/go-ipfs/importer/chunk"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"

	kingpin "github.com/alecthomas/kingpin"
	humanize "github.com/dustin/go-humanize"
	randbo "github.com/dustin/randbo"
	context "golang.org/x/net/context"
)

var WriteSizes = []int64{
	1024,
	1024 * 256,
	1024 * 512,
	1024 * 1024,
	1024 * 1024 * 2,
	1024 * 1024 * 16,
}

type BenchCfg struct {
	Blocksize int64
}

func (bcfg *BenchCfg) String() string {
	return fmt.Sprintf("Bench Config:\n\tBlocksize: %d\n", bcfg.Blocksize)
}

func main() {
	bsize := kingpin.Flag("blocksize", "blocksize to test with").Default("262144").Int64()
	kingpin.Parse()

	home := os.Getenv("HOME")

	ipfsdir := path.Join(home, ".ipfs")
	r, err := fsrepo.Open(ipfsdir)
	if err != nil {
		fmt.Printf("Failed to open ipfs repo at: %s/.ipfs: %s\n", home, err)
		fmt.Println("Please ensure ipfs has been initialized and that no daemon is running")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nd, err := core.NewNode(ctx, &core.BuildCfg{
		Repo: r,
	})
	if err != nil {
		fmt.Printf("failed to create node: %s\n", err)
		return
	}

	cfg := &BenchCfg{
		Blocksize: *bsize,
	}

	fmt.Println(cfg)
	err = BenchmarkBlockRewrites(nd, cfg)
	if err != nil {
		panic(err)
	}

	err = BenchmarkRandomBlockWrites(nd, cfg)
	if err != nil {
		panic(err)
	}

	err = BenchmarkAdd(nd, cfg)
	if err != nil {
		panic(err)
	}

	err = BenchmarkDiskWrites(ipfsdir)
	if err != nil {
		panic(err)
	}
}

func BenchmarkRandomBlockWrites(n *core.IpfsNode, cfg *BenchCfg) error {
	buf := make([]byte, cfg.Blocksize)
	read := randbo.New()

	var keys []key.Key
	f := func(b *testing.B) {
		b.SetBytes(cfg.Blocksize)
		for i := 0; i < b.N; i++ {
			read.Read(buf)
			blk := blocks.NewBlock(buf)
			k, err := n.Blocks.AddBlock(blk)
			if err != nil {
				b.Fatal(err)
			}

			keys = append(keys, k)
		}
	}

	br := testing.Benchmark(f)
	fmt.Printf("RandomBlockWrites:\t%s\n", br)

	// clean up
	for _, k := range keys {
		err := n.Blocks.DeleteBlock(k)
		if err != nil {
			return err
		}
	}

	return nil
}

func BenchmarkBlockRewrites(n *core.IpfsNode, cfg *BenchCfg) error {
	buf := make([]byte, cfg.Blocksize)
	randbo.New().Read(buf)

	blk := blocks.NewBlock(buf)
	// write the block first, before starting the benchmark.
	// we're just looking at the time it takes to write a block thats already
	// been written
	k, err := n.Blocks.AddBlock(blk)
	if err != nil {
		return err
	}

	f := func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := n.Blocks.AddBlock(blk)
			if err != nil {
				b.Fatal(err)
			}
		}
	}

	br := testing.Benchmark(f)
	fmt.Printf("BlockRewrites:\t\t%s\n", br)

	// clean up
	err = n.Blocks.DeleteBlock(k)
	if err != nil {
		return err
	}

	return nil
}

func BenchmarkAdd(n *core.IpfsNode, cfg *BenchCfg) error {
	for _, s := range WriteSizes {
		err := benchAddSize(n, cfg, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func benchAddSize(n *core.IpfsNode, cfg *BenchCfg, size int64) error {
	f := func(b *testing.B) {
		b.SetBytes(size)
		for i := 0; i < b.N; i++ {
			r := io.LimitReader(randbo.New(), size)
			spl := chunk.NewSizeSplitter(r, cfg.Blocksize)
			_, err := importer.BuildDagFromReader(n.DAG, spl, nil)
			if err != nil {
				fmt.Printf("ERRROR: ", err)
				b.Fatal(err)
			}
		}
	}

	br := testing.Benchmark(f)
	bs := humanize.IBytes(uint64(size))
	fmt.Printf("Add File (%s):\t%s\n", bs, br)

	err := cr.GarbageCollect(n, context.Background())
	if err != nil {
		return err
	}

	return nil
}

func BenchmarkDiskWrites(ipfsdir string) error {
	for _, n := range WriteSizes {
		err := benchDiskWriteSize(ipfsdir, n)
		if err != nil {
			return err
		}
	}
	return nil
}

func benchDiskWriteSize(dir string, size int64) error {
	benchdir := path.Join(dir, fmt.Sprintf("benchfiles-%d", size))
	err := os.Mkdir(benchdir, 0777)
	if err != nil {
		return err
	}

	n := 0
	f := func(b *testing.B) {
		b.SetBytes(size)
		r := randbo.New()
		for i := 0; i < b.N; i++ {
			n++
			fi, err := os.Create(path.Join(dir, fmt.Sprint(n)))
			if err != nil {
				fmt.Println(err)
				b.Fatal(err)
			}

			_, err = io.CopyN(fi, r, size)
			if err != nil {
				fi.Close()
				fmt.Println(err)
				b.Fatal(err)
			}
			fi.Close()
		}
	}

	br := testing.Benchmark(f)
	bs := humanize.IBytes(uint64(size))
	fmt.Printf("DiskWrite (%s):\t%s\n", bs, br)

	err = os.RemoveAll(benchdir)
	if err != nil {
		return err
	}
	return nil
}
