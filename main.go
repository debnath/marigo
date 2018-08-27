package main

import (
	"fmt"
	"image"
	"math/rand"
	"os"

	"time"

	"sync"

	"image/png"

	"bytes"

	"github.com/davecgh/go-spew/spew"
	"github.com/disintegration/gift"
	"github.com/nsf/termbox-go"

	"encoding/base64"
)

const (
	MARIO_STAND = 0
	MARIO_JUMP  = 1
)

// parameters
var windowWidth, windowHeight = 250, 100

// sprites
var src = getImage("imgs/sprite_mario_micro.png")

var background = getImage("imgs/bg.png")

//var marioJump = image.Rect(1350,0, 1400, 75)
//var marioRun1 = image.Rect(1534,0, 1584, 75)
//var marioRun2 = image.Rect(1625,0	, 1675, 75)

/*
var src = getImage("imgs/sprite_mario_small.png")

var marioStand = image.Rect(50, 0, 80, 40)
var marioSquat = image.Rect(955, 0, 977, 38)
var ground = image.Rect(0, 60, 1200, 70) //up to 1200
*/

var marioStand = image.Rect(39, 0, 56, 30)  //50, 0, 80, 40
var marioSquat = image.Rect(37, 0, 60, 30)  //955, 0, 977, 38
var marioJump = image.Rect(514, 0, 534, 30) //955, 0, 977, 38
var jumpq = make(chan bool, 50)

var marioState = MARIO_STAND

var ground = image.Rect(0, 43, 600, 50)

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
	FilterA:  gift.New(gift.Crop(marioSquat)),
	FilterE:  gift.New(gift.Crop(marioJump)),
	Position: image.Pt(10, 65),
	Status:   true,
}

//1204 x 68
var terrain = Sprite{
	size:     ground,
	Filter:   gift.New(gift.Crop(ground)),
	Position: image.Pt(0, 95),
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

	var logFps = true    // enable a counter for number of frames rendered in a given second
	fps := map[int]int{} //keep track of framerate with respect to each second

	events := make(chan termbox.Event, 5000) //capture keystrokes
	refresher := make(chan bool, 5000)       //for keeping track of when to redraw sprites

	go func() {
		for {
			events <- termbox.PollEvent()
		}
	}()

	/*
		// block on the start screen
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
	*/
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for !gameOver {
			select {
			case ev := <-events:
				if ev.Type == termbox.EventKey {
					switch ev.Key {
					case termbox.KeyArrowUp:
						//mario.Position.Y -= 10
						jump()
						refresher <- true
					case termbox.KeyArrowRight:
						mario.Position.X += 10
						refresher <- true
					case termbox.KeyArrowLeft:
						mario.Position.X -= 10
						refresher <- true
					case termbox.KeyArrowDown:
						mario.Position.Y += 10
						refresher <- true
					case termbox.KeyEsc, termbox.KeyCtrlQ, termbox.KeyCtrlC:
						gameOver = true
						refresher <- true

					}
				}
			}
		}
		wg.Done()
	}()

	//render terrain
	wg.Add(1)
	go func() {
		for !gameOver {
			time.Sleep(200 * time.Millisecond) //with keystrokes I can get a reliable 6fps with 200ms
			terrain.Position.X -= 10           //20 shows it as stationary

			refresher <- true
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {

		prv := getCurrentTime()
		fpsloop := 1
		currentfps := 0

		for !gameOver {
			select {
			case <-refresher:
				dst := image.NewRGBA(image.Rect(0, 0, windowWidth, windowHeight))
				//gift.New().Draw(dst, background)

				if marioState == MARIO_JUMP {
					mario.FilterE.DrawAt(dst, src, image.Pt(mario.Position.X, mario.Position.Y), gift.OverOperator)
				} else {
					mario.Filter.DrawAt(dst, src, image.Pt(mario.Position.X, mario.Position.Y), gift.OverOperator)
				}

				terrain.Filter.DrawAt(dst, src, image.Pt(terrain.Position.X, terrain.Position.Y), gift.OverOperator)
				printImage(dst)

				if logFps {
					if prv == getCurrentTime() {
						currentfps++
					} else {
						prv = getCurrentTime()
						fps[fpsloop] = currentfps
						currentfps = 0
						fpsloop++
					}
				}
			}
		}
		wg.Done()
	}()

	wg.Wait()
	termbox.Close()

	if logFps {
		spew.Dump(fps)
	}

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
// https://stackoverflow.com/questions/29585727/how-to-display-an-image-on-windows-with-go try this out.
func printImage(img image.Image) {
	var buf bytes.Buffer
	png.Encode(&buf, img)
	imgBase64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	fmt.Printf("\x1b[2;0H\x1b]1337;File=inline=1:%s\a", imgBase64Str)
	//fmt.Printf("\033]1337;File=inline=1:%s\a", imgBase64Str)
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
	return time.Now().UnixNano() / int64(time.Second)
}

//there is probably a less messy way to do this...
func jump() {
	go func() {
		marioState = MARIO_JUMP
		mario.size = marioJump
		time.Sleep(200 * time.Millisecond)
		mario.Position.Y -= 25
		jumpq <- true
		time.Sleep(200 * time.Millisecond)
		mario.Position.Y -= 20
		jumpq <- true
		time.Sleep(200 * time.Millisecond)
		mario.Position.Y -= 13
		jumpq <- true
		time.Sleep(200 * time.Millisecond)
		mario.Position.Y -= 4
		jumpq <- true
		time.Sleep(200 * time.Millisecond)
		mario.Position.Y += 4
		jumpq <- true
		time.Sleep(200 * time.Millisecond)
		mario.Position.Y += 13
		jumpq <- true
		time.Sleep(200 * time.Millisecond)
		mario.Position.Y += 20
		jumpq <- true
		time.Sleep(200 * time.Millisecond)
		mario.Position.Y += 25
		jumpq <- true
		marioState = MARIO_STAND
	}()
}
