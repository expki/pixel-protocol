package main

import (
	"embed"
	"io"
	"io/fs"
	"log"

	"github.com/klauspost/compress/zstd"
	"github.com/spf13/afero"
)

//go:embed dist/*
var root embed.FS

var dist = func() fs.FS {
	dist, err := fs.Sub(root, "dist")
	if err != nil {
		log.Fatalf("open embedding dist: %v", err)
	}
	return dist
}()

var distZstd = func() fs.FS {
	memfs := afero.NewMemMapFs()
	encoder, err := zstd.NewWriter(nil,
		zstd.WithEncoderLevel(zstd.SpeedDefault),
		zstd.WithEncoderCRC(false),
		zstd.WithEncoderConcurrency(1),
	)
	if err != nil {
		log.Fatalf("initialize embedding compression: %v", err)
	}
	err = fs.WalkDir(dist, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		file, err := dist.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil {
			return err
		}
		content, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		var contentZstd []byte
		contentZstd = encoder.EncodeAll(content, contentZstd)
		return afero.WriteFile(memfs, path, contentZstd, info.Mode().Perm())
	})
	if err != nil {
		log.Fatalf("read embedding dist: %v", err)
	}
	return &compressedFS{Fs: memfs}
}()

type compressedFS struct {
	afero.Fs
}

func (c *compressedFS) Open(name string) (fs.File, error) {
	return c.Fs.Open(name)
}
