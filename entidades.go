package main

import (
    "time"
    "fmt"
)

// Estrutura que representa o Ancião
type Anciao struct {
    X, Y int // Posição atual do ancião
}

// Função que movimenta o ancião de um lado para outro no mapa
func (a *Anciao) mover(jogo *Jogo) {
    direcaoX := 1 // 1 para direita, -1 para esquerda
    direcaoY := 1 // 1 para baixo, -1 para cima

    for {
        time.Sleep(500 * time.Millisecond) // Controla a velocidade do movimento

        // Atualiza a posição do ancião
        a.X += direcaoX
        a.Y += direcaoY

        // Verifica os limites horizontais
        if a.X <= 35 || a.X >= 52 {
            direcaoX *= -1 // Inverte a direção horizontal
        }

        // Verifica os limites verticais
        if a.Y <= 26 || a.Y >= 36 {
            direcaoY *= -1 // Inverte a direção vertical
        }

        // Verifica se o personagem está na mesma posição
        if a.X == jogo.PosX && a.Y == jogo.PosY {
            jogo.StatusMsg = "Você encontrou o ancião e recebeu uma espada!"
        }
        fmt.Printf("Anciao está em (%d, %d)\n", a.X, a.Y)
    }

    
}