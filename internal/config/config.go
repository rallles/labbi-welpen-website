package config

import "os"

type Config struct {
	ServerAddress string
	Neo4jUri      string
	Neo4jUser     string
	Neo4jPassword string
	AdminUser     string
	AdminPassword string
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPassword  string
	ContactMailTo string
	UploadDir     string
	StaticDir     string
	TemplateDir   string
}

func LoadConfig() Config {
	cfg := Config{
		ServerAddress: os.Getenv("SERVER_ADDRESS"),
		Neo4jUri:      os.Getenv("NEO4J_URI"),
		Neo4jUser:     os.Getenv("NEO4J_USER"),
		Neo4jPassword: os.Getenv("NEO4J_PASSWORD"),
		AdminUser:     os.Getenv("ADMIN_USER"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		SMTPHost:      os.Getenv("SMTP_HOST"),
		SMTPPort:      os.Getenv("SMTP_PORT"),
		SMTPUser:      os.Getenv("SMTP_USER"),
		SMTPPassword:  os.Getenv("SMTP_PASSWORD"),
		ContactMailTo: os.Getenv("CONTACT_MAIL_TO"),
		UploadDir:     os.Getenv("UPLOAD_DIR"),
		StaticDir:     os.Getenv("STATIC_DIR"),
		TemplateDir:   os.Getenv("TEMPLATE_DIR"),
	}
	if cfg.UploadDir == "" {
		cfg.UploadDir = "data/uploads"
	}
	if cfg.StaticDir == "" {
		cfg.StaticDir = "static"
	}
	if cfg.TemplateDir == "" {
		cfg.TemplateDir = "internal/templates"
	}
	return cfg
}
