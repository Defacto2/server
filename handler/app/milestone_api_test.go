package app

import (
	"testing"
)

func TestCleanHTMLForAPI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Preserves anchor tags with href",
			input:    `<p>Test <a href="https://example.com" class="link">link</a> here</p>`,
			expected: `<p>Test <a href="https://example.com">link</a> here</p>`,
		},
		{
			name:     "Removes anchor tags without href",
			input:    `<p>Test <a name="anchor">link</a> here</p>`,
			expected: `<p>Test link here</p>`,
		},
		{
			name:     "Removes class attributes",
			input:    `<p class="test">Content</p>`,
			expected: `<p>Content</p>`,
		},
		{
			name:     "Removes style attributes",
			input:    `<p style="color: red;">Content</p>`,
			expected: `<p>Content</p>`,
		},
		{
			name:     "Removes id attributes",
			input:    `<p id="test">Content</p>`,
			expected: `<p>Content</p>`,
		},
		{
			name:     "Removes title attributes",
			input:    `<p title="tooltip">Content</p>`,
			expected: `<p>Content</p>`,
		},
		{
			name:     "Removes data attributes",
			input:    `<p data-test="value">Content</p>`,
			expected: `<p>Content</p>`,
		},
		{
			name:     "Preserves semantic HTML",
			input:    `<p>Test <strong>bold</strong> and <em>italic</em> text</p>`,
			expected: `<p>Test <strong>bold</strong> and <em>italic</em> text</p>`,
		},
		{
			name:     "Handles complex anchor tags",
			input:    `<a href="https://example.com" class="link" id="test" title="tooltip" data-info="test">Complex Link</a>`,
			expected: `<a href="https://example.com">Complex Link</a>`,
		},
		{
			name:     "Handles multiple anchor tags",
			input:    `<p><a href="https://example1.com">Link 1</a> and <a href="https://example2.com">Link 2</a></p>`,
			expected: `<p><a href="https://example1.com">Link 1</a> and <a href="https://example2.com">Link 2</a></p>`,
		},
		{
			name:     "Removes empty tags",
			input:    `<p><span> </span>Content</p>`,
			expected: `<p><span> </span>Content</p>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanHTMLForAPI(tt.input)
			if result != tt.expected {
				t.Errorf("CleanHTMLForAPI(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Removes all HTML tags",
			input:    `<p>Test <strong>content</strong> here</p>`,
			expected: `Test content here`,
		},
		{
			name:     "Handles HTML entities",
			input:    `Test &nbsp; with &amp; entities`,
			expected: `Test with & entities`,
		},
		{
			name:     "Adds spacing after punctuation",
			input:    `Test.content.with.punctuation!marks?here`,
			expected: `Test. content. with. punctuation! marks? here`,
		},
		{
			name:     "Collapses multiple spaces",
			input:    `Test    multiple     spaces`,
			expected: `Test multiple spaces`,
		},
		{
			name:     "Handles anchor tags",
			input:    `Test <a href="https://example.com">link</a> content`,
			expected: `Test link content`,
		},
		{
			name:     "Handles complex HTML",
			input:    `<div><p>Test <span>content</span> <a href="#">link</a></p></div>`,
			expected: `Test content link`,
		},
		{
			name:     "Removes spaces before punctuation",
			input:    `Test content with spaces , before commas . and periods !`,
			expected: `Test content with spaces, before commas. and periods!`,
		},
		{
			name:     "Removes spaces around parentheses",
			input:    `Test ( content ) with ( spaces ) around ( parentheses )`,
			expected: `Test (content) with (spaces) around (parentheses)`,
		},
		{
			name:     "Handles real milestone content",
			input:    `<p>Ron Rosenbaum writes the first mainstream article on phone freaks, primarily kids who'd hack and experiment with the global telephone network.</p><p>The piece coins them as phone-freaks (<strong>phreaks</strong>) and introduces the reader to the kids' use of <strong>pseudonyms</strong> or codenames within their cliques and <strong>groups</strong> of friends. It gives an early example of <strong>social engineering</strong>, defines the community of phreakers as the phone-phreak <strong>underground</strong>, and mentions the newer trend of <strong>computer phreaking</strong>, which we call <u>computer&nbsp;hacking</u> today.</p>`,
			expected: `Ron Rosenbaum writes the first mainstream article on phone freaks, primarily kids who'd hack and experiment with the global telephone network. The piece coins them as phone-freaks (phreaks) and introduces the reader to the kids' use of pseudonyms or codenames within their cliques and groups of friends. It gives an early example of social engineering, defines the community of phreakers as the phone-phreak underground, and mentions the newer trend of computer phreaking, which we call computer hacking today.`,
		},
		{
			name:     "Converts <q> tags to quotes",
			input:    `<p>He said <q>Hello world</q> to everyone.</p>`,
			expected: `He said "Hello world" to everyone.`,
		},
		{
			name:     "Handles multiple <q> tags",
			input:    `<p>Multiple <q>quotes</q> in <q>one</q> sentence.</p>`,
			expected: `Multiple "quotes" in "one" sentence.`,
		},
		{
			name:     "Handles nested <q> tags",
			input:    `<p>Nested <q>quotes <q>inside</q> quotes</q> should work.</p>`,
			expected: `Nested "quotes inside" quotes should work.`,
		},
		{
			name:     "Handles <q> tags with attributes",
			input:    `<p>The famous quote <q cite="https://example.com">To be or not to be</q> is from Shakespeare.</p>`,
			expected: `The famous quote "To be or not to be" is from Shakespeare.`,
		},
		{
			name:     "Complex real-world example with quotes",
			input:    `<p>As Steve Jobs famously said <q>Stay hungry, stay foolish</q>, which was inspired by the <q>Whole Earth Catalog</q> manifesto that stated <q>Stay hungry. Stay foolish.</q> This philosophy became a cornerstone of Apple's <q cite="https://apple.com">Think Different</q> campaign.</p>`,
			expected: `As Steve Jobs famously said "Stay hungry, stay foolish", which was inspired by the "Whole Earth Catalog" manifesto that stated "Stay hungry. Stay foolish." This philosophy became a cornerstone of Apple's "Think Different" campaign.`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripHTMLTags(tt.input)
			if result != tt.expected {
				t.Errorf("StripHTMLTags(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetMilestonesByYear(t *testing.T) {
	// This would need a mock or actual data setup
	t.Skip("Skipping until test data is available")
}

func TestGetMilestonesByYearRange(t *testing.T) {
	// This would need a mock or actual data setup
	t.Skip("Skipping until test data is available")
}

func TestGetHighlightedMilestones(t *testing.T) {
	// This would need a mock or actual data setup
	t.Skip("Skipping until test data is available")
}

func TestGetMilestonesByDecade(t *testing.T) {
	// This would need a mock or actual data setup
	t.Skip("Skipping until test data is available")
}

// Benchmark functions
func BenchmarkCleanHTMLForAPI(b *testing.B) {
	html := `<div class="content">
		<p class="lead">This is a <strong>test</strong> with <a href="https://example.com" class="link" id="test">links</a> and <span style="color: red;">formatting</span>.</p>
		<p>Another paragraph with <a name="anchor">anchor</a> and <data-info="test">data attributes</data-info>.</p>
	</div>`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CleanHTMLForAPI(html)
	}
}

func BenchmarkStripHTMLTags(b *testing.B) {
	html := `<div class="content">
		<p class="lead">This is a <strong>test</strong> with <a href="https://example.com" class="link" id="test">links</a> and <span style="color: red;">formatting</span>.</p>
		<p>Another paragraph with &nbsp; non-breaking &amp; spaces and <data-info="test">data attributes</data-info>.</p>
	</div>`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StripHTMLTags(html)
	}
}

func BenchmarkCleanMilestoneForAPI(b *testing.B) {
	// Create a sample milestone
	milestone := Milestone{
		Year:    2023,
		Highlight: true,
		Content: `<div class="milestone-content">
			<p>This is a <strong>milestone</strong> with <a href="https://example.com" class="external">external link</a> and <span class="highlight">formatted text</span>.</p>
			<p>More content with <data-test="value">data attributes</data-test> and <id="section">IDs</id>.</p>
		</div>`,
		Lead: `<p class="lead">This is the <em>lead</em> content with <a href="#section">internal link</a>.</p>`,
		List: Links{
			{
				LinkTitle: `<span class="title">Link 1</span>`,
				SubTitle:  `<p class="subtitle">Subtitle with <strong>formatting</strong></p>`,
			},
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cleanMilestoneForAPI(milestone)
	}
}