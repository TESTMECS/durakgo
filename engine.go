package main

// NOTE: The following types are assumed to be defined in card.go
// type Card struct { Suit Suit; Rank Rank }
// type Suit int
// type Rank int
// type Player int // 0 for player, 1 for AI

// Move represents a game move.
type Move struct {
	Card   Card
	IsPass bool // True to pass/take cards
}

// Board represents the game state for the engine.
type Board struct {
	PlayerHand   []Card
	OpponentHand []Card
	Table        []Card
	Deck         []Card
	TrumpSuit    Suit
	Attacker     Player
}

func (b *Board) Copy() *Board {
	newB := &Board{
		PlayerHand:   make([]Card, len(b.PlayerHand)),
		OpponentHand: make([]Card, len(b.OpponentHand)),
		Deck:         make([]Card, len(b.Deck)),
		Table:        make([]Card, len(b.Table)),
		TrumpSuit:    b.TrumpSuit,
		Attacker:     b.Attacker,
	}
	copy(newB.PlayerHand, b.PlayerHand)
	copy(newB.OpponentHand, b.OpponentHand)
	copy(newB.Deck, b.Deck)
	copy(newB.Table, b.Table)
	return newB
}

// PlayMove is used by the MCTS algorithm to simulate moves.
func (b *Board) PlayMove(move Move) {
	// This is a simplified simulation for the AI.
	// A real implementation would be more complex.
	if !move.IsPass {
		b.Table = append(b.Table, move.Card)
	}
}

type Engine struct {
	ai AI
}

func NewEngine(depth int) *Engine {
	engine := &Engine{}
	mcts := NewMCTS(engine, depth)
	engine.ai = mcts
	return engine
}

// HandleAITurn processes the AI's entire turn, whether attacking or defending.
func (e *Engine) HandleAITurn(board *Board) *Board {
	if board.Attacker == 1 { // AI is Attacker
		move := e.ai.Solve(board.Copy()) // AI decides what card to attack with
		if !move.IsPass {
			board.Table = append(board.Table, move.Card)
			board.OpponentHand = removeCard(board.OpponentHand, move.Card)
		}
	} else { // AI is Defender
		move := e.ai.Solve(board.Copy()) // AI decides to defend or take

		if move.IsPass {
			// AI takes cards
			board.OpponentHand = append(board.OpponentHand, board.Table...)
			board.Table = []Card{}
		} else {
			// AI plays a defending card
			attackingCard := board.Table[len(board.Table)-1]
			if e.CanBeat(attackingCard, move.Card, board.TrumpSuit) {
				board.Table = append(board.Table, move.Card)
				board.OpponentHand = removeCard(board.OpponentHand, move.Card)

				// Successful defense, round ends.
				board.Table = []Card{}                         // Discard cards
				board.Attacker = e.GetOpponent(board.Attacker) // AI becomes attacker
			} else {
				// Invalid move from AI, treat as taking cards.
				board.OpponentHand = append(board.OpponentHand, board.Table...)
				board.Table = []Card{}
			}
		}
		e.DrawCards(board)
	}
	return board
}

// CanBeat checks if a defending card can beat an attacking card.
func (e *Engine) CanBeat(attack Card, defense Card, trump Suit) bool {
	if attack.Suit == defense.Suit {
		return defense.Rank > attack.Rank
	}
	if defense.Suit == trump {
		return attack.Suit != trump
	}
	return false
}

// DrawCards refills players' hands from the deck up to 6 cards.
func (e *Engine) DrawCards(board *Board) {
	for len(board.PlayerHand) < 6 && len(board.Deck) > 0 {
		board.PlayerHand = append(board.PlayerHand, board.Deck[0])
		board.Deck = board.Deck[1:]
	}
	for len(board.OpponentHand) < 6 && len(board.Deck) > 0 {
		board.OpponentHand = append(board.OpponentHand, board.Deck[0])
		board.Deck = board.Deck[1:]
	}
}

func (e *Engine) PlayMove(board *Board, move Move) {
	board.PlayMove(move)
}

func (e *Engine) GetOpponent(player Player) Player {
	if player == 0 {
		return 1
	}
	return 0
}

func (e *Engine) CheckGameOver(board *Board, move Move) (bool, Player) {
	if len(board.Deck) == 0 {
		if len(board.PlayerHand) == 0 && len(board.OpponentHand) == 0 {
			return true, -1 // Draw
		}
		if len(board.PlayerHand) == 0 {
			return true, 0 // Player wins
		}
		if len(board.OpponentHand) == 0 {
			return true, 1 // AI wins
		}
	}
	return false, -1
}

// removeCard is a helper function to remove a card from a hand.
func removeCard(hand []Card, cardToRemove Card) []Card {
	for i, c := range hand {
		if c == cardToRemove {
			return append(hand[:i], hand[i+1:]...)
		}
	}
	return hand
}

