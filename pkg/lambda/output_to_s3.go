package lambda

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type csvFormat struct {
	imageName   string
	imageECRURL string
	imageTag    string
}

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

func buildCSVFile(imageName string, options syncOptions, env environmentVars) (csvContent []csvFormat, err error) {
	for _, tag := range options.tags {
		csvContent = append(csvContent, csvFormat{
			imageName:   imageName,
			imageECRURL: env.awsAccount + `.dkr.ecr.` + env.awsRegion + `.amazonaws.com/` + options.ecrRepoPrefix + `/` + options.ecrImageName,
			imageTag:    tag,
		})
	}

	return csvContent, err
}

func writeCSVFile(csvContent *[]csvFormat, csvFileName string) (err error) {
	if csvContent != nil {
		file, err := os.Create(csvFileName)
		if err != nil {
			fmt.Println("Error creating file")
		}
		defer file.Close()
		writer := csv.NewWriter(file)

		for _, value := range *csvContent {
			data := []string{value.imageName, value.imageECRURL, value.imageTag}

			if err := writer.Write(data); err != nil {
				fmt.Println("Error write file")
			}
		}
		writer.Flush()

		if err != nil {
			fmt.Println(err)
		}
	}

	return err
}

func outputToS3Bucket(csvContent []csvFormat, csvFileName, zipFileName, region, bucket string) (err error) {

	if err := writeCSVFile(&csvContent, csvFileName); err != nil {
		return err
	}
	err = createZipFile(csvFileName, zipFileName)
	if err != nil {
		fmt.Println(err)
	}
	err = addFileToS3Bucket(zipFileName, region, bucket)
	if err != nil {
		fmt.Println(err)
	}
	defer os.Remove(csvFileName)
	defer os.Remove(zipFileName)
	return err
}
