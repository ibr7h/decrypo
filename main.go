package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ajdnik/decrypo/build"
	"github.com/ajdnik/decrypo/decryptor"
	"github.com/ajdnik/decrypo/file"
	"github.com/ajdnik/decrypo/pluralsight"
	"github.com/cheggaaa/pb/v3"
)

func main() {
	defClip, err := pluralsight.GetClipPath()
	if err != nil {
		panic(err)
	}
	defDb, err := pluralsight.GetDbPath()
	if err != nil {
		panic(err)
	}
	defOut := "./Pluralsight Courses/"
	if runtime.GOOS == "windows" {
		abs, err := filepath.Abs(defOut)
		if err != nil {
			panic(err)
		}
		defOut = file.ToUNC(abs)
	}
	clips := flag.String("clips", defClip, "location of clip .psv files")
	db := flag.String("db", defDb, "location of sqlite file")
	output := flag.String("output", defOut, "location of decrypted courses")
	version := flag.Bool("v", false, "print tool version")
	flag.Parse()

	if *version {
		fmt.Println(build.Version)
		os.Exit(0)
	}

	if runtime.GOOS == "windows" {
		abs, err := filepath.Abs(*output)
		if err != nil {
			panic(err)
		}
		*output = file.ToUNC(abs)
	}

	courses := pluralsight.CourseRepository{
		Path: *db,
	}
	clipCount, err := courses.ClipCount()
	if err != nil {
		panic(err)
	}
	svc := decryptor.Service{
		Decoder: &pluralsight.Decoder{},
		Storage: &file.Storage{
			Path:      *output,
			MkdirAll:  os.MkdirAll,
			WriteFile: ioutil.WriteFile,
		},
		CaptionEncoder: &file.SrtEncoder{},
		Clips: &pluralsight.ClipRepository{
			Path:       *clips,
			FileOpen:   os.Open,
			FileExists: file.Exists,
		},
		Courses: &courses,
	}
	fmt.Println("Found", clipCount, "clips in database.")
	fmt.Println("Decrypting clips and extracting captions...")
	bar := pb.StartNew(clipCount)
	successCount := 0
	err = svc.DecryptAll(func(_ decryptor.Clip, _ *string) {
		bar.Increment()
		successCount++
	})
	bar.Finish()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully decrypted", successCount, "of", clipCount, "clips.")
	fmt.Println("You can find them at", *output)
}
