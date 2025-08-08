package main

import (
	"log"
	"math"
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
	// GetOpponent returns the other player
	GetOpponent(player Player) Player
}

type mcts struct {
	engine GameEngine
	depth  int
}

// Node represents a node in the Monte Carlo search tree
type Node struct {
	move         Move
	parent       *Node
	children     []*Node
	wins         float64
	visits       int
	untriedMoves []Move
	playerToMove Player
}

func NewMCTS(engine GameEngine, depth int) AI {
	return &mcts{engine, depth}
}

func (m *mcts) Solve(board *Board) Move {
	legalMoves := m.engine.GetLegalMoves(board)
	if len(legalMoves) == 0 {
		return Move{take: true}
	}
	if len(legalMoves) == 1 {
		return legalMoves[0]
	}

	root := &Node{
		untriedMoves: m.engine.GetLegalMoves(board),
		playerToMove: board.Attacker,
	}

	for i := 0; i < m.depth; i++ {
		node := root
		simulationBoard := board.Copy()

		// 1. Selection
		for len(node.untriedMoves) == 0 && len(node.children) > 0 {
			node = node.selectChild()
			m.engine.PlayMove(simulationBoard, node.move)
		}

		// 2. Expansion
		if len(node.untriedMoves) > 0 {
			moveIndex := rand.IntN(len(node.untriedMoves))
			move := node.untriedMoves[moveIndex]
			node.untriedMoves = append(node.untriedMoves[:moveIndex], node.untriedMoves[moveIndex+1:]...)

			m.engine.PlayMove(simulationBoard, move)
			childNode := &Node{
				move:         move,
				parent:       node,
				playerToMove: simulationBoard.Attacker,
				untriedMoves: m.engine.GetLegalMoves(simulationBoard),
			}
			node.children = append(node.children, childNode)
			node = childNode
		}

		// 3. Simulation
		for {
			gameOver, _ := m.engine.CheckGameOver(simulationBoard)
			if gameOver {
				break
			}
			moves := m.engine.GetLegalMoves(simulationBoard)
			if len(moves) == 0 {
				log.Println("MCTS Simulation: No legal moves but game not over. Breaking.")
				// This should ideally not happen if GetLegalMoves always provides an option (e.g. take/pass)
				break
			}
			randomMove := moves[rand.IntN(len(moves))]
			m.engine.PlayMove(simulationBoard, randomMove)
		}

		// 4. Backpropagation
		_, winner := m.engine.CheckGameOver(simulationBoard)
		for node != nil {
			var playerWhoMovedToThisState Player
			if node.parent != nil {
				playerWhoMovedToThisState = node.parent.playerToMove
			} else {
				playerWhoMovedToThisState = m.engine.GetOpponent(node.playerToMove)
			}
			node.update(winner, playerWhoMovedToThisState)
			node = node.parent
		}
	}

	// Return the move from the most visited child node
	var bestMove Move
	maxVisits := -1
	for _, child := range root.children {
		if child.visits > maxVisits {
			maxVisits = child.visits
			bestMove = child.move
		}
	}

	if maxVisits == -1 {
		log.Println("MCTS: No children expanded, returning 'take' move.")
		for _, move := range legalMoves {
			if move.take {
				return move
			}
		}
		// Fallback, though should be unreachable if GetLegalMoves is correct
		return Move{take: true}
	}

	// The AI is player 1. It is defending if the attacker is player 0.
	isAIDefending := board.Attacker == 0

	if isAIDefending && !bestMove.take && len(board.Table) > 0 {
		// This is a defense move with a card. Let's validate it.
		// The attacking card is the last one played on the table.
		attackingCard := board.Table[len(board.Table)-1].c
		chosenCard := bestMove.Card[0]
		if !m.engine.CanBeat(attackingCard, chosenCard, board.TrumpSuit) {
			log.Printf("MCTS chose an invalid move %v to beat %v. Forcing 'take'.", chosenCard, attackingCard)
			// The chosen move is invalid. Find the 'take' move in legalMoves and return it.
			for _, move := range legalMoves {
				if move.take {
					return move
				}
			}
			// Fallback
			return Move{take: true}
		}
	}

	return bestMove
}

// selectChild selects a child node to explore using the UCB1 formula.
func (n *Node) selectChild() *Node {
	const c = 1.414 // sqrt(2)
	bestScore := -1.0
	var bestChild *Node

	for _, child := range n.children {
		var score float64
		if child.visits == 0 {
			score = math.MaxFloat64 // Prioritize unvisited nodes
		} else {
			exploit := child.wins / float64(child.visits)
			explore := c * math.Sqrt(math.Log(float64(n.visits))/float64(child.visits))
			score = exploit + explore
		}

		if score > bestScore {
			bestScore = score
			bestChild = child
		}
	}
	return bestChild
}

// update updates the node's statistics from a simulation result.
func (n *Node) update(winner Player, playerWhoMovedToThisState Player) {
	n.visits++
	if winner == playerWhoMovedToThisState {
		n.wins += 1.0
	} else if winner == -1 { // Draw
		n.wins += 0.5
	}
}

