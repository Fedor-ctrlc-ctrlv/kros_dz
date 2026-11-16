package store

import (
	"errors"
	"sync"
)

type Order struct {
	ID       int64  `json:"id"`
	Product  string `json:"product"`
	Quantity int    `json:"quantity"`
}

type Orders struct {
	mu     sync.RWMutex
	nextID int64
	items  map[int64]Order
}

var ErrNotFound = errors.New("order not found")

func NewOrders() *Orders { return &Orders{items: make(map[int64]Order)} }

func (s *Orders) Create(product string, quantity int) Order {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	o := Order{ID: s.nextID, Product: product, Quantity: quantity}
	s.items[o.ID] = o
	return o
}

func (s *Orders) Get(id int64) (Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.items[id]
	if !ok {
		return Order{}, ErrNotFound
	}
	return o, nil
}
