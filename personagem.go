// personagem.go - Funções para movimentação e ações do personagem
package main

import (
	"fmt"
	"math/rand"
)

// Atualiza a posição do personagem com base na tecla pressionada (WASD)
func personagemMover(tecla rune, jogo *Jogo) {
	dx, dy := 0, 0
	switch tecla {
	case 'w':
		dy = -1 // Move para cima
	case 'a':
		dx = -1 // Move para a esquerda
	case 's':
		dy = 1 // Move para baixo
	case 'd':
		dx = 1 // Move para a direita
	}

	// Verificar se tem pulos duplos disponíveis
	stepSize := 1
	if jogo.DoubleJumps > 0 {
		stepSize = 2
		jogo.StatusMsg = fmt.Sprintf("Pulo duplo! Restam %d pulos", jogo.DoubleJumps-1)
	}

	nx, ny := jogo.PosX+(dx*stepSize), jogo.PosY+(dy*stepSize)

	// Verifica se o movimento é permitido e realiza a movimentação
	if jogoPodeMoverPara(jogo, nx, ny) {
		// Se for pulo duplo, verificar se a posição intermediária também é válida
		if stepSize == 2 {
			intermediateX, intermediateY := jogo.PosX+dx, jogo.PosY+dy
			if !jogoPodeMoverPara(jogo, intermediateX, intermediateY) {
				// Se a posição intermediária não é válida, fazer movimento normal
				nx, ny = jogo.PosX+dx, jogo.PosY+dy
				stepSize = 1
				if !jogoPodeMoverPara(jogo, nx, ny) {
					return // Não pode mover
				}
				jogo.StatusMsg = fmt.Sprintf("Pulo duplo bloqueado! Restam %d pulos", jogo.DoubleJumps)
			} else {
				// Pulo duplo realizado com sucesso, decrementar contador
				jogo.DoubleJumps--
				if jogo.DoubleJumps == 0 {
					jogo.StatusMsg = "Último pulo duplo usado!"
				}
			}
		}

		jogoMoverElemento(jogo, jogo.PosX, jogo.PosY, dx*stepSize, dy*stepSize)
		jogo.PosX, jogo.PosY = nx, ny

		// DETECÇÃO DIRETA: Verificar se coletou estrela
		coletouEstrela := ConsumirItemEstrela(jogo)
		if coletouEstrela {
			jogo.DoubleJumps = 3 // Concede 3 pulos duplos
			jogo.StatusMsg = "Estrela coletada! 3 pulos duplos concedidos!"
		}

		// DETECÇÃO DIRETA: Verificar se coletou item de invisibilidade
		coletouInvisibilidade := ConsumirItemInvisibilidade(jogo)
		if coletouInvisibilidade {
			jogo.InvisibleSteps = InvisibilityDuration
			jogo.StatusMsg = "Invisibilidade coletada!"
		}

		// CANAIS: Enviar evento de coleta para os itens de invisibilidade (para concorrência)
		collectEvent := PlayerCollect{
			X: jogo.PosX,
			Y: jogo.PosY,
		}
		select {
		case jogo.PlayerCollects <- collectEvent:
			// Evento de coleta enviado com sucesso
		default:
			// Canal cheio, pular este envio
		}

		// Atualizar contador de invisibilidade (apenas se não coletou neste turno)
		if !coletouInvisibilidade && !coletouEstrela && jogo.InvisibleSteps > 0 {
			jogo.InvisibleSteps--
			if jogo.InvisibleSteps == 0 {
				jogo.StatusMsg = "Invisibilidade expirou"
			} else {
				jogo.StatusMsg = fmt.Sprintf("Invisível: %d movimentos restantes", jogo.InvisibleSteps)
			}
		}
	} else {
		// Se não pode mover com pulo duplo, tentar movimento normal
		if stepSize == 2 {
			nx, ny = jogo.PosX+dx, jogo.PosY+dy
			if jogoPodeMoverPara(jogo, nx, ny) {
				jogoMoverElemento(jogo, jogo.PosX, jogo.PosY, dx, dy)
				jogo.PosX, jogo.PosY = nx, ny
				jogo.StatusMsg = fmt.Sprintf("Pulo duplo bloqueado - movimento normal. Restam %d pulos", jogo.DoubleJumps)

				// Verificar coletas mesmo com movimento normal
				coletouEstrela := ConsumirItemEstrela(jogo)
				if coletouEstrela {
					jogo.DoubleJumps = 3
					jogo.StatusMsg = "Estrela coletada! 3 pulos duplos concedidos!"
				}

				coletouInvisibilidade := ConsumirItemInvisibilidade(jogo)
				if coletouInvisibilidade {
					jogo.InvisibleSteps = InvisibilityDuration
					jogo.StatusMsg = "Invisibilidade coletada!"
				}
			}
		}
	}
} // Define o que ocorre quando o jogador pressiona a tecla de interação
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
		personagemMover(ev.Tecla, jogo)
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
		personagemMover(ev.Tecla, jogo)

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
