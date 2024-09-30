package main

import (
	"fmt"
	"golang.org/x/image/draw"
	"image"
	"image/color"
	"image/png"
	"log"
	"net/http"
	"os"
)

// Desaturate converts an RGB image to a desaturated grayscale image
func Desaturate(img image.Image) *image.Gray {
	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			maxVal := uint32(max(max(r, g), b)) >> 8
			minVal := uint32(min(min(r, g), b)) >> 8
			gray := uint8((maxVal + minVal) / 2)
			grayImg.Set(x, y, color.Gray{Y: gray})
		}
	}
	return grayImg
}

// AdjustLightness adjusts the lightness of a grayscale image
func AdjustLightness(img *image.Gray, ratio float64) *image.Gray {
	bounds := img.Bounds()
	adjusted := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := img.GrayAt(x, y).Y
			var newGray uint8
			if ratio > 0 {
				newGray = uint8(float64(gray)*(1-ratio) + 255*ratio)
			} else {
				newGray = uint8(float64(gray) * (1 + ratio))
			}
			adjusted.Set(x, y, color.Gray{Y: newGray})
		}
	}
	return adjusted
}

// Invert inverts the color of the grayscale image
func Invert(img *image.Gray) *image.Gray {
	bounds := img.Bounds()
	inverted := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := img.GrayAt(x, y).Y
			inverted.Set(x, y, color.Gray{Y: 255 - gray})
		}
	}
	return inverted
}

// LinearDodgeBlend blends two grayscale images
func LinearDodgeBlend(imgX, imgY *image.Gray) *image.Gray {
	bounds := imgX.Bounds()
	result := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			grayX := imgX.GrayAt(x, y).Y
			grayY := imgY.GrayAt(x, y).Y
			newGray := uint8(clamp(int(grayX)+int(grayY), 0, 255))
			result.Set(x, y, color.Gray{Y: newGray})
		}
	}
	return result
}

// DivideBlend blends two grayscale images in 'divide' mode
func DivideBlend(imgX, imgY *image.Gray) *image.Gray {
	bounds := imgX.Bounds()
	result := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			grayX := imgX.GrayAt(x, y).Y
			grayY := imgY.GrayAt(x, y).Y
			var newGray uint8
			if grayX == 0 {
				newGray = 255
			} else {
				newGray = uint8(clamp(int(grayY)*255/int(grayX), 0, 255))
			}
			result.Set(x, y, color.Gray{Y: newGray})
		}
	}
	return result
}

// AddMask adds an alpha channel to the grayscale image
func AddMask(imgX, imgY *image.Gray) *image.NRGBA {
	bounds := imgX.Bounds()
	result := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := imgX.GrayAt(x, y).Y
			alpha := imgY.GrayAt(x, y).Y
			result.Set(x, y, color.NRGBA{R: gray, G: gray, B: gray, A: alpha})
		}
	}
	return result
}

// Build creates the 'mirage tank' image
func Build(sourceX, sourceY, targetName string, shrink float64) {
	fmt.Println("Start processing")
	imgAFile, err := os.Open(sourceX)
	if err != nil {
		log.Fatal(err)
	}
	defer imgAFile.Close()

	imgBFile, err := os.Open(sourceY)
	if err != nil {
		log.Fatal(err)
	}
	defer imgBFile.Close()

	imgA, _, err := image.Decode(imgAFile)
	if err != nil {
		log.Fatal(err)
	}

	imgB, _, err := image.Decode(imgBFile)
	if err != nil {
		log.Fatal(err)
	}

	width := int(float64(imgA.Bounds().Max.X) * shrink)
	height := int(float64(imgA.Bounds().Max.Y) * shrink)

	imgA = resize(imgA, width, height)
	imgB = resize(imgB, width, height)

	// 类型转换
	grayImgA := Desaturate(imgA)
	grayImgB := Desaturate(imgB)

	imgA = Invert(AdjustLightness(grayImgA, 0.5))
	imgB = AdjustLightness(grayImgB, -0.5)

	// 将灰度图像转换为*image.Gray
	linearDodge := LinearDodgeBlend(imgA.(*image.Gray), imgB.(*image.Gray))
	divided := DivideBlend(linearDodge, imgB.(*image.Gray))

	finalImage := AddMask(divided, linearDodge)

	outputFile, err := os.Create(targetName)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	if err := png.Encode(outputFile, finalImage); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Finished")
}

// Resize resizes the image to the specified width and height.
func resize(img image.Image, width, height int) image.Image {
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(newImg, newImg.Bounds(), img, img.Bounds(), draw.Over, nil)
	return newImg
}

// Helper functions
func max(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func min(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	} else if value > max {
		return max
	}
	return value
}

// Main function
func main() {

	Build("cmd20-mirage-tank-images/1724382048281.png",
		"cmd20-mirage-tank-images/1726296462076.png",
		"cmd20-mirage-tank-images/target_image.png", 1)
}
func main1() {
	//println(time.Now().Add(time.Hour * 120).Unix())
	//return

	// 设置静态文件目录
	staticDir := "/Users/bytedance/GolandProjects/awesomeProject/cmd20-mirage-tank-images" // 替换为你的静态文件目录

	// 创建一个新的 HTTP 处理器
	http.Handle("/", http.FileServer(http.Dir(staticDir)))

	// 监听端口 8080
	fmt.Println("服务器已启动，访问 http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("启动服务器失败:", err)
	}
}
