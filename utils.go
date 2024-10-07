package restapi

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func GetVideoStreamHandler(filePath string) RouteHandlerFunc {

	return func(w http.ResponseWriter, r *http.Request, rc *RouteContext) {
		videoFile := filePath
		file, err := os.Open(videoFile)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		defer file.Close()

		// Get file size
		stat, err := file.Stat()
		if err != nil {
			http.Error(w, "Unable to get file info", http.StatusInternalServerError)
			return
		}
		fileSize := stat.Size()

		// Parse the Range header
		rangeHeader := r.Header.Get("Range")
		if rangeHeader == "" {
			// If there's no Range header, return the full file
			w.Header().Set("Content-Type", "video/mp4")
			w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
			http.ServeFile(w, r, videoFile)
			return
		}

		// Example: "bytes=0-1024"
		rangeParts := strings.Split(rangeHeader, "=")
		if len(rangeParts) != 2 || rangeParts[0] != "bytes" {
			http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
			return
		}

		byteRange := strings.Split(rangeParts[1], "-")
		start, err := strconv.ParseInt(byteRange[0], 10, 64)
		if err != nil {
			http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
			return
		}

		// End might be omitted (client requesting until the end)
		var end int64
		if len(byteRange) == 2 && byteRange[1] != "" {
			end, err = strconv.ParseInt(byteRange[1], 10, 64)
			if err != nil {
				http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
				return
			}
		} else {
			end = fileSize - 1
		}

		if start > end || end >= fileSize {
			http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
			return
		}

		// Set headers for partial content
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
		w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
		w.WriteHeader(http.StatusPartialContent)

		// Seek to the requested range and write the chunk
		file.Seek(start, 0)
		buffer := make([]byte, end-start+1)
		file.Read(buffer)
		w.Write(buffer)
	}
}
