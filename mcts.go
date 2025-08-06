package main

import "log"

type AI interface {
	Solve(board *Board) Move
}

type MCTS struct {
	engine *Engine
	depth  int
}

func NewMCTS(engine *Engine, depth int) *MCTS {
	return &MCTS{
		engine: engine,
		depth:  depth,
	}
}

func (m *MCTS) Solve(board *Board) Move {
	//TODO: Implement MCTS logic here
	log.Println("MCTS logic not implemented yet")
	return Move{IsPass: true}
}

