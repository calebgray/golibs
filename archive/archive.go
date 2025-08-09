package archive

import (
	"archive/zip"
	"io"
	"os"
	"path"
	"runtime"
	"sync"
)

func Unzip(zipFilename string, outDir string, verbose bool) {
	// Open zip file once
	zipReader, err := zip.OpenReader(zipFilename)
	if err != nil {
		os.Stderr.WriteString("Can't read zip (" + zipFilename + "):" + err.Error() + "\n")
		return
	}
	defer zipReader.Close()

	// Create a channel to distribute work
	fileChan := make(chan *zip.File, len(zipReader.File))

	// Queue all non-directory files for processing
	for _, zipReaderFile := range zipReader.File {
		// Create directories ahead of time
		if zipReaderFile.Mode().IsDir() {
			os.MkdirAll(path.Join(outDir, zipReaderFile.Name), 0755)
			continue
		}

		// Make the parent directory
		os.MkdirAll(path.Dir(path.Join(outDir, zipReaderFile.Name)), 0755)

		// Queue work for each file
		fileChan <- zipReaderFile
	}
	close(fileChan) // Close channel after all work is queued

	// Thread per CPU to extract files from zip
	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()

			// Process files from the channel
			for zipFile := range fileChan {
				// Open the file in the zip for reading
				rc, err := zipFile.Open()
				if err != nil {
					os.Stderr.WriteString("Can't read file (" + zipFile.Name + "):" + err.Error() + "\n")
					continue
				}
				defer rc.Close()

				// Determine the output filename
				zipEntryFile := path.Join(outDir, zipFile.Name)
				if verbose {
					println(zipEntryFile)
				}

				// Create Destination File
				if w, err := os.Create(zipEntryFile); err != nil {
					os.Stderr.WriteString("Can't write file (" + zipEntryFile + "):" + err.Error() + "\n")
				} else {
					// Extract the File
					if _, err := io.Copy(w, rc); err != nil {
						os.Stderr.WriteString("Can't extract file (" + zipFile.Name + ") to (" + zipEntryFile + "): " + err.Error() + "\n")
					}

					// Clean up
					w.Close()
				}
			}
		}()
	}

	// Wait for all workers to finish
	wg.Wait()
}
