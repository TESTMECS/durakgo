package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Table struct {
	Attack  Card
	Defense Card
}
type Player int // 0 = player 1 = Computer
type Game struct {
	deck        []Card
	player1Hand []Card
	player2Hand []Card
	table       []Card
	engine      *Engine
	cursor      int
	winner      Player
	turn        Player // Who is the attacker
	gameover    bool
	colors      map[string]lipgloss.Style
	trump       Suit
}

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

func initialGame() Game {
	deck := NewDeck()
	ShuffleDeck(deck)

	trumpCard := deck[len(deck)-1]

	player1Hand := deck[:6]
	deck = deck[6:]

	player2Hand := deck[:6]
	deck = deck[6:]

	return Game{
		deck:        deck,
		player1Hand: player1Hand,
		player2Hand: player2Hand,
		table:       []Card{},
		cursor:      0,
		engine:      NewEngine(1000),
		turn:        0, // Player starts as attacker
		trump:       trumpCard.Suit,
	}
}

func (g Game) Init() tea.Cmd {
	return tea.SetWindowTitle("Durak")
}

type passTurnToAI struct{}
type aiFinishedTurn struct{}

func (g *Game) ToBoard() *Board {
	return &Board{
		PlayerHand:   g.player1Hand,
		OpponentHand: g.player2Hand,
		Table:        g.table,
		Deck:         g.deck,
		TrumpSuit:    g.trump,
		Attacker:     g.turn,
	}
}

func (g *Game) FromBoard(b *Board) {
	g.player1Hand = b.PlayerHand
	g.player2Hand = b.OpponentHand
	g.table = b.Table
	g.deck = b.Deck
	g.turn = b.Attacker
}

func aiMove(g *Game) tea.Cmd {
	return func() tea.Msg {
		board := g.ToBoard()
		updatedBoard := g.engine.HandleAITurn(board)
		g.FromBoard(updatedBoard)

		isover, win := g.engine.CheckGameOver(g.ToBoard(), Move{})
		if isover {
			g.gameover = true
			g.winner = win
		}

		// If AI is now the attacker, it needs to make another move (an attack).
		if g.turn == 1 && len(g.table) == 0 {
			return passTurnToAI{}
		}

		return aiFinishedTurn{}
	}
}

func (g Game) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case passTurnToAI:
		time.Sleep(time.Second * 1)
		return g, aiMove(&g)
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
		case " ", "enter":
			// Player attacks
			if g.turn == 0 && len(g.table) == 0 {
				if len(g.player1Hand) > 0 {
					card := g.player1Hand[g.cursor]
					g.table = append(g.table, card)
					g.player1Hand = append(g.player1Hand[:g.cursor], g.player1Hand[g.cursor+1:]...)
					if g.cursor >= len(g.player1Hand) && len(g.player1Hand) > 0 {
						g.cursor = len(g.player1Hand) - 1
					}
					return g, func() tea.Msg { return passTurnToAI{} }
				}
			}
		}
	}
	return g, nil
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

func (g Game) View() string {
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
		s.WriteString(renderCards(g.table, -1, false)) // No cursor for table
	}
	s.WriteString("\n\n")

	s.WriteString("Player 1's hand:\n")
	s.WriteString(renderCards(g.player1Hand, g.cursor, true))
	s.WriteString("\n\n")

	s.WriteString(fmt.Sprintf("Trump suit: %s\n", g.trump.String()))
	if g.turn == 0 {
		s.WriteString("Your turn to attack.\n")
	} else {
		s.WriteString("AI's turn to attack.\n")
	}

	s.WriteString("Use left/right arrows to move, space/enter to select, 'q' to quit.")

	return s.String()
}

func main() {
	p := tea.NewProgram(initialGame())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

