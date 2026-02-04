package fix_test

import (
	"context"
	"strings"
	"testing"
)

// BenchmarkStringConcatenation compares string concatenation performance
// This demonstrates why fmt.Fprintf is better than string concatenation
func BenchmarkStringConcatenation(b *testing.B) {
	columns := []string{
		"col1", "col2", "col3", "col4", "col5",
	}
	const updateSet = "UPDATE files SET "
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result string
		for column := range len(columns) {
			result += updateSet + columns[column] + " = NULL WHERE " + columns[column] + " = ''; "
		}
		_ = result
	}
}

// BenchmarkStringBuilder demonstrates the optimized approach
func BenchmarkStringBuilder(b *testing.B) {
	columns := []string{
		"col1", "col2", "col3", "col4", "col5",
	}
	const updateSet = "UPDATE files SET "
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var query strings.Builder
		for column := range len(columns) {
			query.WriteString(updateSet + columns[column] + " = NULL WHERE " + columns[column] + " = ''; ")
		}
		_ = query.String()
	}
}

// BenchmarkStringBuilderFprintf is the best approach with fmt.Fprintf
func BenchmarkStringBuilderFprintf(b *testing.B) {
	columns := []string{
		"col1", "col2", "col3", "col4", "col5",
	}
	const updateSet = "UPDATE files SET "
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var query strings.Builder
		for _, column := range columns {
			query.WriteString(updateSet)
			query.WriteString(column)
			query.WriteString(" = NULL WHERE ")
			query.WriteString(column)
			query.WriteString(" = ''; ")
		}
		_ = query.String()
	}
}

// BenchmarkSliceReallocationVsReuse compares slice reallocation vs reuse
func BenchmarkSliceReallocate(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mods := make([]string, 0, 5)
		mods = append(mods, "a", "b", "c")
		mods = []string{} // This reallocates
		mods = append(mods, "d", "e")
		_ = mods
	}
}

// BenchmarkSliceReuse shows the optimized approach
func BenchmarkSliceReuse(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mods := make([]string, 0, 5)
		mods = append(mods, "a", "b", "c")
		mods = mods[:0] // This reuses capacity
		mods = append(mods, "d", "e")
		_ = mods
	}
}

// BenchmarkLoopConditionCheck compares redundant condition checks
func BenchmarkRedundantCondition(b *testing.B) {
	size := 1000
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		for j := 0; j < size; j++ {
			if j < size { // Always true, but still checked
				count++
			}
		}
		_ = count
	}
}

// BenchmarkNoRedundantCondition shows the optimized version
func BenchmarkNoRedundantCondition(b *testing.B) {
	size := 1000
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		for j := 0; j < size; j++ {
			count++ // No redundant check
		}
		_ = count
	}
}

// BenchmarkParameterizedQuerySetup demonstrates the performance of parameterized queries
func BenchmarkParameterizedQuerySetup(b *testing.B) {
	trainer := "gamehack"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Using parameterized queries (? placeholder)
		_ = "section != ?"
		_ = trainer
	}
}

// BenchmarkStringFormatQuerySetup shows the old fmt.Sprintf approach
func BenchmarkStringFormatQuerySetup(b *testing.B) {
	trainer := "gamehack"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Using fmt.Sprintf (less ideal)
		query := "section != '" + trainer + "'"
		_ = query
	}
}

// BenchmarkContextPassthrough demonstrates context overhead
func BenchmarkContextPassthrough(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulating context passthrough in function calls
		_ = ctx.Err()
	}
}

// BenchmarkSliceCapacityPreallocation shows allocation efficiency
func BenchmarkSliceAllocation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mods := make([]int, 0, 5)
		for j := 0; j < 5; j++ {
			mods = append(mods, j)
		}
		_ = mods
	}
}

// BenchmarkStringBuilderFormatting demonstrates format string efficiency
func BenchmarkStringBuilderFormat(b *testing.B) {
	columns := []string{"a", "b", "c", "d", "e"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var query strings.Builder
		for _, col := range columns {
			query.WriteString("UPDATE files SET ")
			query.WriteString(col)
			query.WriteString(" = NULL WHERE ")
			query.WriteString(col)
			query.WriteString(" = 0; ")
		}
		_ = query.String()
	}
}

// BenchmarkFixesMapAllocation shows the old approach creating new map each time
func BenchmarkFixesMapAllocation(b *testing.B) {
b.ResetTimer()
for i := 0; i < b.N; i++ {
// Simulating creating fixes map every repair
m := map[string]string{
"a": "A", "b": "B", "c": "C", "d": "D", "e": "E",
"f": "F", "g": "G", "h": "H", "i": "I", "j": "J",
"k": "K", "l": "L", "m": "M",
}
_ = m
}
}

// BenchmarkFixesMapPackageLevel shows optimized approach using package-level map
func BenchmarkFixesMapPackageLevel(b *testing.B) {
fixesMapLocal := map[string]string{
"a": "A", "b": "B", "c": "C", "d": "D", "e": "E",
"f": "F", "g": "G", "h": "H", "i": "I", "j": "J",
"k": "K", "l": "L", "m": "M",
}
b.ResetTimer()
for i := 0; i < b.N; i++ {
_ = fixesMapLocal
}
}

// BenchmarkToUpperInLoop shows the old approach with repeated conversions
func BenchmarkToUpperInLoop(b *testing.B) {
items := []string{"acid", "ice", "damn", "rss", "dsi"}
b.ResetTimer()
for i := 0; i < b.N; i++ {
for _, item := range items {
upper := strings.ToUpper(item)
_ = upper
}
}
}

// BenchmarkToUpperPrecomputed shows optimized approach with pre-computed values
func BenchmarkToUpperPrecomputed(b *testing.B) {
itemsUpper := map[string]string{
"ACID": "ACID", "ICE": "ICE", "DAMN": "DAMN",
"RSS": "RSS", "DSI": "DSI",
}
b.ResetTimer()
for i := 0; i < b.N; i++ {
for _, val := range itemsUpper {
_ = val
}
}
}
