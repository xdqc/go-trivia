package device

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"sync"

	"github.com/ngaut/log"
)

func SaveImage(png image.Image, cfg *Config, c1 chan<- string, c2 chan<- string) error {
	/* 	go func() {
		screenshotPath := fmt.Sprintf("%sscreenshot.png", util.ImagePath)
		err := util.SavePNG(screenshotPath, png)
		if err != nil {
			log.Errorf("保存截图失败，%v", err)
		}
		log.Debugf("保存完整截图成功，%s", screenshotPath)
	}() */

	//裁剪图片
	questionImg, answerImg, answer1Img, answer2Img, answer3Img, answer4Img, err := splitImage(png, cfg)
	if err != nil {
		return fmt.Errorf("截图失败，%v", err)
	}

	var wg sync.WaitGroup
	wg.Add(6)

	go func() {
		defer wg.Done()
		// pic := thresholdingImage(questionImg)
		err = savePNG(QuestionImage, questionImg)
		if err != nil {
			log.Errorf("保存question截图失败，%v", err)
		}
		c1 <- QuestionImage
		log.Debugf("保存question截图成功")
	}()

	go func() {
		defer wg.Done()
		// pic := thresholdingImage(answerImg)
		err = savePNG(AnswerImage, answerImg)
		if err != nil {
			log.Errorf("保存answer截图失败，%v", err)
		}
		c2 <- AnswerImage
		log.Debugf("保存answer截图成功")
	}()

	go func() {
		defer wg.Done()
		// pic := thresholdingImage(answerImg)
		err = savePNG(Answer1Image, answer1Img)
		if err != nil {
			log.Errorf("保存answer截图失败，%v", err)
		}
		log.Debugf("保存answer截图成功")
	}()

	go func() {
		defer wg.Done()
		// pic := thresholdingImage(answerImg)
		err = savePNG(Answer2Image, answer2Img)
		if err != nil {
			log.Errorf("保存answer截图失败，%v", err)
		}
		log.Debugf("保存answer截图成功")
	}()

	go func() {
		defer wg.Done()
		// pic := thresholdingImage(answerImg)
		err = savePNG(Answer3Image, answer3Img)
		if err != nil {
			log.Errorf("保存answer截图失败，%v", err)
		}
		log.Debugf("保存answer截图成功")
	}()

	go func() {
		defer wg.Done()
		// pic := thresholdingImage(answerImg)
		err = savePNG(Answer4Image, answer4Img)
		if err != nil {
			log.Errorf("保存answer截图失败，%v", err)
		}
		log.Debugf("保存answer截图成功")
	}()

	wg.Wait()
	return nil
}

//SavePNG 保存png图片
func savePNG(filename string, pic image.Image) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, pic)
}

//CutImage 裁剪图片
func cutImage(src image.Image, x, y, w, h int) (image.Image, error) {
	var subImg image.Image

	if rgbImg, ok := src.(*image.YCbCr); ok {
		subImg = rgbImg.SubImage(image.Rect(x, y, x+w, y+h)).(*image.YCbCr) //图片裁剪x0 y0 x1 y1
	} else if rgbImg, ok := src.(*image.RGBA); ok {
		subImg = rgbImg.SubImage(image.Rect(x, y, x+w, y+h)).(*image.RGBA) //图片裁剪x0 y0 x1 y1
	} else if rgbImg, ok := src.(*image.NRGBA); ok {
		subImg = rgbImg.SubImage(image.Rect(x, y, x+w, y+h)).(*image.NRGBA) //图片裁剪x0 y0 x1 y1
	} else {
		return subImg, errors.New("图片解码失败")
	}

	return subImg, nil
}

//裁剪图片
func splitImage(src image.Image, cfg *Config) (questionImg image.Image, answerImg image.Image, answer1Img image.Image, answer2Img image.Image, answer3Img image.Image, answer4Img image.Image, err error) {
	var qX, qY, qW, qH, aX, aY, aW, aH, a1X, a1Y, a1W, a1H, a2X, a2Y, a2W, a2H, a3X, a3Y, a3W, a3H, a4X, a4Y, a4W, a4H int
	switch cfg.APP {
	case "xigua":
		qX, qY, qW, qH = cfg.XgQx, cfg.XgQy, cfg.XgQw, cfg.XgQh
		aX, aY, aW, aH = cfg.XgAx, cfg.XgAy, cfg.XgAw, cfg.XgAh
		a1X, a1Y, a1W, a1H = cfg.NsA1x, cfg.NsA1y, cfg.NsA1w, cfg.NsA1h
		a2X, a2Y, a2W, a2H = cfg.NsA2x, cfg.NsA2y, cfg.NsA2w, cfg.NsA2h
		a3X, a3Y, a3W, a3H = cfg.NsA3x, cfg.NsA3y, cfg.NsA3w, cfg.NsA3h
		a4X, a4Y, a4W, a4H = cfg.NsA4x, cfg.NsA4y, cfg.NsA4w, cfg.NsA4h
	case "cddh":
		qX, qY, qW, qH = cfg.CdQx, cfg.CdQy, cfg.CdQw, cfg.CdQh
		aX, aY, aW, aH = cfg.CdAx, cfg.CdAy, cfg.CdAw, cfg.CdAh
		a1X, a1Y, a1W, a1H = cfg.NsA1x, cfg.NsA1y, cfg.NsA1w, cfg.NsA1h
		a2X, a2Y, a2W, a2H = cfg.NsA2x, cfg.NsA2y, cfg.NsA2w, cfg.NsA2h
		a3X, a3Y, a3W, a3H = cfg.NsA3x, cfg.NsA3y, cfg.NsA3w, cfg.NsA3h
		a4X, a4Y, a4W, a4H = cfg.NsA4x, cfg.NsA4y, cfg.NsA4w, cfg.NsA4h
	case "huajiao":
		qX, qY, qW, qH = cfg.HjQx, cfg.HjQy, cfg.HjQw, cfg.HjQh
		aX, aY, aW, aH = cfg.HjAx, cfg.HjAy, cfg.HjAw, cfg.HjAh
		a1X, a1Y, a1W, a1H = cfg.NsA1x, cfg.NsA1y, cfg.NsA1w, cfg.NsA1h
		a2X, a2Y, a2W, a2H = cfg.NsA2x, cfg.NsA2y, cfg.NsA2w, cfg.NsA2h
		a3X, a3Y, a3W, a3H = cfg.NsA3x, cfg.NsA3y, cfg.NsA3w, cfg.NsA3h
		a4X, a4Y, a4W, a4H = cfg.NsA4x, cfg.NsA4y, cfg.NsA4w, cfg.NsA4h
	case "zscr":
		qX, qY, qW, qH = cfg.ZsQx, cfg.ZsQy, cfg.ZsQw, cfg.ZsQh
		aX, aY, aW, aH = cfg.ZsAx, cfg.ZsAy, cfg.ZsAw, cfg.ZsAh
		a1X, a1Y, a1W, a1H = cfg.NsA1x, cfg.NsA1y, cfg.NsA1w, cfg.NsA1h
		a2X, a2Y, a2W, a2H = cfg.NsA2x, cfg.NsA2y, cfg.NsA2w, cfg.NsA2h
		a3X, a3Y, a3W, a3H = cfg.NsA3x, cfg.NsA3y, cfg.NsA3w, cfg.NsA3h
		a4X, a4Y, a4W, a4H = cfg.NsA4x, cfg.NsA4y, cfg.NsA4w, cfg.NsA4h
	case "nexusq":
		qX, qY, qW, qH = cfg.NsQx, cfg.NsQy, cfg.NsQw, cfg.NsQh
		aX, aY, aW, aH = cfg.NsAx, cfg.NsAy, cfg.NsAw, cfg.NsAh
		a1X, a1Y, a1W, a1H = cfg.NsA1x, cfg.NsA1y, cfg.NsA1w, cfg.NsA1h
		a2X, a2Y, a2W, a2H = cfg.NsA2x, cfg.NsA2y, cfg.NsA2w, cfg.NsA2h
		a3X, a3Y, a3W, a3H = cfg.NsA3x, cfg.NsA3y, cfg.NsA3w, cfg.NsA3h
		a4X, a4Y, a4W, a4H = cfg.NsA4x, cfg.NsA4y, cfg.NsA4w, cfg.NsA4h
	}

	var wg sync.WaitGroup
	wg.Add(6)

	go func() {
		defer wg.Done()
		questionImg, err = cutImage(src, qX, qY, qW, qH)
		if err != nil {
			return
		}
	}()

	go func() {
		defer wg.Done()
		answerImg, err = cutImage(src, aX, aY, aW, aH)
		if err != nil {
			return
		}
	}()

	go func() {
		defer wg.Done()
		answer1Img, err = cutImage(src, a1X, a1Y, a1W, a1H)
		if err != nil {
			return
		}
	}()

	go func() {
		defer wg.Done()
		answer2Img, err = cutImage(src, a2X, a2Y, a2W, a2H)
		if err != nil {
			return
		}
	}()

	go func() {
		defer wg.Done()
		answer3Img, err = cutImage(src, a3X, a3Y, a3W, a3H)
		if err != nil {
			return
		}
	}()

	go func() {
		defer wg.Done()
		answer4Img, err = cutImage(src, a4X, a4Y, a4W, a4H)
		if err != nil {
			return
		}
	}()

	wg.Wait()
	return
}

//二值化图片
func thresholdingImage(img image.Image) image.Image {
	size := img.Bounds()
	pic := image.NewGray(size)
	draw.Draw(pic, size, img, size.Min, draw.Src)

	width := size.Dx()
	height := size.Dy()
	zft := make([]int, 256) //用于保存每个像素的数量，注意这里用了int类型，在某些图像上可能会溢出。
	var idx int
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			idx = i*height + j
			zft[pic.Pix[idx]]++ //image对像有一个Pix属性，它是一个slice，里面保存的是所有像素的数据。
		}
	}

	fz := getOSTUThreshold(zft)
	for i := 0; i < len(pic.Pix); i++ {
		if int(pic.Pix[i]) > fz {
			pic.Pix[i] = 255
		} else {
			pic.Pix[i] = 0
		}
	}
	return pic
}

//getOSTUThreshold OSTU大律法 计算阀值
func getOSTUThreshold(HistGram []int) int {
	var Y, Amount int
	var PixelBack, PixelFore, PixelIntegralBack, PixelIntegralFore, PixelIntegral int
	var OmegaBack, OmegaFore, MicroBack, MicroFore, SigmaB, Sigma float64 // 类间方差;
	var MinValue, MaxValue int
	var Threshold int
	for MinValue = 0; MinValue < 256 && HistGram[MinValue] == 0; MinValue++ {
	}
	for MaxValue = 255; MaxValue > MinValue && HistGram[MinValue] == 0; MaxValue-- {
	}
	if MaxValue == MinValue {
		return MaxValue // 图像中只有一个颜色
	}
	if MinValue+1 == MaxValue {
		return MinValue // 图像中只有二个颜色
	}
	for Y = MinValue; Y <= MaxValue; Y++ {
		Amount += HistGram[Y] //  像素总数
	}
	PixelIntegral = 0
	for Y = MinValue; Y <= MaxValue; Y++ {
		PixelIntegral += HistGram[Y] * Y
	}
	SigmaB = -1
	for Y = MinValue; Y < MaxValue; Y++ {
		PixelBack = PixelBack + HistGram[Y]
		PixelFore = Amount - PixelBack
		OmegaBack = float64(PixelBack) / float64(Amount)
		OmegaFore = float64(PixelFore) / float64(Amount)
		PixelIntegralBack += HistGram[Y] * Y
		PixelIntegralFore = PixelIntegral - PixelIntegralBack
		MicroBack = float64(PixelIntegralBack) / float64(PixelBack)
		MicroFore = float64(PixelIntegralFore) / float64(PixelFore)
		Sigma = OmegaBack * OmegaFore * (MicroBack - MicroFore) * (MicroBack - MicroFore)
		if Sigma > SigmaB {
			SigmaB = Sigma
			Threshold = Y
		}
	}
	return Threshold
}
