package cfg

import (
	"os"
	"path/filepath"
	"testing"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
)

const TEST_CONF_FILE = CONFIG_FILENAME + "." + CONFIG_FILETYPE
const TEST_URL = "http://localhost:1234/go"
const TEST_USER = "admin"
const TEST_PASSWORD = "badger"

func TestGetBasicAuth(t *testing.T) {
	as := asserts(t)
	c := testConf(true)
	c.native.Set("auth.type", "basic")
	c.native.Set("auth.user", TEST_USER)
	c.native.Set("auth.password", TEST_PASSWORD)
	c.native.WriteConfig()

	as.deepEq(map[string]string{
		"type":     "basic",
		"user":     TEST_USER,
		"password": TEST_PASSWORD,
	}, c.GetAuth())
}

func TestSetBasicAuth(t *testing.T) {
	as := asserts(t)
	c := testConf(true)

	as.configEq(make(dict, 0), c.fs)

	as.ok(c.SetBasicAuth(TEST_USER, TEST_PASSWORD))

	as.configEq(dict{
		"auth": map[string]string{
			"type":     "basic",
			"user":     TEST_USER,
			"password": TEST_PASSWORD,
		},
	}, c.fs)
}

func TestGetTokenAuth(t *testing.T) {
	as := asserts(t)
	c := testConf(true)

	c.native.Set("auth.type", "token")
	c.native.Set("auth.token", "hello!")
	c.native.WriteConfig()

	as.deepEq(map[string]string{
		"type":  "token",
		"token": "hello!",
	}, c.GetAuth())
}

func TestSetTokenAuth(t *testing.T) {
	as := asserts(t)
	c := testConf(true)

	as.configEq(make(dict, 0), c.fs)

	as.ok(c.SetTokenAuth(`gah!`))

	as.configEq(dict{
		"auth": map[string]string{
			"type":  "token",
			"token": "gah!",
		},
	}, c.fs)
}

func TestSettingAuthClearsPreviousSetting(t *testing.T) {
	as := asserts(t)
	c := testConf(true)
	c.SetBasicAuth(TEST_USER, TEST_PASSWORD)

	as.deepEq(map[string]string{
		"type":     "basic",
		"user":     TEST_USER,
		"password": TEST_PASSWORD,
	}, c.GetAuth())

	as.configEq(dict{
		"auth": map[string]string{
			"type":     "basic",
			"user":     TEST_USER,
			"password": TEST_PASSWORD,
		},
	}, c.fs)

	as.ok(c.SetTokenAuth(`gah!`))

	as.deepEq(map[string]string{
		"type":  "token",
		"token": "gah!",
	}, c.GetAuth())

	as.configEq(dict{
		"auth": map[string]string{
			"type":  "token",
			"token": "gah!",
		},
	}, c.fs)

	c.native.ReadInConfig() // ensure we refresh from disk to exclude override register

	as.deepEq(map[string]string{
		"type":  "token",
		"token": "gah!",
	}, c.GetAuth())

	as.configEq(dict{
		"auth": map[string]string{
			"type":  "token",
			"token": "gah!",
		},
	}, c.fs)

	as.eq(nil, c.native.Get("auth.password"))
}

func TestSetBasicAuthShouldValidatePresenceOfUserAndPassword(t *testing.T) {
	as := asserts(t)
	c := testConf(true)

	as.err("Must specify user and password", c.SetBasicAuth("", TEST_PASSWORD))
	as.err("Must specify user and password", c.SetBasicAuth(TEST_USER, ""))
	as.err("Must specify user and password", c.SetBasicAuth("", ""))
	as.eq(0, len(c.GetAuth()))
}

func TestSetServerURL(t *testing.T) {
	as := asserts(t)
	c := testConf(true)

	as.configEq(make(dict, 0), c.fs)

	as.ok(c.SetServerUrl(TEST_URL))

	as.configEq(dict{
		"server": map[string]string{
			"url": TEST_URL,
		},
	}, c.fs)
}

func TestGetServerURL(t *testing.T) {
	as := asserts(t)
	c := testConf(true)
	c.native.Set("server.url", TEST_URL)
	c.native.WriteConfig()

	as.eq(TEST_URL, c.GetServerUrl())
}

func TestSetServerURLValidatesURL(t *testing.T) {
	as := asserts(t)
	c := testConf(true)

	as.err("Must specify a url", c.SetServerUrl(""))
	as.err("server-url must include protocol and hostname", c.SetServerUrl("foo.bar"))
	as.err("server-url must include protocol and hostname", c.SetServerUrl("http://"))
	as.err("parse http://localhost:foo/bar: invalid port \":foo\" after host", c.SetServerUrl("http://localhost:foo/bar"))
	as.err("server-url must end with /go", c.SetServerUrl("http://localhost:8080/bar"))
	as.eq("", c.GetServerUrl())
}

func TestConsumeCreatesDefaultConfigFileIfNotExist(t *testing.T) {
	as := asserts(t)
	c := testConf(false)

	baseDir, e := homedir.Dir()
	as.ok(e)

	expectedPath := filepath.Join(baseDir, CONFIG_DIRNAME, TEST_CONF_FILE)
	exists, err := afero.Exists(c.fs, expectedPath)

	as.ok(err)
	as.not(exists)

	as.ok(c.Consume(""))

	exists, err = afero.Exists(c.fs, expectedPath)
	as.ok(err)
	as.is(exists)

	isDir, er := afero.IsDir(c.fs, expectedPath)
	as.ok(er)
	as.not(isDir)
}

func TestConsumeReturnsErrorIfSpecifiedConfigFileNotExists(t *testing.T) {
	as := asserts(t)
	c := testConf(false)

	expectedPath := "foo.yaml"
	exists, err := afero.Exists(c.fs, expectedPath)

	as.ok(err)
	as.not(exists)

	as.err("open "+expectedPath+": file does not exist", c.Consume(expectedPath))

	exists, err = afero.Exists(c.fs, expectedPath)
	as.ok(err)
	as.not(exists)
}

func TestConsumeLoadsConfigFileAtDefaultPathWhenExists(t *testing.T) {
	as := asserts(t)
	c := testConf(false)

	baseDir, e := homedir.Dir()
	as.ok(e)

	expectedPath := filepath.Join(baseDir, CONFIG_DIRNAME, TEST_CONF_FILE)

	as.ok(writeContent(c.fs, expectedPath, `server:
  url: http://test-server:8080
`))

	exists, err := afero.Exists(c.fs, expectedPath)
	as.ok(err)
	as.is(exists)

	as.ok(c.Consume(""))
	as.eq("http://test-server:8080", c.GetServerUrl())
}

func TestConsumeLoadsSpecifiedConfigFileWhenExists(t *testing.T) {
	as := asserts(t)
	c := testConf(false)

	expectedPath := "/etc/foo.yaml"

	as.ok(writeContent(c.fs, expectedPath, `server:
  url: http://test-server:8080
`))

	exists, err := afero.Exists(c.fs, expectedPath)
	as.ok(err)
	as.is(exists)

	as.ok(c.Consume(expectedPath))
	as.eq("http://test-server:8080", c.GetServerUrl())
}

func TestUnsetRemovesConfiguredValue(t *testing.T) {
	as := asserts(t)
	c, err := makeConf(`---
config_version: 1
auth:
  type: basic
  user: admin
  password: badger
server:
  url: http://test/foo
`)
	as.ok(err)
	as.eq(`http://test/foo`, c.GetServerUrl())

	as.ok(c.Unset(`server-url`))
	as.eq("", c.GetServerUrl())

	// validate that only server-url was affected
	as.deepEq(dict{
		`type`:     `basic`,
		`user`:     `admin`,
		`password`: `badger`,
	}, c.GetAuth())

	// check disk contents for expected structure
	as.configEq(dict{
		`config_version`: 1,
		`auth`: dict{
			`type`:     `basic`,
			`user`:     `admin`,
			`password`: `badger`,
		},
	}, c.fs)
}

func TestUnsetLeavesOverridesOfOtherKeysIntact(t *testing.T) {
	as := asserts(t)
	c, err := makeConf(`---
config_version: 1
auth:
  type: basic
  user: admin
  password: badger
server:
  url: http://test/foo
`)
	as.ok(err)

	as.deepEq(dict{
		`type`:     `basic`,
		`user`:     `admin`,
		`password`: `badger`,
	}, c.GetAuth())

	as.eq(`http://test/foo`, c.GetServerUrl())

	// yes, currently this cannot happen as we do not expose `Set()` directly,
	// but that may change, and we should validate that there are no side-effects
	// on unrelated keys.
	c.native.Set(`server.url`, `http://test/bar`)

	as.eq(`http://test/bar`, c.GetServerUrl())
	as.ok(c.Unset(`auth`))

	// override value is still honored for this key
	as.eq(`http://test/bar`, c.GetServerUrl())

	// The original on-disk value of server.url is preserved, which is different
	// from the override value returned by Get(). We should only persist the
	// override value if explicitly flush that key to disk, but not as a side-effect
	// of writing or clearing another key.
	as.configEq(dict{
		`config_version`: 1,
		`server`: dict{
			`url`: `http://test/foo`,
		},
	}, c.fs)
}

func TestEnvVariableConfig(t *testing.T) {
	os.Clearenv()

	defer os.Clearenv()

	as := asserts(t)
	c := testConf(true)

	c.LayerConfigs()

	as.eq(``, c.GetServerUrl())
	os.Setenv(`GOCDCLI_SERVER.URL`, `http://from.env`)
	as.eq(`http://from.env`, c.GetServerUrl())

	as.eq(0, len(c.GetAuth()))
	os.Setenv(`GOCDCLI_AUTH.TYPE`, `basic`)
	os.Setenv(`GOCDCLI_AUTH.USER`, `jbond`)
	os.Setenv(`GOCDCLI_AUTH.PASSWORD`, `007`)
	as.deepEq(map[string]string{
		`type`:     `basic`,
		`user`:     `jbond`,
		`password`: `007`,
	}, c.GetAuth())
}
