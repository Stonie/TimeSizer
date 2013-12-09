package main

import (
	"bufio"
	"flag"
	//"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var newSize int
var wg sync.WaitGroup

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU()) // use all processors - maybe a bit heavy handed.

	// parse args
	flag.Parse()
	var err error
	newSize, err = strconv.Atoi(flag.Arg(0)) // the new size for the longest side
	if err != nil {
		log.Fatal(err) // exit if size not specified
	}

	// look for jpegs using the current working directory as the start point 
	// traverse all folders downward! For eva! 
	var workingDir, _ = os.Getwd()
	err = filepath.Walk(workingDir, traversed)

	wg.Wait() // wait for all go routines to complete

	log.Println("Processing done...")
}

///
/// Called every time a file or dir is walked.
/// starts a resize go routine for each jpeg found
///
func traversed(path string, f os.FileInfo, err error) error {

	if !f.IsDir() && strings.HasSuffix(strings.ToLower(f.Name()), ".jpg") {
		log.Println("Found Jpeg: %s\n", path)

		// resize

		//go goWorker(resizePreserveModTime(imageName, newSize, resize.NearestNeighbor), &wg)
		//go goWorker(resizePreserveModTime(imageName, newSize, resize.Bilinear), &wg)
		//go goWorker(resizePreserveModTime(imageName, newSize, resize.Bicubic), &wg)
		//go goWorker(resizePreserveModTime(imageName, newSize, resize.MitchellNetravali), &wg)
		//go goWorker(resizePreserveModTime(imageName, newSize, resize.Lanczos2Lut), &wg)
		//go goWorker(resizePreserveModTime(imageName, newSize, resize.Lanczos2), &wg)
		go goWorker(resizePreserveModTime(path, newSize, resize.Lanczos3Lut), &wg)
		//go goWorker(resizePreserveModTime(imageName, newSize, resize.Lanczos3), &wg)
	}
	return nil
}

///
/// Helper to track goroutines
///
func goWorker(f func() int, waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)
	defer waitGroup.Done()
	f() // call the passed function
}

///
/// Resize while preserving original timestamps
///
func resizePreserveModTime(imageName string, newSize int, interpolation resize.InterpolationFunction) func() int {

	n := 0
	return func() int {
		// load file
		fileInfo, err := os.Stat(imageName)
		if err != nil {
			log.Fatal(err)
		}

		// grab the Original mod time
		OriginalModificationTime := fileInfo.ModTime()

		// record it's modification time & create time
		log.Println("existing time-stamp: " + OriginalModificationTime.String())

		resizeImage(imageName, newSize, interpolation)

		// write the old time stamps back
		err = os.Chtimes(imageName, OriginalModificationTime, OriginalModificationTime)
		if err != nil {
			log.Fatal(err)
		}

		return n
	}
}

///
/// Resize and write image
///
func resizeImage(jpegFileName string, newSize int, interpolation resize.InterpolationFunction) {

	imageConfig, _, err := decodeConfig(jpegFileName)
	if err == nil {
		log.Println("Image details: %s width: %d height: %d\n", jpegFileName, imageConfig.Width, imageConfig.Height)
	} else {
		log.Println("Error: %v\n", err)
	}

	// decode jpeg into image.Image
	img, _, err := decode(jpegFileName)
	if err != nil {
		log.Fatal(err)
	}

	// resize based on the long side.
	var resizedImage image.Image

	log.Println("using: ", GetFunctionName(interpolation))

	if imageConfig.Width > imageConfig.Height {
		// landscape
		resizedImage = resize.Resize(uint(newSize), 0, img, interpolation)

	} else {
		// portrait or square
		resizedImage = resize.Resize(0, uint(newSize), img, interpolation)
	}

	// write the resized image out to disk replacing the existing image
	encodeImage(resizedImage, jpegFileName) //strings.Replace(GetFunctionName(interpolation), "/", ".", -1)+
}

///
/// Decode image file
///
func decode(filename string) (image.Image, string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	return image.Decode(bufio.NewReader(f))
}

///
/// Decode image information
///
func decodeConfig(filename string) (image.Config, string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return image.Config{}, "", err
	}

	defer f.Close()
	return image.DecodeConfig(bufio.NewReader(f))
}

///
/// Encode image file
///
func encodeImage(image image.Image, filename string) {
	// create file stub
	out, err := os.Create(filename)
	defer out.Close()
	if err != nil {
		log.Fatal(err)
	}

	// write new image to file
	jpeg.Encode(out, image, nil)
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
