package main

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/manifoldco/promptui"
	"log"
	"strings"
	"time"
)

func init() {
	// Basic S3 info
	userConfig = UserConfig{
		// This need to be modofy
		EndPoint:            "s3-cn-east-1.amazon.com",
		Region:              "cn-east-1",
		BucketName:          "a8uv9rgwj2",
		AccessKeyID:         "67OfTSCSFSsx-onGnTx46t-qTZHUHoVfaq",
		SecretKeyID:         "swYKTptok14veaFQoqKEFXMw3EWq5vr3Wr",
		SignedUrlExpiretion: 24 * 7 * time.Hour,
	}
}

func main() {
	objects, err := ListObjects(1000, "", "")
	if err != nil {
		log.Fatalf("Get objects list failed, %v", err)
	}

	var fileNames []string
	for _, object := range objects {
		fileNames = append(fileNames, *object.Key)
	}

	prompt := promptui.Select{
		Label: "Select which file to download ...",
		Items: fileNames,
		Size:  20,
		// search matched
		Searcher: func(input string, index int) bool {
			item := fileNames[index]
			content := fmt.Sprintf("%s", item)
			if strings.Contains(input, " ") {
				for _, key := range strings.Split(input, " ") {
					key = strings.TrimSpace(key)
					if key != "" {
						if !strings.Contains(content, key) {
							return false
						}
					}
				}
				return true
			}
			if strings.Contains(content, input) {
				return true
			}
			return false
		},
	}

	_, selectedFileName, err := prompt.Run()

	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
		return
	}
	Choose(selectedFileName)
}

func Choose(name string) {
	prompt := promptui.Select{
		Label: "Which operation",
		Items: []string{
			"download it",
			"get signed url",
			//"set object public read",
		},
	}

	_, result, err := prompt.Run()

	if err != nil {
		log.Printf("Prompt failed %v\n", err)
		return
	}

	//fmt.Printf("You choose %q\n", result)
	switch result {
	case "download it":
		fileSize := StatFile(name).ContentLength
		log.Printf("Begin to download %q (%s) ...\n", name, humanize.Bytes(uint64(*fileSize)))
		DownloadObject(name)
	case "get signed url":
		SignS3Url(name)
		//case "set object public read":
		//	SetObjectPublicRead(name)
	}
}
