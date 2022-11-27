package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/robfig/cron/v3"
	"go-jike-sdk/jike"
	"log"
)

var (
	areaCode string
	phone    string
	password string
)

func init() {
	flag.StringVar(&areaCode, "areaCode", "+86", "areaCode for Jike.")
	flag.StringVar(&phone, "phone", "17820160690", "phone for Jike.")
	flag.StringVar(&password, "password", "981475", "password for Jike.")
	flag.Parse()
}

func main() {
	c := cron.New(cron.WithSeconds())
	_, err := c.AddFunc("0 0 */1 * * ?", Wallpapers)
	if err != nil {
		log.Fatal(err)
	}
	c.Start()
	Wallpapers()
	select {}
}

func Wallpapers() {
	ctx, client, err := Init()
	if err != nil {
		return
	}
	wallpapers, err := jike.GetWallpapers()
	if err != nil {
		return
	}
	pictureKeys := jike.UploadWallpapersParallel(wallpapers)
	// try again
	if len(pictureKeys) == 0 {
		log.Println("upload wallpapers parallel failed")
		log.Println("start to try to upload serial...")
		pictureKeys = jike.UploadWallpapersSerial(wallpapers)
	}
	if len(pictureKeys) == 0 {
		log.Println("all ways upload failed, will not try")
		return
	}
	_, err = client.UserService.Create(ctx, pictureKeys)
	if err != nil {
		log.Println("create fail, because of: ", err)
	} else {
		log.Println("create success")
	}
}

func Init() (context.Context, *jike.Jike, error) {
	content := context.Background()
	if phone == "" || password == "" {
		flag.PrintDefaults()
		return content, nil, nil
	}

	client := jike.NewJike(areaCode, phone)

	loginOutput, err := client.UserService.PasswordLogin(content, areaCode, phone, password)
	if err != nil {
		log.Println(err)
		return content, nil, err
	}
	fmt.Printf("Username: %s \nScreenName: %s\n", loginOutput.User.Username, loginOutput.User.ScreenName)
	return content, client, nil
}
