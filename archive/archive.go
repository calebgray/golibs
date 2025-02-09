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
	// Create a channel for work items
	type workItem struct {
		file    *zip.File
		outpath string
	}
	workChan := make(chan workItem)

	// Create worker pool
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU()
	wg.Add(numWorkers)

	// Start workers
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for work := range workChan {
				if verbose {
					println("Extracting:", work.file.Name)
				}

				// Open a Zip Entry
				rc, err := work.file.Open()
				if err != nil {
					println("Can't open a zip entry:", err)
					continue
				}

				// Create Destination File
				w, err := os.Create(work.outpath)
				if err != nil {
					rc.Close()
					println("Can't open a file to write:", err)
					continue
				}

				// Extract the File
				io.Copy(w, rc)

				// Clean up
				rc.Close()
				w.Close()
			}
		}()
	}

	// Process each zip file
	for _, file := range files {
		// Open Reader
		r, err := zip.OpenReader(file)
		if err != nil {
			println("Can't open zip file:", err)
			continue
		}

		// Queue work for all files
		for _, f := range r.File {
			if f.Mode().IsDir() {
				continue
			}

			name := f.FileHeader.Name
			outpath := path.Join(outdir, name)
			os.MkdirAll(path.Dir(outpath), 0755)

			workChan <- workItem{
				file:    f,
				outpath: outpath,
			}
		}

		r.Close()
		println("Queued:", file)
	}

	// Signal workers to stop
	close(workChan)

	// Wait for all workers to finish
	wg.Wait()
	println("All files extracted")
}
