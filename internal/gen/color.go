package gen

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type RGB struct{ R, G, B float64 }
type HSL struct{ H, S, L float64 }

func ParseHex(hex string) (RGB, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return RGB{}, fmt.Errorf("invalid hex color: %s", hex)
	}
	r, _ := strconv.ParseInt(hex[0:2], 16, 32)
	g, _ := strconv.ParseInt(hex[2:4], 16, 32)
	b, _ := strconv.ParseInt(hex[4:6], 16, 32)
	return RGB{float64(r) / 255.0, float64(g) / 255.0, float64(b) / 255.0}, nil
}

func (c RGB) Hex() string {
	r := clampInt(int(c.R*255), 0, 255)
	g := clampInt(int(c.G*255), 0, 255)
	b := clampInt(int(c.B*255), 0, 255)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func (c RGB) ToHSL() HSL {
	r, g, b := c.R, c.G, c.B
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	l := (max + min) / 2

	if max == min {
		return HSL{0, 0, l}
	}

	d := max - min
	var s float64
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}

	var h float64
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	case b:
		h = (r-g)/d + 4
	}
	h /= 6

	return HSL{h, s, l}
}

func (hsl HSL) ToRGB() RGB {
	h, s, l := hsl.H, hsl.S, hsl.L

	if s == 0 {
		return RGB{l, l, l}
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	return RGB{
		hueToRGB(p, q, h+1.0/3.0),
		hueToRGB(p, q, h),
		hueToRGB(p, q, h-1.0/3.0),
	}
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

func (c RGB) Luminance() float64 {
	r := linearize(c.R)
	g := linearize(c.G)
	b := linearize(c.B)
	return 0.2126*r + 0.7152*g + 0.0722*b
}

func linearize(v float64) float64 {
	if v <= 0.03928 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func ContrastRatio(a, b RGB) float64 {
	l1 := a.Luminance()
	l2 := b.Luminance()
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

func (c RGB) WithLightness(newL float64) RGB {
	h := c.ToHSL()
	h.L = clamp(newL, 0, 1)
	return h.ToRGB()
}

func (c RGB) WithSaturation(newS float64) RGB {
	h := c.ToHSL()
	h.S = clamp(newS, 0, 1)
	return h.ToRGB()
}

func clamp(v, lo, hi float64) float64 {
	return math.Max(lo, math.Min(hi, v))
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
