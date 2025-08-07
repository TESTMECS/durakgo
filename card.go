package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Suit int

const (
	Clubs Suit = iota
	Diamonds
	Hearts
	Spades
)

// String returns the the unicode representation of the card suit.
func (s Suit) String() string {
	switch s {
	case Clubs:
		return "♣"
	case Diamonds:
		return "♦"
	case Hearts:
		return "♥"
	case Spades:
		return "♠"
	default:
		return "?"
	}
}

type Rank int

const (
	Two Rank = iota + 2
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	Ace
)

// String returns the the unicode representation of the card rank.
func (r Rank) String() string {
	switch r {
	case Two:
		return "2"
	case Three:
		return "3"
	case Four:
		return "4"
	case Five:
		return "5"
	case Six:
		return "6"
	case Seven:
		return "7"
	case Eight:
		return "8"
	case Nine:
		return "9"
	case Ten:
		return "10"
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	case Ace:
		return "A"
	default:
		return "?"
	}
}

type Card struct {
	Suit Suit
	Rank Rank
}

// String returns the the unicode representation of the card.
func (c Card) String() string {
	return fmt.Sprintf("%s%s", c.Rank, c.Suit)
}

// NewDeck returns a unshuffled deck of cards
func NewDeck() []Card {
	cards := make([]Card, 0, 52)
	for suit := Clubs; suit <= Spades; suit++ {
		for rank := Two; rank <= Ace; rank++ {
			cards = append(cards, Card{Suit: suit, Rank: rank})
		}
	}
	return cards
}

// ShuffleDeck shuffles a deck of cards.
func ShuffleDeck(cards []Card) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})
}
