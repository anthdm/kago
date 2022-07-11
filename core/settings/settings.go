package settings

var GlobalConfig = &config{}

type config struct {
	Host           string
	Port           string
	Profiler       bool
	Docs           bool
	Monitoring     bool
	EmbedStatic    bool
	EmbedTemplates bool
	DbType         string
	DbDSN          string
	DbName         string
	SmtpEmail      string
	SmtpPass       string
	SmtpHost       string
	SmtpPort       string
	Secret         string
}