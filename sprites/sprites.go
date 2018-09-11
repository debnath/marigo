package sprites

import (
	"image"

	"github.com/disintegration/gift"
)

const (
	TERRAIN_HEIGHT       = 90
	MARIO_RESTING_HEIGHT = 60
	PIPE_HEIGHT          = 53
	TREE_HEIGHT          = 62
)

//sprites
var marioStand = image.Rect(39, 0, 56, 30)
var marioSquat = image.Rect(715, 10, 733, 30)
var marioJump = image.Rect(514, 0, 534, 30)
var marioRun1 = image.Rect(582, 0, 600, 30)
var marioRun2 = image.Rect(615, 0, 631, 30)
var ground = image.Rect(0, 43, 900, 50)

/** obstacles, in order of width **/
//width 1
var gPipe = image.Rect(268, 0, 286, 37) //g
var gtree = image.Rect(172, 0, 182, 27) //g

//width 2
var goPipe = image.Rect(268, 0, 304, 37) //go
var tree2 = image.Rect(172, 0, 195, 27)  //gw
var tree2b = image.Rect(196, 0, 220, 27) //Gg
var tree2c = image.Rect(220, 0, 247, 27) //WG

//width 3
var brick = image.Rect(304, 0, 361, 37)   // pipewall
var gogPipe = image.Rect(268, 0, 322, 37) //gog
var tree3 = image.Rect(172, 0, 208, 27)   //gwG
var tree3b = image.Rect(210, 0, 246, 27)  //gWG

// Sprite represents a sprite in the game
type Sprite struct {
	Region   image.Rectangle
	FilterS  *gift.GIFT //standing
	FilterC  *gift.GIFT //crouching (squat)
	FilterJ  *gift.GIFT //jumping
	FilterR1 *gift.GIFT //run1
	FilterR2 *gift.GIFT //run2
	Position image.Point
	Status   bool
	Points   int
}

var Mario = Sprite{
	//Region:   marioStand,
	FilterS:  gift.New(gift.Crop(marioStand)),
	FilterC:  gift.New(gift.Crop(marioSquat)),
	FilterJ:  gift.New(gift.Crop(marioJump)),
	FilterR1: gift.New(gift.Crop(marioRun1)),
	FilterR2: gift.New(gift.Crop(marioRun2)),
	Position: image.Pt(10, MARIO_RESTING_HEIGHT),
	Status:   true,
}

/*
	micro image: 903 x 51
	terrain: (0,43), (903,51)
*/
var Terrain = Sprite{
	Region:   ground,
	FilterS:  gift.New(gift.Crop(ground)),
	Position: image.Pt(0, TERRAIN_HEIGHT),
}

//1
var GreenPipe = Sprite{
	Region:   gPipe,
	FilterS:  gift.New(gift.Crop(gPipe)),
	Position: image.Pt(340, PIPE_HEIGHT),
	Status:   true,
}

var GreenTree = Sprite{
	Region:   gtree,
	FilterS:  gift.New(gift.Crop(gtree)),
	Position: image.Pt(350, TREE_HEIGHT),
	Status:   true,
}

//2
var GOPipe = Sprite{
	Region:   goPipe,
	FilterS:  gift.New(gift.Crop(goPipe)),
	Position: image.Pt(380, PIPE_HEIGHT),
	Status:   true,
}

var GWTrees = Sprite{
	Region:   tree2,
	FilterS:  gift.New(gift.Crop(tree2)),
	Position: image.Pt(400, TREE_HEIGHT),
	Status:   true,
}

var BGGTrees = Sprite{
	Region:   tree2b,
	FilterS:  gift.New(gift.Crop(tree2b)),
	Position: image.Pt(430, TREE_HEIGHT),
	Status:   true,
}

var BWBGTrees = Sprite{
	Region:   tree2c,
	FilterS:  gift.New(gift.Crop(tree2c)),
	Position: image.Pt(460, TREE_HEIGHT),
	Status:   true,
}

//3
var GOGPipe = Sprite{
	Region:   gogPipe,
	FilterS:  gift.New(gift.Crop(gogPipe)),
	Position: image.Pt(480, PIPE_HEIGHT),
	Status:   true,
}

var GWGTree = Sprite{
	Region:   tree3,
	FilterS:  gift.New(gift.Crop(tree3)),
	Position: image.Pt(520, TREE_HEIGHT),
	Status:   true,
}

var GBWGTree = Sprite{
	Region:   tree3b,
	FilterS:  gift.New(gift.Crop(tree3b)),
	Position: image.Pt(550, TREE_HEIGHT),
	Status:   true,
}

var BrickWall = Sprite{
	Region:   brick,
	FilterS:  gift.New(gift.Crop(brick)),
	Position: image.Pt(570, PIPE_HEIGHT),
	Status:   true,
}
