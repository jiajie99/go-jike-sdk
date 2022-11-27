package jike

import (
	"fmt"
	"strings"
	"testing"
)

func TestUploadWallpapers(t *testing.T) {
	wallpapers, _ := GetWallpapers()
	UploadWallpapersParallel(wallpapers)
}

func TestGetWallpaperName(t *testing.T) {
	str := "https://w.wallhaven.cc/full/wy/wallhaven-wyl93r.png"
	split := strings.Split(str, "/")
	fmt.Println(split)
}
