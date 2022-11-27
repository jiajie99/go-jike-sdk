package jike

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

func GetWallpapers() (map[string]string, error) {
	result := make(map[string]string, 0)
	// Request the HTML page.
	res, err := http.Get("https://wallhaven.cc/random")
	if err != nil {
		log.Println("get wallpaper website error, because of: ", err)
		return result, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Printf("status code error: %d %s\n", res.StatusCode, res.Status)
		return result, err
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Printf("load html document fail, because of: %s\n", err)
		return result, err
	}

	for i := 1; i <= 9; i++ {
		url, ok := doc.Find("#thumbs > section > ul > li:nth-child(" + strconv.Itoa(i) + ") > figure > a").Attr("href")
		if !ok {
			continue
		}
		res, err = http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		if res.StatusCode != 200 {
			log.Printf("status code error: %d %s\n", res.StatusCode, res.Status)
			return result, err
		}

		tempDoc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			log.Printf("load html document fail, because of: %s\n", err)
			return result, err
		}
		url, ok = tempDoc.Find("#wallpaper").Attr("src")
		if ok {
			key := strings.Split(url, "/")
			result[key[len(key)-1]] = url
		}
	}
	return result, nil
}

type GetUploadTokenResp struct {
	UpToken string `json:"uptoken"`
}

func UploadWallpapersParallel(wallpapers map[string]string) []string {
	pictureKeys := make([]string, 0)
	var res *http.Response
	var err error
	var wg sync.WaitGroup
	for key, val := range wallpapers {
		res, err = http.Get("https://upload.ruguoapp.com/1.0/misc/qiniu_uptoken")
		if err != nil {
			log.Fatal(err)
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		getUploadTokenResp := GetUploadTokenResp{}
		err = json.Unmarshal(body, &getUploadTokenResp)
		if err != nil {
			log.Fatal(err)
		}

		wg.Add(1)
		go func(key string, val string) {

			defer wg.Done()
			uploadFileResp, err := UploadFile(UploadFileInfo{
				Token:       getUploadTokenResp.UpToken,
				FName:       key,
				OriginalUrl: val,
			})
			if err != nil {
				log.Println("upload file error")
				log.Println(err)
			}
			if uploadFileResp != nil && uploadFileResp.Success {
				pictureKeys = append(pictureKeys, uploadFileResp.Key)
			}
		}(key, val)
	}
	wg.Wait()
	if res != nil {
		res.Body.Close()
	}
	return pictureKeys
}

func UploadWallpapersSerial(wallpapers map[string]string) []string {
	pictureKeys := make([]string, 0)
	var res *http.Response
	var err error
	for key, val := range wallpapers {
		res, err = http.Get("https://upload.ruguoapp.com/1.0/misc/qiniu_uptoken")
		if err != nil {
			log.Fatal(err)
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		getUploadTokenResp := GetUploadTokenResp{}
		err = json.Unmarshal(body, &getUploadTokenResp)
		if err != nil {
			log.Fatal(err)
		}

		uploadFileResp, err := UploadFile(UploadFileInfo{
			Token:       getUploadTokenResp.UpToken,
			FName:       key,
			OriginalUrl: val,
		})
		if err != nil {
			log.Println("upload file error")
			log.Println(err)
		}
		if uploadFileResp != nil && uploadFileResp.Success {
			pictureKeys = append(pictureKeys, uploadFileResp.Key)
		}
	}
	if res != nil {
		res.Body.Close()
	}
	return pictureKeys
}
