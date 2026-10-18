package auth

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// atomicWriteData writes b to f. It is a package-level seam so tests can
// inject a partial-write failure; production code always uses f.Write, which
// returns a non-nil error on any short write.
var atomicWriteData = func(f *os.File, b []byte) error {
	_, err := f.Write(b)
	return err
}

// writeFileAtomic replaces the file at path with data so that readers only
// ever observe either the previous complete contents or the new complete
// contents, never an empty or partial file.
//
// The data is written to a temporary file in the same directory (so the
// final rename stays on one filesystem), the temp file's mode is set to perm
// before any bytes are written, the data is fsynced, and the temp file is
// renamed over path. The resulting file always has mode perm, even if a
// previous file at path had a different mode.
//
// On any failure before the rename completes, the temp file is removed, the
// previous file at path is left untouched, and a wrapped error naming the
// failed step is returned. After a successful rename, the parent directory is
// fsynced on a best-effort basis; a failure there is logged but not returned
// because the new contents are already in place.
func writeFileAtomic(path string, data []byte, perm os.FileMode) (err error) {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file in %s: %w", dir, err)
	}
	tmp := f.Name()
	closed := false

	defer func() {
		if err == nil {
			return
		}
		if !closed {
			if closeErr := f.Close(); closeErr != nil {
				log.Printf("[auth] atomic write: closing temp file %s after error: %v", tmp, closeErr)
			}
		}
		if rmErr := os.Remove(tmp); rmErr != nil && !os.IsNotExist(rmErr) {
			log.Printf("[auth] atomic write: removing temp file %s after error: %v", tmp, rmErr)
		}
	}()

	if err = f.Chmod(perm); err != nil {
		return fmt.Errorf("setting mode on temp file %s: %w", tmp, err)
	}
	if err = atomicWriteData(f, data); err != nil {
		return fmt.Errorf("writing temp file %s: %w", tmp, err)
	}
	if err = f.Sync(); err != nil {
		return fmt.Errorf("syncing temp file %s: %w", tmp, err)
	}
	closed = true
	if err = f.Close(); err != nil {
		return fmt.Errorf("closing temp file %s: %w", tmp, err)
	}
	if err = os.Rename(tmp, path); err != nil {
		return fmt.Errorf("renaming temp file %s to %s: %w", tmp, path, err)
	}

	syncDir(dir)
	return nil
}

// syncDir fsyncs dir so a preceding rename within it is durable. It is best
// effort: failures are logged, not returned, because the rename has already
// taken effect.
func syncDir(dir string) {
	d, err := os.Open(dir)
	if err != nil {
		log.Printf("[auth] atomic write: opening dir %s for sync: %v", dir, err)
		return
	}
	if err := d.Sync(); err != nil {
		log.Printf("[auth] atomic write: syncing dir %s: %v", dir, err)
	}
	if err := d.Close(); err != nil {
		log.Printf("[auth] atomic write: closing dir %s after sync: %v", dir, err)
	}
}
