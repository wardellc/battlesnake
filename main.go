package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
)

type Game struct {
	ID      string `json:"id"`
	Timeout int32  `json:"timeout"`
}

type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Battlesnake struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Health int32   `json:"health"`
	Body   []Coord `json:"body"`
	Head   Coord   `json:"head"`
	Length int32   `json:"length"`
	Shout  string  `json:"shout"`
}

type Board struct {
	Height int           `json:"height"`
	Width  int           `json:"width"`
	Food   []Coord       `json:"food"`
	Snakes []Battlesnake `json:"snakes"`
}

type BattlesnakeInfoResponse struct {
	APIVersion string `json:"apiversion"`
	Author     string `json:"author"`
	Color      string `json:"color"`
	Head       string `json:"head"`
	Tail       string `json:"tail"`
}

type GameRequest struct {
	Game  Game        `json:"game"`
	Turn  int         `json:"turn"`
	Board Board       `json:"board"`
	You   Battlesnake `json:"you"`
}

type MoveResponse struct {
	Move  string `json:"move"`
	Shout string `json:"shout,omitempty"`
}

// HandleIndex is called when your Battlesnake is created and refreshed
// by play.battlesnake.com. BattlesnakeInfoResponse contains information about
// your Battlesnake, including what it should look like on the game board.
func HandleIndex(w http.ResponseWriter, r *http.Request) {
	response := BattlesnakeInfoResponse{
		APIVersion: "1",
		Author:     "",          // TODO: Your Battlesnake username
		Color:      "#FFA500",   // TODO: Personalize
		Head:       "safe",      // TODO: Personalize
		Tail:       "round-bum", // TODO: Personalize
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Fatal(err)
	}
}

func isPartOfSnake(x int, y int, populatedBoard *[][]string) bool {
	partsOfSnake := [...]string{"body", "head", "myBody", "myHead"}
	for _, part := range partsOfSnake {
		if (*populatedBoard)[x][y] == part {
			return true
		}
	}
	return false
}

func getValidMoves(myBattleSnake Battlesnake, board Board, populatedBoard *[][]string) (possibleMoves []string) {

	if myBattleSnake.Head.X < (board.Width - 1) {
		if !isPartOfSnake(myBattleSnake.Head.X+1, myBattleSnake.Head.Y, populatedBoard) {
			possibleMoves = append(possibleMoves, "right")
		}
	}
	if myBattleSnake.Head.X > 0 {
		if !isPartOfSnake(myBattleSnake.Head.X-1, myBattleSnake.Head.Y, populatedBoard) {
			possibleMoves = append(possibleMoves, "left")
		}
	}
	if myBattleSnake.Head.Y < (board.Height - 1) {
		if !isPartOfSnake(myBattleSnake.Head.X, myBattleSnake.Head.Y+1, populatedBoard) {
			possibleMoves = append(possibleMoves, "up")
		}
	}
	if myBattleSnake.Head.Y > 0 {
		if !isPartOfSnake(myBattleSnake.Head.X, myBattleSnake.Head.Y-1, populatedBoard) {
			possibleMoves = append(possibleMoves, "down")
		}
	}

	// Doesn't matter - means it's cornered
	// if len(possibleMoves) == 0 {
	// 	possibleMoves = append(possibleMoves, "up")
	// }

	fmt.Println("Possible moves:", possibleMoves)
	return
}

func populateBoardWithSnakes(populatedBoard *[][]string, myBattleSnake Battlesnake, board Board) {
	for _, coord := range myBattleSnake.Body {
		(*populatedBoard)[coord.X][coord.Y] = "myBody"
	}
	(*populatedBoard)[myBattleSnake.Head.X][myBattleSnake.Head.Y] = "myHead"

	for _, snake := range board.Snakes {
		for _, coord := range snake.Body {
			(*populatedBoard)[coord.X][coord.Y] = "body"
		}
		(*populatedBoard)[snake.Head.X][snake.Head.Y] = "head"
	}
}

// func getForwardDirection(myBattleSnake Battlesnake) string {
// 	if myBattleSnake.Length > 1 {
// 		head := myBattleSnake.Body[0]
// 		neck := myBattleSnake.Body[1]
// 		if head.X > neck.X {
// 			return "right"
// 		}
// 		if head.X < neck.X {
// 			return "left"
// 		}
// 		if head.Y > neck.Y {
// 			return "up"
// 		}
// 		if head.Y < neck.Y {
// 			return "down"
// 		}
// 	}
// 	return ""
// }

// HandleStart is called at the start of each game your Battlesnake is playing.
// The GameRequest object contains information about the game that's about to start.
// TODO: Use this function to decide how your Battlesnake is going to look on the board.
func HandleStart(w http.ResponseWriter, r *http.Request) {
	request := GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}

	// Nothing to respond with here
	fmt.Print("START\n")
}

func isWithinBounds(x int, y int, width int, height int) bool {
	if x < 0 || y < 0 {
		return false
	}
	if x > width-1 {
		return false
	}
	if y > height-1 {
		return false
	}
	return true
}

func absoluteValue(i int) int {
	if i < 0 {
		return i * -1
	}
	return i
}

func populateScoreBoard(scoreBoard *[][]int, board *Board, populatedBoard *[][]string, mySnake *Battlesnake) {
	// Initialise with zeroes
	for i := 0; i < board.Width; i++ {
		for j := 0; j < board.Height; j++ {
			(*scoreBoard)[i][j] = 0
		}
	}

	allFoodCoords := (*board).Food
	maxFoodRadius := 5
	// maxFoodRadius := int(math.Round(float64((*board).Height)))
	startFoodPoints := int(math.Pow(2, float64(maxFoodRadius*2)))

	// Food scoring - max score is board 2 ^ (width * height)
	// This is the score for a square which contains food
	// As distance increases by one, the score gets halved
	// This means up, down, left, right score is halved (not diagnols as these are 2 steps away)
	for _, food := range allFoodCoords {
		iStart := food.X - maxFoodRadius
		iEnd := food.X + maxFoodRadius
		jStart := food.Y - maxFoodRadius
		jEnd := food.Y + maxFoodRadius
		for i := iStart; i <= iEnd; i++ {
			for j := jStart; j <= jEnd; j++ {
				if isWithinBounds(i, j, (*board).Width, (*board).Height) {
					distance := absoluteValue(food.X-i) + absoluteValue(food.Y-j)
					scoreToAdd := 0
					if distance == 0 {
						scoreToAdd += startFoodPoints
					} else {
						// halves score for every step away from food
						scoreToAdd += startFoodPoints / int(math.Pow(2, float64(distance)))
					}
					(*scoreBoard)[i][j] = scoreToAdd

				}
			}
		}
	}

	mySnakeID := (*mySnake).ID
	for _, snake := range (*board).Snakes {
		for _, bodyCoord := range snake.Body {
			(*scoreBoard)[bodyCoord.X][bodyCoord.Y] = 0
		}

		// avoid head of other snakes
		if snake.ID != mySnakeID {
			for i := snake.Head.X - 1; i <= snake.Head.X+1; i++ {
				for j := snake.Head.Y - 1; j <= snake.Head.Y+1; j++ {
					if isWithinBounds(i, j, (*board).Width, (*board).Height) {
						(*scoreBoard)[i][j] = 0
					}
				}
			}
		}
	}

	fmt.Println("After snake head avoidance")
	for j := board.Height - 1; j >= 0; j-- {
		for i := 0; i < board.Width; i++ {
			fmt.Print((*scoreBoard)[i][j], "\t")
		}
		fmt.Print("\n")
	}

	// Halve score of square for every adjacent square that is:
	// out of bounds
	// part of a snake
	for i := 0; i < board.Width; i++ {
		for j := 0; j < board.Height; j++ {
			adjacentSquares := []Coord{{X: -1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 0}, {X: 0, Y: -1}}
			numberSidesBlocked := 0
			for _, adjacentSquare := range adjacentSquares {
				x := i + adjacentSquare.X
				y := j + adjacentSquare.Y
				// check that isn't where head is - alter number sides blocked
				if !isWithinBounds(x, y, board.Width, board.Height) || isPartOfSnake(x, y, populatedBoard) {
					(*scoreBoard)[i][j] /= 2
					numberSidesBlocked++
				}
			}
			if numberSidesBlocked == 4 {
				(*scoreBoard)[i][j] = 0
			}

		}
	}

	fmt.Println("After out of bound/snake decrease")
	for j := board.Height - 1; j >= 0; j-- {
		for i := 0; i < board.Width; i++ {
			fmt.Print((*scoreBoard)[i][j], "\t")
		}
		fmt.Print("\n")
	}
}

func getNextMove(scoreBoard *[][]int, possibleMoves []string, head Coord) (highestScoringDirection string) {
	// Doesn't matter if no spaces available
	if len(possibleMoves) == 0 {
		return "up"
	}
	// Assume first possible move is highest scoring
	currentHighestScore := -1
	for _, move := range possibleMoves {
		var nextMoveCoords Coord
		switch move {
		case "up":
			nextMoveCoords = Coord{X: head.X, Y: head.Y + 1}
		case "down":
			nextMoveCoords = Coord{X: head.X, Y: head.Y - 1}
		case "left":
			nextMoveCoords = Coord{X: head.X - 1, Y: head.Y}
		case "right":
			nextMoveCoords = Coord{X: head.X + 1, Y: head.Y}

		}
		fmt.Printf("Current direction: %s  Current score: %d  Test direction: %s  Test score: %d\n", highestScoringDirection, currentHighestScore, move, (*scoreBoard)[nextMoveCoords.X][nextMoveCoords.Y])

		if (*scoreBoard)[nextMoveCoords.X][nextMoveCoords.Y] > currentHighestScore {
			currentHighestScore = (*scoreBoard)[nextMoveCoords.X][nextMoveCoords.Y]
			highestScoringDirection = move
		}

	}

	return

}

// HandleMove is called for each turn of each game.
// Valid responses are "up", "down", "left", or "right".
// TODO: Use the information in the GameRequest object to determine your next move.
func HandleMove(w http.ResponseWriter, r *http.Request) {
	request := GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}

	populatedBoard := make([][]string, request.Board.Width)
	for index := range populatedBoard {
		populatedBoard[index] = make([]string, request.Board.Height)
	}

	scoreBoard := make([][]int, request.Board.Width)
	for index := range scoreBoard {
		scoreBoard[index] = make([]int, request.Board.Height)
	}
	populateBoardWithSnakes(&populatedBoard, request.You, request.Board)
	// forwardDirection := getForwardDirection(request.You)
	// Choose a random direction to move in

	// possibleMoves is anything in-bounds and not a snake
	possibleMoves := getValidMoves(request.You, request.Board, &populatedBoard)

	// Get random move then check if there's a more sensible one
	// move := possibleMoves[rand.Intn(len(possibleMoves))]
	populateScoreBoard(&scoreBoard, &request.Board, &populatedBoard, &request.You)
	fmt.Println("Populated score board")
	move := getNextMove(&scoreBoard, possibleMoves, request.You.Head)

	// Check if can continue forward
	// for _, possibleMove := range possibleMoves {
	// 	if possibleMove == forwardDirection {
	// 		move = possibleMove
	// 		break
	// 	}
	// }
	// []string{"up", "down", "left", "right"}

	response := MoveResponse{
		Move: move,
	}

	fmt.Printf("MOVE: %s\n", response.Move)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Fatal(err)
	}
}

// HandleEnd is called when a game your Battlesnake was playing has ended.
// It's purely for informational purposes, no response required.
func HandleEnd(w http.ResponseWriter, r *http.Request) {
	request := GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}

	// Nothing to respond with here
	fmt.Print("END\n")
}

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	http.HandleFunc("/", HandleIndex)
	http.HandleFunc("/start", HandleStart)
	http.HandleFunc("/move", HandleMove)
	http.HandleFunc("/end", HandleEnd)

	fmt.Printf("Starting Battlesnake Server at http://0.0.0.0:%s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
