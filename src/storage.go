package calibreApi

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/studio-b12/gowebdav"
	"io"
	"os"
	"path"
	"time"
)

type FileClient interface {
	Stat(path string) (os.FileInfo, error)
	ReadStream(path string) (io.ReadCloser, error)
}

type LocalFileClient struct {
	dir string
}

func (l LocalFileClient) Stat(p string) (os.FileInfo, error) {
	return os.Stat(path.Join(l.dir, p))
}

func (l LocalFileClient) ReadStream(p string) (io.ReadCloser, error) {
	file, err := os.ReadFile(path.Join(l.dir, p))
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(file)), nil

}

type MinioFileClient struct {
	config      Minio
	minioClient *minio.Client
	ctx         context.Context
}

func (m MinioFileClient) Stat(p string) (os.FileInfo, error) {
	stat, err := m.minioClient.StatObject(m.ctx, m.config.BucketName, path.Join(m.config.Path, p), minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return &FileInfo{stat: stat}, nil
}

func (m MinioFileClient) ReadStream(p string) (io.ReadCloser, error) {
	object, err := m.minioClient.GetObject(m.ctx, m.config.BucketName, path.Join(m.config.Path, p), minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return object, err
}

func NewWebDavClient(webdav Webdav) FileClient {
	return gowebdav.NewClient(webdav.Host, webdav.User, webdav.Password)
}

func NewLocalClient(local Local) FileClient {
	return &LocalFileClient{
		dir: local.Path,
	}
}

func NewMinioClient(config Minio, ctx context.Context) (FileClient, error) {
	// Initialize minio client object.
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
	})
	return &MinioFileClient{
		config:      config,
		minioClient: minioClient,
		ctx:         ctx,
	}, err
}

type FileInfo struct {
	stat minio.ObjectInfo
}

// Name Implements os.FileInfo
func (s *FileInfo) Name() string {
	return path.Base(s.stat.Key)
}
func (s *FileInfo) Mode() os.FileMode {
	return os.FileMode(int(0444))
}
func (s *FileInfo) ModTime() time.Time {

	return s.stat.LastModified
}
func (s *FileInfo) IsDir() bool {
	return false
}
func (s *FileInfo) Sys() interface{} { return s.stat }
func (s *FileInfo) Size() int64 {
	return s.stat.Size
}
