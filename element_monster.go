package main

import (
	"context"
	"math"
	"math/rand"
	"time"
)

func (m *Monster) Run(ctx context.Context, out chan<- GameEvent, alerts <-chan PlayerAlert, pstate <-chan PlayerState) {
	// Timer para controlar velocidade do monstro (100ms)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Jogo terminou, sair da goroutine
			return

		case playerState := <-pstate:
			// Recebeu nova posição do jogador
			m.updatePlayerPosition(playerState)

		case <-ticker.C:
			// Hora de tentar mover o monstro
			if m.shouldMove() {
				m.processMovement(out)
			}
		}
	}
}

// Controla velocidade: monstro move a cada 2 turnos (50% mais lento)
func (m *Monster) shouldMove() bool {
	m.shift_count++
	if m.shift_count >= 2 {
		m.shift_count = 0 // Reset contador
		return true
	}
	return false
}

// Processa nova posição do jogador e detecta se pode ver (20 células)
func (m *Monster) updatePlayerPosition(playerState PlayerState) {
	playerPos := Position{X: playerState.X, Y: playerState.Y}

	// Calcula distância até o jogador
	if m.canSeePlayer(playerPos) {
		// Pode ver o jogador - mudar para modo caça
		m.state = Hunting
		m.last_seen = playerPos
		m.destiny_position = playerPos // Perseguir diretamente
	} else if m.state == Hunting {
		// Perdeu o jogador - continuar indo para última posição vista
		m.destiny_position = m.last_seen
		// Se chegou perto da última posição, voltar a patrulhar
		if m.distanceTo(m.last_seen) < 2 {
			m.state = Patrolling
			m.generateRandomDestiny()
		}
	}
}

// Executa movimento baseado no estado atual
func (m *Monster) processMovement(out chan<- GameEvent) {
	// Se está patrulhando e chegou no destino, gerar novo destino
	if m.state == Patrolling && m.distanceTo(m.destiny_position) < 1 {
		m.generateRandomDestiny()
	}

	// Mover em direção ao destino
	m.moveTowards(m.destiny_position)

	// Enviar evento de movimento para o jogo principal
	event := GameEvent{
		Type: "monster_moved",
		Data: map[string]int{
			"x": m.current_position.X,
			"y": m.current_position.Y,
		},
	}

	select {
	case out <- event:
		// Evento enviado com sucesso
	default:
		// Canal cheio, pular este evento
	}
}

// Verifica se pode ver o jogador (máximo 20 células)
func (m *Monster) canSeePlayer(playerPos Position) bool {
	distance := m.distanceTo(playerPos)
	return distance <= 20.0
}

// Calcula distância euclidiana entre monstro e uma posição
func (m *Monster) distanceTo(pos Position) float64 {
	dx := float64(m.current_position.X - pos.X)
	dy := float64(m.current_position.Y - pos.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

// Move o monstro um passo em direção ao alvo
func (m *Monster) moveTowards(target Position) {
	// Calcula direção (normalizada para um passo)
	dx := target.X - m.current_position.X
	dy := target.Y - m.current_position.Y

	// Move apenas 1 célula por vez
	if dx != 0 {
		if dx > 0 {
			m.current_position.X++
		} else {
			m.current_position.X--
		}
	} else if dy != 0 {
		if dy > 0 {
			m.current_position.Y++
		} else {
			m.current_position.Y--
		}
	}
}

// Gera destino aleatório para patrulha
func (m *Monster) generateRandomDestiny() {
	// Gera posição aleatória em um raio de 10 células da posição atual
	radius := 10
	angle := rand.Float64() * 2 * math.Pi
	distance := rand.Float64() * float64(radius)

	newX := m.current_position.X + int(distance*math.Cos(angle))
	newY := m.current_position.Y + int(distance*math.Sin(angle))

	m.destiny_position = Position{X: newX, Y: newY}
}
