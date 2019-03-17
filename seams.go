package seam

import (
	"image"
	"image/draw"
	"math"
)

func energy(img image.Image, x, y int) float32 {
	neighbours := [8]float32{
		luminance(img, x-1, y-1),
		luminance(img, x, y-1),
		luminance(img, x+1, y-1),
		luminance(img, x-1, y),
		luminance(img, x+1, y),
		luminance(img, x-1, y+1),
		luminance(img, x, y+1),
		luminance(img, x+1, y+1),
	}

	eX := neighbours[0] + neighbours[3] + neighbours[5] - neighbours[2] - neighbours[4] - neighbours[7]
	eY := neighbours[0] + neighbours[1] + neighbours[2] - neighbours[5] - neighbours[6] - neighbours[7]

	return float32(math.Abs(float64(eX)) + math.Abs(float64(eY)))
}

func luminance(img image.Image, x, y int) float32 {
	r, g, b, _ := img.At(x, y).RGBA()
	return 0.2126*float32(r)/0xffff + 0.7152*float32(g)/0xffff + 0.0722*float32(b)/0xffff
}

func RemoveVerticalSeams(img image.Image, seamsToRemove int) image.Image {
	imgBounds := img.Bounds()

	resultImg := image.NewNRGBA(image.Rect(0, 0, imgBounds.Dx(), imgBounds.Dy()))
	resultBounds := resultImg.Bounds()
	draw.Draw(resultImg, resultBounds, img, image.Pt(0, 0), draw.Src)

	energyWidth := imgBounds.Dx()
	energies := make([]float32, imgBounds.Dx()*imgBounds.Dy(), imgBounds.Dx()*imgBounds.Dy())
	accumulatedEnergies := make([]float32, len(energies), len(energies))
	seamPositions := make([]int, imgBounds.Dy(), imgBounds.Dy())

	// Calculate initial energy map
	for i := imgBounds.Min.Y; i < imgBounds.Max.Y; i++ {
		for j := imgBounds.Min.X; j < imgBounds.Max.X; j++ {
			offset := (i-imgBounds.Min.Y)*energyWidth + (j - imgBounds.Min.X)
			energies[offset] = energy(img, j, i)
		}
	}

	for seamCount := 0; seamCount < seamsToRemove; seamCount++ {
		resultBounds.Max.X--

		// Calculate accumulated energies
		for j := 0; j < imgBounds.Dx(); j++ {
			accumulatedEnergies[j] = energies[j]
		}
		for i := 1; i < imgBounds.Dy(); i++ {
			for j := 0; j < imgBounds.Dx(); j++ {
				offset := i*energyWidth + j
				northOffset := offset - energyWidth

				minE := accumulatedEnergies[northOffset]
				if j > 0 && accumulatedEnergies[northOffset-1] < minE {
					minE = accumulatedEnergies[northOffset-1]
				}
				if j < imgBounds.Dx()-1 && accumulatedEnergies[northOffset+1] < minE {
					minE = accumulatedEnergies[northOffset+1]
				}

				accumulatedEnergies[offset] = energies[offset] + minE
			}
		}

		// Find beginning of optimal seam
		rowOffset := (imgBounds.Dy() - 1) * energyWidth
		seamMinE := accumulatedEnergies[rowOffset]
		seamX := 0
		for j := 1; j < imgBounds.Dx(); j++ {
			energy := accumulatedEnergies[rowOffset+j]
			if energy < seamMinE {
				seamMinE = energy
				seamX = j
			}
		}
		seamPositions[imgBounds.Dy()-1] = seamX

		// Trace seam upwards
		for i := imgBounds.Dy() - 2; i >= 0; i-- {
			prevRow := i*energyWidth + seamX

			minE := accumulatedEnergies[prevRow]
			if seamX > 0 && accumulatedEnergies[prevRow-1] < minE {
				minE = accumulatedEnergies[prevRow-1]
				seamX--
			}
			if seamX < imgBounds.Dx()-1 && accumulatedEnergies[prevRow+1] < minE {
				minE = accumulatedEnergies[prevRow+1]
				seamX++
			}

			seamPositions[i] = seamX
		}

		// Shift row segments over the seam
		for i := 0; i < imgBounds.Dy(); i++ {
			for j := seamPositions[i]; j < resultBounds.Max.X; j++ {
				resultImg.Set(j, i, img.At(imgBounds.Min.X+j+1, imgBounds.Min.Y+i))
				energies[i*energyWidth+j] = energies[i*energyWidth+j+1]
			}
		}

		img = resultImg
		imgBounds = img.Bounds()

		// Update energies along seam
		for i := 0; i < imgBounds.Dy(); i++ {
			j := seamPositions[i]

			if j < imgBounds.Dx() {
				energies[i*energyWidth+j] = energy(img, j, i)
			}
			if j > 0 {
				energies[i*energyWidth+j-1] = energy(img, j-1, i)
			}
		}
	}

	croppedResult := image.NewNRGBA(image.Rect(0, 0, resultBounds.Dx(), resultBounds.Dy()))
	draw.Draw(croppedResult, croppedResult.Bounds(), resultImg, image.Pt(0, 0), draw.Src)

	return croppedResult
}
