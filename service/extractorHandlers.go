package service

import (
	"strings"

	"github.com/gocolly/colly"
)

func EffectDescription(interactionDiv *colly.HTMLElement) string {
	var effectDescription string
	interactionDiv.ForEachWithBreak("p", func(i int, effect *colly.HTMLElement) bool {
		// to skip first P (which contains name of the interactions)
		if i == 1 {
			effectDescription = effect.Text
			return false
		}
		return true
	})

	return effectDescription
}

func InteractionName(interactionDiv *colly.HTMLElement) string {
	var name string
	interactionDiv.ForEachWithBreak(".interactions-reference-header h3", func(i int, h *colly.HTMLElement) bool {
		name = h.Text
		return false
	})
	name = strings.Join(strings.Fields(name), " ")
	return name
}
