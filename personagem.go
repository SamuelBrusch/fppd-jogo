// personagem.go - Funções para movimentação e ações do personagem
package main

import (
	"fmt"
	"math/rand"
)

// Estrutura para coleta do jogador (direção da coleta)
// A definição de PlayerCollect foi movida para types.go

// Atualiza a posição do personagem com base na tecla pressionada (WASD)
func personagemMover(tecla rune, jogo *Jogo, starEvents <-chan GameEvent, collected chan<- PlayerCollect) {
	dx, dy := 0, 0
	switch tecla {
	case 'w':
		dy = -1
	case 'a':
		dx = -1
	case 's':
		dy = 1
	case 'd':
		dx = 1
	}

	nx, ny := jogo.PosX+dx, jogo.PosY+dy

	if jogoPodeMoverPara(jogo, nx, ny) {
		jogoMoverElemento(jogo, jogo.PosX, jogo.PosY, dx, dy)
		jogo.PosX, jogo.PosY = nx, ny

		// Se pisou na estrela, envia para o canal collected
		if jogo.Mapa[ny][nx].simbolo == '*' {
			collected <- PlayerCollect{X: dx, Y: dy}
			fmt.Println("Estrela coletada! Direção:", dx, dy)

		}
	}

	// Processa eventos de estrela
	select {
	case ev := <-starEvents:
		if ev.Type == "STAR_BONUS" {
			collect := ev.Data.(PlayerCollect)
			// Move o personagem 5 casas na direção coletada
			for i := 0; i < 70; i++ {
				nx, ny := jogo.PosX+collect.X, jogo.PosY+collect.Y
				if !jogoPodeMoverPara(jogo, nx, ny) {
					break
				}
				jogoMoverElemento(jogo, jogo.PosX, jogo.PosY, collect.X, collect.Y)
				jogo.PosX, jogo.PosY = nx, ny
			}
		}
	default:
		// Nenhum evento
	}
}
// Define o que ocorre quando o jogador pressiona a tecla de interação
// Neste exemplo, apenas exibe uma mensagem de status
// Você pode expandir essa função para incluir lógica de interação com objetos
func personagemInteragir(jogo *Jogo) {
	// Atualmente apenas exibe uma mensagem de status
	jogo.StatusMsg = fmt.Sprintf("Interagindo em (%d, %d)", jogo.PosX, jogo.PosY)
}

// Processa o evento do teclado e executa a ação correspondente
func personagemExecutarAcao(ev EventoTeclado, jogo *Jogo) bool {
	switch ev.Tipo {
	case "sair":
		// Retorna false para indicar que o jogo deve terminar
		return false
	case "interagir":
		// Executa a ação de interação
		personagemInteragir(jogo)
	case "mover":
		// Move o personagem com base na tecla
		// Adapte para passar o canal correto de coleta de estrela
personagemMover(ev.Tecla, jogo, jogo.StarEvents, jogo.Collected)
		// Envia estado atualizado do jogador
		jogoEnviarEstadoJogador(jogo)

		// Ocasionalmente enviar alerta para o monstro (simula fazer barulho)
		// 20% de chance de fazer barulho ao se mover
		if rand.Float32() < 0.2 {
			jogoEnviarAlerta(jogo, "noise")
		}
	}
	return true // Continua o jogo
}

// Versão com canal para comunicação concorrente
func personagemExecutarAcaoComCanal(ev EventoTeclado, jogo *Jogo, playerChannel chan<- PlayerState) bool {
	switch ev.Tipo {
	case "sair":
		// Retorna false para indicar que o jogo deve terminar
		return false
	case "interagir":
		// Executa a ação de interação
		personagemInteragir(jogo)
	case "mover":
		// Move o personagem com base na tecla
personagemMover(ev.Tecla, jogo, jogo.StarEvents, jogo.Collected)
		
		// Enviar nova posição do jogador para o monstro
		playerState := PlayerState{
			X: jogo.PosX,
			Y: jogo.PosY,
		}

		select {
		case playerChannel <- playerState:
			// Posição enviada com sucesso
		default:
			// Canal cheio, pular este envio
		}
	}
	return true // Continua o jogo
}
