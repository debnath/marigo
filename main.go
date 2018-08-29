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
	MARIO_SQUAT = 2
	MARIO_RUN1  = 3
	MARIO_RUN2  = 4

	TERRAIN_HEIGHT       = 90
	MARIO_RESTING_HEIGHT = 60
)

// parameters
var windowWidth, windowHeight = 250, 100

// sprites
var src = getImage("imgs/sprite_mario_micro.png")
var background = getImage("imgs/bg.png")

//sprites
var marioStand = image.Rect(39, 0, 56, 30)    //50, 0, 80, 40
var marioSquat = image.Rect(715, 10, 733, 30) //955, 0, 977, 38
var marioJump = image.Rect(514, 0, 534, 30)   //955, 0, 977, 38
var marioRun1 = image.Rect(582, 0, 600, 30)   //955, 0, 977, 38
var marioRun2 = image.Rect(615, 0, 631, 30)   //955, 0, 977, 38
var ground = image.Rect(0, 43, 600, 50)

//mario state
var marioState = MARIO_RUN1

// Sprite represents a sprite in the game
type Sprite struct {
	size        image.Rectangle
	StandFilter *gift.GIFT
	SquatFilter *gift.GIFT
	JumpFilter  *gift.GIFT
	Run1Filter  *gift.GIFT
	Run2Filter  *gift.GIFT
	Position    image.Point
	Status      bool
	Points      int
}

var mario = Sprite{
	size:        marioStand,
	StandFilter: gift.New(gift.Crop(marioStand)),
	SquatFilter: gift.New(gift.Crop(marioSquat)),
	JumpFilter:  gift.New(gift.Crop(marioJump)),
	Run1Filter:  gift.New(gift.Crop(marioRun1)),
	Run2Filter:  gift.New(gift.Crop(marioRun2)),
	Position:    image.Pt(10, MARIO_RESTING_HEIGHT),
	Status:      true,
}

//1204 x 68
var terrain = Sprite{
	size:        ground,
	StandFilter: gift.New(gift.Crop(ground)),
	Position:    image.Pt(0, TERRAIN_HEIGHT),
	Status:      true,
}

var events = make(chan termbox.Event, 100)
var refresher = make(chan bool, 100) //for keeping track of when to redraw sprites

var gameOver = false // end of game
var logFps = true    // enable a counter for number of frames rendered in a given second

var fps = map[int]int{} //keep track of framerate with respect to each second

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	// game variables
	score := 0 // number of points scored in the game so far

	//var loggedMotions []int

	go pollEvents() //poll for termbox events
	startScreen()   //block on start screen until user presses 's' to start or 'q' to quit

	var wg sync.WaitGroup
	wg.Add(1)
	go handleKeystrokes(&wg)

	wg.Add(1)
	go scrollTerrain(&wg) //shift the terrain on the X axis to simulate movement

	wg.Add(1)
	go renderSprites(&wg) //print out all sprites with respect to time

	wg.Wait()
	termbox.Close()

	if logFps {
		spew.Dump(fps)
		//spew.Dump(loggedMotions)
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

//use maths and make this nicer...
func jump() {
	marioState = MARIO_JUMP
	time.Sleep(200 * time.Millisecond)
	mario.Position.Y -= 25
	time.Sleep(200 * time.Millisecond)
	mario.Position.Y -= 20
	time.Sleep(200 * time.Millisecond)
	mario.Position.Y -= 13
	time.Sleep(200 * time.Millisecond)
	mario.Position.Y -= 4
	time.Sleep(200 * time.Millisecond)
	mario.Position.Y += 4
	time.Sleep(200 * time.Millisecond)
	mario.Position.Y += 13
	time.Sleep(200 * time.Millisecond)
	mario.Position.Y += 20
	time.Sleep(200 * time.Millisecond)
	mario.Position.Y += 25
	marioState = MARIO_RUN1
}

//*nix terminals do not have a keyup event, nor does it repeat keystrokes when held down.
//... so mario can only squat for 300ms on 1 keystroke. no more, no less.
func squat() {
	if mario.Position.Y == MARIO_RESTING_HEIGHT {
		marioState = MARIO_SQUAT //@todo running squats... mario wants to get ripped
		mario.Position.Y += 10
		time.Sleep(300 * time.Millisecond)
		mario.Position.Y = MARIO_RESTING_HEIGHT
		marioState = MARIO_RUN1
	}
}

func pollEvents() {
	for {
		events <- termbox.PollEvent()
	}
}

func startScreen() {
	startScreen := getImage("imgs/start_micro.png")
	printImage(startScreen)
start:
	for { //start screen loop
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
}

func scrollTerrain(wg *sync.WaitGroup) {
	for !gameOver {
		time.Sleep(200 * time.Millisecond) //with keystrokes I can get a reliable 5fps with 200ms
		terrain.Position.X -= 10           //20 shows it as stationary

		refresher <- true
	}
	wg.Done()
}

func renderSprites(wg *sync.WaitGroup) {
	prv := getCurrentTime()
	fpsloop := 1
	currentfps := 0

	for !gameOver {
		select {
		case <-refresher:
			dst := image.NewRGBA(image.Rect(0, 0, windowWidth, windowHeight))
			//gift.New().Draw(dst, background)

			//loggedMotions = append(loggedMotions, marioState)
			if marioState == MARIO_JUMP {
				mario.JumpFilter.DrawAt(dst, src, image.Pt(mario.Position.X, mario.Position.Y), gift.OverOperator)
			} else if marioState == MARIO_SQUAT {
				mario.SquatFilter.DrawAt(dst, src, image.Pt(mario.Position.X, mario.Position.Y), gift.OverOperator)
			} else if marioState == MARIO_RUN1 {
				marioState = MARIO_RUN2
				mario.Run2Filter.DrawAt(dst, src, image.Pt(mario.Position.X, mario.Position.Y), gift.OverOperator)
			} else { //if marioState == MARIO_RUN2 {
				marioState = MARIO_RUN1
				mario.Run1Filter.DrawAt(dst, src, image.Pt(mario.Position.X, mario.Position.Y), gift.OverOperator)
			}

			terrain.StandFilter.DrawAt(dst, src, image.Pt(terrain.Position.X, terrain.Position.Y), gift.OverOperator)
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

}

func handleKeystrokes(wg *sync.WaitGroup) {
	for !gameOver {
		select {
		case ev := <-events:
			if ev.Type == termbox.EventKey {
				switch ev.Key {
				case termbox.KeyArrowUp:
					if marioState == MARIO_RUN1 || marioState == MARIO_RUN2 {
						go jump()
						refresher <- true
					}
				case termbox.KeyArrowDown:
					if marioState != MARIO_JUMP {
						go squat()
						refresher <- true
					}
					//mario.Position.Y += 10
					//refresher <- true
				case termbox.KeyEsc, termbox.KeyCtrlQ, termbox.KeyCtrlC:
					gameOver = true
					refresher <- true

				}
			}
		}
	}
	wg.Done()
}
