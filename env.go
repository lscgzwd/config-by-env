package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	defaultSep = "_"
)
var env *Env

type Env struct {
	prefix       string
	ignorePrefix bool
}

func init() {
	env = &Env{prefix: "", ignorePrefix: false}
}

func upper(v string) string {
	return strings.ToUpper(v)
}
func SetPrefix(prefix string) {
	env.ignorePrefix = true
	env.prefix = upper(prefix)
}
func IgnorePrefix() {
	env.ignorePrefix = true
	env.prefix = ""
}

func Fill(v interface{}) error {
	return env.Fill(v)
}

func (e *Env) Fill(v interface{}) error {
	ind := reflect.Indirect(reflect.ValueOf(v))
	if reflect.ValueOf(v).Kind() != reflect.Ptr || ind.Kind() != reflect.Struct {
		return fmt.Errorf("only the pointer to a struct is supported")
	}

	prefix := upper(ind.Type().Name())
	if e.ignorePrefix {
		prefix = e.prefix
	}
	err := fill(prefix, ind)
	if err != nil {
		return err
	}
	return nil
}

func combine(p, n string, sep string, ok bool) string {
	if p == "" {
		return n
	}
	if !ok {
		return p + defaultSep + n
	}
	return p + sep + n
}

func parseBool(v string) (bool, error) {
	if v == "" {
		return false, nil
	}
	return strconv.ParseBool(v)
}

func fill(pf string, ind reflect.Value) error {
	for i := 0; i < ind.NumField(); i++ {
		f := ind.Type().Field(i)
		name := f.Name
		envName, exist := f.Tag.Lookup("env")
		if exist {
			name = envName
		}
		s, exist := f.Tag.Lookup("sep")
		p := combine(pf, upper(name), s, exist)
		switch ind.Field(i).Kind() {
		case reflect.Struct:
			err := fill(p, ind.Field(i))
			if err != nil {
				return err
			}
		default:
			err := parse(p, ind.Field(i), f)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func parse(prefix string, f reflect.Value, sf reflect.StructField) error {
	df := sf.Tag.Get("default")
	isRequire, err := parseBool(sf.Tag.Get("require"))
	if err != nil {
		return fmt.Errorf("the value of %s is not a valid `member` of bool ，only "+
			"[1 0 t f T F true false TRUE FALSE True False] are supported", prefix)
	}
	ev, exist := os.LookupEnv(prefix)

	if !exist && isRequire {
		return fmt.Errorf("%s is required, but has not been set", prefix)
	}
	if !exist && df != "" {
		ev = df
	}

	switch f.Kind() {
	case reflect.String:
		f.SetString(ev)
	case reflect.Int8:
		iv, err := strconv.ParseInt(ev, 10, 8)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetInt(iv)
	case reflect.Int16:
		iv, err := strconv.ParseInt(ev, 10, 16)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetInt(iv)
	case reflect.Int32:
		iv, err := strconv.ParseInt(ev, 10, 32)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetInt(iv)
	case reflect.Int:
		iv, err := strconv.ParseInt(ev, 10, 32)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetInt(iv)
	case reflect.Int64:
		if f.Type().String() == "time.Duration" {
			t, err := time.ParseDuration(ev)
			if err != nil {
				return fmt.Errorf("%s:%s", prefix, err)
			}
			f.Set(reflect.ValueOf(t))
		} else {
			iv, err := strconv.ParseInt(ev, 10, 64)
			if err != nil {
				return fmt.Errorf("%s:%s", prefix, err)
			}
			f.SetInt(iv)
		}
	case reflect.Uint8:
		uiv, err := strconv.ParseUint(ev, 10, 8)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetUint(uiv)
	case reflect.Uint16:
		uiv, err := strconv.ParseUint(ev, 10, 16)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetUint(uiv)
	case reflect.Uint:
		uiv, err := strconv.ParseUint(ev, 10, 32)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetUint(uiv)
	case reflect.Uint32:
		uiv, err := strconv.ParseUint(ev, 10, 32)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetUint(uiv)
	case reflect.Uint64:
		uiv, err := strconv.ParseUint(ev, 10, 64)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetUint(uiv)
	case reflect.Float32:
		f32, err := strconv.ParseFloat(ev, 32)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetFloat(f32)
	case reflect.Float64:
		f64, err := strconv.ParseFloat(ev, 64)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetFloat(f64)
	case reflect.Bool:
		b, err := parseBool(ev)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetBool(b)
	case reflect.Slice:
		sep := ";"
		s, exist := sf.Tag.Lookup("slice_sep")
		if exist && s != "" {
			sep = s
		}
		if ev == "" {
			return nil
		}
		values := strings.Split(ev, sep)
		switch f.Type() {
		case reflect.TypeOf([]string{}):
			f.Set(reflect.ValueOf(values))
		case reflect.TypeOf([]int{}):
			t := make([]int, len(values))
			for i, v := range values {
				val, err := strconv.ParseInt(v, 10, 32)
				if err != nil {
					return fmt.Errorf("%s:%s", prefix, err)
				}
				t[i] = int(val)
			}
		case reflect.TypeOf([]int64{}):
			t := make([]int64, len(values))
			for i, v := range values {
				val, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return fmt.Errorf("%s:%s", prefix, err)
				}
				t[i] = val
			}
		case reflect.TypeOf([]uint{}):
			t := make([]uint, len(values))
			for i, v := range values {
				val, err := strconv.ParseUint(v, 10, 32)
				if err != nil {
					return fmt.Errorf("%s:%s", prefix, err)
				}
				t[i] = uint(val)
			}
		case reflect.TypeOf([]uint64{}):
			t := make([]uint64, len(values))
			for i, v := range values {
				val, err := strconv.ParseUint(v, 10, 64)
				if err != nil {
					return fmt.Errorf("%s:%s", prefix, err)
				}
				t[i] = val
			}
		case reflect.TypeOf([]float32{}):
			t := make([]float32, len(values))
			for i, v := range values {
				val, err := strconv.ParseFloat(v, 32)
				if err != nil {
					return fmt.Errorf("%s:%s", prefix, err)
				}
				t[i] = float32(val)
			}
		case reflect.TypeOf([]float64{}):
			t := make([]float64, len(values))
			for i, v := range values {
				val, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return fmt.Errorf("%s:%s", prefix, err)
				}
				t[i] = val
			}
		case reflect.TypeOf([]bool{}):
			t := make([]bool, len(values))
			for i, v := range values {
				val, err := parseBool(v)
				if err != nil {
					return fmt.Errorf("%s:%s", prefix, err)
				}
				t[i] = val
			}
		}
	}
	return nil
}
