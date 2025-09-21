// element_star.go - Implementação do elemento estrela concorrente
package main

import (
	"context"
	"math/rand"
	"time"
)

// Estados da estrela
type StarState int

const (
	StarVisible   StarState = iota // Estrela visível e coletável
	StarInvisible                  // Estrela invisível (não coletável)
	StarPulsing                    // Estrela pulsando (alternando visibilidade)
	StarCharging                   // Estrela carregando energia
)

// Tipos de eventos da estrela
const (
	EventStarCollected    = "StarCollected"
	EventStarStateChange  = "StarStateChange"
	EventStarPulse        = "StarPulse"
	EventStarCharged      = "StarCharged"
	EventStarTimeout      = "StarTimeout"
	EventStarCommunicate  = "StarCommunicate"
	EventRequestMapAccess = "RequestMapAccess"
	EventMapAccessGranted = "MapAccessGranted"
)

// Duração dos estados da estrela
const (
	StarVisibilityDuration = 8 * time.Second  // Tempo visível
	StarInvisibleDuration  = 4 * time.Second  // Tempo invisível
	StarPulseDuration      = 2 * time.Second  // Duração de uma pulsação
	StarChargeDuration     = 10 * time.Second // Tempo para carregar
	StarTimeoutDuration    = 15 * time.Second // Timeout para mudança de comportamento
)

// Dados para eventos da estrela
type StarCollectedData struct {
	X, Y      int
	BonusType string // "score", "life", "power"
	Value     int
}

type StarStateChangeData struct {
	X, Y     int
	OldState StarState
	NewState StarState
	StarID   string
}

type StarPulseData struct {
	X, Y       int
	IsVisible  bool
	PulseCount int
}

type StarChargedData struct {
	X, Y     int
	Energy   int
	Duration time.Duration
}

type StarTimeoutData struct {
	X, Y    int
	Message string
	Action  string
	StarID  string
}

type StarCommunicationData struct {
	FromStarID string
	ToStarID   string
	Message    string
	Data       interface{}
}

// Comando para controle da estrela
type StarCommand struct {
	Type   string      // "change_state", "pulse", "charge", "communicate"
	Target string      // ID da estrela alvo
	Data   interface{} // Dados específicos do comando
}

// Estrutura da estrela (já definida em types.go, mas vamos estender)
type Star struct {
	X, Y          int            // Posição da estrela
	State         StarState      // Estado atual
	ID            string         // ID único da estrela
	IsVisible     bool           // Se está visível no momento
	Energy        int            // Energia acumulada
	PulseCount    int            // Contador de pulsações
	LastPlayerPos Position       // Última posição conhecida do jogador
	MapAccess     chan chan bool // Canal para requisitar acesso ao mapa
}

// Cria uma nova estrela
func NewStar(x, y int, id string) *Star {
	return &Star{
		X:         x,
		Y:         y,
		State:     StarVisible,
		ID:        id,
		IsVisible: true,
		Energy:    0,
		MapAccess: make(chan chan bool, 1),
	}
}

// Goroutine principal da estrela - atende TODOS os requisitos de concorrência
func (s *Star) Run(ctx context.Context, gameEvents chan<- GameEvent, playerState <-chan PlayerState,
	playerCollects <-chan PlayerCollect, starCommands <-chan StarCommand, mapMutex chan chan bool) {

	// Timers para diferentes comportamentos
	visibilityTimer := time.NewTimer(StarVisibilityDuration)
	pulseTimer := time.NewTimer(StarPulseDuration)
	chargeTimer := time.NewTimer(StarChargeDuration)
	timeoutTimer := time.NewTimer(StarTimeoutDuration)

	defer visibilityTimer.Stop()
	defer pulseTimer.Stop()
	defer chargeTimer.Stop()
	defer timeoutTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		// REQUISITO: Escuta concorrente de múltiplos canais
		case playerPos := <-playerState:
			s.LastPlayerPos = Position{X: playerPos.X, Y: playerPos.Y}
			s.handlePlayerMovement(gameEvents, playerPos)

		case collect := <-playerCollects:
			if collect.X == s.X && collect.Y == s.Y && s.IsVisible && s.State == StarVisible {
				s.handleCollection(gameEvents, collect)
				return // Estrela coletada, termina goroutine
			}

		case command := <-starCommands:
			s.handleStarCommand(gameEvents, command)

		// REQUISITO: Comunicação com timeout
		case <-timeoutTimer.C:
			s.handleTimeout(gameEvents)
			timeoutTimer.Reset(StarTimeoutDuration)

		case <-visibilityTimer.C:
			s.toggleVisibility(gameEvents, mapMutex)
			visibilityTimer.Reset(s.getNextVisibilityDuration())

		case <-pulseTimer.C:
			if s.State == StarPulsing {
				s.handlePulse(gameEvents, mapMutex)
				pulseTimer.Reset(StarPulseDuration)
			}

		case <-chargeTimer.C:
			if s.State == StarCharging {
				s.handleChargeComplete(gameEvents)
				chargeTimer.Reset(StarChargeDuration)
			}

		// REQUISITO: Exclusão mútua usando canais
		case responseChan := <-s.MapAccess:
			s.requestMapAccess(mapMutex, responseChan)
		}
	}
}

// Manipula movimento do jogador próximo à estrela
func (s *Star) handlePlayerMovement(gameEvents chan<- GameEvent, playerPos PlayerState) {
	distance := abs(playerPos.X-s.X) + abs(playerPos.Y-s.Y) // Distância Manhattan

	// Se jogador está próximo (distância 1), muda comportamento
	if distance <= 1 && s.State != StarPulsing {
		s.changeState(StarPulsing, gameEvents)
	} else if distance > 3 && s.State == StarPulsing {
		s.changeState(StarVisible, gameEvents)
	}
}

// Manipula coleta da estrela
func (s *Star) handleCollection(gameEvents chan<- GameEvent, collect PlayerCollect) {
	bonusType := "score"
	value := 100

	// Bônus especial se coletada em estado especial
	switch s.State {
	case StarPulsing:
		bonusType = "power"
		value = 300
	case StarCharging:
		bonusType = "life"
		value = 1
	default:
		// Bônus baseado na energia acumulada
		value += s.Energy * 10
	}

	gameEvents <- GameEvent{
		Type: EventStarCollected,
		Data: StarCollectedData{
			X:         s.X,
			Y:         s.Y,
			BonusType: bonusType,
			Value:     value,
		},
	}

	// Remover estrela do mapa
	gameEvents <- GameEvent{
		Type: EventRemoveElement,
		Data: StarBonus{X: s.X, Y: s.Y},
	}
}

// Manipula comandos de outras estrelas ou elementos
func (s *Star) handleStarCommand(gameEvents chan<- GameEvent, command StarCommand) {
	switch command.Type {
	case "change_state":
		if data, ok := command.Data.(StarState); ok {
			s.changeState(data, gameEvents)
		}
	case "pulse":
		if s.State != StarPulsing {
			s.changeState(StarPulsing, gameEvents)
		}
	case "charge":
		if s.State != StarCharging {
			s.changeState(StarCharging, gameEvents)
		}
	case "communicate":
		if data, ok := command.Data.(StarCommunicationData); ok {
			s.handleCommunication(gameEvents, data)
		}
	}
}

// REQUISITO: Comunicação com timeout
func (s *Star) handleTimeout(gameEvents chan<- GameEvent) {
	// Comportamento alternativo quando não recebe interação por tempo limite
	actions := []string{"charge", "pulse", "hide", "energy_burst"}
	action := actions[rand.Intn(len(actions))]

	switch action {
	case "charge":
		s.changeState(StarCharging, gameEvents)
	case "pulse":
		s.changeState(StarPulsing, gameEvents)
	case "hide":
		s.changeState(StarInvisible, gameEvents)
	case "energy_burst":
		s.Energy += 50
		gameEvents <- GameEvent{
			Type: EventStarCharged,
			Data: StarChargedData{
				X:        s.X,
				Y:        s.Y,
				Energy:   s.Energy,
				Duration: StarChargeDuration,
			},
		}
	}

	gameEvents <- GameEvent{
		Type: EventStarTimeout,
		Data: StarTimeoutData{
			X:       s.X,
			Y:       s.Y,
			Message: "Estrela mudou comportamento por timeout",
			Action:  action,
			StarID:  s.ID,
		},
	}
}

// Alterna visibilidade da estrela
func (s *Star) toggleVisibility(gameEvents chan<- GameEvent, mapMutex chan chan bool) {
	// REQUISITO: Exclusão mútua para acesso ao mapa
	responseChan := make(chan bool)
	mapMutex <- responseChan
	<-responseChan // Aguarda liberação do acesso

	s.IsVisible = !s.IsVisible

	if s.IsVisible {
		s.changeState(StarVisible, gameEvents)
	} else {
		s.changeState(StarInvisible, gameEvents)
	}
}

// Manipula pulsação da estrela
func (s *Star) handlePulse(gameEvents chan<- GameEvent, mapMutex chan chan bool) {
	responseChan := make(chan bool)
	mapMutex <- responseChan
	<-responseChan // Exclusão mútua

	s.IsVisible = !s.IsVisible
	s.PulseCount++

	gameEvents <- GameEvent{
		Type: EventStarPulse,
		Data: StarPulseData{
			X:          s.X,
			Y:          s.Y,
			IsVisible:  s.IsVisible,
			PulseCount: s.PulseCount,
		},
	}

	// Após 10 pulsações, volta ao estado normal
	if s.PulseCount >= 10 {
		s.PulseCount = 0
		s.changeState(StarVisible, gameEvents)
	}
}

// Manipula carregamento completo de energia
func (s *Star) handleChargeComplete(gameEvents chan<- GameEvent) {
	s.Energy += 100

	gameEvents <- GameEvent{
		Type: EventStarCharged,
		Data: StarChargedData{
			X:        s.X,
			Y:        s.Y,
			Energy:   s.Energy,
			Duration: StarChargeDuration,
		},
	}

	// Volta ao estado visível após carregar
	s.changeState(StarVisible, gameEvents)
}

// Manipula comunicação entre estrelas
func (s *Star) handleCommunication(gameEvents chan<- GameEvent, data StarCommunicationData) {
	// Processa mensagem de outra estrela
	switch data.Message {
	case "sync_pulse":
		s.changeState(StarPulsing, gameEvents)
	case "share_energy":
		if energy, ok := data.Data.(int); ok {
			s.Energy += energy / 2 // Recebe metade da energia
		}
	case "warning":
		// Muda para modo defensivo
		s.changeState(StarCharging, gameEvents)
	}

	gameEvents <- GameEvent{
		Type: EventStarCommunicate,
		Data: data,
	}
}

// Muda estado da estrela
func (s *Star) changeState(newState StarState, gameEvents chan<- GameEvent) {
	oldState := s.State
	s.State = newState

	// Ajusta visibilidade baseada no estado
	switch newState {
	case StarVisible:
		s.IsVisible = true
	case StarInvisible:
		s.IsVisible = false
	case StarPulsing:
		// Mantém visibilidade atual
	case StarCharging:
		s.IsVisible = true
	}

	gameEvents <- GameEvent{
		Type: EventStarStateChange,
		Data: StarStateChangeData{
			X:        s.X,
			Y:        s.Y,
			OldState: oldState,
			NewState: newState,
			StarID:   s.ID,
		},
	}
}

// REQUISITO: Exclusão mútua usando canais
func (s *Star) requestMapAccess(mapMutex chan chan bool, responseChan chan bool) {
	// Solicita acesso ao mapa
	mapMutex <- responseChan
}

// Calcula próxima duração de visibilidade
func (s *Star) getNextVisibilityDuration() time.Duration {
	if s.IsVisible {
		return StarVisibilityDuration
	}
	return StarInvisibleDuration
}

// Função auxiliar para valor absoluto
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Função compatível com a interface existente (StarBonus)
func (s *StarBonus) Run(ctx context.Context, out chan<- GameEvent, collected <-chan PlayerCollect) {
	// Converte StarBonus para Star para usar a implementação completa
	star := NewStar(s.X, s.Y, "legacy_star")

	// Criar canais necessários
	playerState := make(chan PlayerState, 10)
	starCommands := make(chan StarCommand, 10)
	mapMutex := make(chan chan bool, 1)

	// Goroutine para gerenciar exclusão mútua
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case responseChan := <-mapMutex:
				responseChan <- true // Simula liberação imediata
			}
		}
	}()

	// Executar a implementação completa
	star.Run(ctx, out, playerState, collected, starCommands, mapMutex)
}
