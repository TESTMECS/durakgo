package main

import (
	"log"
	"math/rand/v2"
)

type AI interface {
	Solve(board *Board) Move
}

type GameEngine interface {
	// Returns gameover (bool) & a value if there's a winner
	CheckGameOver(board *Board) (bool, Player)
	// Get all available moves
	GetLegalMoves(board *Board) []Move
	// Play a move on the board
	PlayMove(board *Board, move Move)
	// Check if a move is valid
	CanBeat(attack Card, defense Card, trump Suit) bool
}

type mcts struct {
	engine GameEngine
	depth  int
}

func NewMCTS(engine GameEngine, depth int) AI {
	return &mcts{engine, depth}
}

func (m *mcts) Solve(board *Board) Move {
	legalMoves := m.engine.GetLegalMoves(board)
	log.Println("Legal moves:", legalMoves)
	if len(legalMoves) == 0 {
		return Move{take: true}
	}
	if legalMoves[0].Card != nil {
		return legalMoves[0]
	}
	// Return a random move from the list of legal moves.
	return legalMoves[rand.IntN(len(legalMoves))]
}

func (e *mcts) CheckGameOver(board *Board) (bool, Player) {
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

// Logic for Moves
// Checks table and own hand.
func (m *mcts) GetLegalMoves(board *Board) []Move {
	var moves []Move
	var hand []Card
	if board.Attacker == 0 {
		hand = board.PlayerHand
	} else {
		hand = board.OpponentHand
	}
	if board.Attacker == 1 { // AI is Attacker
		if len(board.Table) == 0 {
			// Can play any card
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
		}
		// "take" move is to pass
		moves = append(moves, Move{take: true})
	} else { // AI is Defender
		if len(board.Table) > 0 {
			attackingCard := board.Table[len(board.Table)-1].c
			for _, card := range hand {
				if m.CanBeat(attackingCard, card, board.TrumpSuit) {
					moves = append(moves, Move{Card: []Card{card}})
				}
			}
		}
		// "take" move is to take cards
		moves = append(moves, Move{take: true})
	}

	return moves
}
func (m *mcts) PlayMove(board *Board, move Move) {}

// CanBeat checks if a defending card can beat an attacking card.
func (e *mcts) CanBeat(attack Card, defense Card, trump Suit) bool {
	if attack.Suit == defense.Suit {
		return defense.Rank > attack.Rank
	}
	if defense.Suit == trump {
		return attack.Suit != trump
	}
	return false
}

