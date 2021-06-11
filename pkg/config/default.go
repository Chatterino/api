package config

var defaultConf = APIConfig{
	BaseURL:          "",
	BindAddress:      ":1234",
	MaxContentLength: 5 * 1024 * 1024,
	EnableLilliput:   true,

	OembedProvidersPath: "./providers.json",
}
