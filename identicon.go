package identicon

import (
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"image"
	"image/color"
	"math"
	"math/rand"
	"strings"

	"github.com/fogleman/gg"
)

func hexConvert(code string) (red uint8, green uint8, blue uint8, err error) {
	code = strings.TrimPrefix(code, "#")
	b, err := hex.DecodeString(code)
	if err != nil {
		return 0, 0, 0, err
	}
	if len(b) != 3 {
		return 0, 0, 0, errors.New("code must be 6 digits long hex string")
	}
	return b[0], b[1], b[2], nil
}

// Code derives a code for use with Render.
func Code(str string) uint64 {
	buf := sha512.Sum512([]byte(str))
	return binary.BigEndian.Uint64(buf[56:])
}

type Color struct {
	Name string
	Code string
}

var DefaultColorPalette = []Color{
	{Name: "lightBlack", Code: "#2c2c2c"},
	{Name: "lightBlackIntense", Code: "#232323"},
	{Name: "turquoise", Code: "#00bf93"},
	{Name: "turquoiseIntense", Code: "#16a086"},
	{Name: "mint", Code: "#2dcc70"},
	{Name: "mintIntense", Code: "#27ae61"},
	{Name: "green", Code: "#42e453"},
	{Name: "greenIntense", Code: "#24c333"},
	{Name: "yellow", Code: "#ffff25"},
	{Name: "yellowIntense", Code: "#d9d921"},
	{Name: "yellowOrange", Code: "#f1c40f"},
	{Name: "yellowOrangeIntense", Code: "#f39c11"},
	{Name: "brown", Code: "#e67f22"},
	{Name: "brownIntense", Code: "#d25400"},
	{Name: "orange", Code: "#ff944e"},
	{Name: "orangeIntense", Code: "#ff5500"},
	{Name: "red", Code: "#e84c3d"},
	{Name: "redIntense", Code: "#c1392b"},
	{Name: "blue", Code: "#3598db"},
	{Name: "blueIntense", Code: "#297fb8"},
	{Name: "darkBlue", Code: "#34495e"},
	{Name: "darkBlueIntense", Code: "#2d3e50"},
	{Name: "lightGrey", Code: "#ecf0f1"},
	{Name: "lightGreyIntense", Code: "#bec3c7"},
	{Name: "grey", Code: "#95a5a5"},
	{Name: "greyIntense", Code: "#7e8c8d"},
	{Name: "magenta", Code: "#ef3e96"},
	{Name: "magentaIntense", Code: "#e52383"},
	{Name: "violet", Code: "#df21b9"},
	{Name: "violetIntense", Code: "#be127e"},
	{Name: "purple", Code: "#9a59b5"},
	{Name: "purpleIntense", Code: "#8d44ad"},
	{Name: "lightBlue", Code: "#7dc2d2"},
	{Name: "lightBlueIntense", Code: "#1cabbb"},
	{Name: "white", Code: "#ffffff"},
	{Name: "whiteIntense", Code: "#f3f5f7"},
	{Name: "black", Code: "#000000"},
}
var DefaultBackgroundColor = "#2d3e50"

type Settings struct {
	// TwoColor specifies if the identicon should be
	// generated using one or two colors.
	TwoColor bool

	// Alpha specifies the transparency of the generated identicon.
	Alpha uint8

	TransparentBackground bool
	ColorPalette          []Color
	BackgroundColors      []string
}

// DefaultSettings returns a Settings object with the recommended settings.
func DefaultSettings() *Settings {
	palette := DefaultColorPalette
	// backgroundColors := []string{DefaultBackgroundColor}
	backgroundColors := []string{"#f3f5f7", "#ecf0f1", "#2d3e50", "#393939"}
	return &Settings{
		TwoColor:              true,
		Alpha:                 255,
		ColorPalette:          palette,
		TransparentBackground: false,
		BackgroundColors:      backgroundColors,
	}
}

// Render generates an identicon.
// code is a code derived by the Code function.
// totalSize specifies the total size in pixels. It is recommended that
// this is divisible by 3.
func Render(code uint64, totalSize int, settings *Settings) (image.Image, error) {
	rnd := rand.New(rand.NewSource(int64(code % math.MaxInt32)))

	penWidth := 1

	middleType := int(code & 0x03)
	middleInvert := code>>2&0x01 == 1

	cornerType := int(code >> 3 & 0x0f)
	cornerInvert := code>>7&0x01 == 1
	cornerTurn := int(code >> 8 & 0x03)

	sideType := int(code >> 10 & 0x0f)
	sideInvert := code>>14&0x01 == 1
	sideTurn := int(code >> 15 & 0x03)

	swapCross := code>>47&0x01 == 1

	middleType = middlePatchSet[middleType]

	randomFirstColor := settings.ColorPalette[rnd.Intn(len(settings.ColorPalette))]
	red, green, blue, err := hexConvert(randomFirstColor.Code)
	if err != nil {
		return nil, err
	}
	randomSecondColor := settings.ColorPalette[rnd.Intn(len(settings.ColorPalette))]
	secondRed, secondGreen, secondBlue, err := hexConvert(randomSecondColor.Code)
	if err != nil {
		return nil, err
	}

	foreColor := color.RGBA{R: red, G: green, B: blue, A: settings.Alpha}
	var secondColor color.RGBA
	if settings.TwoColor {
		secondColor = color.RGBA{R: secondRed, G: secondGreen, B: secondBlue, A: settings.Alpha}
	} else {
		secondColor = foreColor
	}
	var middleColor color.Color
	if swapCross {
		middleColor = foreColor
	} else {
		middleColor = secondColor
	}
	image := gg.NewContext(totalSize, totalSize)
	patchSize := float64(totalSize) / 3

	if !settings.TransparentBackground {
		randomBackgroundColor := settings.BackgroundColors[rnd.Intn(len(settings.BackgroundColors))]
		bgRed, bgGreen, bgBlue, err := hexConvert(randomBackgroundColor)
		if err != nil {
			return nil, err
		}
		image.DrawRectangle(0, 0, float64(totalSize), float64(totalSize))
		image.SetRGB255(int(bgRed), int(bgGreen), int(bgBlue))
		image.Fill()
	}

	drawPatch(gg.Point{X: 1, Y: 1}, 0, middleInvert, middleType, image, patchSize, middleColor, penWidth)
	for i, p := range []gg.Point{{X: 1, Y: 0}, {X: 2, Y: 1}, {X: 1, Y: 2}, {X: 0, Y: 1}} {
		drawPatch(p, sideTurn+1+i, sideInvert, sideType, image, patchSize, foreColor, penWidth)
	}
	for i, p := range []gg.Point{{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}} {
		drawPatch(p, cornerTurn+1+i, cornerInvert, cornerType, image, patchSize, secondColor, penWidth)
	}
	return image.Image(), nil
}

func drawPatch(pos gg.Point, turn int, invert bool, type_ int, image *gg.Context, patchSize float64, foreColor color.Color, penWidth int) {
	path := pathSet[type_]
	turn %= 4
	image.Push()
	image.Translate(pos.X*patchSize+float64(penWidth)/2, pos.Y*patchSize+float64(penWidth)/2)
	image.RotateAbout(float64(turn)*math.Pi/2, patchSize/2, patchSize/2)
	for _, p := range path {
		image.LineTo(p.X/4*patchSize, p.Y/4*patchSize)
	}
	image.ClosePath()
	if invert {
		image.MoveTo(0, 0)
		image.LineTo(0, patchSize)
		image.LineTo(patchSize, patchSize)
		image.LineTo(patchSize, 0)
		image.ClosePath()
	}
	image.SetColor(foreColor)
	image.Fill()
	image.Pop()
}

var pathSet = [][]gg.Point{
	// [0] full square:
	{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}},
	// [1] right-angled triangle pointing top-left:
	{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 0, Y: 4}},
	// [2] upwardy triangle:
	{{X: 2, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}},
	// [3] left half of square, standing rectangle:
	{{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 4}, {X: 0, Y: 4}},
	// [4] square standing on diagonale:
	{{X: 2, Y: 0}, {X: 4, Y: 2}, {X: 2, Y: 4}, {X: 0, Y: 2}},
	// [5] kite pointing topleft:
	{{X: 0, Y: 0}, {X: 4, Y: 2}, {X: 4, Y: 4}, {X: 2, Y: 4}},
	// [6] Sierpinski triangle, fractal triangles:
	{{X: 2, Y: 0}, {X: 4, Y: 4}, {X: 2, Y: 4}, {X: 3, Y: 2}, {X: 1, Y: 2}, {X: 2, Y: 4}, {X: 0, Y: 4}},
	// [7] sharp angled lefttop pointing triangle:
	{{X: 0, Y: 0}, {X: 4, Y: 2}, {X: 2, Y: 4}},
	// [8] small centered square:
	{{X: 1, Y: 1}, {X: 3, Y: 1}, {X: 3, Y: 3}, {X: 1, Y: 3}},
	// [9] two small triangles:
	{{X: 2, Y: 0}, {X: 4, Y: 0}, {X: 0, Y: 4}, {X: 0, Y: 2}, {X: 2, Y: 2}},
	// [10] small topleft square:
	{{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}},
	// [11] downpointing right-angled triangle on bottom:
	{{X: 0, Y: 2}, {X: 4, Y: 2}, {X: 2, Y: 4}},
	// [12] uppointing right-angled triangle on bottom:
	{{X: 2, Y: 2}, {X: 4, Y: 4}, {X: 0, Y: 4}},
	// [13] small rightbottom pointing right-angled triangle on topleft:
	{{X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}},
	// [14] small lefttop pointing right-angled triangle on topleft:
	{{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 0, Y: 2}},
	// [15] empty:
	{},
}

// get the [0] full square, [4] square standing on diagonale, [8] small centered square, or [15] empty tile:
var middlePatchSet = []int{0, 4, 8, 15}
