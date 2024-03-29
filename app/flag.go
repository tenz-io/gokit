package app

import (
	"flag"
	"fmt"
	"os"
)

var (
	defaultFlags = []Flag{
		&StringFlag{
			Name:  "config",
			Value: "app.yaml",
			Usage: "Conf file",
		},
		&StringFlag{
			Name:  "log",
			Value: "log",
			Usage: "Log output directory",
		},
		&IntFlag{
			Name:  "admin-port",
			Value: 8085,
			Usage: "Admin HTTP port",
		},
		&BoolFlag{
			Name:  "verbose",
			Value: false,
			Usage: "Verbose mode(true/false)",
		},
	}
)

type Flag interface {
	GetName() string
	GetValue() any
	Apply(*flag.FlagSet)
}

type StringFlag struct {
	Name  string
	Value string
	Usage string
}

type IntFlag struct {
	Name  string
	Value int
	Usage string
}

type BoolFlag struct {
	Name  string
	Value bool
	Usage string
}

func (f *StringFlag) Apply(fs *flag.FlagSet) {
	fs.StringVar(&f.Value, f.Name, f.Value, f.Usage)
}

func (f *StringFlag) GetName() string {
	return f.Name
}

func (f *StringFlag) GetValue() any {
	return f.Value
}

func (f *IntFlag) Apply(fs *flag.FlagSet) {
	fs.IntVar(&f.Value, f.Name, f.Value, f.Usage)
}

func (f *IntFlag) GetName() string {
	return f.Name
}

func (f *IntFlag) GetValue() any {
	return f.Value
}

func (f *BoolFlag) Apply(fs *flag.FlagSet) {
	fs.BoolVar(&f.Value, f.Name, f.Value, f.Usage)
}

func (f *BoolFlag) GetName() string {
	return f.Name
}

func (f *BoolFlag) GetValue() any {
	return f.Value
}

type Flags struct {
	flags map[string]Flag
}

// Apply applies the flags to the flag set
func (fs *Flags) Apply(flagSet *flag.FlagSet) {
	for _, f := range fs.flags {
		f.Apply(flagSet)
	}
}

// Int retrieves the int value of a flag
// will return an error if the flag is not found or not an int flag
func (fs *Flags) Int(name string) (int, error) {
	f, ok := fs.flags[name]
	if !ok {
		return 0, fmt.Errorf("flag %s not found", name)
	}

	if intFlag, ok := f.(*IntFlag); ok {
		return intFlag.Value, nil
	}

	return 0, fmt.Errorf("flag %s is not an int flag", name)
}

// IsSet determines if the flag was actually set
func (fs *Flags) IsSet(name string) bool {
	_, ok := fs.flags[name]
	return ok
}

func (fs *Flags) Set(name string, flag Flag) {
	fs.flags[name] = flag
}

// String retrieves the string value of a flag
// will return an error if the flag is not found or not a string flag
func (fs *Flags) String(name string) (string, error) {
	f, ok := fs.flags[name]
	if !ok {
		return "", fmt.Errorf("flag %s not found", name)
	}

	if stringFlag, ok := f.(*StringFlag); ok {
		return stringFlag.Value, nil
	}

	return "", fmt.Errorf("flag %s is not a string flag", name)
}

// Bool retrieves the bool value of a flag
// will return an error if the flag is not found or not a bool flag
func (fs *Flags) Bool(name string) (bool, error) {
	f, ok := fs.flags[name]
	if !ok {
		return false, fmt.Errorf("flag %s not found", name)
	}

	if boolFlag, ok := f.(*BoolFlag); ok {
		return boolFlag.Value, nil
	}

	return false, fmt.Errorf("flag %s is not a bool flag", name)
}

// Print prints the flags
func (fs *Flags) Print() {
	fmt.Println("args: ==================")
	for _, f := range fs.flags {
		fmt.Printf("%s: %v\n", f.GetName(), f.GetValue())
	}
	fmt.Println("==================")
}

// NewFlags creates a new Flags
func NewFlags(flags []Flag) (*Flags, error) {
	fs := &Flags{
		flags: make(map[string]Flag),
	}
	for _, f := range flags {
		if f.GetName() == "" {
			return nil, fmt.Errorf("flag name is empty")
		}

		if fs.IsSet(f.GetName()) {
			return nil, fmt.Errorf("flag %s is already exists", f.GetName())
		}

		fs.Set(f.GetName(), f)
	}
	return fs, nil
}

func Parse(name string, flags *Flags) error {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	flags.Apply(fs)
	return fs.Parse(os.Args[1:])
}
