package archive

import (
	"archive/zip"
	"io"
	"os"
	"path"
	"runtime"
	"sync"
)

func Unzip(files []string, outdir string, verbose bool) {
	// Determine the number of available CPU threads
	maxGoroutines := runtime.NumCPU()
	sem := make(chan struct{}, maxGoroutines) // Semaphore to limit concurrency

	for _, file := range files {
		// Open Reader
		r, err := zip.OpenReader(file)
		if err != nil {
			println("Can't open zip file:", err)
			continue
		}
		defer r.Close()

		// Extract Files in Parallel
		var wg sync.WaitGroup
		for _, f := range r.File {
			if f.Mode().IsDir() {
				continue
			}

			name := f.FileHeader.Name
			outpath := path.Join(outdir, name)
			os.MkdirAll(path.Dir(outpath), 0755)

			wg.Add(1)
			sem <- struct{}{} // Acquire a semaphore slot
			go func(file *zip.File) {
				defer wg.Done()
				defer func() { <-sem }() // Release the semaphore slot

				if verbose {
					println("Extracting:", file.Name)
				}

				// Open a Zip Entry
				rc, err := file.Open()
				if err != nil {
					println("Can't open a zip entry:", err)
					return
				}
				defer rc.Close()

				// Create Destination File
				w, err := os.Create(outpath)
				if err != nil {
					println("Can't open a file to write:", err)
					return
				}
				defer w.Close()

				// Extract the File
				io.Copy(w, rc)
			}(f)
		}
		wg.Wait()

		// Finished
		println("Extracted:", file)
	}
}
