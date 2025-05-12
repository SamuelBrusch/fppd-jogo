// personagem.go - Funções para movimentação e ações do personagem
package main

// Atualiza a posição do personagem com base na tecla pressionada (WASD)
func personagemMover(tecla rune, jogo *Jogo) {
	dx, dy := 0, 0
	switch tecla {
	case 'w': dy = -1 
	case 'a': dx = -1
	case 's': dy = 1  
	case 'd': dx = 1 
	}

	nx, ny := jogo.PosX+dx, jogo.PosY+dy
	if jogoPodeMoverPara(jogo, nx, ny) {
		jogoMoverElemento(jogo, jogo.PosX, jogo.PosY, dx, dy)
		jogo.PosX, jogo.PosY = nx, ny
    }
}


func personagemExecutarAcao(ev EventoTeclado, jogo *Jogo, inimigos []*Inimigo) bool {
    switch ev.Tipo {
    case "sair":
        return false
        
    case "interagir":
        if jogo.Mapa[jogo.PosY][jogo.PosX].simbolo == AnciaoElem.simbolo {
            for _, inimigo := range inimigos {
                inimigo.ativo = true
            }
            jogo.StatusMsg = "Ancião: Cuidado com os inimigos!"
        }
        
        for _, inimigo := range inimigos {
            if distancia(inimigo.X, inimigo.Y, jogo.PosX, jogo.PosY) <= 2 {
                inimigo.comando <- "parar"
                jogo.Mapa[inimigo.Y][inimigo.X] = Vazio
                jogo.StatusMsg = "Inimigo eliminado!"
            }
        }
        return true
        
    case "mover":
        personagemMover(ev.Tecla, jogo)
        return true
        
    default:
        return true
    }
}
