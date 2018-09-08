package main

import (
	"fmt"
	"image"
	"math/rand"
	"sync"
	"time"

	"bytes"
	"encoding/base64"
	"image/png"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/debnath/marigo/sprites"
	"github.com/disintegration/gift"
	"github.com/nsf/termbox-go"
)

const (
	MARIO_STAND = 0
	MARIO_JUMP  = 1
	MARIO_SQUAT = 2
	MARIO_RUN1  = 3
	MARIO_RUN2  = 4
)

//parameters
var windowWidth, windowHeight = 250, 100

//obstacles := map[int][]Sprite {
//	1: {
//
//	},
//}

//source images
var src = GetImage("imgs/sprite_mario_micro.png")
var background = GetImage("imgs/bg.png")

//mario state
var marioState = MARIO_RUN1

//var obstacles []Sprite

var events = make(chan termbox.Event, 100)
var refresher = make(chan bool, 100) //for keeping track of when to redraw sprites

var gameOver = false // end of game
var logFps = true    // enable a counter for number of frames rendered in a given second

var fps = map[int]int{} //keep track of framerate with respect to each second

var x_dist = 0

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

////maybe cleaner to use termbox's collide() function
//func collide(s1, s2 sprites.Sprite) bool {
//	spriteA := image.Rect(s1.Position.X, s1.Position.Y, s1.Position.X+s1.Region.Dx(), s1.Position.Y+s1.Region.Dy())
//	spriteB := image.Rect(s2.Position.X, s2.Position.Y, s2.Position.X+s1.Region.Dx(), s2.Position.Y+s1.Region.Dy())
//	if spriteA.Min.X < spriteB.Max.X && spriteA.Max.X > spriteB.Min.X &&
//		spriteA.Min.Y < spriteB.Max.Y && spriteA.Max.Y > spriteB.Min.Y {
//		return true
//	}
//	return false
//}

func getCurrentTime() int64 {
	return time.Now().UnixNano() / int64(time.Second)
}

//use maths and make this nicer...
func jump() {
	marioState = MARIO_JUMP
	time.Sleep(200 * time.Millisecond)
	sprites.Mario.Position.Y -= 25
	time.Sleep(200 * time.Millisecond)
	sprites.Mario.Position.Y -= 20
	time.Sleep(200 * time.Millisecond)
	sprites.Mario.Position.Y -= 13
	time.Sleep(200 * time.Millisecond)
	sprites.Mario.Position.Y -= 4
	time.Sleep(200 * time.Millisecond)
	sprites.Mario.Position.Y += 4
	time.Sleep(200 * time.Millisecond)
	sprites.Mario.Position.Y += 13
	time.Sleep(200 * time.Millisecond)
	sprites.Mario.Position.Y += 20
	time.Sleep(200 * time.Millisecond)
	sprites.Mario.Position.Y += 25
	marioState = MARIO_RUN1
}

//*nix terminals do not have a keyup event, nor does it repeat keystrokes when held down.
//... so mario can only squat for 300ms on 1 keystroke. no more, no less.
func squat() {
	if sprites.Mario.Position.Y == sprites.MARIO_RESTING_HEIGHT {
		marioState = MARIO_SQUAT //@todo running squats... mario wants to get ripped
		sprites.Mario.Position.Y += 10
		time.Sleep(300 * time.Millisecond)
		sprites.Mario.Position.Y = sprites.MARIO_RESTING_HEIGHT
		marioState = MARIO_RUN1
	}
}

func pollEvents() {
	for {
		events <- termbox.PollEvent()
	}
}

func startScreen() {
	startScreen := GetImage("imgs/start_micro.png")
	PrintImage(startScreen)
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
		sprites.Terrain.Position.X -= 10   //20 shows it as stationary

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
			dst := image.NewRGBA(image.Rect(-50, -125, windowWidth, windowHeight))
			//gift.New().Draw(dst, background)

			//loggedMotions = append(loggedMotions, marioState)
			if marioState == MARIO_JUMP {
				sprites.Mario.FilterJ.DrawAt(dst, src, image.Pt(sprites.Mario.Position.X, sprites.Mario.Position.Y), gift.OverOperator)
			} else if marioState == MARIO_SQUAT {
				sprites.Mario.FilterC.DrawAt(dst, src, image.Pt(sprites.Mario.Position.X, sprites.Mario.Position.Y), gift.OverOperator)
			} else if marioState == MARIO_RUN1 {
				marioState = MARIO_RUN2
				sprites.Mario.FilterR2.DrawAt(dst, src, image.Pt(sprites.Mario.Position.X, sprites.Mario.Position.Y), gift.OverOperator)
			} else { //if marioState == MARIO_RUN2 {
				marioState = MARIO_RUN1
				sprites.Mario.FilterR1.DrawAt(dst, src, image.Pt(sprites.Mario.Position.X, sprites.Mario.Position.Y), gift.OverOperator)
			}

			sprites.Terrain.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X, sprites.Terrain.Position.Y), gift.OverOperator)

			sprites.GreenPipe.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+100, sprites.GreenPipe.Position.Y), gift.OverOperator)
			sprites.GreenTree.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+150, sprites.GreenTree.Position.Y), gift.OverOperator)
			sprites.GreenOrangePipe.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+200, sprites.GreenOrangePipe.Position.Y), gift.OverOperator)
			sprites.GreenWhiteTrees.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+250, sprites.GreenWhiteTrees.Position.Y), gift.OverOperator)
			sprites.BGreenGreenTrees.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+300, sprites.BGreenGreenTrees.Position.Y), gift.OverOperator)
			sprites.BWhiteBGreenTrees.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+350, sprites.BWhiteBGreenTrees.Position.Y), gift.OverOperator)
			sprites.BrickWall.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+400, sprites.BrickWall.Position.Y), gift.OverOperator)
			sprites.GreenOrangeGreenPipe.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+450, sprites.GreenOrangeGreenPipe.Position.Y), gift.OverOperator)
			sprites.GreenWhiteGreenTree.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+500, sprites.GreenWhiteGreenTree.Position.Y), gift.OverOperator)
			sprites.GreenBWhiteGreenTree.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+550, sprites.GreenBWhiteGreenTree.Position.Y), gift.OverOperator)

			PrintImage(dst)

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

func renderObstacles(wg *sync.WaitGroup) {
	for !gameOver {

		time.Sleep(1 * time.Second) //with keystrokes I can get a reliable 5fps with 200ms
		//terrain.Position.X -= 10    //20 shows it as stationary

		refresher <- true
	}
	wg.Done()
}

// this only works for iTerm2!
// https://stackoverflow.com/questions/29585727/how-to-display-an-image-on-windows-with-go try this out.
func PrintImage(img image.Image) {
	var buf bytes.Buffer
	png.Encode(&buf, img)
	imgBase64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	fmt.Printf("\x1b[2;0H\x1b]1337;File=inline=1:%s\a", imgBase64Str)
}

func GetImage(filePath string) image.Image {
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
