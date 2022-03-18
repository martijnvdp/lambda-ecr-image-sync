package handlers

import (
	"encoding/csv"
	"fmt"
	"os"
)

type csvFormat struct {
	imageName   string
	imageECRURL string
	imageTag    string
}

func buildCSVFile(imageName, ecrImageName, ecrRepoPrefix string, resultsFromPublicRepo *[]string, awsAccount, awsRegion string) (csvContent []csvFormat, err error) {
	for _, tag := range *resultsFromPublicRepo {
		csvContent = append(csvContent, csvFormat{
			imageName:   imageName,
			imageECRURL: awsAccount + `.dkr.ecr.` + awsRegion + `.amazonaws.com/` + ecrRepoPrefix + `/` + ecrImageName,
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
