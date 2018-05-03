package dna

import "errors"

//const N = 6				// Dimension del array de ADN
const M = 4				// Longitud de la secuencia buscada
const Q = 2				// Cantidad de secuencias buscadas

type Dna struct {
	Matrix  []string `json:"dna"`
}


func isValidIndex(index int, dim int) bool {
	return index >= 0 && index < dim
}


// Verifica si el array recibido representa a una matriz cuadrada
func isValidDNA(dna Dna) bool {
	n := len(dna.Matrix)
	for i := 0; i < n; i++ {
		if len(dna.Matrix[i]) != n {
			return false
		}
	}
	return true
}


// Determina si un ADN es mutante. Para esto, busca una secuencia de 4 elementos iguales consecutivos
// en sentido horizontal, vertical o en diagonal. Para determinar la dimension de la matriz, mira solamente
// la cantidad de strings en el paramtro 'dna' y asume que la longitud de todos los strings es la adecuada
func IsMutant(dna Dna) (bool, error) {

	found := 0											// Cantidad de secuencias encontradas
	moveRow := [4]int {0, 1, 1, -1}						// Arrays con las direcciones de los desplazamientos desde de un elemento
	moveCol := [4]int {1, 0, 1, 1}

	if !isValidDNA(dna) {								// Valida las dimensiones de la matriz de ADN
		return false, errors.New("Las dimensiones del ADN no son validas")
	}

	dimDNA := len(dna.Matrix)							// Dimension de la matriz de ADN
	for i := 0; i < dimDNA; i++ {						// i para recorrer las filas
		for j := 0; j < dimDNA && found < Q; j++ {		// j para recorrer las columnas
			elem := dna.Matrix[i][j]
			for k := 0; k<4; k++ {						// k para recorrer los arrays de direcciones
				ok := true
				for m := 1; m < M && ok; m++ {			// m es el multiplicador (para moverse a lo largo de una direccion)
					row := i+m*moveRow[k]				// Determina fila y columna del elemento adyacente en la direccion k
					col := j+m*moveCol[k]
					ok = (isValidIndex(row, dimDNA) && isValidIndex(col, dimDNA) && (elem == dna.Matrix[row][col]))
				}
				if ok {found++}
			}
		}
	}
	return found >= Q, nil
}
