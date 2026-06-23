package menubar

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/haritabh17/theirtime/internal/slack"
)

var testBackground = color.RGBA{R: 100, G: 120, B: 140, A: 255}

func solidPNG(size int, c color.Color) []byte {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func pixel(img image.Image, x, y int) color.RGBA {
	r, g, b, a := img.At(x, y).RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}

func isBackground(c color.RGBA) bool {
	return c.R == testBackground.R && c.G == testBackground.G && c.B == testBackground.B
}

func solidJPEG(size int, c color.Color) []byte {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func TestApplyPresenceBadgeJPEG(t *testing.T) {
	raw := solidJPEG(48, testBackground)
	got, err := applyPresenceBadge(raw, slack.PresenceActive)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(got, raw) {
		t.Fatal("expected JPEG avatar to be badged")
	}
}

func TestApplyPresenceBadgeUnknownReturnsOriginal(t *testing.T) {
	raw := solidPNG(48, testBackground)
	got, err := applyPresenceBadge(raw, slack.Presence("offline"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, raw) {
		t.Fatal("expected original bytes for unknown presence")
	}
}

func TestApplyPresenceBadgeActiveDiffersFromAway(t *testing.T) {
	raw := solidPNG(48, testBackground)

	active, err := applyPresenceBadge(raw, slack.PresenceActive)
	if err != nil {
		t.Fatal(err)
	}
	away, err := applyPresenceBadge(raw, slack.PresenceAway)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(active, raw) {
		t.Fatal("active badge should modify image")
	}
	if bytes.Equal(away, raw) {
		t.Fatal("away badge should modify image")
	}
	if bytes.Equal(active, away) {
		t.Fatal("active and away badges should differ")
	}

	activeImg, err := png.Decode(bytes.NewReader(active))
	if err != nil {
		t.Fatal(err)
	}
	corner := pixel(activeImg, 45, 45)
	if corner.G <= corner.R || corner.G <= corner.B {
		t.Fatalf("active badge near corner should be green-dominant, got %#v", corner)
	}
}

func TestBadgeOverflowsCorner(t *testing.T) {
	raw := solidPNG(48, testBackground)
	badged, err := applyPresenceBadge(raw, slack.PresenceActive)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(badged))
	if err != nil {
		t.Fatal(err)
	}

	b := img.Bounds()
	if b.Dx() <= 48 || b.Dy() <= 48 {
		t.Fatalf("badged image should be larger than avatar, got %dx%d", b.Dx(), b.Dy())
	}

	// Avatar unchanged in top-left.
	if !isBackground(pixel(img, 0, 0)) {
		t.Fatal("avatar top-left should be unchanged")
	}

	// Badge extends past original avatar edge into padding (full circle, not clipped).
	if isBackground(pixel(img, 55, 48)) {
		t.Fatal("expected badge pixels in right padding beyond avatar edge")
	}
	if isBackground(pixel(img, 48, 55)) {
		t.Fatal("expected badge pixels in bottom padding beyond avatar edge")
	}

	// A point in the overflow padding (outside the 48×48 avatar) should show badge color.
	overflowPixel := pixel(img, 58, 52)
	if overflowPixel.A == 0 {
		t.Fatal("badge should render in overflow padding, not be clipped")
	}

	const minBadgePixels = 80
	count := 0
	for y := 24; y < b.Dy(); y++ {
		for x := 24; x < b.Dx(); x++ {
			if !isBackground(pixel(img, x, y)) {
				count++
			}
		}
	}
	if count < minBadgePixels {
		t.Fatalf("badge area too small in bottom-right region: %d pixels", count)
	}
}

func TestDisplayAvatarRespectsToggle(t *testing.T) {
	raw := solidPNG(48, testBackground)
	got := displayAvatar(false, raw, slack.PresenceActive)
	if !bytes.Equal(got, raw) {
		t.Fatal("expected raw avatar when presence display disabled")
	}
}

func TestRebuildDisplayAvatarsKeepsRawAvatar(t *testing.T) {
	raw := solidPNG(48, testBackground)
	a := &app{
		cfg:            defaultStatusCfg(),
		rawAvatarCache: map[string][]byte{"u1": raw},
	}

	a.rebuildDisplayAvatars(nil)

	got := a.avatarCache["u1"]
	if !bytes.Equal(got.data, raw) {
		t.Fatal("expected menu bar avatar to remain unbadged")
	}
	if got.contentSize != 0 {
		t.Fatalf("content size got %d want 0", got.contentSize)
	}
}
