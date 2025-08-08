package main

import (
	"fmt"
	"log"
	"os"
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
		log.Println("--- AI Turn ---")
		board := g.ToBoard()
		log.Printf("AI Turn Start: Attacker %d, Player Hand %v, AI Hand %v, Table %v", board.Attacker, board.PlayerHand, board.OpponentHand, board.Table)

		g.engine.AITurn(board)
		log.Printf("Board state after AITurn: Attacker %d, Player Hand %v, AI Hand %v, Table %v", board.Attacker, board.PlayerHand, board.OpponentHand, board.Table)

		g.FromBoard(board)
		log.Printf("Game state after FromBoard: Attacker %d, Player Hand %v, AI Hand %v, Table %v", g.attacker, g.player1Hand, g.player2Hand, g.table)

		isover, win := g.engine.CheckGameOver(g.ToBoard())
		if isover {
			g.gameover = true
			g.winner = win
			return aiFinishedTurn{}
		}

		// If AI defended successfully, it becomes the new attacker and must attack immediately.
		if g.attacker == 1 && len(g.table) == 0 {
			log.Println("AI is new attacker, starting another turn to attack.")
			return passTurnToAI{}
		}

		g.turn = 0
		log.Println("--- AI Turn End: Passing to player ---")
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
							log.Println("--- Player Defends ---")
							log.Printf("Player Hand Before: %v", g.player1Hand)
							// Player defends successfully, round ends.
							g.table = []TableCards{}
							g.player1Hand = append(g.player1Hand[:g.cursor], g.player1Hand[g.cursor+1:]...)
							if g.cursor >= len(g.player1Hand) && len(g.player1Hand) > 0 {
								g.cursor = len(g.player1Hand) - 1
							}
							log.Printf("Player Hand After Card Removal: %v", g.player1Hand)

							board := g.ToBoard()
							g.engine.DrawCards(board)
							log.Printf("Player Hand After Draw (on board): %v", board.PlayerHand)
							log.Printf("AI Hand After Draw (on board): %v", board.OpponentHand)
							g.FromBoard(board)
							log.Printf("Player Hand After FromBoard: %v", g.player1Hand)

							g.attacker = 0 // Player becomes the new attacker
							g.turn = 0     // It's player's turn to attack
							log.Println("--- Player Defends End: Player is new attacker ---")
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

func renderCardsLipGloss(cards []Card, cursor int, selected bool) string {
	if len(cards) == 0 {
		return ""
	}

	// Styles
	borderColor := lipgloss.Color("240")
	cursorBorder := lipgloss.Color("69")
	selectedBorder := lipgloss.Color("190")

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	cursorStyle := cardStyle.Copy().
		BorderForeground(cursorBorder).
		Bold(true)

	selectedStyle := cardStyle.Copy().
		BorderForeground(selectedBorder).
		Bold(true).
		Background(lipgloss.Color("236"))

	// Build cards
	cardViews := make([]string, len(cards))
	for i, card := range cards {
		rank := card.Rank.String()
		suit := card.Suit.String()

		cardText := fmt.Sprintf("%-2s\n  %s\n%2s", rank, suit, rank)

		style := cardStyle
		if i == cursor {
			if selected {
				style = selectedStyle
			} else {
				style = cursorStyle
			}
		}

		cardViews[i] = style.Render(cardText)
	}

	// Join cards horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, cardViews...)
}

func (g *Game) View() string {
	// ===== Styles =====
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")) // pinkish

	sectionStyle := lipgloss.NewStyle().
		MarginBottom(1) // space between sections

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")) // gray

	gameOverStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("9")) // bright red

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")). // green
		MarginTop(1)

	// ===== Game over view =====
	if g.gameover {
		var endMsg string
		switch g.winner {
		case 0:
			endMsg = "You win!"
		case 1:
			endMsg = "You lose!"
		default:
			endMsg = "It's a draw!"
		}

		return lipgloss.JoinVertical(lipgloss.Left,
			gameOverStyle.Render("Game Over!"),
			endMsg,
		)
	}

	// ===== Sections =====
	player2 := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Player 2's hand:"),
		renderCardsLipGloss(g.player2Hand, -1, false),
	)

	table := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Table:"),
		func() string {
			if len(g.table) == 0 {
				return infoStyle.Render("[empty]")
			}
			return renderCardsLipGloss(extractCards(g.table), -1, false)
		}(),
	)

	player1 := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Player 1's hand:"),
		renderCardsLipGloss(g.player1Hand, g.cursor, true),
	)

	gameInfo := infoStyle.Render(
		fmt.Sprintf("Trump suit: %s | Deck: %d", g.trump.String(), len(g.deck)),
	)

	// ===== Turn prompt =====
	var prompt string
	if g.turn == 0 {
		if g.attacker == 0 {
			if len(g.table) > 0 {
				prompt = "Your turn to continue attack. (space/enter to play, 'p' to pass)"
			} else {
				prompt = "Your turn to attack."
			}
		} else {
			prompt = "Your turn to defend. (space/enter to play, 't' to take)"
		}
	} else {
		if g.attacker == 1 {
			prompt = "AI is attacking..."
		} else {
			prompt = "AI is defending..."
		}
	}

	status := statusStyle.Render(prompt)
	controls := infoStyle.Render("Use left/right arrows to move, 'q' to quit.")

	// ===== Final layout =====
	return lipgloss.JoinVertical(lipgloss.Left,
		sectionStyle.Render(player2),
		sectionStyle.Render(table),
		sectionStyle.Render(player1),
		sectionStyle.Render(gameInfo),
		status,
		controls,
	)
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
