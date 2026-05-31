package completion

import (
	"fmt"
	"math"

	"github.com/brogergvhs/value-dsl/internal/values"
)

func formatValueDocumentation(definition values.Definition) string {
	return fmt.Sprintf(
		"**%s**\n\nCategory: `%s`\n\nAngle: `%.2f rad` (`%.1f deg`)\n\nRadius: `%.2f`",
		definition.Name,
		definition.Category,
		definition.Angle,
		definition.Angle*180/math.Pi,
		definition.Radius,
	)
}
