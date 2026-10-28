package stdlib

import (
	"math"
	"strings"

	"github.com/Jelsin29/cr7Script/internal/value"
)

func (l *Library) installModules() {
	champions := &namespace{exports: make(map[string]value.Value)}
	for _, name := range []string{"soy_el_mejor", "minimo", "goles_totales", "promedio"} {
		champions.exports[name] = function(name, 1, statistics(name))
	}
	champions.exports["pichichi"] = function("pichichi", 1, func(a []value.Value) (value.Value, error) {
		list, ok := a[0].List()
		if !ok {
			return wrong()
		}
		if list.Len() == 0 {
			return fail("ListaVacia", "pichichi requires at least one player")
		}
		var winner value.Value
		best := math.Inf(-1)
		for _, player := range list.Snapshot() {
			d, ok := player.Dictionary()
			if !ok {
				return wrong()
			}
			goals, found := d.Get("goles")
			if !found {
				return fail("ClaveAusente", "player is missing goles")
			}
			score, ok := goals.Number()
			if !ok {
				return wrong()
			}
			if score > best {
				best = score
				winner = player
			}
		}
		return winner, nil
	})
	l.modules["mr_champions"] = champions
	relato := &namespace{exports: make(map[string]value.Value)}
	relato.exports["minusculas"] = function("minusculas", 1, textUnary(strings.ToLower))
	relato.exports["separa"] = function("separa", 2, func(a []value.Value) (value.Value, error) {
		text, ok := a[0].Text()
		sep, ok2 := a[1].Text()
		if !ok || !ok2 || sep == "" {
			return wrong()
		}
		parts := strings.Split(text, sep)
		items := make([]value.Value, len(parts))
		for i, s := range parts {
			items[i] = value.NewText(s)
		}
		return value.NewList(items...), nil
	})
	relato.exports["une"] = function("une", 2, func(a []value.Value) (value.Value, error) {
		list, ok := a[0].List()
		sep, ok2 := a[1].Text()
		if !ok || !ok2 || sep == "" {
			return wrong()
		}
		parts := make([]string, list.Len())
		for i, v := range list.Snapshot() {
			s, ok := v.Text()
			if !ok {
				return wrong()
			}
			parts[i] = s
		}
		return value.NewText(strings.Join(parts, sep)), nil
	})
	relato.exports["reemplaza"] = function("reemplaza", 3, func(a []value.Value) (value.Value, error) {
		text, ok := a[0].Text()
		old, ok2 := a[1].Text()
		replacement, ok3 := a[2].Text()
		if !ok || !ok2 || !ok3 || old == "" {
			return wrong()
		}
		return value.NewText(strings.ReplaceAll(text, old, replacement)), nil
	})
	l.modules["relato"] = relato
	marcador := &namespace{exports: make(map[string]value.Value)}
	for name, operation := range map[string]func(float64) float64{"absoluto": math.Abs, "redondea": math.Round} {
		marcador.exports[name] = function(name, 1, func(a []value.Value) (value.Value, error) {
			n, ok := a[0].Number()
			if !ok {
				return wrong()
			}
			return number(operation(n))
		})
	}
	marcador.exports["potencia"] = function("potencia", 2, func(a []value.Value) (value.Value, error) {
		base, ok := a[0].Number()
		exponent, ok2 := a[1].Number()
		if !ok || !ok2 {
			return wrong()
		}
		return number(math.Pow(base, exponent))
	})
	l.modules["marcador"] = marcador
	l.modules["rueda_de_prensa"] = &namespace{exports: map[string]value.Value{
		"RICO_GUAPO_Y_GRAN_JUGADOR": value.NewText("soy guapo, rico y un gran jugador"),
		"ESTOY_TRISTE":              value.NewText("No lo hago cuando estoy triste"),
		"TODOS_A_MI_NIVEL":          value.NewText("Si todos estuviesen en mi nivel seríamos los primeros"),
		// Q4 is a historical Juventus-era remark, not a Madrid-era statistic.
		"CINCO_CHAMPIONS": value.NewText("Cinco Champions tengo yo y el Atleti, cero"),
		"THE_BEST_SOY_YO": value.NewText("el premio The Best fue para el 'the best', que soy yo"),
		"SOY_EL_MEJOR":    value.NewText("Soy el mejor jugador de la historia"),
		// Q7 comes from the 2009 public presentation, not an interview.
		"HALA_MADRID": value.NewText("Uno, dos, tres: ¡Hala Madrid!"),
	}}
}

func statistics(operation string) func([]value.Value) (value.Value, error) {
	return func(a []value.Value) (value.Value, error) {
		list, ok := a[0].List()
		if !ok {
			return wrong()
		}
		if list.Len() == 0 && operation != "goles_totales" {
			return fail("ListaVacia", "operation requires at least one number")
		}
		result := 0.0
		for i, v := range list.Snapshot() {
			n, ok := v.Number()
			if !ok {
				return wrong()
			}
			switch operation {
			case "soy_el_mejor":
				if i == 0 || n > result {
					result = n
				}
			case "minimo":
				if i == 0 || n < result {
					result = n
				}
			default:
				result += n
				if math.IsInf(result, 0) || math.IsNaN(result) {
					return fail("NumeroNoFinito", "numeric result must be finite")
				}
			}
		}
		if operation == "promedio" {
			result /= float64(list.Len())
		}
		return number(result)
	}
}
