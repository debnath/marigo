package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"math/rand"
	"os"

	"time"

	"sync"

	"github.com/disintegration/gift"
	termbox "github.com/nsf/termbox-go"
)

// parameters
var windowWidth, windowHeight = 400, 300

// sprites
var src = getImage("imgs/sprite_mario_small.png")

//var background = getImage("imgs/bg.png")

//var marioJump = image.Rect(1350,0, 1400, 75)
//var marioRun1 = image.Rect(1534,0, 1584, 75)
//var marioRun2 = image.Rect(1625,0	, 1675, 75)
var marioStand = image.Rect(50, 0, 80, 40)
var marioSquat = image.Rect(955, 0, 977, 38)
var ground = image.Rect(0, 60, 1200, 70)

// Sprite represents a sprite in the game
type Sprite struct {
	size     image.Rectangle // the sprite size
	Filter   *gift.GIFT      // normal filter used to draw the sprite
	FilterA  *gift.GIFT      // alternate filter used to draw the sprite
	FilterE  *gift.GIFT      // exploded filter used to draw the sprite
	Position image.Point     // top left position of the sprite
	Status   bool            // alive or dead
	Points   int             // number of points if destroyed
}

var mario = Sprite{
	size:     marioStand,
	Filter:   gift.New(gift.Crop(marioStand)),
	FilterE:  gift.New(gift.Crop(marioSquat)),
	Position: image.Pt(50, 240),
	Status:   true,
}

var terrain = Sprite{
	size:     ground,
	Filter:   gift.New(gift.Crop(ground)),
	Position: image.Pt(0, 280),
	Status:   true,
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	// game variables
	gameOver := false // end of game
	score := 0        // number of points scored in the game so far

	// poll for keyboard events in another goroutine
	events := make(chan termbox.Event, 1000)
	refresher := make(chan bool, 50)

	go func() {
		for {
			events <- termbox.PollEvent()
			refresher <- true
		}
	}()

	// block on the start screen the start screen
	startScreen := getImage("imgs/start.png")
	printImage(startScreen)
start:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Ch == 's' || ev.Ch == 'S' {
				break start
			}

			if ev.Ch == 'q' || ev.Ch == 'Q' || ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlQ || ev.Key == termbox.KeyCtrlC {
				gameOver = true
				break start

			}
		}
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for !gameOver {
			select {
			// create background
			case <-refresher:
				dst := image.NewRGBA(image.Rect(0, 0, windowWidth, windowHeight))

				mario.Filter.DrawAt(dst, src, image.Pt(mario.Position.X, mario.Position.Y), gift.OverOperator)

				terrain.Filter.DrawAt(dst, src, image.Pt(terrain.Position.X, terrain.Position.Y), gift.OverOperator)
				printImage(dst)
			}
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for !gameOver {
			select {
			case ev := <-events:
				if ev.Type == termbox.EventKey {
					switch ev.Key {
					case termbox.KeyEsc, termbox.KeyCtrlQ, termbox.KeyCtrlC:
						gameOver = true
					case termbox.KeyArrowRight:
						mario.Position.X += 10
					case termbox.KeyArrowLeft:
						mario.Position.X -= 10
					case termbox.KeyArrowUp:
						mario.Position.Y -= 10
					case termbox.KeyArrowDown:
						mario.Position.Y += 10
					}
				}
			}
		}
		wg.Done()
	}()

	wg.Wait()
	termbox.Close()
	fmt.Println("\nGAME OVER!\nFinal score:", score)
}

//maybe cleaner to use termbox's collide() function
func collide(s1, s2 Sprite) bool {
	spriteA := image.Rect(s1.Position.X, s1.Position.Y, s1.Position.X+s1.size.Dx(), s1.Position.Y+s1.size.Dy())
	spriteB := image.Rect(s2.Position.X, s2.Position.Y, s2.Position.X+s1.size.Dx(), s2.Position.Y+s1.size.Dy())
	if spriteA.Min.X < spriteB.Max.X && spriteA.Max.X > spriteB.Min.X &&
		spriteA.Min.Y < spriteB.Max.Y && spriteA.Max.Y > spriteB.Min.Y {
		return true
	}
	return false
}

// this only works for iTerm2!
func printImage(img image.Image) {
	var buf bytes.Buffer
	png.Encode(&buf, img)
	imgBase64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	fmt.Printf("\x1b[2;0H\x1b]1337;File=inline=1:%s\a", imgBase64Str)
}

func getImage(filePath string) image.Image {
	imgFile, err := os.Open(filePath)
	defer imgFile.Close()
	if err != nil {
		fmt.Println("Cannot read file:", err)
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		fmt.Println("Cannot decode file:", err)
	}
	return img
}

func getCurrentTime() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
