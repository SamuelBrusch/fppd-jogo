package main

import (
	"time"
	"os"
	"sync"
)

type Inimigo struct {
	X, Y      int
	direcaoX  int
	ativo     bool
	comando   chan string  // Canal de comunicação
	waitGroup *sync.WaitGroup  // Para sincronização
}

func novoInimigo(x, y int, wg *sync.WaitGroup) *Inimigo {
	return &Inimigo{
		X:         x,
		Y:         y,
		direcaoX:  1,
		ativo:     false,
		comando:   make(chan string),  // Canal sem buffer
		waitGroup: wg,
	}
}

func (i *Inimigo) mover(jogo *Jogo, startChan <-chan struct{}) {
	defer i.waitGroup.Done()

	// Espera o sinal para começar a se mover
	<-startChan
	i.ativo = true

	for {
		select {
		case cmd := <-i.comando:
			if cmd == "parar" {
				// Remove do mapa antes de sair
				jogo.Mutex.Lock()
				if i.Y >= 0 && i.Y < len(jogo.Mapa) && i.X >= 0 && i.X < len(jogo.Mapa[i.Y]) {
					jogo.Mapa[i.Y][i.X] = Vazio
				}
				jogo.Mutex.Unlock()
				return
			}

		default:
			jogo.Mutex.Lock()

			// Remove da posição atual
			if i.Y >= 0 && i.Y < len(jogo.Mapa) && i.X >= 0 && i.X < len(jogo.Mapa[i.Y]) {
				jogo.Mapa[i.Y][i.X] = Vazio
			}

			// Calcula nova posição
			novaPosX := i.X + i.direcaoX

			// Verifica colisão com paredes ou limites
			if novaPosX >= 0 && novaPosX < len(jogo.Mapa[0]) {
				if !jogo.Mapa[i.Y][novaPosX].tangivel {
					i.X = novaPosX
				} else {
					i.direcaoX *= -1
				}
			} else {
				i.direcaoX *= -1
			}

			// Atualiza no mapa
			if i.Y >= 0 && i.Y < len(jogo.Mapa) && i.X >= 0 && i.X < len(jogo.Mapa[i.Y]) {
				jogo.Mapa[i.Y][i.X] = InimigoElem
			}

			// Verifica colisão com jogador
			if i.X == jogo.PosX && i.Y == jogo.PosY {
				jogo.StatusMsg = "Você foi derrotado!"
				jogo.Mutex.Unlock()
				time.Sleep(2 * time.Second)
				os.Exit(0)
			}

			jogo.Mutex.Unlock()
			time.Sleep(300 * time.Millisecond)  // Controle de velocidade
		}
	}
}