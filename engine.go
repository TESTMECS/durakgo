package main

import "log"

type Move struct {
	Card []Card
	take bool
}
type Deck struct {
	Cards []Card
	trump Suit
}
type Engine struct {
	PlayerHand   []Card
	OpponentHand []Card
	Table        []TableCards
	Deck         []Card
	TrumpSuit    Suit
	Attacker     Player
}

// type GameEngine interface {
// }

func NewEngine() *Engine {
	engine := &Engine{}
	return engine
}

// Copy retuurns a copy of the board for easy modification
func (b *Engine) Copy() *Engine {
	newB := &Engine{
		PlayerHand:   make([]Card, len(b.PlayerHand)),
		OpponentHand: make([]Card, len(b.OpponentHand)),
		Deck:         make([]Card, len(b.Deck)),
		Table:        make([]TableCards, len(b.Table)),
		TrumpSuit:    b.TrumpSuit,
		Attacker:     b.Attacker,
	}
	copy(newB.PlayerHand, b.PlayerHand)
	copy(newB.OpponentHand, b.OpponentHand)
	copy(newB.Deck, b.Deck)
	copy(newB.Table, b.Table)
	return newB
}

func (e *Engine) AITurn() {
	mcts := NewMCTS(e.Clone())
	bestMove := mcts.Search(1000) // 1000 iterations, for example
	e.MakeMove(bestMove)
}

// It copies the board and hands it to the GameEngine for simulation, then returns an updated board.
func (e *GameEngine) HandleAITurn(board *Board) *Board {
	log.Printf("Current board in AI turn: %%#v: %#v\n", board)
	if board.Attacker == 1 { // AI is Attacker
		log.Println("Ai is deciding attacking...")
		move := e.ai.Solve(board.Copy())
		log.Printf("Ai decided to attack with %s\n", move)
		if !move.IsPass {
			board.Table = append(board.Table, move.Card)
			board.OpponentHand = removeCard(board.OpponentHand, move.Card)
		} else {
			// If Pass then take cards from the Table
			board.OpponentHand = append(board.OpponentHand, board.Table...)
			board.Table = []Card{}
		}
	} else {
		log.Println("Ai is defending...")
		move := e.ai.Solve(board.Copy())
		if move.IsPass {
			board.OpponentHand = append(board.OpponentHand, board.Table...)
			board.Table = []Card{}
		} else {
			log.Printf("Ai decided to defend with %s\n", move)
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
func (e *Engine) CheckGameOver(board *Board) (bool, Player) {
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
func (e *Engine) GetLegalMoves(board *Board) []Move {
	var moves []Move
	for _, card := range board.Table {

	}
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
