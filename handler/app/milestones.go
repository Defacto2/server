package app

// Milestone is an accomplishment for a year and optional month.
type Milestone struct {
	Year      int    // Year of the milestone.
	Month     int    // Month of the milestone.
	Day       int    // Day of the milestone.
	Prefix    string // Prefix replacement for the month, such as 'Early', 'Mid' or 'Late'.
	Title     string // Title of the milestone should be the accomplishment.
	Lead      string // Lead paragraph, is optional and should usually be the product.
	Content   string // Content is the main body of the milestone and can be HTML.
	Link      string // Link is the URL to an article about the milestone or the product.
	LinkTitle string // LinkTitle is the title of the Link.
}

// Milestones is a collection of Milestone.
type Milestones []Milestone

// Len is the number of Milestones.
func (m Milestones) Len() int {
	return len(m)
}

func ByDecade1970s() Milestones {
	m := []Milestone{
		{
			Year: 1971, Month: 10, Title: "Secrets of the Little Blue Box",
			Lead: "Esquire October 1971", LinkTitle: "the complete article",
			Link: "https://www.slate.com/articles/technology/the_spectator/2011/10/the_article_that_inspired_steve_jobs_secrets_of_the_little_blue_.html",
			Content: "Ron Rosenbaum writes the first mainstream article on phone freaks, primarily kids who'd hack and experiment with the global telephone network.<br>" +
				"The piece coins them as phone-<strong>phreaks</strong> and introduces the reader to the kids' use of <strong>pseudonyms</strong> or codenames within their regional <strong>groups</strong> of friends." +
				"It gives an early example of <strong>social engineering</strong>, defines the community of phreakers as the phone-phreak <strong>underground</strong>, and mentions the newer trend of <strong>computer phreaking</strong>, which we call computer hacking today.",
		},
		{
			Year: 1971, Month: 11, Day: 15, Title: "The first microcomputer",
			Lead: "Intel 4004", LinkTitle: "The Story of the Intel 4004",
			Link:    "https://www.intel.com/content/www/us/en/history/museum-story-of-intel-4004.html",
			Content: "Intel advertises the first-to-market general-purpose programmable processor or microprocessor, the 4-bit Intel 4004.",
		},
		{
			Year: 1972, Month: 4, Title: "The first 8-bit microprocessor",
			Lead: "Intel 8008", LinkTitle: "The Story of the Intel 8008",
			Link:    "https://www.intel.com/content/www/us/en/history/museum-story-of-intel-8008.html",
			Content: "Intel releases the world's first 8-bit microprocessor, the Intel 8008.",
		},
		{
			Year: 1972, Prefix: "Early", Title: "Blue boxes",
			Link: "https://explodingthephone.com/", LinkTitle: "about the hackers of the telephone network",
			Content: "Inspired by The Secrets of the Little Blue Box article, Steve Wozniak and a teenage Steve Jobs team up to build and sell 40-100, Wozniak-designed blue boxes to the students of Berkeley University." +
				"The devices allowed users to hack and manipulate the electromechanical machines that operated the national telephone network.",
		},
		{
			Year: 1974, Month: 4, Title: "The first CPU for microcomputers",
			Lead: "Intel 8080", LinkTitle: "about The Intel 8008 and 8080",
			Link: "https://www.intel.com/content/www/us/en/history/museum-story-of-intel-8008.html",
			Content: "Intel releases the 8-bit 8080 CPU, its second but more successful 8-bit programmable microprocessor." +
				"This CPU became the processing heart of the earliest popular microcomputers, the Altair 8800, the Sol-20 and the IMSAI.",
		},
	}
	return m
}
