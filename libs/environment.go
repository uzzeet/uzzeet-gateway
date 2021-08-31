package libs

const (
	EnvProduction  = "production"
	EnvStaging     = "staging"
	EnvDevelopment = "development"
	EnvLocal       = "local"
)

const (
	ClusterKubernetes = "kubernetes"
	ClusterLocal      = "local"
)

const (
	NamespaceDefault = "default"
)

const (
	AppEnv          = "APP_ENV"
	AppKey          = "APP_KEY"
	AppName         = "APP_NAME"
	AppVersion      = "APP_VERSION"
	AppHost         = "APP_HOST"
	AppPort         = "APP_PORT"
	AppEndpoint     = "APP_ENDPOINT"
	AppOpenEndpoint = "APP_OPEN_ENDPOINT"
	AppBasepoint    = "APP_BASEPOINT"
	AppRegistryAddr = "APP_REGISTRY_ADDR"
	AppRegistryPwd  = "APP_REGISTRY_PWD"
	AppTimezone     = "APP_TIMEZONE"
	AppNamespace    = "APP_NAMESPACE"
	AppCluster      = "APP_CLUSTER"

	DBEngine       = "DB_ENGINE"
	DBHost         = "DB_HOST"
	DBPort         = "DB_PORT"
	DBUser         = "DB_USER"
	DBPwd          = "DB_PWD"
	DBName         = "DB_NAME"
	DBSSLMode      = "DB_SSL_MODE"
	DBConnStr      = "DB_CONN_STR"
	DBConnLifetime = "DB_CONN_LIFETIME"
	DBConnMaxIdle  = "DB_CONN_MAX_IDLE"
	DBConnMaxOpen  = "DB_CONN_MAX_OPEN"

	CKEngine       = "CK_ENGINE"
	CKHost         = "CK_HOST"
	CKPort         = "CK_PORT"
	CKUser         = "CK_USER"
	CKPwd          = "CK_PWD"
	CKName         = "CK_NAME"
	CKSSLMode      = "CK_SSL_MODE"
	CKConnStr      = "CK_CONN_STR"
	CKConnLifetime = "CK_CONN_LIFETIME"
	CKConnMaxIdle  = "CK_CONN_MAX_IDLE"
	CKConnMaxOpen  = "CK_CONN_MAX_OPEN"

	BrokerAddr = "BROKER_ADDR"
	BrokerPwd  = "BROKER_PWD"

	RethinkDBHost = "RETHINKDB_HOST"
	RethinkDBName = "RETHINKDB_NAME"
)
