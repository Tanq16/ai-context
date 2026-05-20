package utils

// GlobalDebugFlag is set by cobra root command when --debug is passed
var GlobalDebugFlag bool

// GlobalForAIFlag is set by cobra root command when --for-ai is passed
// When true, output uses plain text with prefixes and input reads from stdin pipe
var GlobalForAIFlag bool