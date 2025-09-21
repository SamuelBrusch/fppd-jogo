package main

import "context"

const InvisibilityDuration = 20

// Tipos de eventos produzidos por este elemento
const (
	EventApplyInvisibility = "ApplyInvisibility"
	EventRemoveElement     = "RemoveElement"
)

// Payload para aplicação de invisibilidade
type InvisibilityApplied struct {
	Duration int
}

func (i *Invisibility) Run(ctx context.Context, out chan<- GameEvent, picked <-chan PlayerCollect) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-picked:
			if !ok {
				return
			}
			// Ignora coletas que não são nesta posição
			if ev.X != i.X || ev.Y != i.Y {
				continue
			}

			// 1) Solicita remoção do item do mapa
			out <- GameEvent{
				Type: EventRemoveElement,
				Data: Invisibility{X: i.X, Y: i.Y},
			}

			// 2) Aplica o buff de invisibilidade ao jogador
			out <- GameEvent{
				Type: EventApplyInvisibility,
				Data: InvisibilityApplied{Duration: InvisibilityDuration},
			}

			// Item é one-shot
			return
		}
	}
}

// NOVO: remove o item “sob” o jogador, substituindo-o por Vazio.
func ConsumirItemInvisibilidade(jogo *Jogo) bool {
	// Se o elemento “embaixo do jogador” for o item, consome-o
	if jogo.UltimoVisitado.simbolo == InvisibilityItem.simbolo {
		jogo.UltimoVisitado = Vazio
		return true
	}
	return false
}
