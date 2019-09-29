package utils

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"gitlab.10101111.com/oped/DBMS_LIBS/blowfish"
)

const (
	Byte = 1
	KiB  = 1024 * Byte
	MiB  = 1024 * KiB
	GiB  = 1024 * MiB
)

// NOTE: pass in the pointer of type cfgStruct
// NOTE: to make the verification work, every field in type cfgStruct must be tagged as toml
func LoadTomlCfg(configFile string, cfgStruct interface{}) error {
	var (
		meta toml.MetaData
		err  error
	)
	meta, err = toml.DecodeFile(configFile, cfgStruct)
	if err != nil {
		return err
	}
	if err = structFieldsDefinedInConfig(&meta, reflect.TypeOf(cfgStruct).Elem(), []string{}); err != nil {
		return err
	}
	return nil
}

// NOTE: pass in the pointer of type cfgStruct
func DumpCfg2Toml(cfgStruct interface{}) string {
	var b bytes.Buffer
	e := toml.NewEncoder(&b)
	e.Indent = "    "
	e.Encode(cfgStruct)
	return b.String()
}

func structFieldsDefinedInConfig(meta *toml.MetaData, structType reflect.Type, hierachyTomlKeys []string) error {
	hierachyTomlKeys = append(hierachyTomlKeys, "")
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if tomlKey, tomlKeyOK := field.Tag.Lookup("toml"); tomlKeyOK {
			hierachyTomlKeys[len(hierachyTomlKeys)-1] = tomlKey
			if _, ok := field.Tag.Lookup("required"); ok {
				if !meta.IsDefined(hierachyTomlKeys...) {
					return fmt.Errorf("\"%s\" not defined in config file", strings.Join(hierachyTomlKeys, "."))
				}
			}
			// make sure all sub-config-block is mapped to a struct's pointer type
			if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
				if err := structFieldsDefinedInConfig(meta, field.Type.Elem(), hierachyTomlKeys); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

// make sure your utils.Duration fields are not defined as pointer type,
// because a type "*Duration" won't match against toml.TextMarshaler interface,
// since we implement it on type "Duration"
func (d Duration) MarshalText() (text []byte, err error) {
	return []byte(d.String()), nil
}

type Bytes struct {
	int64
}

func NewBytes(size int64) Bytes {
	return Bytes{size}
}

func (b *Bytes) UnmarshalText(text []byte) error {
	v := string(text)
	unit := Byte
	subIdx := len(v)
	if strings.HasSuffix(v, "k") {
		unit = KiB
		subIdx = subIdx - 1
	} else if strings.HasSuffix(v, "kb") {
		unit = KiB
		subIdx = subIdx - 2
	} else if strings.HasSuffix(v, "m") {
		unit = MiB
		subIdx = subIdx - 1
	} else if strings.HasSuffix(v, "mb") {
		unit = MiB
		subIdx = subIdx - 2
	} else if strings.HasSuffix(v, "g") {
		unit = GiB
		subIdx = subIdx - 1
	} else if strings.HasSuffix(v, "gb") {
		unit = GiB
		subIdx = subIdx - 2
	}
	i, err := strconv.ParseInt(v[:subIdx], 10, 64)
	if err == nil {
		b.int64 = i * int64(unit)
	}
	return err
}

func (b *Bytes) Int64() int64 {
	return b.int64
}

func (b *Bytes) Int() int {
	return int(b.int64)
}

// make sure your utils.Bytes fields are not defined as pointer type
// because a type "*Bytes" won't match against toml.TextMarshaler interface,
// since we implement it on type "Bytes"
func (b Bytes) MarshalText() (text []byte, err error) {
	return []byte(b.String()), nil
}

// make sure your utils.Bytes fields are not defined as pointer type
// because a type "*Bytes" won't match against GoString interface,
// since we implement it on type "Bytes"
func (b Bytes) String() string {
	unit := ""
	value := float32(b.int64)

	switch {
	case b.int64 >= GiB:
		unit = "GiB"
		value = value / GiB
	case b.int64 >= MiB:
		unit = "MiB"
		value = value / MiB
	case b.int64 >= KiB:
		unit = "KiB"
		value = value / KiB
	case b.int64 >= Byte:
		unit = "B"
	case b.int64 == 0:
		return "0"
	}
	stringValue := fmt.Sprintf("%.1f", value)
	stringValue = strings.TrimSuffix(stringValue, ".0")
	return fmt.Sprintf("%s%s", stringValue, unit)
}

const (
	Holding_main_speaker = "devops2015"
	Make_fun_with_jokes  = ""
)

var (
	MiKey  string = Holding_main_speaker
	MiSkey string = Make_fun_with_jokes
)

func SetCipherKey(key, salt string) {
	MiKey = key
	MiSkey = salt
}

type Ciphertext struct {
	Citext string
}

func (d *Ciphertext) UnmarshalText(text []byte) error {
	var err error
	d.Citext, err = blowfish.Encrypt(d.Citext, []byte(MiKey), []byte(MiSkey))
	return err
}

func (d Ciphertext) MarshalText() (text []byte, err error) {
	sercet, err := blowfish.Encrypt(d.Citext, []byte(MiKey), []byte(MiSkey))
	return []byte(sercet), err
}
