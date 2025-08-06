package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	deck        []Card
	player1Hand []Card
	player2Hand []Card
	table       []Card
	cursor      int
}

func initialModel() model {
	deck := NewDeck()
	ShuffleDeck(deck)

	player1Hand := deck[:6]
	deck = deck[6:]

	player2Hand := deck[:6]
	deck = deck[6:]

	return model{
		deck:        deck,
		player1Hand: player1Hand,
		player2Hand: player2Hand,
		table:       []Card{},
		cursor:      0,
	}
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("Durak")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "right", "l":
			if m.cursor < len(m.player1Hand)-1 {
				m.cursor++
			}
		case "left", "h":
			if m.cursor > 0 {
				m.cursor--
			}
		case " ", "enter":
			// Move card from hand to table
			if len(m.player1Hand) > 0 {
				card := m.player1Hand[m.cursor]
				m.table = append(m.table, card)
				m.player1Hand = append(m.player1Hand[:m.cursor], m.player1Hand[m.cursor+1:]...)
				if m.cursor >= len(m.player1Hand) && len(m.player1Hand) > 0 {
					m.cursor = len(m.player1Hand) - 1
				}
			}
		}
	}
	return m, nil
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

func (m model) View() string {
	var s strings.Builder

	s.WriteString("Player 2's hand:\n")
	s.WriteString(renderCards(m.player2Hand, -1, false)) // No cursor for player 2
	s.WriteString("\n\n")

	s.WriteString("Table:\n")
	if len(m.table) == 0 {
		s.WriteString("[empty]\n")
	} else {
		s.WriteString(renderCards(m.table, -1, false)) // No cursor for table
	}
	s.WriteString("\n\n")

	s.WriteString("Player 1's hand:\n")
	s.WriteString(renderCards(m.player1Hand, m.cursor, true))
	s.WriteString("\n\n")

	s.WriteString("Use left/right arrows to move, space/enter to select, 'q' to quit.")

	return s.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
