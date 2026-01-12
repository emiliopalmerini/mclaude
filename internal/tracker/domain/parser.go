package domain

// TranscriptParser defines the interface for parsing session transcripts
type TranscriptParser interface {
	Parse(transcriptPath string) (Statistics, error)
}
