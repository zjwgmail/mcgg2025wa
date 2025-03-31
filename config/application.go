package config

import (
	"go-fission-activity/activity/runtime"
	"time"
)

type Application struct {
	Host                string
	IsDebug             bool
	IsActivityServer    bool
	Port                int64
	Name                string
	Redis               RedisConfig
	Datasource          DatasourceConfig
	Timer               map[string]TimerConfig
	Rsa                 Rsa
	Nx                  Nx
	Wa                  Wa
	Feishu              Feishu
	MethodInsertMsgInfo map[string]MethodInsertMsgInfo
	Activity            Activity
	S3Config            S3Config
	EmailConfig         EmailConfig
}

type Activity struct {
	Id                     int
	Name                   string
	Scheme                 string
	NeedSubscribe          bool
	LanguageList           []string
	LanguageNameMap        map[string]string
	ChannelList            []string
	ChannelNameMap         map[string]string
	HelpTextList           []HelpText
	UnRedPacketMinute      int
	SendRedPacketMinute    int
	TwoStartGroupMinute    int
	Stage1Award            StageInfo
	Stage2Award            StageInfo
	Stage3Award            StageInfo
	WaIdPrefixList         []string
	InsertOtherRsvMsgTable string
	FreeCdkSendDelayHour   int
	WaRedirectListPrefix   string
}

type HelpText struct {
	Id       string                       `json:"id"`
	BodyText map[string]map[string]string `json:"bodyText"`
	Weight   int                          `json:"weight"`
}

type StageInfo struct {
	HelpNum   int
	AwardName map[string]string
	AwardLink map[string]string
}

type S3Config struct {
	Bucket          string
	Region          string
	DonAmin         string
	AccessKeyID     string
	SecretAccessKey string
	PreSignUrl      string
}

type MethodInsertMsgInfo struct {
	UserAttendPrefixList       []string
	UserAttendOfHelpPrefixList []string
	RenewFreePrefixList        []string
}
type Feishu struct {
	WebHook string
}
type Nx struct {
	Ak            string
	Sk            string
	AppKey        string
	BusinessPhone string
	CallBackUrl   string
	IsVerifySign  bool
}
type Wa struct {
	McggShortProject     string
	McggShortLinkGenUrl  string
	McggShortLinkPrefix  string
	McggShortLinkSignKey string
	MlbbShortProject     string
	MlbbShortLinkGenUrl  string
	MlbbShortLinkPrefix  string
	MlbbShortLinkSignKey string
}
type TimerConfig struct {
	TimerCorn string
}
type Rsa struct {
	PrivateKey string
	PublicKey  string
}
type DatasourceConfig struct {
	XmlPrefix      string
	DriverName     string
	DataSourceLink string
	MaxIdleCount   int           // zero means defaultMaxIdleConns; negative means 0
	MaxOpen        int           // <= 0 means unlimited
	MaxLifetime    time.Duration // maximum amount of time a connection may be reused
	MaxIdleTime    time.Duration // maximum amount of time a connection may be idle before being closed
	LogEnable      bool          //是否打印日志 true打印，false不打印
}

type RedisConfig struct {
	// Maximum number of idle connections in the pool.
	MaxIdle int

	// 最大活跃连接数
	MaxActive int

	// 空闲时间
	IdleTimeout time.Duration

	//redis地址
	Address string

	Username string

	Password string

	Database int
}

type PprofConfig struct {
	Enable             bool
	Port               int64
	ListenEnableSecond int64
}

type TracingConfig struct {
	Enable    bool
	AgentHost string
	Sampler   SamplerConfig
}

type SamplerConfig struct {
	Type  string
	Param float64
}

type SentryConfig struct {
	Enable      bool
	Dsn         string
	SamplerRate float64
	IsDebug     bool
}

type EmailConfig struct {
	ServerHost    string
	ServerPort    int
	FromAddress   string
	ApiUser       string
	ApiKey        string
	ToAddressList []string
}

var ApplicationConfig = new(Application) //&Application{Minio: &MinioConfig{}, Grpc: &GrpcConfig{}, Redis: &RedisConfig{}, Aes: &AesConfig{}}

var Runtime runtime.Runtime = runtime.NewApplication()
