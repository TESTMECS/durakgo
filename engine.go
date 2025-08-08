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
	AI           AI
}

func NewEngine() *Engine {
	engine := &Engine{}
	engine.AI = NewMCTS(engine, 100)
	return engine
}

// Copy returns a copy of the board for easy modification
func (b *Board) Copy() *Board {
	newB := &Board{
		PlayerHand:   make([]Card, len(b.PlayerHand)),
		OpponentHand: make([]Card, len(b.OpponentHand)),
		Table:        make([]TableCards, len(b.Table)),
		TrumpSuit:    b.TrumpSuit,
		Attacker:     b.Attacker,
		Deck:         make([]Card, len(b.Deck)),
	}
	copy(newB.PlayerHand, b.PlayerHand)
	copy(newB.OpponentHand, b.OpponentHand)
	copy(newB.Table, b.Table)
	copy(newB.Deck, b.Deck)
	return newB
}

func (e *Engine) AITurn(board *Board) {
	bestMove := e.AI.Solve(board.Copy())
	log.Println("AI Move:", bestMove)
	e.PlayMove(board, bestMove)
}

// GetLegalMoves returns all legal moves for the current player.
func (e *Engine) GetLegalMoves(board *Board) []Move {
	var moves []Move
	hand := board.OpponentHand // AI's hand

	isAIAttacking := board.Attacker == 1

	if isAIAttacking { // AI is Attacking
		if len(board.Table) == 0 {
			// Can play any card to start attack
			for _, card := range hand {
				moves = append(moves, Move{Card: []Card{card}})
			}
		} else {
			// Can add any card with the same rank as cards on the table
			tableRanks := make(map[Rank]bool)
			for _, tableCard := range board.Table {
				tableRanks[tableCard.c.Rank] = true
			}
			for _, card := range hand {
				if tableRanks[card.Rank] {
					moves = append(moves, Move{Card: []Card{card}})
				}
			}
			// "take" move is to pass and end the attack round
			moves = append(moves, Move{take: true})
		}
	} else { // AI is Defending
		if len(board.Table) > 0 {
			attackingCard := board.Table[len(board.Table)-1].c
			for _, card := range hand {
				if e.CanBeat(attackingCard, card, board.TrumpSuit) {
					moves = append(moves, Move{Card: []Card{card}})
				}
			}
		}
		// "take" move is to take cards
		moves = append(moves, Move{take: true})
	}

	return moves
}

// CanBeat checks if a defending card can beat an attacking card.
// Bug here
func (e *Engine) CanBeat(attack Card, defense Card, trump Suit) bool {
	if attack.Suit == defense.Suit {
		return defense.Rank > attack.Rank
	}
	if attack.Suit == trump {
		return defense.Suit != trump
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
	if move.take {
		if board.Attacker == 1 { // AI is attacker and passes
			board.Attacker = e.GetOpponent(board.Attacker)
			board.Table = []TableCards{} // Round ends, cards are discarded
		} else { // AI is defender and takes
			if len(board.Table) > 0 {
				for _, tc := range board.Table {
					board.OpponentHand = append(board.OpponentHand, tc.c)
				}
				board.Table = []TableCards{}
			}
		}
		e.DrawCards(board)
		return
	}

	card := move.Card[0]
	if board.Attacker == 1 { // AI is Attacker
		board.Table = append(board.Table, TableCards{c: card})
		board.OpponentHand = removeCard(board.OpponentHand, card)
		// Attacker does not change, turn passes to player to defend.
	} else { // AI is Defender
		if len(board.Table) > 0 {
			attackingCard := board.Table[len(board.Table)-1].c
			if e.CanBeat(attackingCard, card, board.TrumpSuit) {
				// Successful defense, round ends.
				board.Table = []TableCards{}
				board.OpponentHand = removeCard(board.OpponentHand, card)
				e.DrawCards(board)
				board.Attacker = e.GetOpponent(board.Attacker) // AI becomes attacker
			} else {
				// Invalid move, treat as taking cards.
				for _, tc := range board.Table {
					board.OpponentHand = append(board.OpponentHand, tc.c)
				}
				board.Table = []TableCards{}
				e.DrawCards(board)
			}
		}
	}
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

// removeCard is a helper function to remove a card from a hand.
func removeCard(hand []Card, cardToRemove Card) []Card {
	for i, c := range hand {
		if c == cardToRemove {
			return append(hand[:i], hand[i+1:]...)
		}
	}
	return hand
}
