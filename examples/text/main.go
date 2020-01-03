package main

import (
	"github.com/maja42/nora"
	"github.com/maja42/nora/builtin/shader"
	"github.com/maja42/nora/builtin/shapes"
	"github.com/maja42/nora/color"
	"github.com/maja42/nora/math"
	"github.com/sirupsen/logrus"
)

func main() {
	err := run()
	if err != nil {
		logrus.Fatalln(err)
	}
}

func run() error {
	if err := nora.Init(); err != nil {
		return err
	}
	defer nora.Destroy()

	n, err := nora.Run(math.Vec2i{1920, 1080}, "Demo", nil, nil, nora.ResizeKeepAspectRatio)
	if err != nil {
		return err
	}
	defer n.Wait()

	if err := n.Shaders.LoadAll(shader.Builtins("builtin/shader")); err != nil {
		logrus.Errorf("Failed to load builtin shaders: %s", err)
	}

	n.SetClearColor(color.Gray(0.1))

	roboto, err := nora.LoadFont("builtin/fonts/roboto", "roboto_regular_65.xml")
	if err != nil {
		logrus.Errorf("Failed to load font: %s", err)
	}

	monospace, err := nora.LoadFont("builtin/fonts/ibm plex mono", "ibm_plex_mono_regular_32.xml")
	if err != nil {
		logrus.Errorf("Failed to load font: %s", err)
	}

	txt := shapes.NewText(roboto, "Nora rendering engine")
	txt.SetUniformScale(0.0008)
	txt.MoveXY(-1.01, 0.775)
	n.Scene.Attach(txt)

	txt = shapes.NewText(monospace, "To-Do:\n"+
		"\t ✓ Support textures\n"+
		"\t ✓ Support fonts\n"+
		"\t ❌ Support truetype fonts with kerning and hinting\n"+
		"\t ✓ Feed the dog\n"+
		"\t ❌ Water the plant\n"+
		"\t ✓ Test with special characters: ⇆ ‡ ⅑ ↶ ₹\n"+
		"\t - Make something nice with it\n"+
		"\t - Buy new plant\n"+
		"")
	txt.SetUniformScale(0.0008)
	txt.MoveXY(-1.01, 0.65)
	n.Scene.Attach(txt)

	return nil
}
