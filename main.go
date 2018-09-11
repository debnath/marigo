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
var windowWidth, windowHeight = 250, 100 //keep low as Iterm2 hack is not designed for animations... bigger resolutions = lower fps

//source images
var src = GetImage("imgs/sprite_mario_micro.png")

//var background = GetImage("imgs/bg.png")

//mario state
var marioState = MARIO_RUN1

//var obstacles []Sprite

var events = make(chan termbox.Event, 500) //Keep track of keyboard events
var refresher = make(chan bool, 500)       //for keeping track of when to redraw sprites
var gameOver = false                       // end of game
var logFps = true                          // enable a counter for number of frames rendered in a given second
var totalfps = 0                           //total number of frames rendered
var fps = map[int]int{}                    //keep track of framerate with respect to each second

//roughly the obstacles are more more difficult the further in it goes.
var obstacles = []sprites.Sprite{
	sprites.GreenPipe,
	sprites.GreenTree,
	sprites.GOPipe,   //GreenOrange Pipe
	sprites.GWTrees,  //Green White Tree
	sprites.BGGTrees, //Big Green and small Green tree, etc
	sprites.BWBGTrees,
	sprites.BrickWall,
	sprites.GOGPipe,
	sprites.GWGTree,
	sprites.GBWGTree,
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	// game variables
	score := 0 // number of points scored in the game so far

	var wg sync.WaitGroup
	wg.Add(1)
	go pollEvents(&wg) //poll for termbox events
	startScreen()      //block on start screen until user presses 's' to start or 'q' to quit

	wg.Add(1)
	go handleKeystrokes(&wg)

	wg.Add(1)
	go scrollTerrain(&wg) //shift the terrain on the X axis to simulate movement

	wg.Add(1)
	go renderSprites(&wg) //print out all sprites with respect to time

	wg.Wait()
	termbox.Close()

	if logFps {
		fmt.Println(fmt.Sprintf("Ran for %d seconds with an average framerate of %dfps", len(fps), totalfps/len(fps)))
		spew.Dump(fps)
	}

	fmt.Println("\nGAME OVER!\nFinal score:", score)
}

//maybe cleaner to use termbox's collide() function
func collide(s1, s2 sprites.Sprite) bool {
	spriteA := image.Rect(s1.Position.X, s1.Position.Y, s1.Position.X+s1.Region.Dx(), s1.Position.Y+s1.Region.Dy())
	spriteB := image.Rect(s2.Position.X, s2.Position.Y, s2.Position.X+s1.Region.Dx(), s2.Position.Y+s1.Region.Dy())
	if spriteA.Min.X < spriteB.Max.X && spriteA.Max.X > spriteB.Min.X &&
		spriteA.Min.Y < spriteB.Max.Y && spriteA.Max.Y > spriteB.Min.Y {
		return true
	}
	return false
}

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
//... so mario can only crouch for 300ms on 1 keystroke. no more, no less.
func crouch() {
	if sprites.Mario.Position.Y == sprites.MARIO_RESTING_HEIGHT {
		marioState = MARIO_SQUAT //@todo running squats... mario wants to get ripped
		sprites.Mario.Position.Y += 10
		time.Sleep(300 * time.Millisecond)
		sprites.Mario.Position.Y = sprites.MARIO_RESTING_HEIGHT
		marioState = MARIO_RUN1
	}
}

func pollEvents(wg *sync.WaitGroup) {
	for !gameOver {
		events <- termbox.PollEvent()
	}

	wg.Done()
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
		time.Sleep(100 * time.Millisecond) //with keystrokes I can get a reliable 5fps with 200ms

		//start obstacle sprites from X=300
		if sprites.Terrain.Position.X < -620 {
			sprites.Terrain.Position.X = -50
		}

		sprites.Terrain.Position.X -= 10

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

			//start obstacle sprites from X=300

			//sprites.GreenPipe.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+100, sprites.GreenPipe.Position.Y), gift.OverOperator)
			//sprites.GreenTree.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+150, sprites.GreenTree.Position.Y), gift.OverOperator)
			//sprites.GOPipe.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+200, sprites.GOPipe.Position.Y), gift.OverOperator)
			//sprites.GWTrees.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+250, sprites.GWTrees.Position.Y), gift.OverOperator)
			//sprites.BGGTrees.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+300, sprites.BGGTrees.Position.Y), gift.OverOperator)
			//sprites.BWBGTrees.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+350, sprites.BWBGTrees.Position.Y), gift.OverOperator)
			//sprites.BrickWall.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+400, sprites.BrickWall.Position.Y), gift.OverOperator)
			//sprites.GOGPipe.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+450, sprites.GOGPipe.Position.Y), gift.OverOperator)
			//sprites.GWGTree.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+500, sprites.GWGTree.Position.Y), gift.OverOperator)
			//sprites.GBWGTree.FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+550, sprites.GBWGTree.Position.Y), gift.OverOperator)

			for i := 0; i < len(obstacles); i++ {
				if obstacles[i].Status {
					obstacles[i].FilterS.DrawAt(dst, src, image.Pt(sprites.Terrain.Position.X+(obstacles[i].Position.X), obstacles[i].Position.Y), gift.OverOperator)
				}
			}

			PrintImage(dst)

			if logFps {
				if prv == getCurrentTime() {
					currentfps++
				} else {
					prv = getCurrentTime()
					fps[fpsloop] = currentfps
					totalfps = totalfps + currentfps
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
				//disabling crouch ability as *nix terminals do not have a KEYUP event, nor does holding a key down send events fast enough.
				//case termbox.KeyArrowDown:
				//	if marioState != MARIO_JUMP {
				//		go crouch()
				//		refresher <- true
				//	}
				case termbox.KeyEsc, termbox.KeyCtrlQ, termbox.KeyCtrlC:
					gameOver = true
					refresher <- true

				}
			}
		}
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
