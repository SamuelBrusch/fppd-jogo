// jogo.go - Funções para manipular os elementos do jogo, como carregar o mapa e mover o personagem
package main

import (
	"bufio"
	"os"
)

// Elemento representa qualquer objeto do mapa (parede, personagem, vegetação, etc)
type Elemento struct {
	simbolo  rune
	cor      Cor
	corFundo Cor
	tangivel bool // Indica se o elemento bloqueia passagem
}

// Jogo contém o estado atual do jogo
type Jogo struct {
	Mapa           [][]Elemento // grade 2D representando o mapa
	PosX, PosY     int          // posição atual do personagem
	UltimoVisitado Elemento     // elemento que estava na posição do personagem antes de mover
	StatusMsg      string       // mensagem para a barra de status
	Monstro        *Monster     // instância do monstro
	// Canais de comunicação
	GameEvents   chan GameEvent   // canal para eventos do jogo
	PlayerState  chan PlayerState // canal para estado do jogador
	PlayerAlerts chan PlayerAlert // canal para alertas do jogador
	StarEvents chan GameEvent       // novo canal para eventos de estrela
    Collected  chan PlayerCollect   // novo canal para avisar coleta
}

// Elementos visuais do jogo
var (
	Personagem = Elemento{'☺', CorCinzaEscuro, CorPadrao, true}
	Inimigo    = Elemento{'☠', CorVermelho, CorPadrao, true}
	Parede     = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao  = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio      = Elemento{' ', CorPadrao, CorPadrao, false}
	Estrela	= Elemento{'*', CorAmarelo, CorPadrao, false}
)

// Cria e retorna uma nova instância do jogo
func jogoNovo() Jogo {
	// O ultimo elemento visitado é inicializado como vazio
	// pois o jogo começa com o personagem em uma posição vazia
	return Jogo{
		UltimoVisitado: Vazio,
		GameEvents:     make(chan GameEvent, 10),
		PlayerState:    make(chan PlayerState, 10),
		PlayerAlerts:   make(chan PlayerAlert, 10),
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
			case Personagem.simbolo:
				jogo.PosX, jogo.PosY = x, y // registra a posição inicial do personagem
			
			case Estrela.simbolo:
				e = Estrela
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
	// Verifica se a coordenada Y está dentro dos limites verticais do mapa
	if y < 0 || y >= len(jogo.Mapa) {
		return false
	}

	// Verifica se a coordenada X está dentro dos limites horizontais do mapa
	if x < 0 || x >= len(jogo.Mapa[y]) {
		return false
	}

	// Verifica se o elemento de destino é tangível (bloqueia passagem)
	if jogo.Mapa[y][x].tangivel {
		return false
	}

	// Pode mover para a posição
	return true
}

// Move um elemento para a nova posição
func jogoMoverElemento(jogo *Jogo, x, y, dx, dy int) {
	nx, ny := x+dx, y+dy

	// Obtem elemento atual na posição
	elemento := jogo.Mapa[y][x] // guarda o conteúdo atual da posição

	jogo.Mapa[y][x] = jogo.UltimoVisitado   // restaura o conteúdo anterior
	jogo.UltimoVisitado = jogo.Mapa[ny][nx] // guarda o conteúdo atual da nova posição
	jogo.Mapa[ny][nx] = elemento            // move o elemento
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

// Trata um evento específico do jogo
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
						// Enviar evento de colisão de volta
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
							// Canal cheio
						}
					}
				}
			}
		}
	case "monster_collision":
		// Monstro colidiu com jogador
		jogo.StatusMsg = "GAME OVER! Você foi pego pelo monstro!"
	case "monster_timeout":
		// Monstro entrou em timeout - mostrar mensagem
		if data, ok := event.Data.(map[string]interface{}); ok {
			if message, hasMsg := data["message"]; hasMsg {
				if msgStr, isString := message.(string); isString {
					jogo.StatusMsg = "Alerta: " + msgStr
				}
			}
		}
	}
}

// Envia estado atual do jogador para o monstro
// (Função mantida para uso futuro ou integração, suprimindo erro de função não utilizada)
var _ = jogoEnviarEstadoJogador

func jogoEnviarEstadoJogador(jogo *Jogo) {
	select {
	case jogo.PlayerState <- PlayerState{X: jogo.PosX, Y: jogo.PosY}:
		// Estado enviado com sucesso
	default:
		// Canal cheio, pular este envio
	}
}

// Envia alerta para o monstro
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
		// Alerta enviado com sucesso
	default:
		// Canal cheio, pular este envio
	}
}

// Suprime erro de função não utilizada
var _ = jogoEnviarAlerta
