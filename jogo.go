package main

import (
	"bufio"
	"fmt"
	"os"
)

// Elemento representa qualquer objeto do mapa (parede, personagem, vegetação, etc)
type Elemento struct {
	simbolo  rune
	cor      Cor
	corFundo Cor
	tangivel bool
}

type Jogo struct {
	Mapa           [][]Elemento // grade 2D representando o mapa
	PosX, PosY     int          // posição atual do personagem
	UltimoVisitado Elemento     // elemento que estava na posição do personagem antes de mover
	StatusMsg      string       // mensagem para a barra de status
	InvisibleSteps int          // contador de invisibilidade do personagem (em passos)
	DoubleJumps    int          // contador de pulos duplos restantes
	Monstro        *Monster     // instância do monstro
	InvisibilityItems []*Invisibility // lista de itens de invisibilidade
	Stars []*Star // lista de estrelas
	GameEvents     chan GameEvent     // canal para eventos do jogo
	PlayerState    chan PlayerState   // canal para estado do jogador
	PlayerAlerts   chan PlayerAlert   // canal para alertas do jogador
	PlayerCollects chan PlayerCollect // canal para coletas do jogador
	StarCommands   chan StarCommand   // canal para comandos das estrelas
	MapMutex       chan chan bool     // canal para exclusão mútua do mapa
}

// Elementos visuais do jogo
var (
	Personagem          = Elemento{'☺', CorCinzaEscuro, CorPadrao, true}
	Inimigo             = Elemento{'☠', CorVermelho, CorPadrao, true}
	Parede              = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao           = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio               = Elemento{' ', CorPadrao, CorPadrao, false}
	InvisibilityItem    = Elemento{'¤', CorAmarelo, CorPadrao, false}
	PersonagemInvisivel = Elemento{'☺', CorTexto, CorPadrao, true}
	StarElementVisible   = Elemento{'★', CorAmarelo, CorPadrao, false}
	StarElementInvisible = Elemento{' ', CorPadrao, CorPadrao, false}
	StarElementPulsing   = Elemento{'✦', CorCinzaEscuro, CorPadrao, false}
	StarElementCharging  = Elemento{'◉', CorVermelho, CorPadrao, false}
)

func jogoNovo() Jogo {
	return Jogo{
		UltimoVisitado: Vazio,
		GameEvents:     make(chan GameEvent, 10),
		PlayerState:    make(chan PlayerState, 10),
		PlayerAlerts:   make(chan PlayerAlert, 10),
		PlayerCollects: make(chan PlayerCollect, 10),
		StarCommands:   make(chan StarCommand, 10),
		MapMutex:       make(chan chan bool, 1),
	}
}

// Lê um arquivo texto linha por linha e constrói o mapa do jogo
func jogoCarregarMapa(nome string, jogo *Jogo) error {
	arq, err := os.Open(nome)
	if err != nil {
		return err
	}
	defer arq.Close()

	scanner := bufio.NewScanner(arq)
	y := 0
	for scanner.Scan() {
		linha := scanner.Text()
		var linhaElems []Elemento
		for x, ch := range linha {
			e := Vazio
			switch ch {
			case Parede.simbolo:
				e = Parede
			case Inimigo.simbolo:
				e = Vazio // Não desenhar o inimigo no mapa, será desenhado separadamente
				// Inicializar monstro se ainda não foi criado
				if jogo.Monstro == nil {
					jogo.Monstro = &Monster{
						current_position: Position{X: x, Y: y},
						state:            Patrolling,
						destiny_position: Position{X: x + 5, Y: y + 5},
						id:               "monster_1",
					}
				}
			case Vegetacao.simbolo:
				e = Vegetacao
			case InvisibilityItem.simbolo:
				e = InvisibilityItem
				invisItem := &Invisibility{
					X: x,
					Y: y,
				}
				jogo.InvisibilityItems = append(jogo.InvisibilityItems, invisItem)
			case '★':
				e = StarElementVisible
			case Personagem.simbolo:
				jogo.PosX, jogo.PosY = x, y // registra a posição inicial do personagem
			}
			linhaElems = append(linhaElems, e)
		}
		jogo.Mapa = append(jogo.Mapa, linhaElems)
		y++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// Verifica se o personagem pode se mover para a posição (x, y)
func jogoPodeMoverPara(jogo *Jogo, x, y int) bool {
	if y < 0 || y >= len(jogo.Mapa) {
		return false
	}

	if x < 0 || x >= len(jogo.Mapa[y]) {
		return false
	}

	if jogo.Mapa[y][x].tangivel {
		return false
	}

	return true
}

// Move um elemento para a nova posição
func jogoMoverElemento(jogo *Jogo, x, y, dx, dy int) {
	nx, ny := x+dx, y+dy

	// Obtem elemento atual na posição
	elemento := jogo.Mapa[y][x] // guarda o conteúdo atual da posição

	jogo.Mapa[y][x] = jogo.UltimoVisitado  
	jogo.UltimoVisitado = jogo.Mapa[ny][nx] 
	jogo.Mapa[ny][nx] = elemento    
}

// Retorna o elemento visual do jogador (normal ou invisível)
func (j *Jogo) elementoJogador() Elemento {
	if j.InvisibleSteps > 0 {
		return PersonagemInvisivel
	}
	return Personagem
}

// Processa eventos vindos do monstro
func jogoProcessarEventos(jogo *Jogo) {
	select {
	case event := <-jogo.GameEvents:
		jogoTratarEvento(jogo, event)
	default:
		// Não há eventos para processar
	}
}

func jogoTratarEvento(jogo *Jogo, event GameEvent) {
	switch event.Type {
	case "monster_move":
		// Monstro quer se mover
		if data, ok := event.Data.(MonsterMoveData); ok {
			// Verificar se o movimento é válido
			if jogoPodeMoverPara(jogo, data.NewX, data.NewY) {
				// Atualizar posição do monstro
				if jogo.Monstro != nil && jogo.Monstro.id == data.MonsterID {
					jogo.Monstro.current_position = Position{X: data.NewX, Y: data.NewY}

					// Verificar colisão com jogador
					if data.NewX == jogo.PosX && data.NewY == jogo.PosY {
						// Enviar evento de colisão
						collisionEvent := GameEvent{
							Type: "monster_collision",
							Data: map[string]interface{}{
								"x":    data.NewX,
								"y":    data.NewY,
								"type": "movement",
							},
						}
						select {
						case jogo.GameEvents <- collisionEvent:
						default:
						}
					}
				}
			}
		}
	case "monster_collision":
		jogo.StatusMsg = "Pego pelo monstro!"
	case "monster_timeout":
		if data, ok := event.Data.(map[string]interface{}); ok {
			if message, hasMsg := data["message"]; hasMsg {
				if msgStr, isString := message.(string); isString {
					jogo.StatusMsg = "Alerta: " + msgStr
				}
			}
		}
	case EventApplyInvisibility:
		if data, ok := event.Data.(InvisibilityApplied); ok {
			jogo.InvisibleSteps = data.Duration
			jogo.StatusMsg = "Invisibilidade coletada!"
		}
	case EventRemoveElement:
		// Remover item do mapa
		if data, ok := event.Data.(Invisibility); ok {
			if data.X == jogo.PosX && data.Y == jogo.PosY {
				jogo.UltimoVisitado = Vazio
			} else if data.Y >= 0 && data.Y < len(jogo.Mapa) &&
				data.X >= 0 && data.X < len(jogo.Mapa[data.Y]) {
				jogo.Mapa[data.Y][data.X] = Vazio
			}
		}
		if data, ok := event.Data.(StarBonus); ok {
			if data.X == jogo.PosX && data.Y == jogo.PosY {
				jogo.UltimoVisitado = Vazio
			} else if data.Y >= 0 && data.Y < len(jogo.Mapa) &&
				data.X >= 0 && data.X < len(jogo.Mapa[data.Y]) {
				jogo.Mapa[data.Y][data.X] = Vazio
			}
		}
	case EventStarCollected:
		if data, ok := event.Data.(StarCollectedData); ok {
			jogo.StatusMsg = fmt.Sprintf("Estrela coletada! %s +%d", data.BonusType, data.Value)
		}
	case EventStarStateChange:
		if data, ok := event.Data.(StarStateChangeData); ok {
			jogo.StatusMsg = fmt.Sprintf("Estrela %s mudou de estado", data.StarID)
		}
	case EventStarPulse:
		if data, ok := event.Data.(StarPulseData); ok {
			jogo.StatusMsg = fmt.Sprintf("Estrela pulsando (%d pulsos)", data.PulseCount)
		}
	case EventStarCharged:
		if data, ok := event.Data.(StarChargedData); ok {
			jogo.StatusMsg = fmt.Sprintf("Estrela carregada! Energia: %d", data.Energy)
		}
	case EventStarTimeout:
		if data, ok := event.Data.(StarTimeoutData); ok {
			jogo.StatusMsg = data.Message
		}
	case EventStarCommunicate:
		if data, ok := event.Data.(StarCommunicationData); ok {
			jogo.StatusMsg = fmt.Sprintf("Estrelas comunicando: %s", data.Message)
		}
	case "ApplyDoubleJump":
		// Boost de pulo duplo foi coletado
		if data, ok := event.Data.(DoubleJumpApplied); ok {
			jogo.DoubleJumps = data.Jumps
			jogo.StatusMsg = "Estrela coletada! Pulo duplo ativado!"
		}
	}
}

func jogoEnviarEstadoJogador(jogo *Jogo) {
	select {
	case jogo.PlayerState <- PlayerState{X: jogo.PosX, Y: jogo.PosY}:
	default:
	}
}

func jogoEnviarAlerta(jogo *Jogo, tipoAlerta string) {
	alert := PlayerAlert{
		Type: tipoAlerta,
		Data: map[string]int{
			"x": jogo.PosX,
			"y": jogo.PosY,
		},
	}

	select {
	case jogo.PlayerAlerts <- alert:
	default:
	}
}

// Gerencia exclusão mútua do mapa usando canais
func jogoGerenciarMapMutex(jogo *Jogo) {
	for responseChan := range jogo.MapMutex {
		responseChan <- true
	}
}

func jogoEnviarComandoEstrela(jogo *Jogo, command StarCommand) {
	select {
	case jogo.StarCommands <- command:
	default:
	}
}

// Remoção da estrela "sob" o jogador quando consumida.
func ConsumirItemEstrela(jogo *Jogo) bool {
	if jogo.UltimoVisitado.simbolo == StarElementVisible.simbolo {
		jogo.UltimoVisitado = Vazio
		return true
	}
	return false
}

func jogoGetStarElement(star *Star) Elemento {
	if !star.IsVisible {
		return StarElementInvisible
	}

	switch star.State {
	case StarVisible:
		return StarElementVisible
	case StarInvisible:
		return StarElementInvisible
	case StarPulsing:
		return StarElementPulsing
	case StarCharging:
		return StarElementCharging
	default:
		return StarElementVisible
	}
}
