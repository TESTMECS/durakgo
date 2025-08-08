package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TableCards struct {
	c Card
}
type Player int

// This Struct is the display state.
type Game struct {
	player1Hand []Card
	player2Hand []Card
	table       []TableCards
	engine      *Engine
	cursor      int
	winner      Player
	turn        Player
	attacker    Player
	gameover    bool
	colors      map[string]lipgloss.Style
	trump       Suit
	deck        []Card
}
type Board struct {
	PlayerHand   []Card
	OpponentHand []Card
	Table        []TableCards
	TrumpSuit    Suit
	Attacker     Player
	Deck         []Card
}

// TODO: Use;
const (
	size   = 3
	yellow = "#FF9E3B"
	dark   = "#3C3A32"
	gray   = "#717C7C"
	light  = "#DCD7BA"
	red    = "#E63D3D"
	green  = "#98BB6C"
	blue   = "#7E9CD8"
)

// Initialize Game Shuffles the deck, creates the player's hand
// Creates the Opponent's Hand and Creates the Game Struct
func initialGame() *Game {
	log.Println("Initializing game...")
	deck := NewDeck()
	ShuffleDeck(deck)

	trumpCard := deck[len(deck)-1]

	player1Hand := deck[:6]
	deck = deck[6:]

	player2Hand := deck[:6]
	deck = deck[6:]

	return &Game{
		player1Hand: player1Hand,
		player2Hand: player2Hand,
		table:       []TableCards{},
		cursor:      0,
		engine:      NewEngine(),
		turn:        0, // Player starts as mover
		attacker:    0, // Player starts as attacker
		trump:       trumpCard.Suit,
		deck:        deck,
	}
}

// Tea Init
func (g *Game) Init() tea.Cmd {
	return tea.SetWindowTitle("Durak")
}

// Communication Between Engine and Player
type passTurnToAI struct{}
type aiFinishedTurn struct{}

// Create a simple board to pass to the Engine.
func (g *Game) ToBoard() *Board {
	return &Board{
		PlayerHand:   g.player1Hand,
		OpponentHand: g.player2Hand,
		Table:        g.table,
		TrumpSuit:    g.trump,
		Attacker:     g.attacker,
		Deck:         g.deck,
	}
}

// FromBoard Updates the game state from the board
func (g *Game) FromBoard(b *Board) {
	g.player1Hand = b.PlayerHand
	g.player2Hand = b.OpponentHand
	g.table = b.Table
	g.attacker = b.Attacker
	g.deck = b.Deck
}

// Called after passTurnToAI
func aiMove(g *Game) tea.Cmd {
	return func() tea.Msg {
		board := g.ToBoard()
		g.engine.AITurn(board)
		g.FromBoard(board)
		isover, win := g.engine.CheckGameOver(g.ToBoard())
		if isover {
			g.gameover = true
			g.winner = win
		} else {
			// After AI move, it's player's turn
			g.turn = 0
		}
		return aiFinishedTurn{}
	}
}

// Main Logic Update function
func (g *Game) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case passTurnToAI:
		log.Println("Ai turn in progress...")
		time.Sleep(time.Second * 1)
		return g, aiMove(g)
	case aiFinishedTurn:
		return g, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return g, tea.Quit
		case "right", "l":
			if g.cursor < len(g.player1Hand)-1 {
				g.cursor++
			}
		case "left", "h":
			if g.cursor > 0 {
				g.cursor--
			}
		case "p": // Player passes attack
			if g.turn == 0 && g.attacker == 0 && len(g.table) > 0 {
				// Player is attacker and passes. Round ends.
				board := g.ToBoard()
				board.Table = []TableCards{}
				g.engine.DrawCards(board)
				g.FromBoard(board) // To get drawn cards
				g.attacker = 1     // AI becomes new attacker
				g.turn = 1         // AI's turn to attack
				return g, func() tea.Msg { return passTurnToAI{} }
			}
		case "t": // Player takes cards
			if g.turn == 0 && g.attacker == 1 { // Player is defending
				board := g.ToBoard()
				// Player takes all cards from table
				for _, tc := range board.Table {
					board.PlayerHand = append(board.PlayerHand, tc.c)
				}
				board.Table = []TableCards{}
				g.engine.DrawCards(board)
				g.FromBoard(board)
				// Attacker (AI) gets to attack again.
				g.attacker = 1
				g.turn = 1
				return g, func() tea.Msg { return passTurnToAI{} }
			}
		case " ", "enter":
			if g.turn == 0 { // Player's turn to move
				if g.attacker == 0 { // Player is attacking
					if len(g.player1Hand) > 0 {
						cardToAdd := g.player1Hand[g.cursor]
						canPlay := false
						if len(g.table) == 0 {
							canPlay = true
						} else {
							tableRanks := make(map[Rank]bool)
							for _, tableCard := range g.table {
								tableRanks[tableCard.c.Rank] = true
							}
							if tableRanks[cardToAdd.Rank] {
								canPlay = true
							}
						}

						if canPlay {
							g.table = append(g.table, TableCards{c: cardToAdd})
							g.player1Hand = append(g.player1Hand[:g.cursor], g.player1Hand[g.cursor+1:]...)
							if g.cursor >= len(g.player1Hand) && len(g.player1Hand) > 0 {
								g.cursor = len(g.player1Hand) - 1
							}
							g.turn = 1 // AI's turn to defend
							return g, func() tea.Msg { return passTurnToAI{} }
						}
					}
				} else { // Player is defending
					if len(g.player1Hand) > 0 && len(g.table) > 0 {
						attackingCard := g.table[len(g.table)-1].c
						defendingCard := g.player1Hand[g.cursor]
						if g.engine.CanBeat(attackingCard, defendingCard, g.trump) {
							// Player defends successfully, round ends.
							g.table = []TableCards{}
							g.player1Hand = append(g.player1Hand[:g.cursor], g.player1Hand[g.cursor+1:]...)
							if g.cursor >= len(g.player1Hand) && len(g.player1Hand) > 0 {
								g.cursor = len(g.player1Hand) - 1
							}

							board := g.ToBoard()
							board.PlayerHand = g.player1Hand
							board.Table = g.table
							g.engine.DrawCards(board)
							g.FromBoard(board)

							g.attacker = 0 // Player becomes the new attacker
							g.turn = 0     // It's player's turn to attack
							return g, nil
						}
					}
				}
			}
		}
	}
	return g, nil
}

func extractCards(table []TableCards) []Card {
	cards := make([]Card, len(table))
	for i, tc := range table {
		cards[i] = tc.c
	}
	return cards
}

// renderCards renders cards side-by-side, highlighting the selected card.
func renderCards(cards []Card, cursor int, selected bool) string {
	if len(cards) == 0 {
		return ""
	}
	var lines [6]string
	for i, card := range cards {
		rank := card.Rank.String()
		suit := card.Suit.String()

		lines[0] += "┌─────┐ "
		lines[1] += fmt.Sprintf("│%-2s   │ ", rank)
		lines[2] += fmt.Sprintf("│  %s  │ ", suit)
		lines[3] += fmt.Sprintf("│   %2s│ ", rank)
		lines[4] += "└─────┘ "
		if selected && i == cursor {
			lines[5] += "   ^   "
		} else {
			lines[5] += "       "
		}
	}
	return strings.Join(lines[:], "\n")
}

func (g *Game) View() string {
	var s strings.Builder

	if g.gameover {
		s.WriteString("Game Over!\n")
		switch g.winner {
		case 0:
			s.WriteString("You win!\n")
		case 1:
			s.WriteString("You lose!\n")
		default:
			s.WriteString("It's a draw!\n")
		}
		return s.String()
	}

	s.WriteString("Player 2's hand:\n")
	s.WriteString(renderCards(g.player2Hand, -1, false)) // No cursor for player 2
	s.WriteString("\n\n")

	s.WriteString("Table:\n")
	if len(g.table) == 0 {
		s.WriteString("[empty]\n")
	} else {
		s.WriteString(renderCards(extractCards(g.table), -1, false)) // No cursor for table
	}
	s.WriteString("\n\n")

	s.WriteString("Player 1's hand:\n")
	s.WriteString(renderCards(g.player1Hand, g.cursor, true))
	s.WriteString("\n\n")

	s.WriteString(fmt.Sprintf("Trump suit: %s\n", g.trump.String()))
	if g.turn == 0 {
		if g.attacker == 0 {
			if len(g.table) > 0 {
				s.WriteString("Your turn to continue attack. (space/enter to play, 'p' to pass)\n")
			} else {
				s.WriteString("Your turn to attack.\n")
			}
		} else {
			s.WriteString("Your turn to defend. (space/enter to play, 't' to take)\n")
		}
	} else {
		if g.attacker == 1 {
			s.WriteString("AI is attacking...\n")
		} else {
			s.WriteString("AI is defending...\n")
		}
	}

	s.WriteString("Use left/right arrows to move, 'q' to quit.")

	return s.String()
}

func main() {
	// Logging
	f, err := LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	// Game
	p := tea.NewProgram(initialGame())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}