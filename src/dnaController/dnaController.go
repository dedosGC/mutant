package dnaController

import (
	"fmt"
	"strings"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/gin-gonic/gin"
	"dna"
)

const (
	dnaKind = `DNA`
	counterKind = `Counter`
	mutantDNAType = `mutant`
	humanDNAType = `human`
	invalidDNA = `Error: ADN no valido`
)

type DNAEntity struct {
	Matrix  string
	DnaType string
}

type CounterEntity struct {
	Count int
}

type ResponseObject map[string]interface{}


// Verifica si el ADN enviado es mutante. Devuelve '200 OK' si es mutante,
// '403 Forbidden' si es humano y '422 Unprocessable Entity' si el ADN enviado no es valido
func IsMutant(ctx *gin.Context) {
	var adn dna.Dna
	var err error
	newctx := appengine.NewContext(ctx.Request)
	mutant := false

	if err = ctx.BindJSON(&adn); err == nil {
		log.Infof(newctx, "ADN: %s", strings.Join(adn.Matrix, ";"))

		mutant, err = dna.IsMutant(adn)
		if err != nil {
			log.Errorf(newctx, "ADN no valido: %v", err.Error())
			ctx.Status(http.StatusUnprocessableEntity)		// Si el ADN enviado no es valido devuelve 422-Unprocessable Entity
			return
		}

		if mutant, err = dna.IsMutant(adn); mutant {		// Verifica si el ADN es mutante
			log.Infof(newctx, "ADN mutante")
			ctx.Status(http.StatusOK)						// Si es mutante devuelve 200-Ok
		} else {
			log.Infof(newctx, "ADN humano")
			ctx.Status(http.StatusForbidden)				// Si no es mutante devuelve 403-Forbidden
		}
		SaveDNA(ctx, adn, mutant)							// Guarda el ADN en la base de datos
	} else {
		log.Errorf(newctx, "Error: %v", err.Error())
		ctx.String(http.StatusPreconditionFailed, "Error: %v", err.Error())
	}
}


// Devuelve las estadisticas de los ADNs
func Statistics(ctx *gin.Context) {
	stats := GetStatistics(ctx)
	ctx.JSON(http.StatusOK, ResponseObject{"count_mutant_dna": stats["mutants"], "count_human_dna": stats["humans"], "mutants_ratio": stats["mutants_ratio"], "humans_ratio": stats["humans_ratio"]})
}


// Obtiene los datos estadisiticos de los ADNs
func GetStatistics(ctx *gin.Context) map[string] string {

	st := make(map[string] string)

	mutants := GetDNATypeCount(ctx, "mutant")				// Obtiene la cantidad de ADNs mutantes
	st["mutants"] = fmt.Sprintf("%d", mutants)

	humans := GetDNATypeCount(ctx, "human")					// Obtiene la cantidad de ADNs humanos
	st["humans"] = fmt.Sprintf("%d", humans)

	total := GetDNACount(ctx)								// Obtiene la cantidad total de ADNs
	st["total"] = fmt.Sprintf("%d", total)

	if total != 0 {											// Obtiene los ratios de humanos y mutantes
		st["mutants_ratio"] = fmt.Sprintf("%.2f", float32(mutants)/float32(total))
		st["humans_ratio"] = fmt.Sprintf("%.2f", float32(humans)/float32(total))
	} else {
		st["mutants_ratio"] = "0"
		st["humans_ratio"] = "0"
	}
	return st
}


// Guarda el ADN en la base e incrementa el contador correspondiente
func SaveDNA(ctx *gin.Context, adn dna.Dna, mutant bool) {
	var counter CounterEntity
	var dnaEnt DNAEntity

	newctx := appengine.NewContext(ctx.Request)
	dnaEnt.Matrix = strings.Join(adn.Matrix, ";")							// Arma la entidad ADN para guardar en la base
	if mutant {
		dnaEnt.DnaType = mutantDNAType
	} else {
		dnaEnt.DnaType = humanDNAType
	}

	key := datastore.NewIncompleteKey(newctx, dnaKind, nil)					// Genera la key para guardar el ADN
	_, err := datastore.Put(newctx, key, &dnaEnt)							// Guarda el ADN en la base
	if err == nil {
		log.Infof(newctx, "Se guardo el ADN: %+v", dnaEnt)
	} else {
		log.Errorf(newctx, "Error al guardar el ADN: %v", err.Error())
		return
	}

	key = datastore.NewKey(newctx, counterKind, dnaEnt.DnaType, 0, nil)		// Genera la key para consultar el contador
	err = datastore.Get(newctx, key, &counter)								// Busca el contador en la base

	if err == nil {															// Si encontro el contador, le suma 1
		log.Infof(newctx, "El valor del Contador '%s' es: %d", dnaEnt.DnaType, counter.Count)
		counter.Count++
	} else {
		if strings.HasPrefix(err.Error(), `datastore: no such entity`) {	// Si no encontro el contador, es el primer ADN de ese tipo
			counter.Count = 1												// Inicia el contador en 1
			log.Infof(newctx, "Se inicio el Contador '%s'", dnaEnt.DnaType)
		}
	}
	_, err = datastore.Put(newctx, key, &counter)							// Guarda el nuevo valor de contador
	if err == nil {
		log.Infof(newctx, "El nuevo valor del Contador '%s' es: %d", dnaEnt.DnaType, counter.Count)
	} else {
		log.Errorf(newctx, "Error al guardar el Contador '%s': %v", dnaEnt.DnaType, err.Error())
	}
}


// Devuelve la cantidad de ADNs del tipo dnaType que hay en la base
func GetDNATypeCount(ctx *gin.Context, dnaType string) int {
	var counter CounterEntity
	newctx := appengine.NewContext(ctx.Request)
	key := datastore.NewKey(newctx, counterKind, dnaType, 0, nil)			// Genera la key para consultar el contador
	err := datastore.Get(newctx, key, &counter)								// Busca el contador en la base
	if err != nil {															// Si no encontro el contador, no se guardaron ADNs de ese tipo todavia
		if strings.HasPrefix(err.Error(), `datastore: no such entity`) {
			return 0
		}
	}
	return counter.Count
}


// Devuelve la cantidad total de ADNs que hay en la base
func GetDNACount(ctx *gin.Context) int {
	newctx := appengine.NewContext(ctx.Request)
	total := 0
	query := datastore.NewQuery(counterKind)
	for t := query.Run(newctx); ; {
		var counter CounterEntity
		_, err := t.Next(&counter)
		if err == datastore.Done {
			break
		}

		if err != nil {														// Si no encontro contadores, no se guardaron ADNs todavia
			if strings.HasPrefix(err.Error(), `datastore: no such entity`) {
				return 0
			}
		}
		total += counter.Count
	}
	return total
}


// Elimina todos los ADNs y los contadores de la base. Se hace en bloques ordenados de 500 entidades
func ClearDB(ctx *gin.Context) {

	var lastKey *datastore.Key
	var i int
	newctx := appengine.NewContext(ctx.Request)
	kinds := [2]string {counterKind, dnaKind}

	for k := 0; k < len(kinds); k++ {													// Primero borra los ADN y despues los contadores
		q := datastore.NewQuery(kinds[k]).Limit(500).KeysOnly().Order("__key__")		// Consulta las primeras 500 entidades
		for i = 0;; {
			keys, err := q.GetAll(newctx, nil)
			log.Infof(newctx, "Se obtuvieron %d entidades %s", len(keys), kinds[k])
			if err != nil {
				log.Errorf(newctx, "Error al recuperar las entidades %s: %v", kinds[k], err.Error())
				ctx.Status(http.StatusInternalServerError)								// Si hay errores devuelve 500-Internal server error
				return
			}
			if len(keys) == 0 {
				break
			} else {
				lastKey = keys[len(keys)-1]
				i = i + len(keys)
				if err := datastore.DeleteMulti(newctx, keys); err != nil {				// Borra las 500 entidades recuperadas
					log.Errorf(newctx, "Error al borrar las entidades %s: %v", kinds[k], err.Error())
					ctx.Status(http.StatusInternalServerError)							// Si hay errores devuelve 500-Internal server error
					return
				}
				log.Infof(newctx, "Se borraron %d entidades", len(keys))
			}
			q = datastore.NewQuery(kinds[k]).Limit(500).KeysOnly().Order("__key__").Filter("__key__ >", lastKey)	// Consulta las siguientes 500 entidades
		}
	}
	ctx.Status(http.StatusOK)															// Si pudo borrar todo devuelve 200-Ok
}
