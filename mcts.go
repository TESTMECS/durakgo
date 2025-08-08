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
	engine              GameEngine
	depth               int
	simulationStepLimit int
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
	return &mcts{
		engine:              engine,
		depth:               depth,
		simulationStepLimit: 100, // Limit simulations to 100 moves
	}
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
		for j := 0; j < m.simulationStepLimit; j++ {
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
			node.update(winner)
			node = node.parent
		}
	}

	// Return the move from the child with the highest win rate.
	var bestMove Move
	bestScore := -1.0
	for _, child := range root.children {
		if child.visits > 0 {
			score := child.wins / float64(child.visits)
			if score > bestScore {
				bestScore = score
				bestMove = child.move
			}
		}
	}

	if bestScore == -1.0 {
		log.Println("MCTS: No children visited, returning 'take' move as fallback.")
		for _, move := range legalMoves {
			if move.take {
				return move
			}
		}
		// Fallback if 'take' is not a legal move for some reason.
		if len(legalMoves) > 0 {
			return legalMoves[0]
		}
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
			winRate := child.wins / float64(child.visits)
			// If it's the opponent's turn at this node (player 0), they will try to maximize
			// their win rate, which is 1 - our (AI's) win rate.
			if n.playerToMove != 1 { // AI is player 1
				winRate = 1.0 - winRate
			}

			explore := c * math.Sqrt(math.Log(float64(n.visits)) / float64(child.visits))
			score = winRate + explore
		}

		if score > bestScore {
			bestScore = score
			bestChild = child
		}
	}
	return bestChild
}

// update updates the node's statistics from a simulation result.
// Wins are from the perspective of player 1 (the AI).
func (n *Node) update(winner Player) {
	n.visits++
	if winner == 1 { // AI is player 1
		n.wins += 1.0
	} else if winner == -1 { // Draw
		n.wins += 0.5
	}
}
