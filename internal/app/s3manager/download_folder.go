package s3manager

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
)

// HandleDownloadFolder streams every object below a folder prefix to the client
// as a single ZIP archive.
//
// The archive is written straight to the response as the objects are read, so
// no temporary file is created and memory usage stays independent of the folder
// size. Because the response is streamed, a failure that happens after the
// first bytes were written cannot change the status code; it is logged and the
// (truncated) archive is left for the client to discard.
func HandleDownloadFolder(s3 S3) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bucketName := mux.Vars(r)["bucketName"]
		prefix := mux.Vars(r)["objectName"]

		// A folder prefix always ends with a slash. An empty prefix means the
		// whole bucket.
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}

		archiveName := bucketName
		if prefix != "" {
			archiveName = path.Base(strings.TrimSuffix(prefix, "/"))
		}
		archiveName += ".zip"

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", contentDisposition(archiveName))

		zw := zip.NewWriter(w)
		count := 0

		objectCh := s3.ListObjects(r.Context(), bucketName, minio.ListObjectsOptions{Prefix: prefix, Recursive: true})
		for object := range objectCh {
			if object.Err != nil {
				zipStreamError(w, zw, count, fmt.Errorf("error listing objects to archive: %w", object.Err))
				return
			}

			// Skip the zero-byte marker objects some S3 providers use to
			// represent a folder; the directory structure comes from the
			// entry names instead.
			if strings.HasSuffix(object.Key, "/") {
				continue
			}

			if err := addObjectToZip(r, s3, zw, bucketName, object, prefix); err != nil {
				zipStreamError(w, zw, count, err)
				return
			}
			count++
		}

		if count == 0 {
			// Nothing was written yet, so a proper error response is still possible.
			handleHTTPError(w, fmt.Errorf("%w: no objects found at %q", errEmptyArchive, prefix))
			return
		}

		if err := zw.Close(); err != nil {
			log.Printf("error finalizing zip archive: %v", err)
		}
	}
}

// addObjectToZip copies a single object into the archive, storing it under its
// path relative to the folder prefix.
func addObjectToZip(r *http.Request, s3 S3, zw *zip.Writer, bucketName string, object minio.ObjectInfo, prefix string) error {
	src, err := s3.GetObject(r.Context(), bucketName, object.Key, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("error getting object %q: %w", object.Key, err)
	}
	defer func() {
		if cErr := src.Close(); cErr != nil {
			log.Printf("error closing object %q: %v", object.Key, cErr)
		}
	}()

	entryName := strings.TrimPrefix(object.Key, prefix)
	if entryName == "" {
		entryName = path.Base(object.Key)
	}

	header := &zip.FileHeader{
		Name:     entryName,
		Method:   zip.Deflate,
		Modified: object.LastModified,
	}
	entry, err := zw.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("error creating zip entry for %q: %w", object.Key, err)
	}

	if _, err := io.Copy(entry, src); err != nil {
		return fmt.Errorf("error writing object %q to archive: %w", object.Key, err)
	}
	return nil
}

// zipStreamError reports a failure that happened while building the archive.
// Once the first entry has been written the response body is already a partial
// ZIP stream, so the error can only be logged.
func zipStreamError(w http.ResponseWriter, zw *zip.Writer, written int, err error) {
	if written == 0 {
		handleHTTPError(w, err)
		return
	}
	log.Printf("error streaming zip archive after %d object(s): %v", written, err)
	if cErr := zw.Close(); cErr != nil {
		log.Printf("error closing zip archive: %v", cErr)
	}
}

// contentDisposition builds an attachment header that carries both an ASCII
// fallback and the UTF-8 encoded file name.
func contentDisposition(name string) string {
	ascii := strings.Map(func(r rune) rune {
		if r < 32 || r > 126 || r == '"' || r == '\\' {
			return '_'
		}
		return r
	}, name)
	return fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s", ascii, url.PathEscape(name))
}
