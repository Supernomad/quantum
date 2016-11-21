package common

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vishvananda/netlink"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"os"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	version   = "0.9.0"
	envPrefix = "QUANTUM_"
)

var (
	googleV4   = net.ParseIP("8.8.8.8")
	googleV6   = net.ParseIP("2001:4860:4860::8888")
	loopbackV4 = net.ParseIP("127.0.0.1")
	loopbackV6 = net.ParseIP("::1")
	allV4      = net.ParseIP("0.0.0.0")
	allV6      = net.ParseIP("::")
)

// Config handles marshalling user supplied configuration data
type Config struct {
	ConfFile        string        `skip:"false"  type:"string"    short:"c"    long:"conf-file"         default:""                      description:"The configuration file to use to configure quantum."`
	InterfaceName   string        `skip:"false"  type:"string"    short:"i"    long:"interface-name"    default:"quantum%d"             description:"The name to give the TUN interface quantum uses, append '%d' to have auto incrementing names."`
	NumWorkers      int           `skip:"false"  type:"int"       short:"n"    long:"workers"           default:"0"                     description:"The number of quantum workers to use, set to 0 for a worker per available cpu core."`
	PrivateIP       net.IP        `skip:"false"  type:"ip"        short:"ip"   long:"private-ip"        default:""                      description:"The private ip address to assign this quantum instance."`
	ListenIP        net.IP        `skip:"false"  type:"ip"        short:"lip"  long:"listen-ip"         default:""                      description:"The local server ip to listen on, leave blank of automatic association."`
	ListenPort      int           `skip:"false"  type:"int"       short:"p"    long:"listen-port"       default:"1099"                  description:"The local server port to listen on."`
	PublicIPv4      net.IP        `skip:"false"  type:"ip"        short:"4"    long:"public-v4"         default:""                      description:"The public ipv4 address to associate with this quantum instance, leave blank for automatic association."`
	PublicIPv6      net.IP        `skip:"false"  type:"ip"        short:"6"    long:"public-v6"         default:""                      description:"The public ipv6 address to associate with this quantum instance, leave blank for automatic association."`
	Prefix          string        `skip:"false"  type:"string"    short:"pr"   long:"prefix"            default:"quantum"               description:"The prefix to store quantum configuration data under in the backend key/value store."`
	DataDir         string        `skip:"false"  type:"string"    short:"d"    long:"data-dir"          default:"/var/lib/quantum"      description:"The directory to store local quantum state to."`
	PidFile         string        `skip:"false"  type:"string"    short:"pf"   long:"pid-file"          default:"/var/run/quantum.pid"  description:"The pid file to use for tracking rolling restarts."`
	StatsWindow     time.Duration `skip:"false"  type:"duration"  short:"sw"   long:"stats-window"      default:"10s"                   description:"The interval of stats aggregation."`
	SyncInterval    time.Duration `skip:"false"  type:"duration"  short:"si"   long:"sync-interval"     default:"60s"                   description:"The interval of full backend syncs."`
	RefreshInterval time.Duration `skip:"false"  type:"duration"  short:"ri"   long:"refresh-interval"  default:"120s"                  description:"The interval of dhcp lease refreshes."`
	Datastore       string        `skip:"false"  type:"string"    short:"ds"   long:"datastore"         default:"etcd"                  description:"The type of backend key/value store, this should be 'etcd' or 'consul'."`
	Endpoints       []string      `skip:"false"  type:"list"      short:"e"    long:"endpoints"         default:"127.0.0.1:2379"        description:"A comma delimited list of backend key/value store endpoints, in 'IPADDR:PORT' syntax."`
	Username        string        `skip:"false"  type:"string"    short:"u"    long:"username"          default:""                      description:"The username to use for authentication with the backend datastore."`
	Password        string        `skip:"false"  type:"string"    short:"pw"   long:"password"          default:""                      description:"The password to use for authentication with the backend datastore."`
	TLSSkipVerify   bool          `skip:"false"  type:"bool"      short:"tsv"  long:"tls-skip-verify"   default:"false"                 description:"Whether or not to authenticate the TLS certificates of the backend key/value store."`
	TLSCA           string        `skip:"false"  type:"string"    short:"tca"  long:"tls-ca-cert"       default:""                      description:"The TLS CA certificate to authenticate the TLS certificates of the backend key/value store certificates."`
	TLSCert         string        `skip:"false"  type:"string"    short:"tc"   long:"tls-cert"          default:""                      description:"The TLS client certificate to use to authenticate with the backend key/value store."`
	TLSKey          string        `skip:"false"  type:"string"    short:"tk"   long:"tls-key"           default:""                      description:"The TLS client key to use to authenticate with the backend key/value store."`

	RealInterfaceName string            `skip:"true"`
	ReuseFDS          bool              `skip:"true"`
	MachineID         string            `skip:"true"`
	AuthEnabled       bool              `skip:"true"`
	TLSEnabled        bool              `skip:"true"`
	IsIPv4Enabled     bool              `skip:"true"`
	IsIPv6Enabled     bool              `skip:"true"`
	ListenAddr        syscall.Sockaddr  `skip:"true"`
	NetworkConfig     *NetworkConfig    `skip:"true"`
	PrivateKey        []byte            `skip:"true"`
	PublicKey         []byte            `skip:"true"`
	fileData          map[string]string `skip:"true"`
}

func (cfg *Config) cliArg(short, long string, isFlag bool) (string, bool) {
	for i, arg := range os.Args {
		if arg == "-"+short ||
			arg == "--"+short ||
			arg == "-"+long ||
			arg == "--"+long {
			if !isFlag {
				return os.Args[i+1], true
			}
			return "true", true
		}
	}
	return "", false
}

func (cfg *Config) envArg(long string) (string, bool) {
	env := envPrefix + strings.ToUpper(strings.Replace(long, "-", "_", 10))
	output := os.Getenv(env)
	if output == "" {
		return output, false
	}
	return output, true
}

func (cfg *Config) fileArg(long string) (string, bool) {
	if cfg.fileData == nil {
		return "", false
	}
	value, ok := cfg.fileData[long]
	return value, ok
}

func (cfg *Config) parseFile() error {
	if cfg.ConfFile != "" {
		buf, err := ioutil.ReadFile(cfg.ConfFile)
		if err != nil {
			return err
		}

		data := make(map[string]string)
		ext := path.Ext(cfg.ConfFile)
		switch {
		case ".json" == ext:
			err = json.Unmarshal(buf, &data)
		case ".yaml" == ext || ".yml" == ext:
			err = yaml.Unmarshal(buf, &data)
		default:
			return errors.New("the configuration file is not in a supported format")
		}

		cfg.fileData = data
	}
	return nil
}

func (cfg *Config) parseField(tag reflect.StructTag) (skip, fieldType, short, long, def, description string) {
	skip = tag.Get("skip")
	fieldType = tag.Get("type")
	short = tag.Get("short")
	long = tag.Get("long")
	def = tag.Get("default")
	description = tag.Get("description")
	return
}

func (cfg *Config) usage() {
	fmt.Println("Usage of quantum:")
	st := reflect.TypeOf(*cfg)

	numFields := st.NumField()
	for i := 0; i < numFields; i++ {
		field := st.Field(i)
		skip, fieldType, short, long, def, description := cfg.parseField(field.Tag)
		if skip == "true" {
			continue
		}

		fmt.Printf("\t-%s|--%s  (%s)\n", short, long, fieldType)
		fmt.Printf("\t\t%s (default: '%s')\n", description, def)
	}

	os.Exit(1)
}

func (cfg *Config) version() {
	fmt.Printf("quantum: v%s\n", version)
	os.Exit(0)
}

func (cfg *Config) parseSpecial() {
	for _, arg := range os.Args {
		switch {
		case strings.HasSuffix(arg, "h") || strings.HasSuffix(arg, "help"):
			cfg.usage()
		case strings.HasSuffix(arg, "v") || strings.HasSuffix(arg, "version"):
			cfg.version()
		}
	}
}

func parseArgs() (*Config, error) {
	cfg := Config{}
	cfg.parseSpecial()

	st := reflect.TypeOf(cfg)
	sv := reflect.ValueOf(&cfg).Elem()

	numFields := st.NumField()
	for i := 0; i < numFields; i++ {
		field := st.Field(i)
		fieldValue := sv.Field(i)
		skip, fieldType, short, long, def, _ := cfg.parseField(field.Tag)

		if skip == "true" || !fieldValue.CanSet() {
			continue
		}

		var raw string
		if value, ok := cfg.cliArg(short, long, fieldType == "bool"); ok {
			raw = value
		} else if value, ok := cfg.envArg(long); ok {
			raw = value
		} else if value, ok := cfg.fileArg(long); ok {
			raw = value
		} else {
			raw = def
		}

		switch fieldType {
		case "int":
			i, err := strconv.Atoi(raw)
			if err != nil {
				return nil, err
			}
			fieldValue.Set(reflect.ValueOf(i))
		case "duration":
			dur, err := time.ParseDuration(raw)
			if err != nil {
				return nil, err
			}
			fieldValue.Set(reflect.ValueOf(dur))
		case "ip":
			ip := net.ParseIP(raw)
			if ip == nil && raw != "" {
				return nil, errors.New("Invalid ip address.")
			}
			fieldValue.Set(reflect.ValueOf(ip))
		case "bool":
			b, err := strconv.ParseBool(raw)
			if err != nil {
				return nil, err
			}
			fieldValue.Set(reflect.ValueOf(b))
		case "list":
			list := strings.Split(raw, ",")
			fieldValue.Set(reflect.ValueOf(list))
		case "string":
			fieldValue.Set(reflect.ValueOf(raw))
		default:
			return nil, errors.New("Unknown configuration field type.")
		}

		if field.Name == "ConfFile" {
			cfg.parseFile()
		}
	}

	return &cfg, nil
}

func (cfg *Config) computeArgs() error {
	pubkey, privkey := GenerateECKeyPair()
	cfg.PublicKey = pubkey
	cfg.PrivateKey = privkey

	if (cfg.TLSCert != "" && cfg.TLSKey != "") || cfg.TLSCA != "" {
		cfg.TLSEnabled = true
	}

	if cfg.Username != "" {
		cfg.AuthEnabled = true
	}

	if cfg.NumWorkers == 0 {
		cfg.NumWorkers = runtime.NumCPU()
	}

	os.MkdirAll(cfg.DataDir, os.ModeDir)
	os.MkdirAll(path.Dir(cfg.PidFile), os.ModeDir)

	machineID := make([]byte, 32)
	machineIDPath := path.Join(cfg.DataDir, "machine-id")
	if _, err := os.Stat(machineIDPath); os.IsNotExist(err) {
		rand.Read(machineID)
		ioutil.WriteFile(machineIDPath, machineID, os.ModePerm)
	} else {
		buf, _ := ioutil.ReadFile(machineIDPath)
		machineID = buf
	}
	cfg.MachineID = hex.EncodeToString(machineID)

	cfg.RealInterfaceName = os.Getenv(RealInterfaceNameEnv)
	if cfg.RealInterfaceName != "" {
		cfg.ReuseFDS = true
	}

	if cfg.PublicIPv4 == nil {
		routes, err := netlink.RouteGet(googleV4)
		if err != nil {
			return err
		}
		if !ArrayEquals(routes[0].Src, loopbackV4) {
			cfg.PublicIPv4 = routes[0].Src
			cfg.IsIPv4Enabled = true
		}
	} else {
		cfg.IsIPv4Enabled = true
	}

	if cfg.PublicIPv6 == nil {
		routes, err := netlink.RouteGet(googleV6)
		if err != nil {
			return err
		}
		if !ArrayEquals(routes[0].Src, loopbackV6) {
			cfg.PublicIPv6 = routes[0].Src
			cfg.IsIPv6Enabled = true
		}
	} else {
		cfg.IsIPv6Enabled = true
	}

	if cfg.ListenIP == nil {
		switch {
		case cfg.IsIPv4Enabled && cfg.IsIPv6Enabled:
			fallthrough
		case cfg.IsIPv6Enabled:
			sa := &syscall.SockaddrInet6{Port: cfg.ListenPort}
			copy(sa.Addr[:], allV6.To16()[:])
			cfg.ListenAddr = sa
		case cfg.IsIPv4Enabled:
			sa := &syscall.SockaddrInet4{Port: cfg.ListenPort}
			copy(sa.Addr[:], allV4.To4()[:])
			cfg.ListenAddr = sa
		default:
			return errors.New("Impossible situation occured, neither ipv4 or ipv6 is active. Check your networking configuration you must have public internet access to use autoconfiguration.")
		}
	} else if addr := cfg.ListenIP.To4(); addr != nil {
		sa := &syscall.SockaddrInet4{Port: cfg.ListenPort}
		copy(sa.Addr[:], addr[:])
		cfg.ListenAddr = sa
	} else if addr := cfg.ListenIP.To16(); addr != nil {
		sa := &syscall.SockaddrInet6{Port: cfg.ListenPort}
		copy(sa.Addr[:], addr[:])
		cfg.ListenAddr = sa
	} else {
		return errors.New("Impossible situation occured, neither ipv4 or ipv6 is active. Check your networking configuration you must have public internet access to use autoconfiguration.")
	}

	pid := os.Getpid()
	return ioutil.WriteFile(cfg.PidFile, []byte(strconv.Itoa(pid)), os.ModePerm)
}

// NewConfig object
func NewConfig() (*Config, error) {
	cfg, err := parseArgs()
	if err != nil {
		return nil, err
	}

	if err := cfg.computeArgs(); err != nil {
		return nil, err
	}

	return cfg, nil
}
