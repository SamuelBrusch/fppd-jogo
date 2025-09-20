package main

import (
	"context"
	"math"
	"math/rand"
	"time"
)

func (m *Monster) Run(ctx context.Context, out chan<- GameEvent, alerts <-chan PlayerAlert, pstate <-chan PlayerState) {
	// Timer para controlar velocidade do monstro (30ms - muito rápido)
	ticker := time.NewTicker(30 * time.Millisecond)
	defer ticker.Stop()

	// Timeout para comportamento alternativo se não receber posição do jogador
	playerTimeout := time.NewTimer(3 * time.Second)
	defer playerTimeout.Stop()

	for {
		select {
		case <-ctx.Done():
			// Jogo terminou, sair da goroutine
			return

		case playerState := <-pstate:
			// Recebeu nova posição do jogador
			m.updatePlayerPosition(playerState)

			// Reset do timeout quando recebe posição do jogador
			if !playerTimeout.Stop() {
				select {
				case <-playerTimeout.C:
				default:
				}
			}
			playerTimeout.Reset(3 * time.Second)

		case <-playerTimeout.C:
			// TIMEOUT: Não recebeu posição do jogador por 3 segundos
			// Comportamento alternativo: entrar em modo "alerta" e patrulhar agressivamente
			if m.state == Hunting {
				// Se estava caçando, voltar para patrulha mas com comportamento diferente
				m.state = Patrolling
				m.generateAggressivePatrolDestiny() // Patrulha mais agressiva
			} else {
				// Se estava patrulhando, gerar destino aleatório mais distante
				m.generateRandomDestiny()
			}

			// Enviar evento de timeout
			timeoutEvent := GameEvent{
				Type: "monster_timeout",
				Data: map[string]interface{}{
					"monster_id": m.id,
					"message":    "Monster lost track of player - entering aggressive patrol",
				},
			}

			select {
			case out <- timeoutEvent:
			default:
			}

			// Reset timeout para próxima verificação
			playerTimeout.Reset(3 * time.Second)

		case alert := <-alerts:
			// Recebeu alerta - processar com timeout
			select {
			case <-time.After(500 * time.Millisecond):
				// Timeout ao processar alerta - comportamento alternativo
				m.state = Patrolling
				m.generateRandomDestiny()
			default:
				// Processar alerta normalmente
				m.processAlert(alert)
			}

		case <-ticker.C:
			// Hora de tentar mover o monstro
			if m.shouldMove() {
				m.processMovement(out)
			}
		}
	}
}

// Controla velocidade: monstro move SEMPRE quando caçando
func (m *Monster) shouldMove() bool {
	if m.state == Hunting {
		// Quando caçando, move SEMPRE (extremamente rápido)
		return true
	}

	// Quando patrulhando, ainda move frequentemente
	m.shift_count++
	if m.shift_count >= 1 { // Reduzido de 2 para 1
		m.shift_count = 0
		return true
	}
	return false
} // Processa nova posição do jogador e detecta se pode ver (25 células)
func (m *Monster) updatePlayerPosition(playerState PlayerState) {
	playerPos := Position{X: playerState.X, Y: playerState.Y}

	// Calcula distância até o jogador
	if m.canSeePlayer(playerPos) {
		// Pode ver o jogador - mudar para modo caça IMEDIATAMENTE
		m.state = Hunting
		m.last_seen = playerPos
		m.destiny_position = playerPos // Perseguir diretamente
	} else if m.state == Hunting {
		// Perdeu o jogador - continuar indo para última posição vista
		m.destiny_position = m.last_seen
		// Só para de caçar se chegou MUITO perto da última posição
		if m.distanceTo(m.last_seen) < 0.5 {
			m.state = Patrolling
			m.generateAggressivePatrolDestiny() // Patrulha agressiva
		}
	}

	// Se está caçando, SEMPRE atualizar destino para posição atual do jogador
	if m.state == Hunting {
		m.destiny_position = playerPos
		m.last_seen = playerPos
	}
} // Executa movimento baseado no estado atual
func (m *Monster) processMovement(out chan<- GameEvent) {
	// Se está patrulhando e chegou no destino, gerar novo destino
	if m.state == Patrolling && m.distanceTo(m.destiny_position) < 1 {
		m.generateRandomDestiny()
	}

	// Calcular próxima posição desejada (sem mover ainda)
	oldX, oldY := m.current_position.X, m.current_position.Y
	newPos := m.calculateNextPosition(m.destiny_position)

	// Enviar evento de movimento para o jogo principal validar
	event := GameEvent{
		Type: "monster_move",
		Data: MonsterMoveData{
			OldX:      oldX,
			OldY:      oldY,
			NewX:      newPos.X,
			NewY:      newPos.Y,
			MonsterID: m.id,
		},
	}

	select {
	case out <- event:
		// Evento enviado com sucesso
	default:
		// Canal cheio, pular este evento
	}
}

// Verifica se pode ver o jogador (máximo 25 células - muito agressivo)
func (m *Monster) canSeePlayer(playerPos Position) bool {
	distance := m.distanceTo(playerPos)
	return distance <= 25.0
}

// Calcula distância euclidiana entre monstro e uma posição
func (m *Monster) distanceTo(pos Position) float64 {
	dx := float64(m.current_position.X - pos.X)
	dy := float64(m.current_position.Y - pos.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

// Calcula próxima posição em direção ao alvo (movimento mais inteligente)
func (m *Monster) calculateNextPosition(target Position) Position {
	currentPos := m.current_position

	// Calcula direção
	dx := target.X - currentPos.X
	dy := target.Y - currentPos.Y

	newPos := currentPos

	// Se está caçando, priorizar movimento diagonal para ser mais rápido
	if m.state == Hunting {
		// Movimento diagonal quando possível
		if dx != 0 && dy != 0 {
			// Mover diagonalmente
			if dx > 0 {
				newPos.X++
			} else {
				newPos.X--
			}
			if dy > 0 {
				newPos.Y++
			} else {
				newPos.Y--
			}
		} else if dx != 0 {
			// Só movimento horizontal
			if dx > 0 {
				newPos.X++
			} else {
				newPos.X--
			}
		} else if dy != 0 {
			// Só movimento vertical
			if dy > 0 {
				newPos.Y++
			} else {
				newPos.Y--
			}
		}
	} else {
		// Movimento normal quando patrulhando
		if dx != 0 {
			if dx > 0 {
				newPos.X++
			} else {
				newPos.X--
			}
		} else if dy != 0 {
			if dy > 0 {
				newPos.Y++
			} else {
				newPos.Y--
			}
		}
	}

	return newPos
} // Gera destino aleatório para patrulha
func (m *Monster) generateRandomDestiny() {
	// Gera posição aleatória em um raio de 10 células da posição atual
	radius := 10
	maxTries := 10 // Máximo de tentativas para encontrar posição válida

	for tries := 0; tries < maxTries; tries++ {
		angle := rand.Float64() * 2 * math.Pi
		distance := rand.Float64() * float64(radius)

		newX := m.current_position.X + int(distance*math.Cos(angle))
		newY := m.current_position.Y + int(distance*math.Sin(angle))

		// Verificar se está dentro dos limites do mapa (assumindo 80x30)
		if newX >= 1 && newX < 79 && newY >= 1 && newY < 29 {
			m.destiny_position = Position{X: newX, Y: newY}
			return
		}
	}

	// Se não encontrou posição válida, ficar próximo da posição atual
	m.destiny_position = Position{
		X: m.current_position.X + (rand.Intn(3) - 1), // -1, 0, ou 1
		Y: m.current_position.Y + (rand.Intn(3) - 1), // -1, 0, ou 1
	}
}

// Gera destino para patrulha agressiva (raio maior, movimento mais rápido)
func (m *Monster) generateAggressivePatrolDestiny() {
	// Patrulha agressiva com raio maior (15 células)
	radius := 15
	maxTries := 10

	for tries := 0; tries < maxTries; tries++ {
		angle := rand.Float64() * 2 * math.Pi
		distance := rand.Float64() * float64(radius)

		newX := m.current_position.X + int(distance*math.Cos(angle))
		newY := m.current_position.Y + int(distance*math.Sin(angle))

		// Verificar se está dentro dos limites do mapa
		if newX >= 1 && newX < 79 && newY >= 1 && newY < 29 {
			m.destiny_position = Position{X: newX, Y: newY}
			return
		}
	}

	// Fallback: mover para posição mais distante possível
	m.destiny_position = Position{
		X: m.current_position.X + (rand.Intn(5) - 2), // -2 a 2
		Y: m.current_position.Y + (rand.Intn(5) - 2), // -2 a 2
	}
}

// Processa alertas recebidos
func (m *Monster) processAlert(alert PlayerAlert) {
	// Processar diferentes tipos de alertas
	switch alert.Type {
	case "player_nearby":
		// Jogador detectado próximo
		if data, ok := alert.Data.(map[string]int); ok {
			if x, hasX := data["x"]; hasX {
				if y, hasY := data["y"]; hasY {
					m.state = Hunting
					m.last_seen = Position{X: x, Y: y}
					m.destiny_position = Position{X: x, Y: y}
				}
			}
		}
	case "noise":
		// Som detectado - investigar
		if data, ok := alert.Data.(map[string]int); ok {
			if x, hasX := data["x"]; hasX {
				if y, hasY := data["y"]; hasY {
					m.destiny_position = Position{X: x, Y: y}
				}
			}
		}
	default:
		// Alerta desconhecido - patrulhar
		m.generateRandomDestiny()
	}
}
