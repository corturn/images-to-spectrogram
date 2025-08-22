package wavIO

import (
	// "github.com/go-audio/wav"
	"image"
	"image/png"
	"os"
	"strings"

	"golang.org/x/image/draw"
)

type Image struct {
	Path          string
	Img           *image.Gray
	Bounds        image.Rectangle //The size of the image
	Width, Height int
}

func (img *Image) Load(inPath string, height int) error {
	img.Path = inPath
	inReader, err := os.Open(inPath)

	if err != nil {
		return err
	}
	defer inReader.Close()

	origImg, err := png.Decode(inReader)

	if err != nil {
		return err
	}
	// Scales image to proportional with input height
	width := int(float64(origImg.Bounds().Size().X) * (float64(height) / float64(origImg.Bounds().Size().Y)))
	resized := image.NewGray(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(resized, resized.Bounds(), origImg, origImg.Bounds(), draw.Over, nil)

	img.Img = resized
	img.Bounds = resized.Bounds()
	size := img.Bounds.Size()
	img.Width = size.X
	img.Height = size.Y

	img.Save()

	return nil
}

func (img *Image) IntensityCol(x int) []float64 {
	col := make([]float64, 0, img.Height)
	for y := img.Bounds.Min.Y; y < img.Bounds.Max.Y; y++ {
		// Convert each pixel to grayscale and
		col = append(col, float64(img.Img.GrayAt(x, y).Y)/255.0)
	}
	return col
}

func (img *Image) IntensityRow(y int) []float64 {
	col := make([]float64, 0, img.Height)
	for x := img.Bounds.Min.X; x < img.Bounds.Max.X; x++ {
		// Convert each pixel to grayscale and
		col = append(col, float64(img.Img.GrayAt(x, y).Y)/255.0)
	}
	return col
}

func (img *Image) Save() error {
	outWriter, err := os.Create(strings.TrimSuffix(img.Path, ".png") + "_modified.png")
	if err != nil {
		return err
	}
	defer outWriter.Close()

	err = png.Encode(outWriter, img.Img)
	if err != nil {
		return err
	}
	return nil
}
