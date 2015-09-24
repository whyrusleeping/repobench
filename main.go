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

	humanize "github.com/dustin/go-humanize"
	randbo "github.com/dustin/randbo"
	context "golang.org/x/net/context"
)

type BenchCfg struct {
	Blocksize int64
}

func (bcfg *BenchCfg) String() string {
	return fmt.Sprintf("Bench Config:\n\tBlocksize: %d\n", bcfg.Blocksize)
}

func main() {
	home := os.Getenv("HOME")

	r, err := fsrepo.Open(path.Join(home, ".ipfs"))
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
		Blocksize: 1024 * 256,
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
	sizes := []int64{
		1024,
		1024 * 256,
		1024 * 512,
		1024 * 1024,
		1024 * 1024 * 2,
		1024 * 1024 * 16,
	}
	for _, s := range sizes {
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
