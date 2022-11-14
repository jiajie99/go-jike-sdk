package main

import (
	"context"
	"flag"
	"fmt"
	"go-jike-sdk/jike"
	"log"

	"github.com/robfig/cron/v3"
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
	_, err := c.AddFunc("0 0 */8 * * ?", Wallpapers)
	if err != nil {
		log.Fatal(err)
	}
	c.Start()
	select {}
}

func Wallpapers() {
	ctx, client := Init()

	wallpapers := jike.GetWallpapers()
	pictureKeys := jike.UploadWallpapers(wallpapers)

	_, err := client.UserService.Create(ctx, pictureKeys)
	if err != nil {
		log.Fatalln(err)
	}
}

func Init() (context.Context, *jike.Jike) {
	content := context.Background()
	if phone == "" || password == "" {
		flag.PrintDefaults()
		return content, nil
	}

	client := jike.NewJike(areaCode, phone)

	loginOutput, err := client.UserService.PasswordLogin(content, areaCode, phone, password)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Username: %s \nScreenName: %s\n", loginOutput.User.Username, loginOutput.User.ScreenName)
	return content, client
}
