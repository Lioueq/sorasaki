package main

import (
	"image"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const (
	FOLDER     = "hina"
	dstW, dstH = 1920, 1080
)

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))

	paths, err := listImages(FOLDER)
	if err != nil {
		panic(err)
	}
	if len(paths) == 0 {
		panic("no images found")
	}

	rects := splitCanvas(image.Rect(0, 0, dstW, dstH), len(paths))
	rand.Shuffle(len(paths), func(i, j int) {
		paths[i], paths[j] = paths[j], paths[i]
	})

	var wg sync.WaitGroup
	errCh := make(chan error, len(paths))

	for i, path := range paths {
		wg.Add(1)
		go func(path string, r image.Rectangle) {
			defer wg.Done()

			file, err := os.Open(path)
			if err != nil {
				errCh <- err
				return
			}

			src, err := decodeImage(file)
			file.Close()
			if err != nil {
				errCh <- err
				return
			}

			xdraw.CatmullRom.Scale(dst, r, src, src.Bounds(), draw.Over, nil)
		}(path, rects[i])
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			panic(err)
		}
	}

	outFile, err := os.Create("output.png")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, dst); err != nil {
		panic(err)
	}
}

func listImages(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := strings.ToLower(entry.Name())
		if strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".jpeg") || strings.HasSuffix(name, ".webp") {
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}

	sort.Strings(paths)
	return paths, nil
}

func decodeImage(file *os.File) (image.Image, error) {
	img, _, err := image.Decode(file)
	return img, err
}

func splitCanvas(canvas image.Rectangle, n int) []image.Rectangle {
	if n <= 1 {
		return []image.Rectangle{canvas}
	}

	rects := []image.Rectangle{canvas}

	for len(rects) < n {
		idx := pickLargest(rects)
		current := rects[idx]
		rects = append(rects[:idx], rects[idx+1:]...)

		left, right, ok := splitRect(current)
		if !ok {
			rects = append(rects, current)
			break
		}

		rects = append(rects, left, right)
	}

	return rects
}

func pickLargest(rects []image.Rectangle) int {
	best := 0
	bestArea := area(rects[0])

	for i := 1; i < len(rects); i++ {
		a := area(rects[i])
		if a > bestArea {
			best = i
			bestArea = a
		}
	}

	return best
}

func area(r image.Rectangle) int {
	return r.Dx() * r.Dy()
}

func splitRect(r image.Rectangle) (image.Rectangle, image.Rectangle, bool) {
	w := r.Dx()
	h := r.Dy()

	if w < 2 && h < 2 {
		return image.Rectangle{}, image.Rectangle{}, false
	}

	if w >= h && w >= 2 {
		x := r.Min.X + 1 + rand.Intn(w-1)
		return image.Rect(r.Min.X, r.Min.Y, x, r.Max.Y), image.Rect(x, r.Min.Y, r.Max.X, r.Max.Y), true
	}

	if h >= 2 {
		y := r.Min.Y + 1 + rand.Intn(h-1)
		return image.Rect(r.Min.X, r.Min.Y, r.Max.X, y), image.Rect(r.Min.X, y, r.Max.X, r.Max.Y), true
	}

	return image.Rectangle{}, image.Rectangle{}, false
}
