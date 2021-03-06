package blockfs

import (
	"github.com/dchest/safefile"
	"io"
	"os"
	"path/filepath"
)

const BlocksDir string = "blocks"
const BlockSize int = 1024 * 1024 * 2

type FS struct {
	Path string
}

func mkdirs(paths ...string) error {
	for _, path := range paths {
		err := os.Mkdir(path, 0755)
		if err != nil && !os.IsExist(err) {
			return err
		}
	}
	return nil
}

func NewFS(path string) (*FS, error) {
	err := mkdirs(
		path,
		filepath.Join(path, BlocksDir),
	)
	if err != nil {
		return nil, err
	}
	return &FS{Path: path}, nil
}

func (fs *FS) WriteBlock(b *Block) error {
	path := filepath.Join(fs.Path, BlocksDir, string(b.Checksum))
	f, err := safefile.Create(path, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = b.WriteTo(f)
	if err != nil {
		return err
	}
	return f.Commit()
}

func (fs *FS) Export(r io.Reader) (HashList, error) {
	hashes := HashList{}
	buffer := make([]byte, BlockSize)
	for {
		length, err := io.ReadFull(r, buffer)
		if length == 0 {
			break
		}
		b := NewBlock(buffer[:length])
		fs.WriteBlock(b)
		hashes = append(hashes, b.Checksum)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return hashes, err
		}
	}
	return hashes, nil
}

func (fs *FS) GetBlock(c Checksum) (*Block, error) {
	f, err := os.Open(filepath.Join(fs.Path, BlocksDir, string(c)))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b := make([]byte, BlockSize)
	length, err := f.Read(b)
	if err != nil {
		return nil, err
	}
	return NewBlock(b[:length]), nil
}

func (fs *FS) Exists(c Checksum) bool {
	_, err := os.Stat(filepath.Join(fs.Path, BlocksDir, string(c)))
	if err != nil {
		return false
	}
	return true
}

func (fs *FS) MissingBlocks(h HashList) HashList {
	missing := HashList{}
	for _, checksum := range h {
		if !fs.Exists(checksum) {
			missing = append(missing, checksum)
		}
	}
	return missing
}
