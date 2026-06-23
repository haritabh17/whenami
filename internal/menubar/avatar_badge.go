package menubar

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"math"

	"github.com/fogleman/gg"
	"github.com/haritabh17/theirtime/internal/slack"
)

const badgeRadiusRatio = 0.27

// avatarEntry holds display-ready avatar bytes and the original square avatar
// size before badge padding, used so the menu bar scales the face to full size.
type avatarEntry struct {
	data        []byte
	contentSize int
}

var (
	colorSlackGreen     = color.RGBA{R: 0x2B, G: 0xAC, B: 0x76, A: 0xFF}
	colorActiveHighlight = color.RGBA{R: 0x4C, G: 0xD6, B: 0x94, A: 0xFF}
	colorActiveShadow    = color.RGBA{R: 0x1A, G: 0x7A, B: 0x52, A: 0xFF}
	colorAwayLight       = color.RGBA{R: 0x9A, G: 0x9A, B: 0x9A, A: 0xFF}
	colorAwayDark        = color.RGBA{R: 0x4A, G: 0x4A, B: 0x4A, A: 0xFF}
)

func applyPresenceBadge(avatarPNG []byte, presence slack.Presence) ([]byte, error) {
	switch presence {
	case slack.PresenceActive, slack.PresenceAway:
	default:
		return avatarPNG, nil
	}

	src, _, err := image.Decode(bytes.NewReader(avatarPNG))
	if err != nil {
		return nil, err
	}

	bounds := src.Bounds()
	size := bounds.Dx()
	if size <= 0 {
		return avatarPNG, nil
	}

	pad := badgeOverflowPadding(size)
	canvas := image.NewRGBA(image.Rect(0, 0, size+pad, size+pad))
	draw.Draw(canvas, image.Rect(0, 0, size, size), src, bounds.Min, draw.Src)

	dc := gg.NewContextForImage(canvas)
	drawPresenceBadge(dc, size, presence)

	var out bytes.Buffer
	if err := dc.EncodePNG(&out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func badgeGeometry(size int) (cx, cy, r float64) {
	cx = float64(size)
	cy = float64(size)
	r = float64(size) * badgeRadiusRatio
	return cx, cy, r
}

// badgeOverflowPadding is extra transparent space on the right and bottom so the
// full badge circle can extend past the avatar square without clipping.
func badgeOverflowPadding(size int) int {
	_, _, r := badgeGeometry(size)
	// Shadow (r+1) + offset (1) + white ring (2px) beyond radius r.
	return int(math.Ceil(r + 4))
}

func drawPresenceBadge(dc *gg.Context, size int, presence slack.Presence) {
	cx, cy, r := badgeGeometry(size)

	dc.SetRGBA(0, 0, 0, 0.25)
	dc.DrawCircle(cx+1, cy+1, r+1)
	dc.Fill()

	dc.SetLineWidth(2)
	dc.SetColor(color.White)
	dc.DrawCircle(cx, cy, r+1)
	dc.Stroke()

	switch presence {
	case slack.PresenceActive:
		drawActiveBadge(dc, cx, cy, r)
	case slack.PresenceAway:
		drawAwayBadge(dc, cx, cy, r)
	}
}

func drawActiveBadge(dc *gg.Context, cx, cy, r float64) {
	hx := cx - r*0.35
	hy := cy - r*0.35
	grad := gg.NewRadialGradient(hx, hy, 0, cx, cy, r)
	grad.AddColorStop(0, colorActiveHighlight)
	grad.AddColorStop(0.55, colorSlackGreen)
	grad.AddColorStop(1, colorActiveShadow)
	dc.SetFillStyle(grad)
	dc.DrawCircle(cx, cy, r)
	dc.Fill()

	dc.SetRGBA(1, 1, 1, 0.7)
	dc.DrawCircle(cx-r*0.35, cy-r*0.35, r*0.22)
	dc.Fill()
}

func drawAwayBadge(dc *gg.Context, cx, cy, r float64) {
	grad := gg.NewRadialGradient(cx-r*0.3, cy-r*0.3, 0, cx, cy, r)
	grad.AddColorStop(0, colorAwayLight)
	grad.AddColorStop(0.7, color.RGBA{R: 0x6B, G: 0x6B, B: 0x6B, A: 0xFF})
	grad.AddColorStop(1, colorAwayDark)
	dc.SetFillStyle(grad)
	dc.DrawCircle(cx, cy, r)
	dc.Fill()

	dc.SetColor(color.White)
	dc.DrawCircle(cx, cy, r*0.55)
	dc.Fill()
}

func imageSquareSize(raw []byte) int {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || cfg.Width <= 0 {
		return 0
	}
	return cfg.Width
}

func displayAvatar(showPresence bool, raw []byte, presence slack.Presence) []byte {
	if !showPresence || len(raw) == 0 {
		return raw
	}
	badged, err := applyPresenceBadge(raw, presence)
	if err != nil {
		return raw
	}
	return badged
}
