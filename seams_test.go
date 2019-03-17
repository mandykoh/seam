package seam

import (
	"image/jpeg"
	"image/png"
	"os"
	"path"
	"testing"
)

func TestRemoveVerticalSeams(t *testing.T) {

	const inFileName = "L1000010.jpg"

	imgFile, err := os.Open(path.Join("test-images", inFileName))
	if err != nil {
		t.Fatalf("Error opening test image: %v", err)
	}
	defer imgFile.Close()

	img, err := jpeg.Decode(imgFile)
	if err != nil {
		t.Fatalf("Error decoding test image: %v", err)
	}

	img = RemoveVerticalSeams(img, 512)

	outFile, err := os.Create(path.Join("test-images", inFileName+"-output.png"))
	if err != nil {
		t.Fatalf("Error writing output image: %v", err)
	}
	defer outFile.Close()

	err = png.Encode(outFile, img)
	if err != nil {
		t.Fatalf("Error encoding output image: %v", err)
	}
}
