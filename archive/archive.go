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

				// Create Destination File with temporary permissions
				w, err := os.OpenFile(work.outpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
				if err != nil {
					rc.Close()
					println("Can't open a file to write:", err)
					continue
				}

				// Extract the File
				_, err = io.Copy(w, rc)

				// Close handles before changing permissions
				rc.Close()
				w.Close()

				if err != nil {
					println("Error copying file:", err)
					continue
				}

				// Get original permissions from zip
				mode := work.file.Mode()

				// If the file is marked as executable in the zip, ensure execute bits are set
				if mode&0100 != 0 { // User executable
					mode |= 0111 // Add execute bits for all
				}

				// Apply the permissions
				if err := os.Chmod(work.outpath, mode); err != nil {
					println("Error setting file permissions:", err)
				}
			}
		}()
	}

	for _, file := range files {
		r, err := zip.OpenReader(file)
		if err != nil {
			println("Can't open zip file:", err)
			continue
		}

		for _, f := range r.File {
			if f.Mode().IsDir() {
				name := f.FileHeader.Name
				outpath := path.Join(outdir, name)
				if err := os.MkdirAll(outpath, f.Mode()); err != nil {
					println("Error creating directory:", err)
				}
				continue
			} else {
				name := f.FileHeader.Name
				outpath := path.Join(outdir, name)

				if err := os.MkdirAll(path.Dir(outpath), 0755); err != nil {
					println("Error creating parent directory:", err)
					continue
				}

				workChan <- workItem{
					file:    f,
					outpath: outpath,
				}
			}
		}

		r.Close()
		println("Queued:", file)
	}

	close(workChan)

	wg.Wait()
	println("All files extracted")
}
