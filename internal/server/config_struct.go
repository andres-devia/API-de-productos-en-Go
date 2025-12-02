package server

// Config contiene los parámetros de configuración del servidor.
// Esto ayuda a mantener la configuración separada y mejora la capacidad de prueba.
type Config struct {
	Port   string
	DBPath string
}
