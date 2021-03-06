package storage

import (
	"bufio"
	"context"
	"os"
	"sort"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/helloworldpark/tickle-stock-watcher/logger"
	"google.golang.org/api/iterator"
)

var client *storage.Client
var bucket *storage.BucketHandle

// ObjectAttrs alias of storage.ObjectAttrs
type ObjectAttrs storage.ObjectAttrs

// InitStorage inits google storage
func InitStorage() {
	ctx := context.Background()
	clientYet, err := storage.NewClient(ctx)
	if err != nil {
		logger.Panic(err.Error())
	}
	client = clientYet
	bucket = client.Bucket("ticklemeta-storage")
}

// Remove remove dir and files containing from
func Remove(contains string) error {
	ctx := context.Background()
	it := bucket.Objects(ctx, nil)
	var toDelete []string
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}

		if strings.Contains(attrs.Name, contains) {
			toDelete = append(toDelete, attrs.Name)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(toDelete)))
	var err error
	for _, path := range toDelete {
		err = bucket.Object(path).Delete(ctx)
	}
	return err
}

// Clean cleans including 'contains'
func Clean(contains string) error {
	return Remove(contains)
}

// FilesInDirectory returns attributes in dirname
func FilesInDirectory(dirname string) []storage.ObjectAttrs {
	ctx := context.Background()
	it := bucket.Objects(ctx, nil)
	filesAttrs := make([]storage.ObjectAttrs, 0)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}

		dirs := strings.Split(attrs.Name, "/")
		for _, d := range dirs {
			if d == dirname {
				filesAttrs = append(filesAttrs, *attrs)
			}
		}
	}
	return filesAttrs
}

// Write writes []byte to filename in GCS
func Write(contents []byte, filename string) (string, error) {
	ctx := context.Background()
	filePath := "tickle-stock-watcher/" + filename
	writer := bucket.Object(filePath).NewWriter(ctx)
	writer.Write(contents)
	var err error
	if err = writer.Close(); err == nil {
		err = bucket.Object(filePath).ACL().Set(ctx, storage.AllUsers, storage.RoleReader)
	}

	return filePath, err
}

// PNGtoBytes converts png to []byte
// https://www.socketloop.com/tutorials/golang-convert-an-image-file-to-byte
func PNGtoBytes(filename string) []byte {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	fileInfo, _ := file.Stat()
	var size = fileInfo.Size()
	bytes := make([]byte, size)

	buffer := bufio.NewReader(file)
	_, err = buffer.Read(bytes)

	return bytes
}
