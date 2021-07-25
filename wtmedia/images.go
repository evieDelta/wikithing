package wtmedia

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"git.lan/wikithing/etc/stopwatch"
	"git.lan/wikithing/wterr"
	"github.com/disintegration/imaging"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

type imageFunc func(string, image.Image) (image.Image, error)

var imageProcessingTags = map[string]imageFunc{
	"size": func(s string, img image.Image) (image.Image, error) {
		args := strings.Split(s, "-")
		size := strings.Split(args[0], "x")
		args = args[1:]

		x, y := img.Bounds().Dx(), img.Bounds().Dy()
		if len(size) == 0 {
			return nil, wterr.New(wterr.ErrInvalidInput, "no size given")
		}
		if len(size) == 1 {
			ns, err := strconv.Atoi(size[0])
			if err != nil {
				return nil, wterr.New(wterr.ErrInvalidInput, err)
			}
			if x > y {
				ratio := 1 - ((float64(x) - float64(ns)) / float64(x))
				log.Println(ratio)
				y = int(float64(y) * ratio)
				x = ns
			}
			if x < y {
				ratio := 1 - ((float64(y) - float64(ns)) / float64(y))
				log.Println(ratio)
				x = int(float64(x) * ratio)
				y = ns
			}
		} else {
			var err error
			x, err = strconv.Atoi(size[0])
			if err != nil {
				return nil, err
			}
			y, err = strconv.Atoi(size[1])
			if err != nil {
				return nil, err
			}
		}

		return imaging.Resize(img, x, y, imaging.MitchellNetravali), nil
	},
	"sat": func(s string, img image.Image) (image.Image, error) {
		s = strings.TrimRight(s, "%")
		p, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}

		return imaging.AdjustSaturation(img, p-100), nil
	},
	"gamma": func(s string, img image.Image) (image.Image, error) {
		s = strings.TrimRight(s, "%")
		p, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}

		return imaging.AdjustGamma(img, p), nil
	},
	"brightness": func(s string, img image.Image) (image.Image, error) {
		s = strings.TrimRight(s, "%")
		p, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}

		return imaging.AdjustBrightness(img, p-100), nil
	},
	"chromaticSmear": func(s string, img image.Image) (image.Image, error) {
		s = strings.TrimRight(s, "%")
		p, err := strconv.ParseUint(s, 0, 8)
		if err != nil {
			return nil, err
		}
		sub := float64(p)

		var red float64
		var grn float64
		var blu float64
		var alp float64

		reset := func() {
			sub = float64(p)

			red = 0
			grn = 0
			blu = 0
			alp = 0
		}

		smear := func(c color.NRGBA) color.NRGBA {
			nr := float64(c.R)
			if nr > red {
				red = ((red * (255 - sub)) + (nr * sub)) / 255
				nr = (nr + red*2) / 3
			} else {
				nr = (nr + red*2) / 3
				red = ((red * sub) + (nr * (255 - sub))) / 255
			}
			ng := float64(c.G)
			if ng > grn {
				grn = ((grn * (255 - sub)) + (ng * sub)) / 255
				ng = (ng + grn*2) / 3
			} else {
				ng = (ng + grn*2) / 3
				grn = ((grn * sub) + (ng * (255 - sub))) / 255
			}
			nb := float64(c.B)
			if nb > blu {
				blu = ((blu * (255 - sub)) + (nb * sub)) / 255
				nb = (nb + blu*2) / 3
			} else {
				nb = (nb + blu*2) / 3
				blu = ((blu * sub) + (nb * (255 - sub))) / 255
			}
			na := float64(c.A)
			if na > alp {
				alp = ((alp * (255 - sub)) + (na * sub)) / 255
				na = (na + alp*2) / 3
			} else {
				na = (na + alp*2) / 3
				alp = ((alp * sub) + (na * (255 - sub))) / 255
			}
			//	if c.G > grn {
			//		grn = c.G
			//	} else {
			//		c.G += grn
			//		grn -= sub
			//	}
			//	if c.B > blu {
			//		blu = c.B
			//	} else {
			//		c.B += blu
			//		blu -= sub
			//	}

			c.R = uint8(nr)
			c.G = uint8(ng)
			c.B = uint8(nb)
			c.A = uint8(na)

			return c
		}

		hrimg := image.NewNRGBA(image.Rect(img.Bounds().Min.X, img.Bounds().Min.Y, img.Bounds().Max.X, img.Bounds().Max.Y))
		for y := 0; y < img.Bounds().Dy(); y++ {
			reset()
			for x := 0; x < img.Bounds().Dx(); x++ {
				c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)

				b := smear(c)
				hrimg.Set(x, y, b)
			}
		}

		lhing := image.NewNRGBA(image.Rect(img.Bounds().Min.X, img.Bounds().Min.Y, img.Bounds().Max.X, img.Bounds().Max.Y))
		for y := img.Bounds().Dy(); y >= 0; y-- {
			reset()
			for x := img.Bounds().Dx(); x >= 0; x-- {
				c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)

				b := smear(c)
				lhing.Set(x, y, b)
			}
		}

		dvimg := image.NewNRGBA(image.Rect(img.Bounds().Min.X, img.Bounds().Min.Y, img.Bounds().Max.X, img.Bounds().Max.Y))
		for x := 0; x < img.Bounds().Dx(); x++ {
			reset()
			for y := 0; y < img.Bounds().Dy(); y++ {
				c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)

				b := smear(c)
				dvimg.Set(x, y, b)
			}
		}

		uving := image.NewNRGBA(image.Rect(img.Bounds().Min.X, img.Bounds().Min.Y, img.Bounds().Max.X, img.Bounds().Max.Y))
		for x := img.Bounds().Dx(); x >= 0; x-- {
			reset()
			for y := img.Bounds().Dy(); y >= 0; y-- {
				c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)

				b := smear(c)
				uving.Set(x, y, b)
			}
		}

		fnimg := image.NewNRGBA(image.Rect(img.Bounds().Min.X, img.Bounds().Min.Y, img.Bounds().Max.X, img.Bounds().Max.Y))
		for y := 0; y < img.Bounds().Dy(); y++ {
			for x := 0; x < img.Bounds().Dx(); x++ {
				r := hrimg.NRGBAAt(x, y)
				l := lhing.NRGBAAt(x, y)
				u := uving.NRGBAAt(x, y)
				d := dvimg.NRGBAAt(x, y)

				f := color.NRGBA{}

				f.R = uint8((uint32(l.R) + uint32(r.R) + uint32(u.R) + uint32(d.R)) / 4)
				f.G = uint8((uint32(l.G) + uint32(r.G) + uint32(u.G) + uint32(d.G)) / 4)
				f.B = uint8((uint32(l.B) + uint32(r.B) + uint32(u.B) + uint32(d.B)) / 4)
				f.A = uint8((uint32(l.A) + uint32(r.A) + uint32(u.A) + uint32(d.A)) / 4)
				fnimg.Set(x, y, f)
			}
		}

		return fnimg, nil
	},
	"blur": func(s string, img image.Image) (image.Image, error) {
		s = strings.TrimRight(s, "%")
		p, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}

		return imaging.Blur(img, p), nil
	},
	"colorCruncher": func(s string, img image.Image) (image.Image, error) {
		//    args := strings.Split(s, ",")
		//    space := args[0]
		//    amt

		seeb, err := strconv.ParseInt(s, 0, 64)
		if err != nil {
			return nil, err
		}

		rsrc := rand.New(rand.NewSource(seeb))

		roff := func(n int) int {
			return rsrc.Intn(n) - (n / 2)
		}

		{
			nimg := image.NewNRGBA(image.Rect(img.Bounds().Min.X, img.Bounds().Min.Y, img.Bounds().Max.X, img.Bounds().Max.Y))

			for y := 0; y < img.Bounds().Dy(); y++ {
				var rofx, gofx, bofx, rofy, gofy, bofy int

				rofx = roff(6)
				gofx = roff(6)
				bofx = roff(6)

				for x := 0; x < img.Bounds().Dx(); x++ {

					if x%1 == 0 {
						rofy = roff(6)
						gofy = roff(6)
						bofy = roff(6)
					}

					r, g, b, _ := img.At(x, y).RGBA()

					rs := nimg.NRGBAAt(x+rofx, y+rofy)
					rs.R = uint8Trunc(r/255 + uint32(rs.R))
					rs.A = uint8(255)
					nimg.SetNRGBA(x+rofx, y+rofy, rs)

					gs := nimg.NRGBAAt(x+gofx, y+gofy)
					gs.G = uint8Trunc(g/255 + uint32(gs.G))
					gs.A = uint8(255)
					nimg.SetNRGBA(x+gofx, y+gofy, gs)

					bs := nimg.NRGBAAt(x+bofx, y+bofy)
					bs.B = uint8Trunc(b/255 + uint32(bs.B))
					bs.A = uint8(255)
					nimg.SetNRGBA(x+bofx, y+bofy, bs)
				}
			}

			img = nimg
		}

		return img, nil
	},
	"chromaticAberration": func(s string, img image.Image) (image.Image, error) {
		parseXY := func(s string) (int, int, error) {
			if !strings.Contains(s, "x") {
				var per = false
				if strings.HasSuffix(s, "%") || strings.HasSuffix(s, "p") || strings.HasSuffix(s, "P") {
					per = true
				}
				n, err := strconv.ParseFloat(strings.TrimRight(s, "pP%"), 64)
				if err != nil {
					return 0, 0, wterr.Newln(wterr.ErrInvalidInput, "invalid x/y input", err)
				}
				var xx, yy int
				if per {
					xx = int(float64(img.Bounds().Dx()) * (n / 100))
					yy = int(float64(img.Bounds().Dy()) * (n / 100))
				} else {
					xx = int(n)
					yy = int(n)
				}
				return xx, yy, nil
			}

			l := strings.Split(s, "x")
			if len(l) != 2 {
				return 0, 0, wterr.New(wterr.ErrInvalidInput, "invalid x/y input")
			}
			xper := strings.HasSuffix(l[0], "%") || strings.HasSuffix(l[0], "p") || strings.HasSuffix(l[0], "P")
			x, err := strconv.ParseFloat(strings.TrimRight(l[0], "pP%"), 64)
			if err != nil {
				return 0, 0, err
			}
			l[1] = strings.TrimSuffix(l[1], "y")
			yper := strings.HasSuffix(l[1], "%") || strings.HasSuffix(l[0], "p") || strings.HasSuffix(l[0], "P")
			y, err := strconv.ParseFloat(strings.TrimRight(l[1], "pP%"), 64)
			if err != nil {
				return 0, 0, wterr.Newln(wterr.ErrInvalidInput, "invalid x/y input", err)
			}

			if xper {
				x = float64(img.Bounds().Dx()) * (x / 100)
			}
			if yper {
				y = float64(img.Bounds().Dy()) * (y / 100)
			}

			return int(x), int(y), nil

		}

		var rx, ry, gx, gy, bx, by int

		args := strings.Split(s, ",")
		if len(args) == 0 {
			rx = 5
			by = 5

			goto skipParse
		}
		{
			var set = false
			for _, v := range args {
				if strings.HasPrefix(v, "g:") {
					v := strings.TrimPrefix(v, "g:")
					var err error
					bx, by, err = parseXY(v)
					if err != nil {
						return nil, err
					}
					set = true
				}
				if strings.HasPrefix(v, "g:") {
					v := strings.TrimPrefix(v, "g:")
					var err error
					gx, gy, err = parseXY(v)
					if err != nil {
						return nil, err
					}
					set = true
				}
				if strings.HasPrefix(v, "b:") {
					v := strings.TrimPrefix(v, "b:")
					var err error
					bx, by, err = parseXY(v)
					if err != nil {
						return nil, err
					}
					set = true
				}
			}
			if !set && len(args) == 1 {
				x, y, err := parseXY(s)
				if err != nil {
					return nil, err
				}
				rx = x
				by = y
			}
		}
	skipParse:

		{
			nimg := image.NewNRGBA(image.Rect(img.Bounds().Min.X, img.Bounds().Min.Y, img.Bounds().Max.X, img.Bounds().Max.Y))

			for y := 0; y < img.Bounds().Dy(); y++ {
				for x := 0; x < img.Bounds().Dx(); x++ {

					r, g, b, _ := img.At(x, y).RGBA()

					rs := nimg.NRGBAAt(x+rx, y+ry)
					rs.R = uint8Trunc(r/255 + uint32(rs.R))
					rs.A = uint8(255)
					nimg.SetNRGBA(x+rx, y+ry, rs)

					gs := nimg.NRGBAAt(x+gx, y+gy)
					gs.G = uint8Trunc(g/255 + uint32(gs.G))
					gs.A = uint8(255)
					nimg.SetNRGBA(x+gx, y+gy, gs)

					bs := nimg.NRGBAAt(x+bx, y+by)
					bs.B = uint8Trunc(b/255 + uint32(bs.B))
					bs.A = uint8(255)
					nimg.SetNRGBA(x+bx, y+by, bs)
				}
			}

			img = nimg
		}

		return img, nil
	},
}

func uint8Trunc(i uint32) uint8 {
	if i > 255 {
		return 255
	}
	if i < 0 { // this is impossible
		return 0
	}
	return uint8(i)
}

type imagetask struct {
	fn   imageFunc
	args string

	order int

	name string
}

func (d *DefaultLocal) servImage(hash string, meta ObjectMeta, q QueryData) (io.ReadCloser, string, error) {

	tasks := make([]imagetask, 0)

	var tagSO = map[string]int{}

	if len(q.Values.Get("order")) == 0 {
		goto checkTags
	}

	for i, v := range strings.Split(q.Values.Get("order"), ",") {
		s := strings.SplitN(v, ":", 2)

		if len(s) == 0 {
			return nil, "", wterr.Newf(wterr.ErrInvalidInput, "nameless order key at position %v", i)
		}
		if len(s) == 1 {
			return nil, "", wterr.Newf(wterr.ErrInvalidInput, "Order key `%v` missing number", s[0])
		}

		n, err := strconv.Atoi(s[1])
		if err != nil {
			return nil, "", wterr.Newf(wterr.ErrInvalidInput, "invalid number given in key order `%v`: %v", s[0], err)
		}

		tagSO[s[0]] = n
	}

checkTags:

	for k, v := range imageProcessingTags {
		if len(q.Values.Get(k)) == 0 {
			continue
		}
		tasks = append(tasks, imagetask{
			fn:   v,
			args: q.Values.Get(k),

			order: tagSO[k],

			name: k,
		})
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].order > tasks[j].order
	})

	ext := strings.TrimLeft(q.Extension, ".")
	var formatFunc func(io.Writer, image.Image) error
	var formatMime string
	extmod := ""

	{
		l := strings.Split(ext, ":")
		ext = l[0]
		if len(l) > 1 {
			extmod = l[1]
		}
	}

	if meta.Type.Mime == "image/gif" {
		switch ext {
		case "", "gif":
			goto plain
		default:
			return nil, "", wterr.New(wterr.ErrUnsupported, "converting from gifs is currently unsupported")
		}
	}

	switch ext {
	case "":
		goto plain

	case "png":
		formatFunc = png.Encode
		formatMime = "image/png"

	case "jpg", "jpeg":
		q := 75
		if extmod != "" {
			p, err := strconv.Atoi(extmod)
			if err != nil {
				return nil, "", wterr.Newln(wterr.ErrInvalidInput, "invalid quality modifier:", err)
			}
			q = p
		}
		formatFunc = func(w io.Writer, i image.Image) error {
			return jpeg.Encode(w, i, &jpeg.Options{
				Quality: q,
			})
		}
		formatMime = "image/jpeg"

	case "webp":
		formatMime = "image/webp"
		switch meta.Type.Mime {
		case "image/webp":
			goto plain
		default:
			// unfortunetely x/image/webp doesn't support encoding
			return nil, "", wterr.Newf(wterr.ErrUnsupported, "encoding to webp is not supported")
		}

	case "tiff", "tif":
		formatFunc = func(w io.Writer, i image.Image) error {
			return tiff.Encode(w, i, &tiff.Options{
				Compression: tiff.Deflate,
				Predictor:   extmod == "",
			})
		}
		formatMime = "image/tiff"

	case "bmp":
		formatMime = "image/bmp"
		switch meta.Type.Mime {
		case "image/bmp":
			goto plain
		default:
			// who the heck still uses bmp anyways
			return nil, "", wterr.Newf(wterr.ErrUnsupported, "encoding to bmp is not supported")
		}
	}

	switch meta.Type.Mime {
	case formatMime:
		goto plain
	default:
		goto reformat
	}

plain:
	if len(tasks) > 0 {
		goto reformat
	}

	return d.servBinary(hash, meta)

reformat:
	if formatFunc == nil && formatMime != meta.Type.Mime {
		return nil, "", wterr.Newf(wterr.ErrUnsupported, "encoding images in %v is not supported", formatMime)
	}

	modifiers := q.Values.Encode()

	data, err := d.getFileFromEither(hash)
	if err != nil {
		return nil, "", err
	}

	key := hash + ":" + formatMime + ":" + extmod + ":" + modifiers

	return d.processImages(formatFunc, formatMime, tasks, data, meta.Type.Mime, key)
}

func (d *DefaultLocal) processImages(to func(io.Writer, image.Image) error, toMime string, tasks []imagetask, file []byte, initial, key string) (io.ReadCloser, string, error) {
	if d, ok := d.getFromCache(key); ok {
		log.Println("serving from cache")
		return io.NopCloser(bytes.NewReader(d)), toMime, nil
	}

	var img image.Image
	var dat = bytes.NewBuffer(file)
	var err error

	switch initial {
	case "image/png":
		img, err = png.Decode(dat)
	case "image/jpeg":
		img, err = jpeg.Decode(dat)
	case "image/webp":
		img, err = webp.Decode(dat)
	case "image/tiff":
		img, err = tiff.Decode(dat)
	case "image/bmp":
		img, err = bmp.Decode(dat)
	default:
		img, _, err = image.Decode(dat)
	}
	if err != nil {
		return nil, "", err
	}

	sw := stopwatch.New()

	for _, x := range tasks {
		sw.Start()
		img, err = x.fn(x.args, img)
		if err != nil {
			return nil, "", err
		}
		log.Println(x.name, x.args, "took", sw.Stop().Round(time.Millisecond/10))
	}

	dat.Reset()
	err = to(dat, img)
	if err != nil {
		return nil, "", err
	}

	d.Cache.Add(key, dat.Bytes())

	return io.NopCloser(dat), toMime, nil
}
