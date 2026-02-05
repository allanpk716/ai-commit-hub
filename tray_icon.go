package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"sort"
	"sync"
)

func looksLikeICO(b []byte) bool {
	if len(b) < 4 {
		return false
	}
	reserved := binary.LittleEndian.Uint16(b[0:2])
	typ := binary.LittleEndian.Uint16(b[2:4])
	return reserved == 0 && typ == 1
}

func looksLikePNG(b []byte) bool {
	if len(b) < 8 {
		return false
	}
	return b[0] == 0x89 && b[1] == 0x50 && b[2] == 0x4e && b[3] == 0x47 && b[4] == 0x0d && b[5] == 0x0a && b[6] == 0x1a && b[7] == 0x0a
}

func icoSquareSizes(b []byte) ([]int, error) {
	if len(b) < 6 {
		return nil, fmt.Errorf("ico too small: %d", len(b))
	}
	if !looksLikeICO(b) {
		return nil, fmt.Errorf("not ico header")
	}

	count := int(binary.LittleEndian.Uint16(b[4:6]))
	entriesLen := 6 + count*16
	if count <= 0 || len(b) < entriesLen {
		return nil, fmt.Errorf("invalid ico directory: count=%d len=%d", count, len(b))
	}

	unique := map[int]struct{}{}
	for i := 0; i < count; i++ {
		base := 6 + i*16
		w := int(b[base])
		h := int(b[base+1])
		if w == 0 {
			w = 256
		}
		if h == 0 {
			h = 256
		}
		if w > 0 && h > 0 && w == h {
			unique[w] = struct{}{}
		}
	}

	sizes := make([]int, 0, len(unique))
	for s := range unique {
		sizes = append(sizes, s)
	}
	sort.Ints(sizes)
	return sizes, nil
}

func hasSize(sizes []int, size int) bool {
	for _, s := range sizes {
		if s == size {
			return true
		}
	}
	return false
}

func toNRGBA(img image.Image) *image.NRGBA {
	b := img.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)
	return dst
}

func scaleNearest(src *image.NRGBA, w, h int) *image.NRGBA {
	if w <= 0 || h <= 0 {
		return image.NewNRGBA(image.Rect(0, 0, 0, 0))
	}
	sw := src.Bounds().Dx()
	sh := src.Bounds().Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	if sw == 0 || sh == 0 {
		return dst
	}

	for y := 0; y < h; y++ {
		sy := y * sh / h
		for x := 0; x < w; x++ {
			sx := x * sw / w
			si := src.PixOffset(sx, sy)
			di := dst.PixOffset(x, y)
			copy(dst.Pix[di:di+4], src.Pix[si:si+4])
		}
	}
	return dst
}

func encodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func pngToICO(pngBytes []byte, sizes []int) ([]byte, error) {
	if len(pngBytes) == 0 {
		return nil, fmt.Errorf("png bytes empty")
	}
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, fmt.Errorf("decode png: %w", err)
	}

	src := toNRGBA(img)
	images := make([][]byte, 0, len(sizes))
	for _, size := range sizes {
		scaled := scaleNearest(src, size, size)
		encoded, err := encodePNG(scaled)
		if err != nil {
			return nil, fmt.Errorf("encode png size %d: %w", size, err)
		}
		images = append(images, encoded)
	}

	return buildICO(images, sizes, true)
}

func buildICO(images [][]byte, sizes []int, pngEncoded bool) ([]byte, error) {
	count := len(images)
	if count == 0 {
		return nil, fmt.Errorf("no images")
	}

	headerLen := 6
	dirLen := 16 * count
	offset := headerLen + dirLen

	out := make([]byte, 0, offset+1024)
	header := make([]byte, 6)
	binary.LittleEndian.PutUint16(header[0:2], 0)
	binary.LittleEndian.PutUint16(header[2:4], 1)
	binary.LittleEndian.PutUint16(header[4:6], uint16(count))
	out = append(out, header...)

	dir := make([]byte, dirLen)
	for i, imgData := range images {
		size := sizes[i]
		entry := dir[i*16 : i*16+16]
		if size >= 256 {
			entry[0] = 0
			entry[1] = 0
		} else {
			entry[0] = byte(size)
			entry[1] = byte(size)
		}
		entry[2] = 0
		entry[3] = 0
		if pngEncoded {
			binary.LittleEndian.PutUint16(entry[4:6], 1)
			binary.LittleEndian.PutUint16(entry[6:8], 32)
		} else {
			binary.LittleEndian.PutUint16(entry[4:6], 1)
			binary.LittleEndian.PutUint16(entry[6:8], 32)
		}
		binary.LittleEndian.PutUint32(entry[8:12], uint32(len(imgData)))
		binary.LittleEndian.PutUint32(entry[12:16], uint32(offset))
		offset += len(imgData)
	}
	out = append(out, dir...)

	for _, imgData := range images {
		out = append(out, imgData...)
	}
	return out, nil
}

func encodeICOBMPDIB(img *image.NRGBA) []byte {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	rowSize := w * 4
	xorSize := rowSize * h
	maskStride := ((w + 31) / 32) * 4
	andSize := maskStride * h

	header := make([]byte, 40)
	binary.LittleEndian.PutUint32(header[0:4], 40)
	binary.LittleEndian.PutUint32(header[4:8], uint32(w))
	binary.LittleEndian.PutUint32(header[8:12], uint32(h*2))
	binary.LittleEndian.PutUint16(header[12:14], 1)
	binary.LittleEndian.PutUint16(header[14:16], 32)
	binary.LittleEndian.PutUint32(header[16:20], 0)
	binary.LittleEndian.PutUint32(header[20:24], uint32(xorSize))
	binary.LittleEndian.PutUint32(header[24:28], 0)
	binary.LittleEndian.PutUint32(header[28:32], 0)
	binary.LittleEndian.PutUint32(header[32:36], 0)
	binary.LittleEndian.PutUint32(header[36:40], 0)

	buf := make([]byte, 0, 40+xorSize+andSize)
	buf = append(buf, header...)

	for y := h - 1; y >= 0; y-- {
		for x := 0; x < w; x++ {
			i := img.PixOffset(x, y)
			r := img.Pix[i+0]
			g := img.Pix[i+1]
			b := img.Pix[i+2]
			a := img.Pix[i+3]
			buf = append(buf, b, g, r, a)
		}
	}

	for y := 0; y < h; y++ {
		for x := 0; x < maskStride; x++ {
			buf = append(buf, 0)
		}
	}
	return buf
}

func pngToICOBMP(pngBytes []byte, sizes []int) ([]byte, error) {
	if len(pngBytes) == 0 {
		return nil, fmt.Errorf("png bytes empty")
	}
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, fmt.Errorf("decode png: %w", err)
	}

	src := toNRGBA(img)
	bmpImages := make([][]byte, 0, len(sizes))
	for _, size := range sizes {
		scaled := scaleNearest(src, size, size)
		dib := encodeICOBMPDIB(scaled)
		bmpImages = append(bmpImages, dib)
	}

	return buildICO(bmpImages, sizes, false)
}

var (
	windowsGeneratedICOOnce sync.Once
	windowsGeneratedICO     []byte
	windowsGeneratedICOErr  error
)

func windowsICOFromPNGOnce(pngBytes []byte) ([]byte, error) {
	windowsGeneratedICOOnce.Do(func() {
		windowsGeneratedICO, windowsGeneratedICOErr = pngToICOBMP(pngBytes, []int{16, 32, 48, 256})
	})
	return windowsGeneratedICO, windowsGeneratedICOErr
}

var (
	windowsSmallPNGOnce sync.Once
	windowsSmallPNG     []byte
	windowsSmallPNGErr  error
)

func windows32PNGFromPNGOnce(pngBytes []byte) ([]byte, error) {
	windowsSmallPNGOnce.Do(func() {
		if len(pngBytes) == 0 {
			windowsSmallPNGErr = fmt.Errorf("png bytes empty")
			return
		}
		img, err := png.Decode(bytes.NewReader(pngBytes))
		if err != nil {
			windowsSmallPNGErr = fmt.Errorf("decode png: %w", err)
			return
		}
		src := toNRGBA(img)
		scaled := scaleNearest(src, 32, 32)
		windowsSmallPNG, windowsSmallPNGErr = encodePNG(scaled)
	})
	return windowsSmallPNG, windowsSmallPNGErr
}

// generateRedBoxICO 生成一个纯红色的 16x16 ICO 数据，用于终极兜底
func generateRedBoxICO() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			i := img.PixOffset(x, y)
			img.Pix[i+0] = 255
			img.Pix[i+1] = 0
			img.Pix[i+2] = 0
			img.Pix[i+3] = 255
		}
	}
	pngBytes, _ := encodePNG(img)
	icoBytes, _ := pngToICOBMP(pngBytes, []int{16})
	return icoBytes
}

func generateRedBoxPNG() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			i := img.PixOffset(x, y)
			img.Pix[i+0] = 255
			img.Pix[i+1] = 0
			img.Pix[i+2] = 0
			img.Pix[i+3] = 255
		}
	}
	b, _ := encodePNG(img)
	return b
}
