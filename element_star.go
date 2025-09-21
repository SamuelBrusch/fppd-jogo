package main

import "context" 


// GameEvent é usado para enviar eventos do jogo
// (definido em types.go)

// PlayerCollect é definido em types.go

func (s *StarBonus) Run(ctx context.Context, out chan<- GameEvent, collected <-chan PlayerCollect) {
	for {
		select {
		case <-ctx.Done():
			return
		case collect := <-collected:
			// Quando coletada, envia um evento para mover 5 casas na direção atual
			out <- GameEvent{
				Type: "STAR_BONUS",
				Data: collect,
			}
		}
	}
}

