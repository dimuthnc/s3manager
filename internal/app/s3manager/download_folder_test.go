package s3manager_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudlena/s3manager/internal/app/s3manager"
	"github.com/cloudlena/s3manager/internal/app/s3manager/mocks"
	"github.com/gorilla/mux"
	"github.com/matryer/is"
	"github.com/minio/minio-go/v7"
)

// doDownloadFolder issues a ZIP download request for the given folder prefix.
func doDownloadFolder(t *testing.T, s3 s3manager.S3, bucket, prefix string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/buckets/"+bucket+"/objects/"+prefix+"zip", nil)
	req = mux.SetURLVars(req, map[string]string{"bucketName": bucket, "objectName": prefix})
	rr := httptest.NewRecorder()
	s3manager.HandleDownloadFolder(s3).ServeHTTP(rr, req)
	return rr
}

// objectChan turns a list of object infos into the channel shape ListObjects returns.
func objectChan(objects ...minio.ObjectInfo) <-chan minio.ObjectInfo {
	ch := make(chan minio.ObjectInfo, len(objects))
	for _, o := range objects {
		ch <- o
	}
	close(ch)
	return ch
}

func TestHandleDownloadFolder(t *testing.T) {
	t.Parallel()

	t.Run("returns not found if the folder contains no objects", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s3 := &mocks.S3Mock{
			ListObjectsFunc: func(context.Context, string, minio.ListObjectsOptions) <-chan minio.ObjectInfo {
				return objectChan()
			},
		}

		rr := doDownloadFolder(t, s3, "BUCKET-NAME", "empty-folder/")

		is.Equal(http.StatusNotFound, rr.Code)
		is.Equal(len(s3.GetObjectCalls()), 0)
	})

	t.Run("ignores folder marker objects", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s3 := &mocks.S3Mock{
			ListObjectsFunc: func(context.Context, string, minio.ListObjectsOptions) <-chan minio.ObjectInfo {
				return objectChan(
					minio.ObjectInfo{Key: "photos/"},
					minio.ObjectInfo{Key: "photos/sub/"},
				)
			},
		}

		rr := doDownloadFolder(t, s3, "BUCKET-NAME", "photos/")

		is.Equal(http.StatusNotFound, rr.Code) // markers alone produce no archive
		is.Equal(len(s3.GetObjectCalls()), 0)  // and are never fetched
	})

	t.Run("lists the folder recursively", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s3 := &mocks.S3Mock{
			ListObjectsFunc: func(context.Context, string, minio.ListObjectsOptions) <-chan minio.ObjectInfo {
				return objectChan()
			},
		}

		doDownloadFolder(t, s3, "BUCKET-NAME", "photos/")

		is.Equal(len(s3.ListObjectsCalls()), 1)
		call := s3.ListObjectsCalls()[0]
		is.Equal(call.BucketName, "BUCKET-NAME")
		is.Equal(call.Opts.Prefix, "photos/")
		is.True(call.Opts.Recursive)
	})

	t.Run("normalizes a prefix without a trailing slash", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s3 := &mocks.S3Mock{
			ListObjectsFunc: func(context.Context, string, minio.ListObjectsOptions) <-chan minio.ObjectInfo {
				return objectChan()
			},
		}

		doDownloadFolder(t, s3, "BUCKET-NAME", "photos")

		is.Equal(s3.ListObjectsCalls()[0].Opts.Prefix, "photos/")
	})

	t.Run("names the archive after the folder", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s3 := &mocks.S3Mock{
			ListObjectsFunc: func(context.Context, string, minio.ListObjectsOptions) <-chan minio.ObjectInfo {
				return objectChan()
			},
		}

		rr := doDownloadFolder(t, s3, "BUCKET-NAME", "photos/holiday/")

		is.Equal(rr.Header().Get("Content-Disposition"),
			`attachment; filename="holiday.zip"; filename*=UTF-8''holiday.zip`)
	})

	t.Run("names the archive after the bucket when downloading its root", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s3 := &mocks.S3Mock{
			ListObjectsFunc: func(context.Context, string, minio.ListObjectsOptions) <-chan minio.ObjectInfo {
				return objectChan()
			},
		}

		rr := doDownloadFolder(t, s3, "BUCKET-NAME", "")

		is.Equal(rr.Header().Get("Content-Disposition"),
			`attachment; filename="BUCKET-NAME.zip"; filename*=UTF-8''BUCKET-NAME.zip`)
		is.Equal(s3.ListObjectsCalls()[0].Opts.Prefix, "")
	})

	t.Run("returns an error if listing fails", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s3 := &mocks.S3Mock{
			ListObjectsFunc: func(context.Context, string, minio.ListObjectsOptions) <-chan minio.ObjectInfo {
				return objectChan(minio.ObjectInfo{Err: errS3})
			},
		}

		rr := doDownloadFolder(t, s3, "BUCKET-NAME", "photos/")

		is.Equal(http.StatusInternalServerError, rr.Code)
	})

	t.Run("returns an error if an object cannot be read", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s3 := &mocks.S3Mock{
			ListObjectsFunc: func(context.Context, string, minio.ListObjectsOptions) <-chan minio.ObjectInfo {
				return objectChan(minio.ObjectInfo{Key: "photos/cat.jpg"})
			},
			GetObjectFunc: func(context.Context, string, string, minio.GetObjectOptions) (*minio.Object, error) {
				return nil, errS3
			},
		}

		rr := doDownloadFolder(t, s3, "BUCKET-NAME", "photos/")

		is.Equal(http.StatusInternalServerError, rr.Code)
	})
}
