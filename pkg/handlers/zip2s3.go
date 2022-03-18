package handlers

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func addFileToS3Bucket(keyfile, region, bucket string) error {
	s, _ := session.NewSession(&aws.Config{Region: aws.String(region)})
	file, err := os.Open(keyfile)

	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String("images.zip"),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})

	return err
}

func createZipFile(file string, target string) error {
	createZipFile, _ := os.Create(target)
	z := zip.NewWriter(createZipFile)
	src, err := os.Open(file)

	if err != nil {
		return err
	}
	info, err := src.Stat()

	if err != nil {
		return err
	}
	hdr, err := zip.FileInfoHeader(info)

	if err != nil {
		return err
	}
	hdr.Name = filepath.Base(file)
	dst, err := z.CreateHeader(hdr)

	if err != nil {
		return err
	}
	_, err = io.Copy(dst, src)

	if err != nil {
		return err
	}

	src.Close()
	return z.Close()
}
