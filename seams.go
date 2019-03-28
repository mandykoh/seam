package seam

import (
	"image"
	"image/draw"
	"math"
)

func energy(img *image.RGBA, x, y int) float32 {
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

func luminance(img *image.RGBA, x, y int) float32 {
	c := img.RGBAAt(x, y)
	return 0.2126*float32(c.R) + 0.7152*float32(c.G) + 0.0722*float32(c.B)
}

func RemoveVerticalSeams(img image.Image, seamsToRemove int) image.Image {
	imgBounds := img.Bounds()

	resultImg := image.NewRGBA(image.Rect(0, 0, imgBounds.Dx(), imgBounds.Dy()))
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
			energies[offset] = energy(resultImg, j, i)
		}
	}

	// Calculate accumulated energies
	for j := 0; j < resultBounds.Dx(); j++ {
		accumulatedEnergies[j] = energies[j]
	}
	for i := 1; i < resultBounds.Dy(); i++ {
		for j := 0; j < resultBounds.Dx(); j++ {
			offset := i*energyWidth + j
			northOffset := offset - energyWidth

			minE := accumulatedEnergies[northOffset]
			if j > 0 && accumulatedEnergies[northOffset-1] < minE {
				minE = accumulatedEnergies[northOffset-1]
			}
			if j < resultBounds.Dx()-1 && accumulatedEnergies[northOffset+1] < minE {
				minE = accumulatedEnergies[northOffset+1]
			}

			accumulatedEnergies[offset] = energies[offset] + minE
		}
	}

	for seamCount := 0; seamCount < seamsToRemove; seamCount++ {

		// Find beginning of optimal seam
		rowOffset := (resultBounds.Dy() - 1) * energyWidth
		seamMinE := accumulatedEnergies[rowOffset]
		seamX := 0
		for j := 1; j < resultBounds.Dx(); j++ {
			energy := accumulatedEnergies[rowOffset+j]
			if energy < seamMinE {
				seamMinE = energy
				seamX = j
			}
		}
		seamPositions[resultBounds.Dy()-1] = seamX

		// Trace seam upwards
		for i := resultBounds.Dy() - 2; i >= 0; i-- {
			prevRow := i*energyWidth + seamX

			minE := accumulatedEnergies[prevRow]
			if seamX > 0 && accumulatedEnergies[prevRow-1] < minE {
				minE = accumulatedEnergies[prevRow-1]
				seamX--
			}
			if seamX < resultBounds.Dx()-1 && accumulatedEnergies[prevRow+1] < minE {
				minE = accumulatedEnergies[prevRow+1]
				seamX++
			}

			seamPositions[i] = seamX
		}

		resultBounds.Max.X--

		// Shift row segments over the seam
		for i := 0; i < imgBounds.Dy(); i++ {
			for j := seamPositions[i]; j < resultBounds.Max.X; j++ {
				resultImg.SetRGBA(j, i, resultImg.RGBAAt(imgBounds.Min.X+j+1, imgBounds.Min.Y+i))
			}

			rowOffset := i * energyWidth
			offset := rowOffset + seamPositions[i]
			copy(energies[offset:rowOffset+energyWidth], energies[offset+1:rowOffset+energyWidth])
		}

		// Update energies along seam
		for i := 0; i < resultBounds.Dy(); i++ {
			j := seamPositions[i]

			if j < resultBounds.Dx() {
				energies[i*energyWidth+j] = energy(resultImg, j, i)
			}
			if j > 0 {
				energies[i*energyWidth+j-1] = energy(resultImg, j-1, i)
			}
		}

		// Update the accumulated energies propagating from the seam
		for j := seamPositions[0]; j < resultBounds.Dx(); j++ {
			accumulatedEnergies[j] = energies[j]
		}
		for i := 1; i < resultBounds.Dy(); i++ {
			lowBound := seamPositions[0] - i
			if lowBound < 0 {
				lowBound = 0
			}
			highBound := seamPositions[0] + i
			if highBound > resultBounds.Dx()-1 {
				highBound = resultBounds.Dx() - 1
			}

			for offset, j := i*energyWidth+lowBound, lowBound; j <= highBound; offset, j = offset+1, j+1 {
				northOffset := offset - energyWidth

				minE := accumulatedEnergies[northOffset]
				if j > 0 && accumulatedEnergies[northOffset-1] < minE {
					minE = accumulatedEnergies[northOffset-1]
				}
				if j < resultBounds.Dx()-1 && accumulatedEnergies[northOffset+1] < minE {
					minE = accumulatedEnergies[northOffset+1]
				}

				accumulatedEnergies[offset] = energies[offset] + minE
			}
		}
	}

	return resultImg.SubImage(resultBounds)
}
