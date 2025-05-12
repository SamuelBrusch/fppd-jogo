// jogo.go - Funções para manipular os elementos do jogo, como carregar o mapa e mover o personagem
package main

import (
	"fmt"
	"bufio"
	"os"
	"sync"
)

// Elemento representa qualquer objeto do mapa (parede, personagem, vegetação, etc)
type Elemento struct {
	simbolo   rune
	cor       Cor
	corFundo  Cor
	tangivel  bool // Indica se o elemento bloqueia passagem
}

// Jogo contém o estado atual do jogo
type Jogo struct {
	Mapa            [][]Elemento // grade 2D representando o mapa
	PosX, PosY      int          // posição atual do personagem
	UltimoVisitado  Elemento     // elemento que estava na posição do personagem antes de mover
	StatusMsg       string       // mensagem para a barra de status
	Mutex     sync.Mutex
}

// Elementos visuais do jogo
var (
	AnciaoElem = Elemento{'A', CorVerde, CorPadrao, false} 
	Personagem = Elemento{'☺', CorCinzaEscuro, CorPadrao, true}
	InimigoElem= Elemento{'☠', CorVermelho, CorPadrao, true}
	Parede     = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao  = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio      = Elemento{' ', CorPadrao, CorPadrao, false}
)

// Cria e retorna uma nova instância do jogo
func jogoNovo() Jogo {
	// O ultimo elemento visitado é inicializado como vazio
	// pois o jogo começa com o personagem em uma posição vazia
	return Jogo{UltimoVisitado: Vazio}
}

// Lê um arquivo texto linha por linha e constrói o mapa do jogo
func jogoCarregarMapa(nome string, jogo *Jogo, anciao *Anciao) ([]*Inimigo, chan struct{}, error) {
    startChan := make(chan struct{})
	arq, err := os.Open(nome)
    if err != nil {
		return nil, nil, err
    }
    defer arq.Close()

    var inimigos []*Inimigo
    var mapa [][]Elemento  // Criamos um mapa temporário

    scanner := bufio.NewScanner(arq)
    for y := 0; scanner.Scan(); y++ {
        linha := scanner.Text()
        var linhaElems []Elemento
        
        for x, ch := range linha {
            e := Vazio
            switch ch {
            case AnciaoElem.simbolo:
                anciao.X, anciao.Y = x, y
                e = Vazio  // Remove o ancião do mapa estático
            case Parede.simbolo:
                e = Parede
            case InimigoElem.simbolo:
                inimigos = append(inimigos, &Inimigo{
                    X: x,
                    Y: y,
                    ativo: false,
                })
                e = Vazio  // Remove o inimigo do mapa estático
            case Vegetacao.simbolo:
                e = Vegetacao
            case Personagem.simbolo:
                jogo.PosX, jogo.PosY = x, y
                e = Vazio  // Personagem será desenhado separadamente
            }
            linhaElems = append(linhaElems, e)
        }
        mapa = append(mapa, linhaElems)
    }

    jogo.Mapa = mapa  // Atribui o mapa completo ao jogo
    
    if err := scanner.Err(); err != nil {
        return nil, nil, err
    }
    return inimigos, startChan, nil
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

	jogo.Mapa[y][x] = jogo.UltimoVisitado     // restaura o conteúdo anterior
	jogo.UltimoVisitado = jogo.Mapa[ny][nx]   // guarda o conteúdo atual da nova posição
	jogo.Mapa[ny][nx] = elemento              // move o elemento
}

// Define o que ocorre quando o jogador
func jogoInteragir(jogo *Jogo) {
	// Atualmente apenas exibe uma mensagem de status
	jogo.StatusMsg = fmt.Sprintf("Interagindo em (%d, %d)", jogo.PosX, jogo.PosY)
}
